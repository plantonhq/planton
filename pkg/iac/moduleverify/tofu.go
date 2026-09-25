//go:build !codegen
// +build !codegen

package moduleverify

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/zclconf/go-cty/cty"
)

const (
	variablesFileName = "variables.tf"
	outputsFileName   = "outputs.tf"
)

// verifyTofu runs the OpenTofu/Terraform contract checks: the input surface
// (variables.tf against the kind's schema), every secret home the schema
// declares being read, the outputs contract, and — when the toolchain is
// available — the engine's own validation.
func verifyTofu(kind cloudresourcekind.CloudResourceKind, kindName, moduleDir string, in Input, result *Result) {
	checkTofuVariables(kind, kindName, moduleDir, result)
	checkSecretHomesTofu(declaredSecretHomes(kind), moduleDir, result)

	overrideKind := checkOutputsOverride(kind, moduleDir, in.SampleOutputs, result)
	if overrideKind == noOverride {
		checkTofuOutputNames(kind, kindName, moduleDir, result)
	}

	if in.SkipToolchainChecks {
		result.addNotice("engine validation skipped on request (tofu/terraform validate)")
	} else {
		runTofuValidate(moduleDir, result)
	}
}

// variableDecl is a parsed `variable` block: its type constraint and whether
// a default makes it satisfiable when the caller provides no value.
type variableDecl struct {
	Type       cty.Type
	HasDefault bool
}

// checkTofuVariables compares the module's declared input surface with the
// one the kind's schema defines. The expected surface is rendered through
// the same generator the platform uses (never compared as text): a
// customized module may diverge in everything except what breaks the value
// handoff.
func checkTofuVariables(kind cloudresourcekind.CloudResourceKind, kindName, moduleDir string, result *Result) {
	variablesPath := filepath.Join(moduleDir, variablesFileName)
	src, err := os.ReadFile(variablesPath)
	if err != nil {
		result.addError(variablesFileName, fmt.Sprintf(
			"variables.tf is missing — deployments pass the resource's metadata and spec as variables, so the module must declare both (eject the official %s module to see the expected shape)", kindName))
		return
	}

	actual, err := parseVariableDecls(variablesFileName, src)
	if err != nil {
		result.addError(variablesFileName, fmt.Sprintf("variables.tf could not be parsed: %v", err))
		return
	}

	expected, err := expectedVariableDecls(kind)
	if err != nil {
		result.addError("", fmt.Sprintf("the expected input surface for %s could not be derived from its schema: %v", kindName, err))
		return
	}

	for _, varName := range sortedKeys(expected) {
		expectedDecl := expected[varName]
		actualDecl, declared := actual[varName]
		if !declared {
			if varName == "spec" {
				result.addError(variablesFileName, fmt.Sprintf(
					"the module declares no %q variable — deployments pass the %s configuration through it, so the module cannot receive its configuration; declare it as an object type mirroring the %s spec schema", varName, kindName, kindName))
			} else {
				result.addWarning(variablesFileName, fmt.Sprintf(
					"the module declares no %q variable — deployments always pass it, and the engine will warn about an undeclared value on every run", varName))
			}
			continue
		}
		compareObjectAttrs(varName, expectedDecl.Type, actualDecl.Type, result)
	}

	for _, varName := range sortedKeys(actual) {
		if _, known := expected[varName]; known {
			continue
		}
		if !actual[varName].HasDefault {
			result.addError(variablesFileName, fmt.Sprintf(
				"the variable %q has no default and is not part of the %s contract — deployments never provide it, so every run fails asking for its value; give it a default or remove it", varName, kindName))
		}
	}
}

// compareObjectAttrs compares one contract variable's object type with the
// schema-derived one, one attribute level deep. Severity follows deployment
// impact: rendered values omit unpopulated fields, and the engine's object
// conversion rejects both unexpected attributes in a value and missing
// required ones — so what breaks depends on which side declares what.
func compareObjectAttrs(varName string, expectedType, actualType cty.Type, result *Result) {
	if actualType == cty.DynamicPseudoType {
		// `any` accepts every rendered value; nothing to compare.
		return
	}
	if !actualType.IsObjectType() {
		result.addError(variablesFileName, fmt.Sprintf(
			"the variable %q must be an object type (or any) — deployments pass an object value, which a %s type cannot accept", varName, actualType.FriendlyName()))
		return
	}
	if !expectedType.IsObjectType() {
		return
	}

	expectedAttrs := expectedType.AttributeTypes()
	actualAttrs := actualType.AttributeTypes()

	for _, attrName := range sortedTypeKeys(expectedAttrs) {
		expectedOptional := expectedType.AttributeOptional(attrName)
		if _, declared := actualAttrs[attrName]; !declared {
			if expectedOptional {
				result.addWarning(variablesFileName, fmt.Sprintf(
					"deployments that set %s.%s will fail: the module's %s type does not declare the attribute — declare it as optional()", varName, attrName, varName))
			} else {
				result.addError(variablesFileName, fmt.Sprintf(
					"deployments always provide %s.%s, but the module's %s type does not declare the attribute, so every run fails converting the value — add it", varName, attrName, varName))
			}
			continue
		}
		// Fails only the deployments that leave the field unset (rendered
		// values omit unpopulated fields) — a real gap, but configuration-
		// dependent, so a warning by the severity contract.
		if expectedOptional && !actualType.AttributeOptional(attrName) {
			result.addWarning(variablesFileName, fmt.Sprintf(
				"%s.%s may be absent from the rendered values (it is optional in the schema), but the module requires it — deployments that leave it unset fail; declare it optional() with a default", varName, attrName))
		}
	}

	for _, attrName := range sortedTypeKeys(actualAttrs) {
		if _, known := expectedAttrs[attrName]; known {
			continue
		}
		if actualType.AttributeOptional(attrName) {
			result.addWarning(variablesFileName, fmt.Sprintf(
				"deployments never provide %s.%s — the attribute is unused (or a misspelling of a schema field)", varName, attrName))
		} else {
			result.addError(variablesFileName, fmt.Sprintf(
				"deployments never provide %s.%s, but the module requires it, so every run fails converting the value — declare it optional() with a default or remove it", varName, attrName))
		}
	}
}

// expectedVariableDecls renders the kind's canonical variables.tf through
// the platform's own generator and parses it back — the schema-derived
// oracle for the input surface, guaranteed to track the generator's type
// rules without duplicating them.
func expectedVariableDecls(kind cloudresourcekind.CloudResourceKind) (map[string]*variableDecl, error) {
	instance, err := crkreflect.NewInstance(kind)
	if err != nil {
		return nil, err
	}
	rendered, err := generators.ProtoToVariablesTF(instance)
	if err != nil {
		return nil, err
	}
	return parseVariableDecls("expected-variables.tf", []byte(rendered))
}

// parseVariableDecls reads every `variable` block out of an HCL file.
func parseVariableDecls(filename string, src []byte) (map[string]*variableDecl, error) {
	file, diags := hclparse.NewParser().ParseHCL(src, filename)
	if diags.HasErrors() {
		return nil, errors.New(diags.Error())
	}

	content, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "variable", LabelNames: []string{"name"}}},
	})
	if diags.HasErrors() {
		return nil, errors.New(diags.Error())
	}

	decls := make(map[string]*variableDecl, len(content.Blocks))
	for _, block := range content.Blocks {
		attrs, _, diags := block.Body.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{{Name: "type"}, {Name: "default"}},
		})
		if diags.HasErrors() {
			return nil, errors.Errorf("variable %q: %s", block.Labels[0], diags.Error())
		}

		decl := &variableDecl{Type: cty.DynamicPseudoType}
		if typeAttr, ok := attrs.Attributes["type"]; ok {
			parsedType, _, diags := typeexpr.TypeConstraintWithDefaults(typeAttr.Expr)
			if diags.HasErrors() {
				return nil, errors.Errorf("variable %q has an unparseable type: %s", block.Labels[0], diags.Error())
			}
			decl.Type = parsedType
		}
		_, decl.HasDefault = attrs.Attributes["default"]

		decls[block.Labels[0]] = decl
	}
	return decls, nil
}

// checkTofuOutputNames compares outputs.tf's declared outputs with the
// kind's stack-outputs schema by name — the join the generic transformer
// performs after every deployment. Both directions are warnings: an unknown
// output is silently dropped, and an unpopulated schema field stays empty on
// the deployed resource.
func checkTofuOutputNames(kind cloudresourcekind.CloudResourceKind, kindName, moduleDir string, result *Result) {
	outputsPath := filepath.Join(moduleDir, outputsFileName)
	src, err := os.ReadFile(outputsPath)
	if err != nil {
		result.addWarning(outputsFileName, fmt.Sprintf(
			"the module declares no outputs — every %s stack-outputs field will stay empty on deployed resources", kindName))
		return
	}

	declaredNames, err := parseOutputNames(outputsFileName, src)
	if err != nil {
		result.addError(outputsFileName, fmt.Sprintf("outputs.tf could not be parsed: %v", err))
		return
	}

	schemaFields, err := stackOutputsFieldNames(kind)
	if err != nil {
		result.addNotice(fmt.Sprintf("outputs name check skipped: %v", err))
		return
	}

	var unknown, unpopulated []string
	seen := map[string]bool{}
	for _, name := range declaredNames {
		normalized := strings.ReplaceAll(name, "-", "_")
		if _, ok := schemaFields[normalized]; ok {
			seen[normalized] = true
		} else {
			unknown = append(unknown, name)
		}
	}
	for _, field := range sortedTypeKeysOfSet(schemaFields) {
		if !seen[field] {
			unpopulated = append(unpopulated, field)
		}
	}

	if len(unknown) > 0 {
		result.addWarning(outputsFileName, fmt.Sprintf(
			"these outputs match no %s stack-outputs field and are dropped after deployment: %s", kindName, strings.Join(unknown, ", ")))
	}
	if len(unpopulated) > 0 {
		result.addWarning(outputsFileName, fmt.Sprintf(
			"no output populates these %s stack-outputs fields, so they stay empty on deployed resources: %s", kindName, strings.Join(unpopulated, ", ")))
	}
}

// parseOutputNames reads every `output` block label out of an HCL file.
func parseOutputNames(filename string, src []byte) ([]string, error) {
	file, diags := hclparse.NewParser().ParseHCL(src, filename)
	if diags.HasErrors() {
		return nil, errors.New(diags.Error())
	}

	content, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "output", LabelNames: []string{"name"}}},
	})
	if diags.HasErrors() {
		return nil, errors.New(diags.Error())
	}

	names := make([]string, 0, len(content.Blocks))
	for _, block := range content.Blocks {
		names = append(names, block.Labels[0])
	}
	return names, nil
}

// stackOutputsFieldNames resolves the kind's stack-outputs message and
// returns its top-level field names.
func stackOutputsFieldNames(kind cloudresourcekind.CloudResourceKind) (map[string]struct{}, error) {
	instance, err := crkreflect.NewInstance(kind)
	if err != nil {
		return nil, err
	}
	statusField := instance.ProtoReflect().Descriptor().Fields().ByName("status")
	if statusField == nil || statusField.Message() == nil {
		return nil, errors.Errorf("%s has no status message", kind)
	}
	outputsField := statusField.Message().Fields().ByName("outputs")
	if outputsField == nil || outputsField.Message() == nil {
		return nil, errors.Errorf("%s has no stack-outputs message", kind)
	}

	fields := outputsField.Message().Fields()
	names := make(map[string]struct{}, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		names[string(fields.Get(i).Name())] = struct{}{}
	}
	return names, nil
}

func sortedKeys(m map[string]*variableDecl) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedTypeKeys(m map[string]cty.Type) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedTypeKeysOfSet(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
