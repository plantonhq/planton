//go:build !codegen
// +build !codegen

package moduleverify

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/outputs"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/zclconf/go-cty/cty"
)

// An output the kind's schema marks `sensitive` is a secret the resource generates (a client
// secret, an access key, an admin password). The platform stores it in the organization's secret
// store and keeps only a reference, but the engine sees the value first: an engine that does not
// treat the output as a secret prints it in its own plan, apply and event logs, and Pulumi keeps
// it unencrypted in state. So both engines must export exactly the outputs the schema marks as
// secrets. A marked output exported in the clear is an error -- it leaks on every deployment,
// which is worse than a deployment that fails. An output the engine hides although the schema
// says it is public is a warning: nothing leaks, but the two declarations disagree.

// outputFields maps each top-level outputs field name to whether the schema marks it
// a secret -- the one rule the platform and the CLI read too (outputs.SecretOutputs).
func outputFields(kind catalogkind.CatalogKind) (map[string]bool, error) {
	return outputs.SecretOutputs(kind)
}

// secretOutputExportedInTheClear is the finding for a marked output an engine does not treat as
// a secret.
func secretOutputExportedInTheClear(kindName, output, engine, fix string) string {
	return fmt.Sprintf(
		"the %s output %q is a secret in %s's schema, but the module exports it in the clear -- the engine then "+
			"prints it in its deploy logs and stores it readable in state; %s",
		engine, output, kindName, fix)
}

// publicOutputExportedAsSecret is the finding for an output an engine hides although the schema
// says it is public.
func publicOutputExportedAsSecret(kindName, output, engine, fix string) string {
	return fmt.Sprintf(
		"the %s output %q is exported as a secret, but %s's schema does not mark it sensitive -- the platform "+
			"records it as a public value, so the two declarations disagree; %s",
		engine, output, kindName, fix)
}

// checkSecretOutputsTofu holds outputs.tf's `sensitive` declarations to the schema's marks, in
// both directions.
func checkSecretOutputsTofu(kind catalogkind.CatalogKind, kindName, moduleDir string, result *Result) {
	src, err := os.ReadFile(filepath.Join(moduleDir, outputsFileName))
	if err != nil {
		// checkTofuOutputNames already reports a module that declares no outputs.
		return
	}
	declared, err := parseOutputSensitivity(outputsFileName, src)
	if err != nil {
		result.addError(outputsFileName, fmt.Sprintf("outputs.tf could not be parsed: %v", err))
		return
	}
	schema, err := outputFields(kind)
	if err != nil {
		result.addNotice(fmt.Sprintf("secret-outputs check skipped: %v", err))
		return
	}
	for _, name := range sortedBoolKeys(declared) {
		marked, known := schema[strings.ReplaceAll(name, "-", "_")]
		if !known {
			continue // checkTofuOutputNames reports outputs the schema does not know
		}
		switch sensitive := declared[name]; {
		case marked && !sensitive:
			result.addError(outputsFileName, secretOutputExportedInTheClear(kindName, name, "OpenTofu",
				"declare it with sensitive = true"))
		case !marked && sensitive:
			result.addWarning(outputsFileName, publicOutputExportedAsSecret(kindName, name, "OpenTofu",
				"drop sensitive = true, and wrap the value in nonsensitive() when the provider marks the attribute sensitive"))
		}
	}
}

// parseOutputSensitivity reads every `output` block and whether it declares sensitive = true.
func parseOutputSensitivity(filename string, src []byte) (map[string]bool, error) {
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
	sensitivity := make(map[string]bool, len(content.Blocks))
	for _, block := range content.Blocks {
		name := block.Labels[0]
		attrs, _, diags := block.Body.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{{Name: "sensitive"}},
		})
		if diags.HasErrors() {
			return nil, errors.Errorf("output %q: %s", name, diags.Error())
		}
		attr, ok := attrs.Attributes["sensitive"]
		if !ok {
			sensitivity[name] = false
			continue
		}
		value, diags := attr.Expr.Value(nil)
		if diags.HasErrors() || value.Type() != cty.Bool || !value.IsKnown() || value.IsNull() {
			return nil, errors.Errorf("output %q: sensitive must be the literal true or false", name)
		}
		sensitivity[name] = value.True()
	}
	return sensitivity, nil
}

// pulumiImportPath is the package whose ToSecret marks an output as a secret.
const pulumiImportPath = "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// pulumiExport is one ctx.Export call whose output name the check could resolve.
type pulumiExport struct {
	Name   string
	Secret bool
	File   string
}

// checkPulumiOutputs holds the module's ctx.Export calls to the kind's outputs schema: by
// name (the join the generic transformer performs after every deployment, as outputs.tf is held
// for OpenTofu) and by secrecy (a marked output is exported as pulumi.ToSecret(...), directly at
// the export, and no other output is).
func checkPulumiOutputs(kind catalogkind.CatalogKind, kindName, moduleDir string, result *Result) {
	exports, unresolved, err := collectPulumiExports(moduleDir)
	if err != nil {
		result.addNotice(fmt.Sprintf("outputs checks skipped: %v", err))
		return
	}
	schema, err := outputFields(kind)
	if err != nil {
		result.addNotice(fmt.Sprintf("outputs checks skipped: %v", err))
		return
	}
	if unresolved > 0 {
		result.addNotice(fmt.Sprintf("%d ctx.Export call(s) compute their output name and were not checked", unresolved))
	}

	var unknown []string
	exported := map[string]bool{}
	for _, export := range exports {
		// A nested field is exported by its path ("password_secret.name"), which the transformer
		// flattens back onto the top-level field it names, so the field is the path's first step.
		field, _, _ := strings.Cut(strings.ReplaceAll(export.Name, "-", "_"), ".")
		marked, known := schema[field]
		if !known {
			unknown = append(unknown, export.Name)
			continue
		}
		exported[field] = true
		switch {
		case marked && !export.Secret:
			result.addError(export.File, secretOutputExportedInTheClear(kindName, export.Name, "Pulumi",
				fmt.Sprintf("export it as ctx.Export(%q, pulumi.ToSecret(value))", export.Name)))
		case !marked && export.Secret:
			result.addWarning(export.File, publicOutputExportedAsSecret(kindName, export.Name, "Pulumi",
				"export the value without pulumi.ToSecret, and unwrap it with pulumi.Unsecret when the provider marks it secret"))
		}
	}

	var unpopulated []string
	for _, field := range sortedBoolKeys(schema) {
		if !exported[field] {
			unpopulated = append(unpopulated, field)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		result.addWarning("", fmt.Sprintf(
			"these exports match no %s outputs field and are dropped after deployment: %s", kindName, strings.Join(unknown, ", ")))
	}
	if len(unpopulated) > 0 && unresolved == 0 {
		result.addWarning("", fmt.Sprintf(
			"no export populates these %s outputs fields, so they stay empty on deployed resources: %s", kindName, strings.Join(unpopulated, ", ")))
	}
}

// collectPulumiExports finds every ctx.Export(name, value) call in the module's own Go packages.
// A name is a string literal or a string constant declared in the module (the catalog's Op*
// constants); a computed name is counted as unresolved. An export is a secret when its value is a
// direct pulumi.ToSecret(...) or pulumi.ToSecretWithContext(...) call.
func collectPulumiExports(moduleDir string) ([]pulumiExport, int, error) {
	fset := token.NewFileSet()
	type parsedFile struct {
		path string
		file *ast.File
	}
	var files []parsedFile
	err := filepath.WalkDir(moduleDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != moduleDir && (entry.Name() == "vendor" || strings.HasPrefix(entry.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("%s could not be parsed: %w", path, err)
		}
		files = append(files, parsedFile{path: path, file: file})
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// String constants by name. The catalog declares output names as package-level constants in
	// the module's own package, referenced unqualified or through the module package's name.
	constants := map[string]string{}
	for _, parsed := range files {
		for _, decl := range parsed.file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				valueSpec := spec.(*ast.ValueSpec)
				for i, name := range valueSpec.Names {
					if i < len(valueSpec.Values) {
						if value, ok := stringLiteral(valueSpec.Values[i]); ok {
							constants[name.Name] = value
						}
					}
				}
			}
		}
	}

	var exports []pulumiExport
	unresolved := 0
	for _, parsed := range files {
		pulumiNames := importNamesOf(parsed.file, pulumiImportPath, "pulumi")
		relPath, _ := filepath.Rel(moduleDir, parsed.path)
		ast.Inspect(parsed.file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Export" {
				return true
			}
			name, ok := exportName(call.Args[0], constants)
			if !ok {
				unresolved++
				return true
			}
			exports = append(exports, pulumiExport{
				Name:   name,
				Secret: isToSecretCall(call.Args[1], pulumiNames),
				File:   relPath,
			})
			return true
		})
	}
	sort.SliceStable(exports, func(i, j int) bool { return exports[i].Name < exports[j].Name })
	return exports, unresolved, nil
}

func exportName(expr ast.Expr, constants map[string]string) (string, bool) {
	if value, ok := stringLiteral(expr); ok {
		return value, true
	}
	switch e := expr.(type) {
	case *ast.Ident:
		value, ok := constants[e.Name]
		return value, ok
	case *ast.SelectorExpr:
		value, ok := constants[e.Sel.Name]
		return value, ok
	}
	return "", false
}

func stringLiteral(expr ast.Expr) (string, bool) {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func isToSecretCall(expr ast.Expr, pulumiNames map[string]bool) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	if !ok || !pulumiNames[pkg.Name] {
		return false
	}
	return selector.Sel.Name == "ToSecret" || selector.Sel.Name == "ToSecretWithContext"
}

// importNamesOf returns the names a file refers to an import path by (its alias, or the default
// package name).
func importNamesOf(file *ast.File, importPath, defaultName string) map[string]bool {
	names := map[string]bool{}
	for _, spec := range file.Imports {
		if strings.Trim(spec.Path.Value, `"`) != importPath {
			continue
		}
		if spec.Name != nil {
			names[spec.Name.Name] = true
		} else {
			names[defaultName] = true
		}
	}
	return names
}

func sortedBoolKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
