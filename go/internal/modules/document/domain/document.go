package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

const (
	SchemaLegacy  = "1.0.0" // S1–S2: widget-first nodes with `type` and `geometry`
	SchemaCurrent = "1.1.0" // ADR-01: freeform nodes with `kind`

	MaxNodes         = 5000
	MaxDocumentBytes = 5 << 20 // 5 MiB
	MaxPatchBytes    = 1 << 20 // 1 MiB
	SnapshotEvery    = 50      // checkpoint version every N revisions
)

// EmptyDocument is the content of a page that has never been edited.
const EmptyDocument = `{"schemaVersion":"1.1.0","rootNodeIds":[],"nodes":{}}`

// schemaVersion accepts both the string form ("1.1.0") and the numeric form
// (1) written by the initial database default.
type schemaVersion string

func (v *schemaVersion) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*v = schemaVersion(s)
		return nil
	}
	var n float64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	if n == 1 {
		*v = schemaVersion(SchemaLegacy)
		return nil
	}
	*v = schemaVersion(fmt.Sprintf("%v", n))
	return nil
}

// DesignDocument is the aggregate persisted per page (ADR-05).
type DesignDocument struct {
	SchemaVersion schemaVersion         `json:"schemaVersion"`
	DocumentID    string                `json:"documentId,omitempty"`
	PageID        string                `json:"pageId,omitempty"`
	RootNodeIDs   []string              `json:"rootNodeIds"`
	Nodes         map[string]DesignNode `json:"nodes"`
}

func (d *DesignDocument) Version() string { return string(d.SchemaVersion) }

// ParseDocument decodes raw JSON into the typed model. It does not validate.
func ParseDocument(raw []byte) (*DesignDocument, error) {
	if len(raw) > MaxDocumentBytes {
		return nil, ErrDocumentTooLarge
	}
	var doc DesignDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, &ValidationError{Issues: []ValidationIssue{{
			Code: IssueUnsupportedVersion, Message: "document is not valid JSON of the expected shape: " + err.Error(),
		}}}
	}
	if doc.Nodes == nil {
		doc.Nodes = map[string]DesignNode{}
	}
	if doc.RootNodeIDs == nil {
		doc.RootNodeIDs = []string{}
	}
	return &doc, nil
}

// Validate enforces the ADR-05 invariants and returns *ValidationError.
//  1. every id in children/rootNodeIds exists in nodes
//  2. parentId matches the containing list; roots have no parent
//  3. each node has exactly one parent and the graph has no cycles
//  4. nodes map key equals node.id; every node is reachable from a root
func (d *DesignDocument) Validate() error {
	var issues []ValidationIssue
	add := func(code, nodeID, msg string) {
		issues = append(issues, ValidationIssue{Code: code, NodeID: nodeID, Message: msg})
	}

	switch d.Version() {
	case SchemaLegacy, SchemaCurrent:
	default:
		add(IssueUnsupportedVersion, "", fmt.Sprintf("unsupported schema version %q", d.Version()))
		return &ValidationError{Issues: issues}
	}

	if len(d.Nodes) > MaxNodes {
		add(IssueTooManyNodes, "", fmt.Sprintf("document has %d nodes, limit is %d", len(d.Nodes), MaxNodes))
		return &ValidationError{Issues: issues}
	}

	ids := make([]string, 0, len(d.Nodes))
	for id := range d.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids) // deterministic error output

	parentOf := map[string]string{} // child -> parent ("" for roots)
	claim := func(child, parent string) {
		if prev, seen := parentOf[child]; seen {
			add(IssueMultipleParents, child, fmt.Sprintf("node %q is listed under both %q and %q", child, displayParent(prev), displayParent(parent)))
			return
		}
		parentOf[child] = parent
	}

	for _, rootID := range d.RootNodeIDs {
		n, ok := d.Nodes[rootID]
		if !ok {
			add(IssueMissingNode, rootID, fmt.Sprintf("root node %q does not exist", rootID))
			continue
		}
		claim(rootID, "")
		if n.ParentID != nil && *n.ParentID != "" {
			add(IssueParentMismatch, rootID, fmt.Sprintf("root node %q must not have a parentId", rootID))
		}
	}

	for _, id := range ids {
		n := d.Nodes[id]
		if n.ID != id {
			add(IssueIDMismatch, id, fmt.Sprintf("nodes key %q does not match node.id %q", id, n.ID))
		}
		if d.Version() == SchemaCurrent {
			if !n.Kind.Valid() {
				add(IssueInvalidKind, id, fmt.Sprintf("node %q has invalid kind %q", id, n.Kind))
			}
		} else if n.Type == "" {
			add(IssueInvalidKind, id, fmt.Sprintf("legacy node %q has no type", id))
		}

		seenChild := map[string]struct{}{}
		for _, childID := range n.Children {
			if _, dup := seenChild[childID]; dup {
				add(IssueDuplicateChild, childID, fmt.Sprintf("node %q lists child %q more than once", id, childID))
				continue
			}
			seenChild[childID] = struct{}{}

			child, ok := d.Nodes[childID]
			if !ok {
				add(IssueMissingNode, childID, fmt.Sprintf("child %q of %q does not exist", childID, id))
				continue
			}
			claim(childID, id)
			if child.ParentID == nil || *child.ParentID != id {
				add(IssueParentMismatch, childID, fmt.Sprintf("node %q is a child of %q but its parentId disagrees", childID, id))
			}
		}
	}

	// Reachability and cycles: walk from roots; any node not reached is an
	// orphan or part of a cycle that is detached from the roots.
	reached := map[string]bool{}
	var walk func(id string, stack map[string]bool)
	walk = func(id string, stack map[string]bool) {
		if stack[id] {
			add(IssueCycle, id, fmt.Sprintf("cycle detected at node %q", id))
			return
		}
		if reached[id] {
			return
		}
		reached[id] = true
		n, ok := d.Nodes[id]
		if !ok {
			return
		}
		stack[id] = true
		for _, c := range n.Children {
			walk(c, stack)
		}
		delete(stack, id)
	}
	for _, r := range d.RootNodeIDs {
		walk(r, map[string]bool{})
	}
	for _, id := range ids {
		if !reached[id] {
			add(IssueOrphan, id, fmt.Sprintf("node %q is not reachable from any root", id))
		}
	}

	if len(issues) > 0 {
		return &ValidationError{Issues: issues}
	}
	return nil
}

func displayParent(p string) string {
	if p == "" {
		return "<root>"
	}
	return p
}

// PageDocument is the persisted current state of a page's document.
type PageDocument struct {
	PageID      string          `json:"page_id"`
	WorkspaceID string          `json:"workspace_id"`
	ProjectID   string          `json:"project_id"`
	Revision    int64           `json:"revision"`
	Content     json.RawMessage `json:"document"`
	UpdatedBy   *string         `json:"updated_by,omitempty"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Operation is one durable JSON Patch applied to a page (S5 sync foundation).
type Operation struct {
	ID           string          `json:"id"`
	PageID       string          `json:"page_id"`
	BaseRevision int64           `json:"base_revision"`
	Revision     int64           `json:"revision"`
	Patch        json.RawMessage `json:"patch"`
	CreatedBy    *string         `json:"created_by,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}
