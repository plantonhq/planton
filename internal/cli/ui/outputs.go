package ui

import (
	"fmt"
	"sort"

	"github.com/plantonhq/planton/pkg/outputs"
)

// sensitivePlaceholder is what a sensitive output's value renders as —
// terminal and CI logs are persistent, so secret values never print.
const sensitivePlaceholder = "(sensitive)"

// OutputsSummary displays the stack's captured outputs after a
// successful apply, terraform-style: every output listed, secret values
// masked. Rendering is the leak boundary for captured outputs — the capture
// itself deliberately holds real secret values for downstream use, so this
// is the one function that must never show them.
func OutputsSummary(result *outputs.CaptureResult) {
	lines := FormatOutputLines(result)
	if len(lines) == 0 {
		return
	}
	fmt.Println()
	fmt.Printf("%s  %s\n",
		successIcon.Render("📤"),
		successTitle.Render("Outputs"))
	for _, line := range lines {
		fmt.Printf("   %s\n", line)
	}
	fmt.Println()
}

// FormatOutputLines renders the outputs as sorted "key: value" lines
// with sensitive values masked. Split from the printer so the masking law —
// a sensitive value never appears in rendered output — is unit-testable.
func FormatOutputLines(result *outputs.CaptureResult) []string {
	if result == nil || len(result.Flat) == 0 {
		return nil
	}

	keys := make([]string, 0, len(result.Flat))
	for key := range result.Flat {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		value := result.Flat[key]
		if result.IsSensitive(key) {
			value = sensitivePlaceholder
		}
		lines = append(lines, fmt.Sprintf("%s %s",
			dimStyle.Render(key+":"),
			infoMessage.Render(value)))
	}
	return lines
}
