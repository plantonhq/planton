package presetvalidity

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

// A preset's guide (its .md sidecar) ends with a "Placeholders to Replace"
// table: one row per `<placeholder>` the user must swap before applying the
// manifest. The table is only useful while it names the manifest's own
// placeholders. A row whose placeholder the manifest never carries sends the
// reader hunting for a value that is not there, and usually hides the value
// that IS there under another name (the row survived a rename of the
// placeholder, or a field that moved out of the preset).
//
// The check is textual on purpose: a placeholder is an angle-bracket token
// in the row's first cell, and the manifest carries it when the same token
// appears anywhere in the YAML bytes (a value, a comment, a key). Only
// angle-bracket tokens are checked -- pattern-valid placeholders such as
// 123456789012 are ordinary values the YAML may spell with quotes or inside
// a longer string, and matching them would be guesswork.

// placeholderToken is one angle-bracket placeholder: `<aws-region>`,
// `<private-subnet-a-resource-name>`. It excludes nested brackets so HTML
// or a comparison ("a < b") never reads as a placeholder.
var placeholderToken = regexp.MustCompile(`<[^<>\s|]+>`)

// tableSeparator is a Markdown table's header rule cell: `---`, `:---:`.
var tableSeparator = regexp.MustCompile(`^:?-+:?$`)

// phantomRow is one row of a guide's placeholder table naming a placeholder
// the manifest does not carry.
type phantomRow struct {
	line        int
	placeholder string
}

// phantomPlaceholderRows returns every row of the guide's placeholder table
// whose first cell names an angle-bracket placeholder absent from the
// manifest. The table is the one under the first level-2 heading that begins
// with "Placeholders" ("Placeholders to Replace", "Placeholders"), up to the
// next level-2 heading.
func phantomPlaceholderRows(guide, manifest []byte) []phantomRow {
	var rows []phantomRow
	inTable := false
	for i, line := range strings.Split(string(guide), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			inTable = strings.HasPrefix(strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))), "placeholders")
			continue
		}
		if !inTable || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		first := strings.TrimSpace(strings.Split(strings.Trim(trimmed, "|"), "|")[0])
		if tableSeparator.MatchString(first) || strings.EqualFold(first, "placeholder") {
			continue
		}
		for _, token := range placeholderToken.FindAllString(first, -1) {
			if !strings.Contains(string(manifest), token) {
				rows = append(rows, phantomRow{line: i + 1, placeholder: token})
			}
		}
	}
	return rows
}

// checkPlaceholderTable reports the preset's phantom placeholder rows as one
// violation naming each row. presetPath is the repo-root-relative YAML path;
// the guide is its .md sibling.
func checkPlaceholderTable(presetPath string, guide, manifest []byte) []Violation {
	rows := phantomPlaceholderRows(guide, manifest)
	if len(rows) == 0 {
		return nil
	}
	guidePath := strings.TrimSuffix(presetPath, ".yaml") + ".md"
	named := make([]string, 0, len(rows))
	for _, r := range rows {
		named = append(named, fmt.Sprintf("line %d names %s", r.line, r.placeholder))
	}
	return []Violation{{
		Path: presetPath,
		Rule: RulePhantomPlaceholder,
		Detail: fmt.Sprintf(
			"the placeholder table in %s lists placeholders %s does not contain (%s); "+
				"the guide describes values the manifest does not ask for under those names, so a reader looks for them and misses the ones it does ask for; "+
				"delete each row whose value the manifest does not use, reword a row to name the placeholder the manifest carries for that value, "+
				"or add the placeholder to the manifest where the value belongs (the preset must still validate)",
			guidePath, path.Base(presetPath), strings.Join(named, "; ")),
	}}
}
