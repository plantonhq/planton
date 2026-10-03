//go:build !codegen
// +build !codegen

package moduleverify

import (
	"fmt"

	"github.com/plantonhq/planton/pkg/outputs"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// noOverride mirrors outputs.OverrideNone for callers deciding whether the
// generic by-name transformation applies.
const noOverride = outputs.OverrideNone

// checkOutputsOverride validates the module's output-transformation override
// (when one exists) and, given sample outputs, dry-runs the full
// transformation — the same machinery `planton validate-outputs` fronts.
// Returns which override mechanism was discovered so the caller knows
// whether by-name output matching applies.
func checkOutputsOverride(kind catalogkind.CatalogKind, moduleDir string, sampleOutputs map[string]interface{}, result *Result) outputs.OverrideKind {
	validation, err := outputs.ValidateOverride(kind, moduleDir, sampleOutputs)
	if err != nil {
		result.addError("", fmt.Sprintf("the outputs transformation could not be validated: %v", err))
		return noOverride
	}

	// With no override present the generic transformer applies — the normal
	// case, not a finding (ValidateOverride words it as a warning for its
	// own low-level audience).
	if validation.OverrideType != noOverride {
		for _, schemaErr := range validation.SchemaErrors {
			result.addError("", fmt.Sprintf("outputs override: %s", schemaErr))
		}
		for _, schemaWarn := range validation.SchemaWarnings {
			result.addWarning("", fmt.Sprintf("outputs override: %s", schemaWarn))
		}
	}

	if validation.DryRun != nil {
		for _, dryRunErr := range validation.DryRun.Errors {
			result.addError("", fmt.Sprintf("outputs dry-run: %s", dryRunErr))
		}
		for _, unmatched := range validation.DryRun.UnmappedOutputs {
			result.addWarning("", fmt.Sprintf(
				"outputs dry-run: the sample output %q matched no outputs field and would be dropped", unmatched))
		}
	}

	return validation.OverrideType
}
