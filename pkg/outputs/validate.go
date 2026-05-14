//go:build !codegen
// +build !codegen

package outputs

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ValidationResult holds the outcome of ValidateOverride.
type ValidationResult struct {
	// OverrideType reports which override mechanism was discovered.
	OverrideType OverrideKind

	// SchemaErrors are fatal problems that prevent transformation.
	SchemaErrors []string

	// SchemaWarnings are non-fatal concerns worth investigating.
	SchemaWarnings []string

	// DryRun is populated only when sample outputs are provided.
	DryRun *DryRunResult
}

// DryRunResult holds the outcome of running the full transformation
// pipeline against sample outputs.
type DryRunResult struct {
	// PopulatedFields lists each proto field that received a value.
	PopulatedFields []FieldResult

	// UnmappedOutputs lists output keys that had no matching proto field.
	UnmappedOutputs []string

	// TotalProtoFields is the number of fields on the StackOutputs message.
	TotalProtoFields int

	// PopulatedCount is how many fields were actually set.
	PopulatedCount int

	// Errors lists transformation failures (type coercion, etc.).
	Errors []string
}

// FieldResult describes a single proto field that was populated.
type FieldResult struct {
	ProtoField string
	SourceKey  string
	Value      string
}

// ValidateOverride checks a module directory's output transformation
// override for correctness. It performs schema validation (always) and
// optionally a full dry-run transformation against sample outputs.
//
// If sampleOutputs is nil, only schema validation is performed.
func ValidateOverride(
	kind cloudresourcekind.CloudResourceKind,
	moduleDir string,
	sampleOutputs map[string]interface{},
) (*ValidationResult, error) {
	result := &ValidationResult{}

	result.OverrideType = discoverOverride(moduleDir)

	switch result.OverrideType {
	case OverrideExecutable:
		validateExecutableSchema(moduleDir, result)
	case OverrideMapping:
		validateMappingSchema(kind, moduleDir, result)
	default:
		result.SchemaWarnings = append(result.SchemaWarnings,
			"no override found; the generic reflection transformer will be used")
	}

	if len(result.SchemaErrors) > 0 || sampleOutputs == nil {
		return result, nil
	}

	dryRun, err := runDryRun(kind, moduleDir, sampleOutputs, result.OverrideType)
	if err != nil {
		return result, errors.Wrap(err, "dry-run failed")
	}
	result.DryRun = dryRun

	return result, nil
}

func validateExecutableSchema(moduleDir string, result *ValidationResult) {
	if !isExecutableFile(fmt.Sprintf("%s/%s", moduleDir, executableFileName)) {
		result.SchemaErrors = append(result.SchemaErrors,
			"transform-outputs exists but is not executable (missing +x permission)")
	}
}

func validateMappingSchema(kind cloudresourcekind.CloudResourceKind, moduleDir string, result *ValidationResult) {
	mapping, err := loadMapping(moduleDir)
	if err != nil {
		result.SchemaErrors = append(result.SchemaErrors,
			fmt.Sprintf("failed to load output_transform.yaml: %v", err))
		return
	}

	outputsMsg, resolveErr := resolveStackOutputsMessage(kind)
	if resolveErr != nil {
		result.SchemaErrors = append(result.SchemaErrors,
			fmt.Sprintf("cannot resolve StackOutputs for kind %s: %v", kind.String(), resolveErr))
		return
	}

	protoFields := collectProtoFieldNames(outputsMsg)

	// Validate that all mapping targets are real proto fields.
	targetSeen := make(map[string]string)
	for src, dst := range mapping.Mappings {
		if _, exists := protoFields[dst]; !exists {
			result.SchemaErrors = append(result.SchemaErrors,
				fmt.Sprintf("mapping target %q (from source %q) is not a field on %s StackOutputs",
					dst, src, kind.String()))
		}
		if prevSrc, dup := targetSeen[dst]; dup {
			result.SchemaWarnings = append(result.SchemaWarnings,
				fmt.Sprintf("duplicate mapping target %q: both %q and %q map to it",
					dst, prevSrc, src))
		}
		targetSeen[dst] = src
	}

	// Warn if skip list overlaps with mapping sources.
	skipSet := make(map[string]struct{}, len(mapping.Skip))
	for _, key := range mapping.Skip {
		skipSet[key] = struct{}{}
	}
	for src := range mapping.Mappings {
		if _, overlap := skipSet[src]; overlap {
			result.SchemaWarnings = append(result.SchemaWarnings,
				fmt.Sprintf("source key %q is both mapped and skipped; the rename runs first, so the original key will be gone before skip runs",
					src))
		}
	}
}

func runDryRun(
	kind cloudresourcekind.CloudResourceKind,
	moduleDir string,
	sampleOutputs map[string]interface{},
	override OverrideKind,
) (*DryRunResult, error) {
	opts := &TransformOptions{ModuleDir: moduleDir}
	msg, flatOutputs, transformErr := TransformRaw(kind, sampleOutputs, opts)

	dr := &DryRunResult{}

	if transformErr != nil {
		dr.Errors = append(dr.Errors, transformErr.Error())
		return dr, nil
	}

	ref := msg.ProtoReflect()
	dr.TotalProtoFields = ref.Descriptor().Fields().Len()

	// Collect populated fields.
	populatedSet := make(map[string]bool)
	ref.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fieldName := string(fd.Name())
		populatedSet[fieldName] = true
		dr.PopulatedFields = append(dr.PopulatedFields, FieldResult{
			ProtoField: fieldName,
			SourceKey:  findSourceKey(flatOutputs, fieldName),
			Value:      fmt.Sprintf("%v", v),
		})
		return true
	})
	dr.PopulatedCount = len(dr.PopulatedFields)

	// Find unmapped outputs: keys in flatOutputs that didn't match a proto field.
	allProtoFields := collectProtoFieldNames(msg)
	for key := range flatOutputs {
		topKey := topLevelKey(key)
		if _, found := allProtoFields[topKey]; !found {
			dr.UnmappedOutputs = append(dr.UnmappedOutputs, key)
		}
	}

	return dr, nil
}

// collectProtoFieldNames returns a set of all field names on a proto message.
func collectProtoFieldNames(msg proto.Message) map[string]struct{} {
	fields := msg.ProtoReflect().Descriptor().Fields()
	result := make(map[string]struct{}, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		result[string(fields.Get(i).Name())] = struct{}{}
	}
	return result
}

// findSourceKey returns the flat output key that best matches a proto field name.
func findSourceKey(flatOutputs map[string]string, protoField string) string {
	if _, ok := flatOutputs[protoField]; ok {
		return protoField
	}
	return ""
}

// topLevelKey extracts the first segment of a dot-path key.
func topLevelKey(key string) string {
	for i, c := range key {
		if c == '.' || c == '[' {
			return key[:i]
		}
	}
	return key
}
