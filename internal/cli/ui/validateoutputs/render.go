//go:build !codegen
// +build !codegen

package validateoutputs

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/outputs"
)

var (
	colorRed     = lipgloss.Color("#FF6B6B")
	colorGreen   = lipgloss.Color("#69DB7C")
	colorYellow  = lipgloss.Color("#FFD43B")
	colorBlue    = lipgloss.Color("#74C0FC")
	colorGray    = lipgloss.Color("#868E96")
	colorDimGray = lipgloss.Color("#495057")
	colorWhite   = lipgloss.Color("#DEE2E6")

	errorIcon    = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	errorTitle   = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	errorMsg     = lipgloss.NewStyle().Foreground(colorWhite)
	successIcon  = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	successTitle = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	warningIcon  = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	warningTitle = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	infoIcon     = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	infoMsg      = lipgloss.NewStyle().Foreground(colorWhite)
	pathStyle    = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	cmdStyle     = lipgloss.NewStyle().Foreground(colorYellow)
	dimStyle     = lipgloss.NewStyle().Foreground(colorGray)
	hintStyle    = lipgloss.NewStyle().Foreground(colorDimGray).Italic(true)
	labelStyle   = lipgloss.NewStyle().Foreground(colorGray)
	valueStyle   = lipgloss.NewStyle().Foreground(colorWhite)
)

const (
	iconError   = "✗"
	iconSuccess = "✓"
	iconWarning = "!"
	iconTip     = "💡"
	iconArrow   = "←"
	sepChar     = "═"
	sepLen      = 80
)

func sep(style lipgloss.Style) string {
	return style.Render(strings.Repeat(sepChar, sepLen))
}

// RenderUnknownKind displays a contextual error when the user provides an
// unrecognized CloudResourceKind, with a "did you mean?" suggestion and
// example valid kinds.
func RenderUnknownKind(kindName string) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s  %s\n",
		errorIcon.Render(iconError),
		errorTitle.Render("Unknown Cloud Resource Kind"))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s is not a recognized CloudResourceKind.\n",
		errorTitle.Render(fmt.Sprintf("%q", kindName)))

	if suggestion := suggestSimilarKind(kindName); suggestion != "" {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "   Did you mean: %s\n", cmdStyle.Render(suggestion))
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s Example valid kinds: %s\n",
		dimStyle.Render("Hint:"),
		hintStyle.Render("AwsVpc, Auth0ResourceServer, GcpCloudSql, KubernetesPostgres, AzureAksCluster"))

	fmt.Fprintln(os.Stderr)
}

// RenderModuleDirNotFound displays a helpful error when the module directory
// does not exist, explaining what the directory should contain.
func RenderModuleDirNotFound(moduleDir string) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s  %s\n",
		errorIcon.Render(iconError),
		errorTitle.Render("Module Directory Not Found"))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   The directory does not exist: %s\n", pathStyle.Render(moduleDir))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s The module directory should contain your IaC module files:\n", dimStyle.Render("Expected:"))
	fmt.Fprintf(os.Stderr, "   %s  %s for Terraform modules\n", dimStyle.Render("•"), infoMsg.Render("*.tf files"))
	fmt.Fprintf(os.Stderr, "   %s  %s for output transformation overrides\n", dimStyle.Render("•"), infoMsg.Render("output_transform.yaml"))
	fmt.Fprintf(os.Stderr, "   %s  %s for programmatic overrides\n", dimStyle.Render("•"), infoMsg.Render("transform-outputs"))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s %s\n",
		dimStyle.Render("Hint:"),
		hintStyle.Render(fmt.Sprintf("Verify the path: ls -la %s", moduleDir)))

	fmt.Fprintln(os.Stderr)
}

// RenderSampleFileError displays a helpful error when the sample outputs
// file cannot be read.
func RenderSampleFileError(path string, err error) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s  %s\n",
		errorIcon.Render(iconError),
		errorTitle.Render("Failed to Read Sample Outputs"))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   Could not read file: %s\n", pathStyle.Render(path))
	fmt.Fprintf(os.Stderr, "   %s %s\n", dimStyle.Render("Error:"), errorMsg.Render(err.Error()))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s Provide a JSON file containing raw IaC outputs:\n", dimStyle.Render("Expected format:"))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s\n", cmdStyle.Render(`{"vpc_id": "vpc-123", "port": 5432, "is_public": true}`))

	fmt.Fprintln(os.Stderr)
}

// RenderSampleParseError displays a helpful error when the sample outputs
// file contains invalid JSON.
func RenderSampleParseError(path string, err error) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s  %s\n",
		errorIcon.Render(iconError),
		errorTitle.Render("Invalid Sample Outputs JSON"))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   File: %s\n", pathStyle.Render(path))
	fmt.Fprintf(os.Stderr, "   %s %s\n", dimStyle.Render("Parse error:"), errorMsg.Render(err.Error()))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s The file must contain a JSON object with string keys:\n", dimStyle.Render("Expected format:"))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s\n", cmdStyle.Render(`{"vpc_id": "vpc-123", "port": 5432, "is_public": true}`))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s %s\n",
		dimStyle.Render("Hint:"),
		hintStyle.Render("Values can be strings, numbers, booleans, arrays, or nested objects."))

	fmt.Fprintln(os.Stderr)
}

// RenderValidationInternalError displays an error when the validation library
// itself fails unexpectedly.
func RenderValidationInternalError(err error) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s  %s\n",
		errorIcon.Render(iconError),
		errorTitle.Render("Validation Failed"))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s\n", errorMsg.Render(err.Error()))

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %s %s\n",
		dimStyle.Render("Hint:"),
		hintStyle.Render("This may be a bug. Please report it with the full command and error output."))

	fmt.Fprintln(os.Stderr)
}

// RenderValidationSuccess displays the success banner with optional dry-run
// results in a rich, structured format.
func RenderValidationSuccess(kindName, moduleDir string, result *outputs.ValidationResult) {
	greenSep := sep(successIcon)

	fmt.Println()
	fmt.Println(greenSep)
	fmt.Printf("%s  %s\n",
		successIcon.Render(iconSuccess),
		successTitle.Render("Output Override Validation Passed"))
	fmt.Println(greenSep)

	renderOverrideSummary(kindName, moduleDir, result.OverrideType)

	if len(result.SchemaWarnings) > 0 {
		renderWarnings(result.SchemaWarnings)
	}

	if result.DryRun != nil {
		renderDryRunResults(result.DryRun)
	}

	fmt.Println(greenSep)
}

// RenderValidationFailure displays the failure banner with schema errors,
// warnings, and actionable guidance.
func RenderValidationFailure(kindName, moduleDir string, result *outputs.ValidationResult) {
	redSep := sep(errorIcon)

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, redSep)
	fmt.Fprintf(os.Stderr, "%s  %s\n",
		errorIcon.Render(iconError),
		errorTitle.Render("Output Override Validation Failed"))
	fmt.Fprintln(os.Stderr, redSep)

	renderOverrideSummaryStderr(kindName, moduleDir, result.OverrideType)

	if len(result.SchemaErrors) > 0 {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "   %s\n", errorTitle.Render("Schema Errors:"))
		fmt.Fprintln(os.Stderr)
		for _, e := range result.SchemaErrors {
			fmt.Fprintf(os.Stderr, "   %s %s\n",
				errorIcon.Render(iconError),
				errorMsg.Render(e))
		}
	}

	if len(result.SchemaWarnings) > 0 {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "   %s\n", warningTitle.Render("Warnings:"))
		fmt.Fprintln(os.Stderr)
		for _, w := range result.SchemaWarnings {
			fmt.Fprintf(os.Stderr, "   %s %s\n",
				warningIcon.Render(iconWarning),
				dimStyle.Render(w))
		}
	}

	if result.DryRun != nil && len(result.DryRun.Errors) > 0 {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "   %s\n", errorTitle.Render("Transformation Errors:"))
		fmt.Fprintln(os.Stderr)
		for _, e := range result.DryRun.Errors {
			fmt.Fprintf(os.Stderr, "   %s %s\n",
				errorIcon.Render(iconError),
				errorMsg.Render(e))
		}
	}

	renderFailureTip(kindName, moduleDir, result)

	fmt.Fprintln(os.Stderr, redSep)
}

func renderOverrideSummary(kindName, moduleDir string, override outputs.OverrideKind) {
	fmt.Println()
	fmt.Printf("   %-12s %s\n", labelStyle.Render("Kind:"), valueStyle.Render(kindName))
	fmt.Printf("   %-12s %s\n", labelStyle.Render("Module:"), pathStyle.Render(moduleDir))
	fmt.Printf("   %-12s %s\n", labelStyle.Render("Override:"), valueStyle.Render(overrideLabel(override)))
}

func renderOverrideSummaryStderr(kindName, moduleDir string, override outputs.OverrideKind) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "   %-12s %s\n", labelStyle.Render("Kind:"), valueStyle.Render(kindName))
	fmt.Fprintf(os.Stderr, "   %-12s %s\n", labelStyle.Render("Module:"), pathStyle.Render(moduleDir))
	fmt.Fprintf(os.Stderr, "   %-12s %s\n", labelStyle.Render("Override:"), valueStyle.Render(overrideLabel(override)))
}

func overrideLabel(o outputs.OverrideKind) string {
	switch o {
	case outputs.OverrideExecutable:
		return "transform-outputs (executable)"
	case outputs.OverrideMapping:
		return "output_transform.yaml (mapping)"
	default:
		return "none (generic reflection transformer)"
	}
}

func renderWarnings(warnings []string) {
	fmt.Println()
	for _, w := range warnings {
		fmt.Printf("   %s %s\n",
			warningIcon.Render(iconWarning),
			dimStyle.Render(w))
	}
}

func renderDryRunResults(dr *outputs.DryRunResult) {
	fmt.Println()
	fmt.Printf("   %s %s\n",
		infoIcon.Render("→"),
		infoMsg.Render(fmt.Sprintf("Dry-run: %d/%d proto fields populated",
			dr.PopulatedCount, dr.TotalProtoFields)))

	if len(dr.PopulatedFields) > 0 {
		fmt.Println()

		maxFieldLen := 0
		for _, f := range dr.PopulatedFields {
			if len(f.ProtoField) > maxFieldLen {
				maxFieldLen = len(f.ProtoField)
			}
		}

		for _, f := range dr.PopulatedFields {
			src := f.SourceKey
			if src == "" {
				src = "(indirect)"
			}
			fmt.Printf("   %s %-*s  %s  %s\n",
				successIcon.Render(iconSuccess),
				maxFieldLen,
				valueStyle.Render(f.ProtoField),
				dimStyle.Render(iconArrow),
				dimStyle.Render(src))
		}
	}

	if len(dr.UnmappedOutputs) > 0 {
		fmt.Println()
		fmt.Printf("   %s\n", warningTitle.Render("Unmapped outputs (no matching proto field):"))
		for _, key := range dr.UnmappedOutputs {
			fmt.Printf("   %s %s\n",
				warningIcon.Render(iconWarning),
				dimStyle.Render(key))
		}
	}

	unpopulated := dr.TotalProtoFields - dr.PopulatedCount
	if unpopulated > 0 && dr.PopulatedCount > 0 {
		fmt.Println()
		fmt.Printf("   %s\n",
			dimStyle.Render(fmt.Sprintf("%d proto fields not populated (no matching output key)", unpopulated)))
	}

	if len(dr.UnmappedOutputs) > 0 {
		fmt.Println()
		fmt.Printf("   %s %s\n",
			infoIcon.Render(iconTip),
			hintStyle.Render("Unmapped outputs are safe to ignore if intentional. Add them to"))
		fmt.Printf("   %s\n",
			hintStyle.Render("   the 'skip' list in output_transform.yaml to suppress this message."))
	}

	if len(dr.Errors) > 0 {
		fmt.Println()
		fmt.Printf("   %s\n", errorTitle.Render("Transformation Errors:"))
		for _, e := range dr.Errors {
			fmt.Printf("   %s %s\n",
				errorIcon.Render(iconError),
				errorMsg.Render(e))
		}
	}
}

func renderFailureTip(kindName, moduleDir string, result *outputs.ValidationResult) {
	fmt.Fprintln(os.Stderr)

	switch result.OverrideType {
	case outputs.OverrideMapping:
		fmt.Fprintf(os.Stderr, "   %s %s\n",
			infoIcon.Render(iconTip),
			hintStyle.Render("Check the StackOutputs proto for "+kindName+" to see valid field names."))
		fmt.Fprintf(os.Stderr, "   %s\n",
			hintStyle.Render("   Mapping targets (right side) must match proto field names exactly."))
	case outputs.OverrideExecutable:
		fmt.Fprintf(os.Stderr, "   %s %s\n",
			infoIcon.Render(iconTip),
			hintStyle.Render("Verify the transform-outputs executable runs correctly:"))
		fmt.Fprintf(os.Stderr, "   %s\n",
			cmdStyle.Render(fmt.Sprintf("   echo '{\"kind\":\"%s\",\"outputs\":{}}' | %s/transform-outputs",
				kindName, moduleDir)))
	default:
		fmt.Fprintf(os.Stderr, "   %s %s\n",
			infoIcon.Render(iconTip),
			hintStyle.Render("Add an output_transform.yaml or transform-outputs executable to your module"))
		fmt.Fprintf(os.Stderr, "   %s\n",
			hintStyle.Render("   directory to customize how outputs map to proto fields."))
	}

	fmt.Fprintln(os.Stderr)
}

// suggestSimilarKind finds the closest CloudResourceKind name using
// levenshtein distance against all registered kinds.
func suggestSimilarKind(input string) string {
	inputLower := strings.ToLower(input)
	bestDist := len(input)
	bestMatch := ""

	for _, kind := range crkreflect.KindsList() {
		if kind == cloudresourcekind.CloudResourceKind_unspecified {
			continue
		}
		name := kind.String()
		dist := levenshteinDistance(inputLower, strings.ToLower(name))
		if dist < bestDist {
			bestDist = dist
			bestMatch = name
		}
	}

	if bestDist <= 3 && bestMatch != "" {
		return bestMatch
	}
	return ""
}

func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := matrix[i-1][j] + 1
			ins := matrix[i][j-1] + 1
			sub := matrix[i-1][j-1] + cost
			best := del
			if ins < best {
				best = ins
			}
			if sub < best {
				best = sub
			}
			matrix[i][j] = best
		}
	}

	return matrix[len(a)][len(b)]
}

// RenderAvailableProtoFields prints the available StackOutputs field names
// for a kind, useful when a mapping target is invalid.
func RenderAvailableProtoFields(kindName string, kind cloudresourcekind.CloudResourceKind) {
	fields := collectStackOutputsFieldNames(kind)
	if len(fields) == 0 {
		return
	}

	sort.Strings(fields)
	fmt.Fprintf(os.Stderr, "\n   %s %s\n",
		dimStyle.Render("Available fields:"),
		hintStyle.Render(strings.Join(fields, ", ")))
}

func collectStackOutputsFieldNames(kind cloudresourcekind.CloudResourceKind) []string {
	instance, err := crkreflect.NewInstance(kind)
	if err != nil {
		return nil
	}

	ref := instance.ProtoReflect()
	statusFd := ref.Descriptor().Fields().ByName("status")
	if statusFd == nil {
		return nil
	}
	statusMsg := ref.Mutable(statusFd).Message()
	outputsFd := statusMsg.Descriptor().Fields().ByName("outputs")
	if outputsFd == nil {
		return nil
	}
	outputsDesc := outputsFd.Message()
	if outputsDesc == nil {
		return nil
	}

	var names []string
	for i := 0; i < outputsDesc.Fields().Len(); i++ {
		names = append(names, string(outputsDesc.Fields().Get(i).Name()))
	}
	return names
}
