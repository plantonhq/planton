package generators

import (
	"fmt"
	"strings"
)

// TFType represents a Terraform type expression. Implementations format
// themselves as valid HCL type constraint syntax.
type TFType interface {
	Format(indent int) string
}

// TFPrimitive represents a Terraform primitive type: string, number, or bool.
type TFPrimitive string

func (p TFPrimitive) Format(_ int) string {
	return string(p)
}

// TFList represents a Terraform list type: list(elementType).
type TFList struct {
	Elem TFType
}

func (l TFList) Format(indent int) string {
	return fmt.Sprintf("list(%s)", l.Elem.Format(indent))
}

// TFMap represents a Terraform map type: map(valueType). Proto map<string, V>
// fields map to this type.
type TFMap struct {
	Value TFType
}

func (m TFMap) Format(indent int) string {
	return fmt.Sprintf("map(%s)", m.Value.Format(indent))
}

// TFObject represents a Terraform object type with named fields.
type TFObject struct {
	Fields []TFField
}

// TFField holds a single field within a TFObject.
type TFField struct {
	Name        string
	Description string
	Type        TFType
}

func (o TFObject) Format(indent int) string {
	if len(o.Fields) == 0 {
		return "object({})"
	}

	indentStr := strings.Repeat("  ", indent)
	nextIndent := strings.Repeat("  ", indent+1)

	var lines []string
	for _, f := range o.Fields {
		if f.Description != "" {
			lines = append(lines, "")
			for _, cl := range strings.Split(f.Description, "\n") {
				lines = append(lines, fmt.Sprintf("%s# %s", nextIndent, cl))
			}
		}
		lines = append(lines, fmt.Sprintf("%s%s = %s", nextIndent, f.Name, f.Type.Format(indent+1)))
	}

	return fmt.Sprintf("object({\n%s\n%s})", strings.Join(lines, "\n"), indentStr)
}
