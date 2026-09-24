package explain

import (
	"fmt"
	"strings"
)

// Render formats a report as human-readable reference text, in the visual
// grain of `kubectl explain`: an uppercase header block, then the field tree
// with documentation indented under each field. Machine consumers use the
// JSON report; this view is for people reading a terminal.
func Render(report *Report) string {
	var b strings.Builder

	writeHeader(&b, "KIND", report.Kind)
	if report.ApiVersion != "" {
		writeHeader(&b, "VERSION", report.ApiVersion)
	}
	if report.Path != "" {
		writeHeader(&b, "PATH", report.Path)
	}

	if report.Field != nil {
		b.WriteString("\n")
		renderFieldTree(&b, *report.Field, 0, true)
		return b.String()
	}

	if report.Doc != "" {
		b.WriteString("\nDESCRIPTION:\n")
		writeIndented(&b, report.Doc, 1)
	}
	// The envelope skeleton grounds the one thing no field listing teaches:
	// what the top level of the manifest looks like.
	b.WriteString("\nMANIFEST:\n")
	writeIndented(&b, "apiVersion: "+report.ApiVersion, 1)
	writeIndented(&b, "kind: "+report.Kind, 1)
	writeIndented(&b, "metadata:", 1)
	writeIndented(&b, "name: <resource-name>", 2)
	writeIndented(&b, "spec:", 1)
	writeIndented(&b, "<the fields below>", 2)
	if len(report.SpecRules) > 0 {
		b.WriteString("\nRULES:\n")
		for _, rule := range report.SpecRules {
			writeIndented(&b, "- "+rule.Message, 1)
		}
	}
	if len(report.Spec) > 0 {
		b.WriteString("\nSPEC:\n")
		for _, f := range report.Spec {
			renderFieldTree(&b, f, 1, false)
		}
	}
	if len(report.Outputs) > 0 {
		b.WriteString("\nOUTPUTS (reference as valueFrom fieldPath: status.outputs.<name>):\n")
		for _, f := range report.Outputs {
			renderFieldTree(&b, f, 1, false)
		}
	}
	return b.String()
}

func writeHeader(b *strings.Builder, key, value string) {
	fmt.Fprintf(b, "%-9s%s\n", key+":", value)
}

// renderFieldTree prints one field and its children. expandEnumDocs is true
// only for the single-field view, where per-value documentation is the point
// of the drill-down; in tree views enum values list inline to keep the tree
// scannable.
func renderFieldTree(b *strings.Builder, f Field, depth int, expandEnumDocs bool) {
	var markers []string
	if f.Required {
		markers = append(markers, "-required-")
	}
	if f.Optional {
		markers = append(markers, "-optional-")
	}
	if f.Sensitive {
		markers = append(markers, "(sensitive)")
	}
	if f.SecretHome != "" {
		markers = append(markers, "(no secrets: use "+f.SecretHome+")")
	}
	if f.Provenance != "" {
		markers = append(markers, "("+f.Provenance+")")
	}
	line := fmt.Sprintf("%s <%s>", f.Name, f.Type)
	if len(markers) > 0 {
		line += " " + strings.Join(markers, " ")
	}
	writeIndented(b, line, depth)

	body := depth + 1
	if f.Doc != "" {
		writeIndented(b, f.Doc, body)
	}
	if f.SecretHome != "" {
		writeIndented(b, "secrets: stored where anyone who can view the resource reads it, so a secret reference ($secret/...) here is refused -- put a secret in "+f.SecretHome, body)
	}
	if f.RecommendedDefault != "" {
		writeIndented(b, "default: "+f.RecommendedDefault, body)
	}
	if f.RefKind != "" || f.RefFieldPath != "" {
		ref := "references: " + f.RefKind
		if f.RefFieldPath != "" {
			ref += " (" + f.RefFieldPath + ")"
		}
		writeIndented(b, ref, body)
	}
	if len(f.Enum) > 0 {
		switch {
		case expandEnumDocs:
			writeIndented(b, "values (use exactly as shown):", body)
			for _, v := range f.Enum {
				writeIndented(b, "- "+v.Name, body+1)
				if v.Doc != "" {
					writeIndented(b, v.Doc, body+2)
				}
			}
		// Catalog-scale enums (the cloud-resource kind list is 400+) would
		// bury every other field in a tree view; name the count and the
		// drill-down instead.
		case len(f.Enum) > 15:
			writeIndented(b, fmt.Sprintf("values: <%d values -- drill into this field to list them>", len(f.Enum)), body)
		default:
			names := make([]string, 0, len(f.Enum))
			for _, v := range f.Enum {
				names = append(names, v.Name)
			}
			writeIndented(b, "values: "+strings.Join(names, ", "), body)
		}
	}
	for _, constraint := range f.Constraints {
		writeIndented(b, "rule: "+constraint, body)
	}
	if len(f.Fields) > 0 {
		for _, child := range f.Fields {
			// Per-value enum documentation expands only on the field the
			// user asked about; a nested catalog-scale enum (400+ kinds,
			// each with prose) would otherwise bury the answer.
			renderFieldTree(b, child, body, false)
		}
	}
	b.WriteString("\n")
}

const indentUnit = "  "

func writeIndented(b *strings.Builder, text string, depth int) {
	prefix := strings.Repeat(indentUnit, depth)
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			b.WriteString("\n")
			continue
		}
		b.WriteString(prefix + line + "\n")
	}
}
