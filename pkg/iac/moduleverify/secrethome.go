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
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/shared/options"
	"github.com/zclconf/go-cty/cty"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// secretHome is one place a kind's schema sends secrets: a field whose value every viewer of the
// resource reads (Refused), and the sibling the component stores in a secret store instead
// (Home). The platform refuses a secret reference in the refused field and points the author at
// the home, so a module that never reads the home silently deploys without the value the author
// was told to put there.
type secretHome struct {
	// Refused is the refused field's path from the spec root, e.g. "containers.env.value".
	Refused string
	// Home is the home field's proto name, e.g. "secret_value".
	Home string
	// HomePath is the home's path from the spec root, e.g. "containers.env.secret_value".
	HomePath string
}

// declaredSecretHomes walks the kind's spec and returns every secret home it declares, sorted by
// the home's path. A kind that declares none returns nothing, and the checks below do not run.
func declaredSecretHomes(kind cloudresourcekind.CloudResourceKind) []secretHome {
	instance, err := crkreflect.NewInstance(kind)
	if err != nil {
		return nil
	}
	spec := instance.ProtoReflect().Descriptor().Fields().ByName("spec")
	if spec == nil || spec.Message() == nil {
		return nil
	}
	var homes []secretHome
	visited := map[protoreflect.FullName]bool{}
	var walk func(message protoreflect.MessageDescriptor, prefix string)
	walk = func(message protoreflect.MessageDescriptor, prefix string) {
		if visited[message.FullName()] {
			return
		}
		visited[message.FullName()] = true
		fields := message.Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			path := prefix + string(field.Name())
			if home := secretHomeOf(field); home != "" {
				homes = append(homes, secretHome{Refused: path, Home: home, HomePath: prefix + home})
			}
			if field.IsMap() {
				if value := field.MapValue().Message(); value != nil {
					walk(value, path+".")
				}
			} else if next := field.Message(); next != nil {
				walk(next, path+".")
			}
		}
	}
	walk(spec.Message(), "")
	sort.Slice(homes, func(i, j int) bool { return homes[i].HomePath < homes[j].HomePath })
	return homes
}

func secretHomeOf(field protoreflect.FieldDescriptor) string {
	opts, ok := field.Options().(proto.Message)
	if !ok || opts == nil || !proto.HasExtension(opts, options.E_SecretHome) {
		return ""
	}
	return proto.GetExtension(opts, options.E_SecretHome).(string)
}

// secretHomeUnread is the finding for a home the module never reads. A warning by this package's
// severity rule: only configurations that use the home lose anything, but those lose it silently.
func secretHomeUnread(home secretHome, engine string) string {
	return fmt.Sprintf(
		"the %s module never reads spec.%s -- the platform refuses a secret in spec.%s and sends the author there, "+
			"so a secret put where the platform asks is silently not deployed; read it the way the official module "+
			"does (eject it to see how)", engine, home.HomePath, home.Refused)
}

// checkSecretHomesTofu proves every declared secret home is referenced by the module's
// configuration: some expression outside variables.tf (which only declares the input's type)
// traverses an attribute of the home's name, or names it as a string key.
func checkSecretHomesTofu(homes []secretHome, moduleDir string, result *Result) {
	if len(homes) == 0 {
		return
	}
	read, err := tofuReferencedNames(moduleDir)
	if err != nil {
		result.addNotice(fmt.Sprintf("secret-home check skipped: %v", err))
		return
	}
	for _, home := range homes {
		if !read[home.Home] {
			result.addWarning("", secretHomeUnread(home, "OpenTofu"))
		}
	}
}

// tofuReferencedNames collects every attribute name an expression traverses (env.secret_value,
// var.spec.containers[0].secret_environment) and every string literal, across the module's .tf
// files other than variables.tf.
func tofuReferencedNames(moduleDir string) (map[string]bool, error) {
	files, err := filepath.Glob(filepath.Join(moduleDir, "*.tf"))
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	parser := hclparse.NewParser()
	for _, path := range files {
		if filepath.Base(path) == variablesFileName {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		file, diags := parser.ParseHCL(src, filepath.Base(path))
		if diags.HasErrors() {
			return nil, fmt.Errorf("%s could not be parsed: %s", filepath.Base(path), diags.Error())
		}
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}
		_ = hclsyntax.VisitAll(body, func(node hclsyntax.Node) hcl.Diagnostics {
			switch expr := node.(type) {
			case *hclsyntax.ScopeTraversalExpr:
				addTraversalNames(expr.Traversal, names)
			case *hclsyntax.RelativeTraversalExpr:
				addTraversalNames(expr.Traversal, names)
			case *hclsyntax.LiteralValueExpr:
				if expr.Val.Type() == cty.String && expr.Val.IsKnown() && !expr.Val.IsNull() {
					names[expr.Val.AsString()] = true
				}
			}
			return nil
		})
	}
	return names, nil
}

func addTraversalNames(traversal hcl.Traversal, names map[string]bool) {
	for _, step := range traversal {
		if attr, ok := step.(hcl.TraverseAttr); ok {
			names[attr.Name] = true
		}
	}
}

// checkSecretHomesPulumi proves every declared secret home is read by the module's Go source: a
// selector on the home's generated field (envVar.SecretValue) or its getter (GetSecretValue()).
// Modules often read their spec through shared helpers, so the search follows the module's
// imports that resolve inside its own Go module, two hops deep (module, helper, the helper's
// helper). When the home is not found and the module imports the catalog's helpers from outside
// its own Go module (an ejected module), the check cannot see those helpers and says so as a
// notice instead of warning about code it could not read.
func checkSecretHomesPulumi(homes []secretHome, moduleDir string, result *Result) {
	if len(homes) == 0 {
		return
	}
	read, unreadable, err := goSelectorNames(moduleDir)
	if err != nil {
		result.addNotice(fmt.Sprintf("secret-home check skipped: %v", err))
		return
	}
	for _, home := range homes {
		field := goFieldName(home.Home)
		if read[field] || read["Get"+field] {
			continue
		}
		if len(unreadable) > 0 {
			result.addNotice(fmt.Sprintf("secret-home check could not confirm spec.%s is read: the module reaches it "+
				"through %s, outside this module's own Go module", home.HomePath, strings.Join(unreadable, ", ")))
			continue
		}
		result.addWarning("", secretHomeUnread(home, "Pulumi"))
	}
}

// catalogModulePath is the Go module the catalog's shared helpers live in.
const catalogModulePath = "github.com/plantonhq/planton"

// importHops is how far the search follows imports from the module's own packages: a helper and
// the helper it delegates to.
const importHops = 2

// goSelectorNames collects every selector name in the module's non-test, non-generated Go files
// and in the packages they import from the same Go module, up to importHops away. It also returns the
// catalog packages the module imports but that live outside its Go module, which it cannot read.
func goSelectorNames(moduleDir string) (map[string]bool, []string, error) {
	names := map[string]bool{}
	fset := token.NewFileSet()
	goModRoot, modulePath := enclosingGoModule(moduleDir)

	var frontier []string
	err := filepath.WalkDir(moduleDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "vendor" || strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			imports, err := collectPackage(fset, path, names)
			frontier = append(frontier, imports...)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	seen := map[string]bool{}
	unreadable := map[string]bool{}
	for hop := 0; hop < importHops && len(frontier) > 0; hop++ {
		var next []string
		for _, importPath := range frontier {
			if seen[importPath] {
				continue
			}
			seen[importPath] = true
			if modulePath != "" && strings.HasPrefix(importPath, modulePath+"/") {
				dir := filepath.Join(goModRoot, filepath.FromSlash(strings.TrimPrefix(importPath, modulePath+"/")))
				imports, err := collectPackage(fset, dir, names)
				if err != nil {
					return nil, nil, err
				}
				next = append(next, imports...)
			} else if strings.HasPrefix(importPath, catalogModulePath+"/") {
				unreadable[importPath] = true
			}
		}
		frontier = next
	}
	unreadableList := make([]string, 0, len(unreadable))
	for importPath := range unreadable {
		unreadableList = append(unreadableList, importPath)
	}
	sort.Strings(unreadableList)
	return names, unreadableList, nil
}

// collectPackage adds the selector names of one directory's non-test Go files to names and
// returns the import paths those files declare.
func collectPackage(fset *token.FileSet, dir string, names map[string]bool) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}
	var imports []string
	for _, path := range files {
		// Generated protobuf code reads every field in its own getters; it is the schema, not the
		// module reading it.
		if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("%s could not be parsed: %w", path, err)
		}
		for _, spec := range file.Imports {
			imports = append(imports, strings.Trim(spec.Path.Value, `"`))
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok {
				names[selector.Sel.Name] = true
			}
			return true
		})
	}
	return imports, nil
}

// enclosingGoModule returns the directory of the nearest go.mod at or above dir and the module
// path it declares, or empty strings when there is none.
func enclosingGoModule(dir string) (string, string) {
	for {
		if src, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
			for _, line := range strings.Split(string(src), "\n") {
				if fields := strings.Fields(line); len(fields) == 2 && fields[0] == "module" {
					return dir, fields[1]
				}
			}
			return dir, ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ""
		}
		dir = parent
	}
}

// goFieldName is protoc-gen-go's field name for a snake_case proto field: secret_value ->
// SecretValue.
func goFieldName(protoName string) string {
	var b strings.Builder
	upper := true
	for _, r := range protoName {
		if r == '_' {
			upper = true
			continue
		}
		if upper && r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		upper = r >= '0' && r <= '9'
		b.WriteRune(r)
	}
	return b.String()
}
