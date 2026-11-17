package domain

// NodeKind is the primitive kind of a design object (ADR-01).
// Widgets such as buttons are NOT kinds: they are frame/group nodes promoted to
// components (ComponentRef) or `instance` nodes (ADR-06).
type NodeKind string

const (
	KindFrame    NodeKind = "frame"
	KindGroup    NodeKind = "group"
	KindRect     NodeKind = "rect"
	KindEllipse  NodeKind = "ellipse"
	KindLine     NodeKind = "line"
	KindPolygon  NodeKind = "polygon"
	KindStar     NodeKind = "star"
	KindPath     NodeKind = "path"
	KindText     NodeKind = "text"
	KindImage    NodeKind = "image"
	KindInstance NodeKind = "instance"
)

var allKinds = map[NodeKind]struct{}{
	KindFrame: {}, KindGroup: {}, KindRect: {}, KindEllipse: {}, KindLine: {},
	KindPolygon: {}, KindStar: {}, KindPath: {}, KindText: {}, KindImage: {}, KindInstance: {},
}

func (k NodeKind) Valid() bool {
	_, ok := allKinds[k]
	return ok
}

// Container kinds may own children.
func (k NodeKind) IsContainer() bool {
	return k == KindFrame || k == KindGroup || k == KindInstance
}

// Transform is relative to the parent's local space (ADR-02).
type Transform struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation,omitempty"`
	FlipX    bool    `json:"flipX,omitempty"`
	FlipY    bool    `json:"flipY,omitempty"`
}

type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ComponentRef links a node to a component definition (ADR-06) or, for
// documents migrated from schema 1.0.0, to a builtin widget type.
type ComponentRef struct {
	ComponentID string `json:"componentId,omitempty"`
	BuiltinType string `json:"builtinType,omitempty"`
	Version     string `json:"version,omitempty"`
}

// DesignNode is the typed view of a node used for validation. Documents are
// persisted as the original JSON, so unknown fields are never lost.
type DesignNode struct {
	ID       string   `json:"id"`
	Kind     NodeKind `json:"kind,omitempty"`
	Type     string   `json:"type,omitempty"` // legacy 1.0.0 builtin component type
	Name     string   `json:"name,omitempty"`
	ParentID *string  `json:"parentId,omitempty"`
	Children []string `json:"children"`

	Transform    *Transform     `json:"transform,omitempty"`
	Size         *Size          `json:"size,omitempty"`
	Style        map[string]any `json:"style,omitempty"`
	Layout       map[string]any `json:"layout,omitempty"`
	Constraints  map[string]any `json:"constraints,omitempty"`
	Data         map[string]any `json:"data,omitempty"`
	ComponentRef *ComponentRef  `json:"componentRef,omitempty"`

	Visible *bool `json:"visible,omitempty"`
	Locked  bool  `json:"locked,omitempty"`
}
