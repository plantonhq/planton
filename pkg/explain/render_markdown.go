package explain

import (
	"fmt"
	"strings"
)

// RenderMarkdown formats a ROOT report as a complete, self-contained markdown
// reference page -- the committed per-kind reference file agents read and
// grep instead of invoking the CLI once per kind.
//
// The section headings below are a published search grammar: agents rely on
// patterns like `rg "^## Outputs"` and `rg "^### spec\."` holding across
// every generated file. Headings may be ADDED, but renaming or removing one
// breaks every consumer at once -- treat the vocabulary like a wire format.
//
//	# <Kind>            -- one page per kind, H1 is the kind name
//	## Example          -- minimal validated manifest
//	## Spec Fields      -- one table row per field, dotted YAML paths
//	## Field Details    -- one `### <path>` block per field: docs, rules, enums
//	## Validation Rules -- the spec's cross-field rules
//	## Outputs          -- deployment outputs table
//	## References       -- outbound foreign-key targets
//	## Referenced By    -- inbound foreign keys (catalog-computed, via options)
//	## See Also         -- links to hand-written essays (never duplicated here)
//
// The page head additionally carries a `**Guide**:` line linking the kind's
// authored GUIDE.md when one exists at the component root -- a line, not a
// heading, but its prefix is part of the same stable grammar.
//
// Path spelling is deliberately asymmetric: spec paths use protojson names
// (the exact keys an author writes in manifest YAML), while output paths use
// proto field names (snake_case) -- the canonical valueFrom fieldPath
// spelling the control plane stores and canonicalizes to.
//
// Output must be byte-deterministic for identical descriptors: everything
// renders in descriptor declaration order and nothing here may introduce
// randomness (freshness gates byte-compare committed files against a fresh
// render).
func RenderMarkdown(report *Report, opts MarkdownOptions) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", report.Kind)
	b.WriteString("> Generated from the protobuf schema by `make generate-reference` -- do not\n")
	b.WriteString("> edit by hand. To change a fact on this page, change the proto field comment\n")
	b.WriteString("> or validation rule it is derived from, then regenerate.\n\n")
	if report.ApiVersion != "" {
		fmt.Fprintf(&b, "**apiVersion**: `%s`\n\n", report.ApiVersion)
	}
	if opts.HasGuide {
		// The link target is deliberately hardcoded: the guide is always the
		// component root's GUIDE.md, one level above the version dir holding
		// this page -- in the repo, the zip, the skill, and the vendored copy
		// alike (all preserve the tree shape) -- so the file a contributor
		// edits is the file agents read, and the name never changes between
		// source and artifact.
		b.WriteString("**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.\n\n")
	}
	if report.Doc != "" {
		b.WriteString(strings.TrimSpace(report.Doc))
		b.WriteString("\n\n")
	}

	if opts.ExampleYAML != "" {
		b.WriteString("## Example\n\n")
		b.WriteString("```yaml\n")
		b.WriteString(strings.TrimRight(opts.ExampleYAML, "\n"))
		b.WriteString("\n```\n\n")
	}

	specRows := flattenFields("spec", report.Spec, authoredName)
	if len(specRows) > 0 {
		b.WriteString("## Spec Fields\n\n")
		writeFieldTable(&b, specRows)
		b.WriteString("\n## Field Details\n\n")
		for _, row := range specRows {
			writeFieldDetail(&b, row)
		}
	}

	if len(report.SpecRules) > 0 {
		b.WriteString("## Validation Rules\n\n")
		for _, rule := range report.SpecRules {
			if rule.Id != "" {
				fmt.Fprintf(&b, "- `%s`: %s\n", rule.Id, flattenText(rule.Message))
			} else {
				fmt.Fprintf(&b, "- %s\n", flattenText(rule.Message))
			}
		}
		b.WriteString("\n")
	}

	outputRows := flattenFields("status.outputs", report.Outputs, protoName)
	if len(outputRows) > 0 {
		b.WriteString("## Outputs\n\n")
		b.WriteString("Reference an output from another manifest as `valueFrom: {kind: " + report.Kind +
			", name: <resource-name>, fieldPath: status.outputs.<output>}`.\n\n")
		b.WriteString("| Output | Type | Description |\n")
		b.WriteString("|---|---|---|\n")
		for _, row := range outputRows {
			fmt.Fprintf(&b, "| `%s` | %s | %s |\n",
				row.Path, tableCell("`"+row.Field.Type+"`"), tableCell(row.Field.Doc))
		}
		b.WriteString("\n")
	}

	writeReferences(&b, specRows)
	writeReferencedBy(&b, opts.ReferencedBy)

	if len(opts.SeeAlso) > 0 {
		b.WriteString("## See Also\n\n")
		for _, link := range opts.SeeAlso {
			fmt.Fprintf(&b, "- [%s](%s)\n", link.Title, link.Path)
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

// MarkdownOptions carries what the schema alone cannot supply. The renderer
// owns the entire page format; callers provide content, never markdown
// structure.
type MarkdownOptions struct {
	// ExampleYAML is a minimal manifest embedded verbatim. Callers must only
	// pass YAML that already passed validation -- an invalid example teaches
	// every reader a broken shape.
	ExampleYAML string
	// ReferencedBy lists the catalog's inbound foreign keys into this kind
	// -- knowledge a single kind's schema cannot carry, so the caller (who
	// sees the whole catalog) computes it. Entries name kinds, never file
	// paths: kind names survive every artifact layout, and the catalog
	// indexes own the kind-to-file mapping.
	ReferencedBy []InboundRef
	// SeeAlso links to the kind's hand-written essays. Reference pages link
	// to prose, never duplicate it.
	SeeAlso []MarkdownLink
	// HasGuide adds the head link to the kind's authored GUIDE.md sibling.
	// Set only when the file actually exists beside the rendered page: the
	// line teaches readers a guide is one open away, and a link to a missing
	// file would teach every reader a falsehood. The `**Guide**:` prefix is
	// stable grammar (like the headings above) -- greppable across the pack.
	HasGuide bool
}

// InboundRef is one inbound foreign-key edge: a field on another kind whose
// valueFrom reads this kind.
type InboundRef struct {
	// Kind is the referencing kind.
	Kind string
	// FieldPath is the referencing field on that kind, in its spec-path
	// spelling (protojson, as rendered in that kind's own reference page).
	FieldPath string
	// TargetFieldPath is the field on THIS kind that the reference reads
	// (usually a status.outputs leaf, occasionally metadata.name or a spec
	// field), exactly as authored in the schema.
	TargetFieldPath string
}

// MarkdownLink is one See-Also entry, Path relative to the rendered file.
type MarkdownLink struct {
	Title string
	Path  string
}

// flatField is one field addressed by its full dotted YAML path -- the unit
// both the table and the detail blocks render, so the two views can never
// disagree about which fields exist.
type flatField struct {
	Path  string
	Field Field
}

// authoredName and protoName select which of a field's two names a path is
// built from: spec paths spell what an author writes in manifest YAML
// (protojson), output paths spell the canonical valueFrom fieldPath (proto
// field names -- see the RenderMarkdown doc comment).
func authoredName(f Field) string { return f.Name }
func protoName(f Field) string    { return f.ProtoName }

// flattenFields linearizes a field tree into dotted YAML paths in declaration
// order. List elements keep a `[]` marker (`spec.instances[].name`) and
// message-valued map entries a `.*` marker (`spec.envVars.*.value`), so a
// rendered path always spells out the YAML nesting an author writes.
func flattenFields(prefix string, fields []Field, nameOf func(Field) string) []flatField {
	out := make([]flatField, 0, len(fields))
	for _, f := range fields {
		path := prefix + "." + nameOf(f)
		out = append(out, flatField{Path: path, Field: f})
		childPrefix := path
		switch {
		case strings.HasPrefix(f.Type, "[]"):
			childPrefix += "[]"
		case strings.HasPrefix(f.Type, "map<"):
			childPrefix += ".*"
		}
		out = append(out, flattenFields(childPrefix, f.Fields, nameOf)...)
	}
	return out
}

func writeFieldTable(b *strings.Builder, rows []flatField) {
	b.WriteString("| Path | Type | Required | Default | References |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, row := range rows {
		f := row.Field
		typeCell := "`" + f.Type + "`"
		if f.Sensitive {
			typeCell += " (sensitive)"
		}
		if f.SecretHome != "" {
			typeCell += " (no secrets: use `" + f.SecretHome + "`)"
		}
		required := ""
		if f.Required {
			required = "yes"
		}
		def := ""
		if f.RecommendedDefault != "" {
			def = "`" + f.RecommendedDefault + "`"
		}
		ref := ""
		if f.RefKind != "" {
			ref = f.RefKind
			if f.RefFieldPath != "" {
				ref += " (`" + f.RefFieldPath + "`)"
			}
		}
		fmt.Fprintf(b, "| `%s` | %s | %s | %s | %s |\n",
			row.Path, tableCell(typeCell), required, tableCell(def), ref)
	}
}

func writeFieldDetail(b *strings.Builder, row flatField) {
	f := row.Field
	fmt.Fprintf(b, "### %s\n\n", row.Path)

	meta := []string{"`" + f.Type + "`"}
	if f.Required {
		meta = append(meta, "required")
	}
	if f.Optional {
		meta = append(meta, "optional (explicit presence)")
	}
	if f.Sensitive {
		meta = append(meta, "sensitive")
	}
	if f.SecretHome != "" {
		meta = append(meta, "no secrets")
	}
	if f.Provenance != "" {
		// Provenance names who writes the field when it is not the manifest
		// author -- surfaced so agents do not hand-author computed fields.
		meta = append(meta, f.Provenance+" (not author-written)")
	}
	b.WriteString(strings.Join(meta, " · ") + "\n\n")

	if f.Doc != "" {
		b.WriteString(strings.TrimSpace(f.Doc))
		b.WriteString("\n\n")
	}
	if f.SecretHome != "" {
		fmt.Fprintf(b, "- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `%s`, which keeps it in a secret store the workload reads by reference\n", f.SecretHome)
	}
	if f.RecommendedDefault != "" {
		fmt.Fprintf(b, "- default: `%s`\n", f.RecommendedDefault)
	}
	if f.RefKind != "" || f.RefFieldPath != "" {
		ref := "- references: " + f.RefKind
		if f.RefFieldPath != "" {
			ref += " (`" + f.RefFieldPath + "`)"
		}
		b.WriteString(ref + "\n")
	}
	for _, constraint := range f.Constraints {
		fmt.Fprintf(b, "- rule: %s\n", flattenText(constraint))
	}
	if f.RecommendedDefault != "" || f.RefKind != "" || f.RefFieldPath != "" || len(f.Constraints) > 0 {
		b.WriteString("\n")
	}

	if len(f.Enum) > 0 {
		b.WriteString("Allowed values (use exactly as shown):\n\n")
		for _, v := range f.Enum {
			if v.Doc != "" {
				fmt.Fprintf(b, "- `%s` -- %s\n", v.Name, flattenText(v.Doc))
			} else {
				fmt.Fprintf(b, "- `%s`\n", v.Name)
			}
		}
		b.WriteString("\n")
	}
}

// writeReferences renders the outbound half of the cross-reference story:
// which other kinds this kind's fields can point at. (The inbound half --
// "referenced by" -- needs whole-catalog knowledge and lives in the catalog
// index layer, not on a single kind's page.)
func writeReferences(b *strings.Builder, rows []flatField) {
	var refs []flatField
	for _, row := range rows {
		if row.Field.RefKind != "" {
			refs = append(refs, row)
		}
	}
	if len(refs) == 0 {
		return
	}
	b.WriteString("## References\n\n")
	b.WriteString("Fields that can point at another resource's outputs:\n\n")
	b.WriteString("| Field | Kind | Output |\n")
	b.WriteString("|---|---|---|\n")
	for _, row := range refs {
		fmt.Fprintf(b, "| `%s` | %s | `%s` |\n",
			row.Path, row.Field.RefKind, row.Field.RefFieldPath)
	}
	b.WriteString("\n")
}

// writeReferencedBy renders the inbound half of the cross-reference story:
// which fields elsewhere in the catalog can point at this kind. The caller
// computes the edges (a single kind's schema cannot know them); an empty
// list renders nothing, matching every other absent section.
func writeReferencedBy(b *strings.Builder, refs []InboundRef) {
	if len(refs) == 0 {
		return
	}
	b.WriteString("## Referenced By\n\n")
	b.WriteString("Fields on other kinds that can point at this resource:\n\n")
	b.WriteString("| Kind | Field | Reads |\n")
	b.WriteString("|---|---|---|\n")
	for _, ref := range refs {
		fmt.Fprintf(b, "| %s | `%s` | `%s` |\n", ref.Kind, ref.FieldPath, ref.TargetFieldPath)
	}
	b.WriteString("\n")
}

// ReferenceEdge is one outbound foreign-key edge of a kind: a spec field
// that can read another kind through a valueFrom reference.
type ReferenceEdge struct {
	// FieldPath is the referencing field in its spec-path spelling, exactly
	// as the kind's own reference page renders it.
	FieldPath string
	// Kind is the target kind name.
	Kind string
	// TargetFieldPath is the field the reference reads on the target kind,
	// exactly as authored in the schema (usually a status.outputs leaf,
	// occasionally metadata.name or a spec field).
	TargetFieldPath string
}

// ReferenceEdges extracts a root report's outbound foreign-key edges in
// declaration order. It shares the renderer's path flattening so an edge's
// FieldPath and the page's References table can never spell the same field
// differently.
func ReferenceEdges(report *Report) []ReferenceEdge {
	var edges []ReferenceEdge
	for _, row := range flattenFields("spec", report.Spec, authoredName) {
		if row.Field.RefKind == "" {
			continue
		}
		edges = append(edges, ReferenceEdge{
			FieldPath:       row.Path,
			Kind:            row.Field.RefKind,
			TargetFieldPath: row.Field.RefFieldPath,
		})
	}
	return edges
}

// flattenText collapses text onto one line for bullets and table cells.
func flattenText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// tableCell renders arbitrary text safe inside a markdown table row: one
// line, pipes escaped (type labels like `string | valueFrom` and doc prose
// would otherwise split the row).
func tableCell(text string) string {
	return strings.ReplaceAll(flattenText(text), "|", "\\|")
}
