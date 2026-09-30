// Package catalogschema projects the catalog's protobuf contracts into
// per-file JSON schema documents -- the machine-readable shape consoles and
// tools render API contracts from without compiling protos.
//
// The projection this package produces is a PUBLISHED CONTRACT shipped as
// release cargo (catalog-schemas.zip, one JSON per catalog .proto): extend it
// additively, never rename or repurpose fields casually. Two declared
// fidelity boundaries, carried from the shape's first consumer: message-level
// options (e.g. buf.validate.message CEL rules) are not reconstructed --
// consumers render them from rawContent -- and field-level rawOptions
// reconstruct only the common validate rules; rawContent carries the full
// authored source for everything the structured fields do not.
package catalogschema

// ProtoFile is the top-level representation of one parsed .proto file.
//
// JSON field names use camelCase; null-or-absent semantics use pointer types
// (*ProtoFieldSemantic, *string).
type ProtoFile struct {
	Syntax      string         `json:"syntax"`
	PackageName string         `json:"packageName"`
	Imports     []string       `json:"imports"`
	Messages    []ProtoMessage `json:"messages"`
	Enums       []ProtoEnum    `json:"enums"`
	Options     []string       `json:"options"`
	// RawContent is the authored .proto source, verbatim -- the fidelity
	// escape hatch for everything the structured fields do not reconstruct.
	RawContent string `json:"rawContent"`
}

// ProtoMessage represents a message block with nested messages, enums, and oneofs.
type ProtoMessage struct {
	Syntax         string         `json:"-"`
	Name           string         `json:"name"`
	Comment        string         `json:"comment"`
	Fields         []ProtoField   `json:"fields"`
	NestedMessages []ProtoMessage `json:"nestedMessages"`
	NestedEnums    []ProtoEnum    `json:"nestedEnums"`
	Oneofs         []ProtoOneof   `json:"oneofs"`
	RawOptions     []string       `json:"rawOptions"`
}

// ProtoField represents a single field within a message or oneof.
type ProtoField struct {
	Name         string              `json:"name"`
	Number       int32               `json:"number"`
	Type         string              `json:"type"`
	Label        string              `json:"label"`
	Comment      string              `json:"comment"`
	RawOptions   string              `json:"rawOptions"`
	IsMap        bool                `json:"isMap"`
	MapKeyType   string              `json:"mapKeyType"`
	MapValueType string              `json:"mapValueType"`
	Semantics    *ProtoFieldSemantic `json:"semantics"`
}

// ProtoFieldSemantic holds well-known annotations extracted from field options.
type ProtoFieldSemantic struct {
	IsRequired     bool    `json:"isRequired"`
	DefaultValue   *string `json:"defaultValue"`
	ForeignKeyKind *string `json:"foreignKeyKind"`
	// ForeignKeyCandidates lists every kind and output a reference field
	// composes from (its default first, then each declared candidate), set
	// only when the field declares candidates.
	ForeignKeyCandidates []ForeignKeyCandidate `json:"foreignKeyCandidates,omitempty"`
}

// ForeignKeyCandidate is one kind a reference field can point at and the
// output path it composes from.
type ForeignKeyCandidate struct {
	Kind      string `json:"kind"`
	FieldPath string `json:"fieldPath"`
}

// ProtoEnum represents an enum block.
type ProtoEnum struct {
	Name    string           `json:"name"`
	Comment string           `json:"comment"`
	Values  []ProtoEnumValue `json:"values"`
}

// ProtoEnumValue represents a single value within an enum.
type ProtoEnumValue struct {
	Name    string `json:"name"`
	Number  int32  `json:"number"`
	Comment string `json:"comment"`
}

// ProtoOneof represents a oneof group within a message.
type ProtoOneof struct {
	Name   string       `json:"name"`
	Fields []ProtoField `json:"fields"`
}
