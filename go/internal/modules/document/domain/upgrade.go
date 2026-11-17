package domain

import (
	"encoding/json"
	"fmt"
)

// UpgradeLegacy converts a schema 1.0.0 document (widget nodes with `type` and
// `geometry`) into 1.1.0 (ADR-01): type -> kind=frame + componentRef.builtinType,
// geometry -> transform + size. Unknown fields are preserved. Documents already
// at 1.1.0 are returned unchanged.
func UpgradeLegacy(raw []byte) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}

	switch v := doc["schemaVersion"].(type) {
	case string:
		if v == SchemaCurrent {
			return raw, nil
		}
		if v != SchemaLegacy {
			return nil, fmt.Errorf("%w: unsupported schema version %q", ErrInvalidDocument, v)
		}
	case float64:
		if v != 1 {
			return nil, fmt.Errorf("%w: unsupported schema version %v", ErrInvalidDocument, v)
		}
	default:
		return nil, fmt.Errorf("%w: missing schemaVersion", ErrInvalidDocument)
	}

	nodes, _ := doc["nodes"].(map[string]any)
	for id, rawNode := range nodes {
		node, ok := rawNode.(map[string]any)
		if !ok {
			continue
		}
		if t, ok := node["type"].(string); ok {
			delete(node, "type")
			node["kind"] = string(KindFrame)
			if text, isText := legacyText(t); isText {
				node["kind"] = text
			}
			node["componentRef"] = map[string]any{"builtinType": t}
		}
		if g, ok := node["geometry"].(map[string]any); ok {
			delete(node, "geometry")
			node["transform"] = map[string]any{
				"x": numOr(g["x"], 0), "y": numOr(g["y"], 0), "rotation": numOr(g["rotation"], 0),
			}
			node["size"] = map[string]any{
				"width": numOr(g["width"], 0), "height": numOr(g["height"], 0),
			}
		}
		if rawProps, ok := node["properties"].(map[string]any); ok {
			node["data"] = rawProps
			delete(node, "properties")
		}
		nodes[id] = node
	}
	doc["schemaVersion"] = SchemaCurrent
	return json.Marshal(doc)
}

// legacyText maps the legacy `text` and `image` builtins to their primitive kinds.
func legacyText(t string) (string, bool) {
	switch t {
	case "text":
		return string(KindText), true
	case "image":
		return string(KindImage), true
	}
	return "", false
}

func numOr(v any, def float64) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return def
}
