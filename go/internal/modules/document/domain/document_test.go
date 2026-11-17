package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/klnp-studio/backend/internal/modules/document/domain"
)

func parse(t *testing.T, raw string) *domain.DesignDocument {
	t.Helper()
	doc, err := domain.ParseDocument([]byte(raw))
	require.NoError(t, err)
	return doc
}

func issueCodes(err error) []string {
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		return nil
	}
	codes := make([]string, len(ve.Issues))
	for i, is := range ve.Issues {
		codes[i] = is.Code
	}
	return codes
}

const validDoc = `{
  "schemaVersion":"1.1.0","rootNodeIds":["f"],
  "nodes":{
    "f":{"id":"f","kind":"frame","children":["r","t"]},
    "r":{"id":"r","kind":"rect","parentId":"f","children":[],"transform":{"x":1,"y":2},"size":{"width":10,"height":10}},
    "t":{"id":"t","kind":"text","parentId":"f","children":[],"data":{"text":"hi"}}
  }}`

func TestValidate_ValidDocument(t *testing.T) {
	assert.NoError(t, parse(t, validDoc).Validate())
}

func TestValidate_EmptyDocument(t *testing.T) {
	assert.NoError(t, parse(t, domain.EmptyDocument).Validate())
}

func TestValidate_NumericLegacyVersionFromDatabaseDefault(t *testing.T) {
	doc := parse(t, `{"schemaVersion":1,"revision":1,"nodes":{}}`)
	assert.Equal(t, domain.SchemaLegacy, doc.Version())
	assert.NoError(t, doc.Validate())
}

func TestValidate_Invariants(t *testing.T) {
	cases := []struct {
		name string
		doc  string
		want string
	}{
		{"missing root", `{"schemaVersion":"1.1.0","rootNodeIds":["x"],"nodes":{}}`, domain.IssueMissingNode},
		{"missing child", `{"schemaVersion":"1.1.0","rootNodeIds":["a"],"nodes":{"a":{"id":"a","kind":"frame","children":["b"]}}}`, domain.IssueMissingNode},
		{"id mismatch", `{"schemaVersion":"1.1.0","rootNodeIds":["a"],"nodes":{"a":{"id":"z","kind":"frame","children":[]}}}`, domain.IssueIDMismatch},
		{"parent mismatch", `{"schemaVersion":"1.1.0","rootNodeIds":["a"],"nodes":{"a":{"id":"a","kind":"frame","children":["b"]},"b":{"id":"b","kind":"rect","children":[]}}}`, domain.IssueParentMismatch},
		{"root with parent", `{"schemaVersion":"1.1.0","rootNodeIds":["a"],"nodes":{"a":{"id":"a","kind":"frame","parentId":"q","children":[]}}}`, domain.IssueParentMismatch},
		{"two parents", `{"schemaVersion":"1.1.0","rootNodeIds":["a","c"],"nodes":{
			"a":{"id":"a","kind":"frame","children":["b"]},
			"c":{"id":"c","kind":"frame","children":["b"]},
			"b":{"id":"b","kind":"rect","parentId":"a","children":[]}}}`, domain.IssueMultipleParents},
		{"duplicate child", `{"schemaVersion":"1.1.0","rootNodeIds":["a"],"nodes":{"a":{"id":"a","kind":"frame","children":["b","b"]},"b":{"id":"b","kind":"rect","parentId":"a","children":[]}}}`, domain.IssueDuplicateChild},
		{"orphan", `{"schemaVersion":"1.1.0","rootNodeIds":[],"nodes":{"a":{"id":"a","kind":"rect","children":[]}}}`, domain.IssueOrphan},
		{"cycle", `{"schemaVersion":"1.1.0","rootNodeIds":["a"],"nodes":{
			"a":{"id":"a","kind":"frame","children":["b"]},
			"b":{"id":"b","kind":"frame","parentId":"a","children":["a"]}}}`, domain.IssueCycle},
		{"invalid kind", `{"schemaVersion":"1.1.0","rootNodeIds":["a"],"nodes":{"a":{"id":"a","kind":"button","children":[]}}}`, domain.IssueInvalidKind},
		{"unsupported version", `{"schemaVersion":"9.9.9","rootNodeIds":[],"nodes":{}}`, domain.IssueUnsupportedVersion},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := parse(t, tc.doc).Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrInvalidDocument)
			assert.Contains(t, issueCodes(err), tc.want)
		})
	}
}

func TestValidate_TooManyNodes(t *testing.T) {
	doc := &domain.DesignDocument{SchemaVersion: "1.1.0", RootNodeIDs: []string{}, Nodes: map[string]domain.DesignNode{}}
	for i := 0; i <= domain.MaxNodes; i++ {
		id := string(rune('a'+i%26)) + json.Number(string(rune('0'+i%10))).String() + string(rune(i))
		doc.Nodes[id] = domain.DesignNode{ID: id}
	}
	err := doc.Validate()
	assert.Contains(t, issueCodes(err), domain.IssueTooManyNodes)
}

func TestValidate_LegacyNodesNeedType(t *testing.T) {
	ok := parse(t, `{"schemaVersion":"1.0.0","rootNodeIds":["a"],"nodes":{"a":{"id":"a","type":"container","children":[]}}}`)
	assert.NoError(t, ok.Validate())
	bad := parse(t, `{"schemaVersion":"1.0.0","rootNodeIds":["a"],"nodes":{"a":{"id":"a","children":[]}}}`)
	assert.Contains(t, issueCodes(bad.Validate()), domain.IssueInvalidKind)
}

func TestUpgradeLegacy(t *testing.T) {
	legacy := `{"schemaVersion":"1.0.0","documentId":"d","rootNodeIds":["root"],"nodes":{
	  "root":{"id":"root","type":"container","children":["t","b"],"geometry":{"x":20,"y":40,"width":335,"height":480},"properties":{"p":1},"custom":"keep"},
	  "t":{"id":"t","type":"text","parentId":"root","children":[],"geometry":{"x":0,"y":0,"width":10,"height":5},"properties":{"text":"hi"}},
	  "b":{"id":"b","type":"button","parentId":"root","children":[],"properties":{"label":"Go"}}}}`

	out, err := domain.UpgradeLegacy([]byte(legacy))
	require.NoError(t, err)

	doc := parse(t, string(out))
	assert.Equal(t, domain.SchemaCurrent, doc.Version())
	require.NoError(t, doc.Validate())

	root := doc.Nodes["root"]
	assert.Equal(t, domain.KindFrame, root.Kind)
	assert.Equal(t, "container", root.ComponentRef.BuiltinType)
	assert.Equal(t, 20.0, root.Transform.X)
	assert.Equal(t, 480.0, root.Size.Height)
	assert.Equal(t, domain.KindText, doc.Nodes["t"].Kind)
	assert.Equal(t, "hi", doc.Nodes["t"].Data["text"])
	assert.Equal(t, domain.KindFrame, doc.Nodes["b"].Kind)
	assert.Equal(t, "button", doc.Nodes["b"].ComponentRef.BuiltinType)

	var generic map[string]any
	require.NoError(t, json.Unmarshal(out, &generic))
	assert.Equal(t, "keep", generic["nodes"].(map[string]any)["root"].(map[string]any)["custom"])
}

func TestUpgradeLegacy_Idempotent(t *testing.T) {
	out, err := domain.UpgradeLegacy([]byte(validDoc))
	require.NoError(t, err)
	assert.JSONEq(t, validDoc, string(out))
}

func TestUpgradeLegacy_RejectsUnknownVersion(t *testing.T) {
	_, err := domain.UpgradeLegacy([]byte(`{"schemaVersion":"7.0.0"}`))
	assert.ErrorIs(t, err, domain.ErrInvalidDocument)
}
