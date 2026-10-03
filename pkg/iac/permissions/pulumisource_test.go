package permissions

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	permissionsv1 "github.com/plantonhq/planton/iac/catalogkindpermissions/v1"
)

// This file reads an official Pulumi module's Go source for the Kubernetes
// objects it creates, so the conformance gate can hold the kind's
// manifest to what the provider reads while it waits for them
// (pulumikubernetes.go). It reads constructor calls, not names that happen
// to appear: a call is resolved through the file's own imports, so an alias
// (kubernetescorev1 "…/core/v1", or the bare package name v1) never hides
// one. What it cannot resolve it reports as a problem, which fails the gate:
// a constructor passed around as a value, a raw ctx.RegisterResource of a
// Kubernetes token, a Kubernetes type SDK it does not read, and a package
// outside the module that creates Kubernetes objects on the module's behalf.

// crd2pulumiPrefix is where this repository keeps the typed SDKs generated
// from custom resource definitions. Their constructors register a token the
// gate reads from the generated source itself.
const crd2pulumiPrefix = "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/"

// modulePrefix maps this repository's import paths to directories.
const modulePrefix = "github.com/plantonhq/planton/"

// keptCRDsPackage applies a chart's CustomResourceDefinitions through one
// ConfigGroup each, retained on delete while the kind keeps them on
// uninstall (keptcrds.go). The gate models its calls directly: every CRD it
// applies is a cluster-scoped CustomResourceDefinition.
const keptCRDsPackage = "github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/keptcrds"

// manifestCRPackage applies a Kubernetes manifest projection kind's one
// custom resource (manifestcr.go). Its group and kind are not in the call:
// they are the kind's kubernetes_manifest_projection in the kind
// registry, which the gate reads to model each Apply as that custom resource.
const manifestCRPackage = "github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/manifestcr"

// helperPackages are the repository packages outside a module that a module
// may import although they create Kubernetes-provider resources, each with
// why the gate needs nothing more from them. Any other repository package
// that constructs a Kubernetes object is reported: the objects it creates
// are the module's, and the gate would otherwise not see them.
var helperPackages = map[string]string{
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider": "it constructs only the Kubernetes provider itself, a Pulumi resource that is not a cluster object",
	keptCRDsPackage:   "its CRD ConfigGroups are modelled as the CustomResourceDefinitions they apply (keptCRDsPackage)",
	manifestCRPackage: "its one custom resource is modelled as the kind's kubernetes_manifest_projection from the kind registry (manifestCRPackage)",
}

const (
	annotationSkipAwait = "pulumi.com/skipAwait"
	annotationWaitFor   = "pulumi.com/waitFor"
)

// createdObject is one object a module creates and how the module changes
// what the provider waits for on it.
type createdObject struct {
	where string // file:line of the call, relative to the module directory
	via   string // "" for a constructor; the yaml constructor or helper that applies it otherwise
	// kind is set for a typed SDK constructor (or a yaml child whose kind is
	// in the table); group/kind for a custom resource (untyped or
	// crd2pulumi); self alone for a yaml child the table does not know.
	kind         *PulumiKubernetesKind
	crGroup      string
	crKind       string
	self         *KubernetesResource
	skipAwait    bool // pulumi.com/skipAwait: "true" on the object
	waitFor      bool // pulumi.com/waitFor: a readiness wait on the object's own kind
	retainDelete bool // pulumi.RetainOnDelete(true): the provider never deletes it
}

// yamlCall is one non-Helm yaml constructor: its children are the YAML's
// kinds, which the source does not name.
type yamlCall struct {
	where     string
	via       string
	skipAwait bool // a transformation sets pulumi.com/skipAwait: "true" on every child
	retain    bool // pulumi.RetainOnDelete(true) rides the call or a transformation
}

// moduleScan is everything the gate learns from one kind's module.
type moduleScan struct {
	objects   []createdObject
	yamlCalls []yamlCall
	keptCRDs  []string // keptcrds.Apply calls
	// projectionApplies are manifestcr.Apply calls; their group and kind are
	// the kind's projection, resolved where the kind is known.
	projectionApplies []createdObject
	delegated         map[string]bool // delegated constructors used (keys of PulumiKubernetesDelegated)
	problems          []string        // what the gate cannot resolve -- each fails the gate
	// setsSkipAwait: the module sets pulumi.com/skipAwait to "true" somewhere.
	setsSkipAwait bool
}

// sourcePackage is one Go package's parsed files and package-level names.
type sourcePackage struct {
	files  []*ast.File
	values map[string]ast.Expr
	funcs  map[string]*ast.FuncDecl
}

// scanPulumiModule parses every non-test Go file under dir.
func scanPulumiModule(root, dir string) (*moduleScan, error) {
	scan := &moduleScan{delegated: map[string]bool{}}
	fset := token.NewFileSet()
	packages, err := parseTree(fset, dir)
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(packages))
	for d := range packages {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	outside := map[string]bool{}
	for _, d := range dirs {
		for _, file := range packages[d].files {
			scanFile(root, dir, fset, file, packages[d], scan, outside)
		}
	}
	for _, importPath := range sortedKeys(outside, nil) {
		if reason := constructsKubernetesObjects(root, importPath); reason != "" {
			scan.problems = append(scan.problems, fmt.Sprintf("the module imports %s, which %s -- the objects it creates are the module's, and the gate cannot see them; "+
				"create them in the module, or model the package and say why in helperPackages", importPath, reason))
		}
	}
	return scan, nil
}

func parseTree(fset *token.FileSet, dir string) (map[string]*sourcePackage, error) {
	packages := map[string]*sourcePackage{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		pkg := packages[filepath.Dir(path)]
		if pkg == nil {
			pkg = &sourcePackage{}
			packages[filepath.Dir(path)] = pkg
		}
		pkg.files = append(pkg.files, file)
		return nil
	})
	for _, pkg := range packages {
		pkg.values = packageValues(pkg.files)
		pkg.funcs = packageFuncs(pkg.files)
	}
	return packages, err
}

// isKubernetesTypePackage reports the import paths whose calls create
// Kubernetes objects: the provider's SDK and this repository's generated
// custom-resource SDKs.
func isKubernetesTypePackage(path string) bool {
	return path == PulumiKubernetesSDK || strings.HasPrefix(path, PulumiKubernetesSDK+"/") || strings.HasPrefix(path, crd2pulumiPrefix)
}

// isForeignKubernetesTypePackage reports a Kubernetes type SDK the gate does
// not read: another Pulumi Kubernetes package, or a generated CRD SDK kept
// anywhere but this repository's kubernetestypes tree.
func isForeignKubernetesTypePackage(path string) bool {
	// This repository's own packages are read, not guessed at: a module's
	// subpackages are scanned as the module, any other is checked for the
	// objects it creates (constructsKubernetesObjects).
	if isKubernetesTypePackage(path) || strings.HasPrefix(path, modulePrefix) || strings.HasPrefix(path, "github.com/pulumi/pulumi/") {
		return false
	}
	return strings.Contains(path, "pulumi-kubernetes") ||
		(strings.Contains(path, "pulumi") && strings.Contains(path, "crd")) ||
		strings.Contains(path, "kubernetestypes")
}

func scanFile(root, moduleDir string, fset *token.FileSet, file *ast.File, pkg *sourcePackage, scan *moduleScan, outside map[string]bool) {
	where := func(n ast.Node) string {
		pos := fset.Position(n.Pos())
		rel, _ := filepath.Rel(moduleDir, pos.Filename)
		return fmt.Sprintf("%s:%d", rel, pos.Line)
	}
	imports := map[string]string{} // local name -> import path
	for _, spec := range file.Imports {
		path, _ := strconv.Unquote(spec.Path.Value)
		if isForeignKubernetesTypePackage(path) {
			scan.problems = append(scan.problems, fmt.Sprintf("%s imports %s, a Kubernetes type SDK the gate does not read -- generate the types into %s, or teach the scanner to read this one", where(spec), path, crd2pulumiPrefix))
			continue
		}
		if strings.HasPrefix(path, modulePrefix) && !strings.HasPrefix(path, crd2pulumiPrefix) && helperPackages[path] == "" {
			dir := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(path, modulePrefix)))
			if rel, err := filepath.Rel(moduleDir, dir); err != nil || strings.HasPrefix(rel, "..") {
				outside[path] = true
			}
		}
		if !isKubernetesTypePackage(path) && path != keptCRDsPackage && path != manifestCRPackage {
			continue
		}
		name := filepath.Base(path)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		if name == "_" || name == "." {
			continue
		}
		imports[name] = path
	}
	if setsAnnotation(file, pkg, annotationSkipAwait, "true", 0) {
		scan.setsSkipAwait = true
	}

	// A constructor is only understood where it is called. Every selector
	// that is a call's function is recorded first, so one that is not -- a
	// constructor stored or passed as a value -- is reported, not skipped.
	called := map[*ast.SelectorExpr]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				called[sel] = true
			}
		}
		return true
	})
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok || called[sel] {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && imports[id.Name] != "" && imports[id.Name] != keptCRDsPackage && imports[id.Name] != manifestCRPackage && strings.HasPrefix(sel.Sel.Name, "New") {
			scan.problems = append(scan.problems, fmt.Sprintf("%s: %s.%s is used as a value, so the gate cannot see what it creates -- call the constructor directly", where(sel), imports[id.Name], sel.Sel.Name))
		}
		return true
	})

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			// A package-level value built by a constructor would run at
			// init, outside any function the gate reads.
			if gen, ok := decl.(*ast.GenDecl); ok {
				ast.Inspect(gen, func(n ast.Node) bool {
					if call, ok := n.(*ast.CallExpr); ok {
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if id, ok := sel.X.(*ast.Ident); ok && imports[id.Name] != "" && strings.HasPrefix(sel.Sel.Name, "New") {
								scan.problems = append(scan.problems, fmt.Sprintf("%s: %s.%s runs in a package-level declaration -- create objects inside the module's functions", where(call), imports[id.Name], sel.Sel.Name))
							}
						}
					}
					return true
				})
			}
			continue
		}
		var calls []*ast.CallExpr
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "RegisterResource" && len(call.Args) > 0 {
					if tok, err := stringValue(call.Args[0], pkg.values); err == nil && strings.HasPrefix(tok, "kubernetes:") {
						scan.problems = append(scan.problems, fmt.Sprintf("%s: RegisterResource(%q) creates a Kubernetes object the gate cannot type -- use the SDK's constructor", where(call), tok))
					}
				}
				if id, ok := sel.X.(*ast.Ident); ok && imports[id.Name] != "" && (strings.HasPrefix(sel.Sel.Name, "New") || sel.Sel.Name == "Apply") {
					calls = append(calls, call)
				}
			}
			return true
		})
		// An annotation built outside the constructor call is still the
		// object's when the function it sits in creates nothing else.
		fnSkipAwait := len(calls) == 1 && setsAnnotation(fn.Body, pkg, annotationSkipAwait, "true", 0)
		fnWaitFor := len(calls) == 1 && setsAnnotation(fn.Body, pkg, annotationWaitFor, "", 0)
		for _, call := range calls {
			scanCall(root, call, where(call), imports, pkg, scan, fnSkipAwait, fnWaitFor)
		}
	}
}

func scanCall(root string, call *ast.CallExpr, where string, imports map[string]string, pkg *sourcePackage, scan *moduleScan, fnSkipAwait, fnWaitFor bool) {
	sel := call.Fun.(*ast.SelectorExpr)
	importPath := imports[sel.X.(*ast.Ident).Name]
	constructor := sel.Sel.Name
	skipAwait := fnSkipAwait
	waitFor := fnWaitFor
	retain := false
	for _, arg := range call.Args {
		skipAwait = skipAwait || setsAnnotation(arg, pkg, annotationSkipAwait, "true", 0)
		waitFor = waitFor || setsAnnotation(arg, pkg, annotationWaitFor, "", 0)
		retain = retain || retainsOnDelete(arg, pkg, 0)
	}
	object := createdObject{where: where, skipAwait: skipAwait, waitFor: waitFor, retainDelete: retain}
	key := importPath + "." + constructor

	switch {
	case importPath == keptCRDsPackage:
		if constructor == "Apply" {
			scan.keptCRDs = append(scan.keptCRDs, where)
		}
		return
	case importPath == manifestCRPackage:
		if constructor == "Apply" {
			object.via = "manifestcr.Apply"
			scan.projectionApplies = append(scan.projectionApplies, object)
		}
		return
	case strings.HasPrefix(key, PulumiKubernetesSDK+"/yaml") || key == PulumiKubernetesSDK+"/helm/v3.NewChart":
		if PulumiKubernetesDelegated[key] == "" {
			scan.problems = append(scan.problems, fmt.Sprintf("%s: %s is a yaml constructor the gate does not model", where, key))
			return
		}
		scan.yamlCalls = append(scan.yamlCalls, yamlCall{where: where, via: strings.TrimPrefix(importPath, PulumiKubernetesSDK+"/") + " " + strings.TrimPrefix(constructor, "New"), skipAwait: skipAwait, retain: retain})
		return
	case PulumiKubernetesDelegated[key] != "":
		scan.delegated[key] = true
		return
	case importPath == PulumiKubernetesCustomResourcePackage && constructor == PulumiKubernetesCustomResource:
		group, kind, err := customResourceKind(call, pkg.values)
		if err != nil {
			scan.problems = append(scan.problems, fmt.Sprintf("%s: %v -- the gate reads a custom resource's kind from the ApiVersion and Kind the call passes: a string, a package constant, or a field of a package-level struct literal", where, err))
			return
		}
		object.crGroup, object.crKind = group, kind
	case strings.HasPrefix(importPath, crd2pulumiPrefix):
		tok, err := crd2pulumiToken(root, importPath, constructor)
		if err != nil {
			scan.problems = append(scan.problems, fmt.Sprintf("%s: %s.%s: %v", where, importPath, constructor, err))
			return
		}
		group, kind, err := splitToken(tok)
		if err != nil {
			scan.problems = append(scan.problems, fmt.Sprintf("%s: %v", where, err))
			return
		}
		if strings.HasSuffix(kind, "Patch") || strings.HasSuffix(kind, "List") {
			scan.problems = append(scan.problems, fmt.Sprintf("%s: %s is a patch or list resource, which the gate does not model -- a patch is relinquished on delete rather than awaited; read what the provider does for it and model it before a module uses one", where, tok))
			return
		}
		object.crGroup, object.crKind = group, kind
	default:
		kind, ok := PulumiKubernetesKindByConstructor(importPath, constructor)
		if !ok {
			scan.problems = append(scan.problems, fmt.Sprintf("%s: %s is not in PulumiKubernetesKinds -- read what pulumi-kubernetes %s waits for on this kind (provider/pkg/await) and add its row", where, key, PulumiKubernetesProviderVersion))
			return
		}
		object.kind = &kind
	}
	scan.objects = append(scan.objects, object)
}

// setsAnnotation reports whether node sets the annotation -- as a map
// literal's key or an index assignment, the key and value resolved through
// the package's constants -- to want ("" for any value), directly or in a
// package function it calls (the modules' transformation convention).
func setsAnnotation(node ast.Node, pkg *sourcePackage, annotation, want string, depth int) bool {
	matches := func(key, value ast.Expr) bool {
		k, err := stringValue(key, pkg.values)
		if err != nil || k != annotation {
			return false
		}
		if want == "" {
			return true
		}
		v, err := stringValue(value, pkg.values)
		return err == nil && v == want
	}
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}
		switch e := n.(type) {
		case *ast.KeyValueExpr:
			found = matches(e.Key, e.Value)
		case *ast.AssignStmt:
			for i, lhs := range e.Lhs {
				if index, ok := lhs.(*ast.IndexExpr); ok && i < len(e.Rhs) && matches(index.Index, e.Rhs[i]) {
					found = true
				}
			}
		case *ast.CallExpr:
			if id, ok := e.Fun.(*ast.Ident); ok && depth < 2 {
				if fn := pkg.funcs[id.Name]; fn != nil && fn.Body != nil {
					found = setsAnnotation(fn.Body, pkg, annotation, want, depth+1)
				}
			}
		}
		return !found
	})
	return found
}

// retainsOnDelete reports pulumi.RetainOnDelete(true) in node, directly or in
// a transformation it applies: the engine then drops the object from state
// without asking the provider to delete it, so no delete wait runs.
func retainsOnDelete(node ast.Node, pkg *sourcePackage, depth int) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "RetainOnDelete" && len(call.Args) == 1 {
			if value, ok := call.Args[0].(*ast.Ident); ok && value.Name == "true" {
				found = true
			}
		}
		if id, ok := call.Fun.(*ast.Ident); ok && depth < 2 {
			if fn := pkg.funcs[id.Name]; fn != nil && fn.Body != nil {
				found = found || retainsOnDelete(fn.Body, pkg, depth+1)
			}
		}
		return !found
	})
	return found
}

// constructsKubernetesObjects reports why a repository package outside the
// module would create Kubernetes objects, or "" when it creates none.
func constructsKubernetesObjects(root, importPath string) string {
	dir := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(importPath, modulePrefix)))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if err != nil {
			continue
		}
		names := map[string]bool{}
		for _, spec := range file.Imports {
			path, _ := strconv.Unquote(spec.Path.Value)
			if !isKubernetesTypePackage(path) {
				continue
			}
			name := filepath.Base(path)
			if spec.Name != nil {
				name = spec.Name.Name
			}
			names[name] = true
		}
		reason := ""
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || reason != "" {
				return reason == ""
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if id, ok := sel.X.(*ast.Ident); ok && names[id.Name] && strings.HasPrefix(sel.Sel.Name, "New") {
					reason = fmt.Sprintf("calls %s.%s in %s", id.Name, sel.Sel.Name, entry.Name())
				}
				if sel.Sel.Name == "RegisterResource" && len(call.Args) > 0 {
					if lit, ok := call.Args[0].(*ast.BasicLit); ok && strings.Contains(lit.Value, "kubernetes:") {
						reason = fmt.Sprintf("registers %s in %s", lit.Value, entry.Name())
					}
				}
			}
			return reason == ""
		})
		if reason != "" {
			return reason
		}
	}
	return ""
}

// packageFuncs indexes a package's top-level functions by name.
func packageFuncs(files []*ast.File) map[string]*ast.FuncDecl {
	funcs := map[string]*ast.FuncDecl{}
	for _, file := range files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
				funcs[fn.Name.Name] = fn
			}
		}
	}
	return funcs
}

// packageValues indexes a package's package-level const and var values by
// name, for resolving the strings a custom resource's kind is built from.
func packageValues(files []*ast.File) map[string]ast.Expr {
	values := map[string]ast.Expr{}
	for _, file := range files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || (gen.Tok != token.CONST && gen.Tok != token.VAR) {
				continue
			}
			for _, spec := range gen.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range value.Names {
					if i < len(value.Values) {
						values[name.Name] = value.Values[i]
					}
				}
			}
		}
	}
	return values
}

// customResourceKind reads ApiVersion and Kind from the
// &apiextensions.CustomResourceArgs{...} a NewCustomResource call passes.
func customResourceKind(call *ast.CallExpr, values map[string]ast.Expr) (group, kind string, err error) {
	if len(call.Args) < 3 {
		return "", "", fmt.Errorf("NewCustomResource called with %d arguments", len(call.Args))
	}
	args := call.Args[2]
	if unary, ok := args.(*ast.UnaryExpr); ok {
		args = unary.X
	}
	literal, ok := args.(*ast.CompositeLit)
	if !ok {
		return "", "", fmt.Errorf("NewCustomResource's arguments are not a CustomResourceArgs literal")
	}
	var apiVersion string
	for _, element := range literal.Elts {
		field, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := field.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch key.Name {
		case "ApiVersion":
			if apiVersion, err = stringValue(field.Value, values); err != nil {
				return "", "", fmt.Errorf("ApiVersion: %w", err)
			}
		case "Kind":
			if kind, err = stringValue(field.Value, values); err != nil {
				return "", "", fmt.Errorf("Kind: %w", err)
			}
		}
	}
	if apiVersion == "" || kind == "" {
		return "", "", fmt.Errorf("NewCustomResource names no ApiVersion or no Kind")
	}
	slash := strings.Index(apiVersion, "/")
	if slash < 0 {
		return "", "", fmt.Errorf("apiVersion %q has no group -- a custom resource always has one", apiVersion)
	}
	return apiVersion[:slash], kind, nil
}

// stringValue resolves a string built as a literal, pulumi.String(...), a
// package constant or variable, or a field of a package-level struct literal
// (the modules' `vars` convention).
func stringValue(expr ast.Expr, values map[string]ast.Expr) (string, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			return strconv.Unquote(e.Value)
		}
	case *ast.CallExpr:
		if sel, ok := e.Fun.(*ast.SelectorExpr); ok && (sel.Sel.Name == "String" || sel.Sel.Name == "StringPtr") && len(e.Args) == 1 {
			return stringValue(e.Args[0], values)
		}
	case *ast.Ident:
		if value, ok := values[e.Name]; ok {
			return stringValue(value, values)
		}
	case *ast.SelectorExpr:
		holder, ok := e.X.(*ast.Ident)
		if !ok {
			break
		}
		literal, ok := values[holder.Name].(*ast.CompositeLit)
		if !ok {
			break
		}
		for _, element := range literal.Elts {
			if field, ok := element.(*ast.KeyValueExpr); ok {
				if key, ok := field.Key.(*ast.Ident); ok && key.Name == e.Sel.Name {
					return stringValue(field.Value, values)
				}
			}
		}
	}
	return "", fmt.Errorf("cannot read %T as a string from source", expr)
}

// crd2pulumiToken reads the type token a generated constructor registers,
// from the generated package's source.
func crd2pulumiToken(root, importPath, constructor string) (string, error) {
	dir := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(importPath, modulePrefix)))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if err != nil {
			return "", err
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != constructor || fn.Body == nil {
				continue
			}
			tok := ""
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || tok != "" {
					return tok == ""
				}
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "RegisterResource" && len(call.Args) > 0 {
					if lit, ok := call.Args[0].(*ast.BasicLit); ok {
						tok, _ = strconv.Unquote(lit.Value)
					}
				}
				return tok == ""
			})
			if tok == "" {
				return "", fmt.Errorf("the generated constructor registers no literal token")
			}
			return tok, nil
		}
	}
	return "", fmt.Errorf("no constructor %s in %s", constructor, dir)
}

// splitToken reads "kubernetes:<group>/<version>:<Kind>".
func splitToken(tok string) (group, kind string, err error) {
	parts := strings.Split(tok, ":")
	if len(parts) != 3 || parts[0] != "kubernetes" || !strings.Contains(parts[1], "/") {
		return "", "", fmt.Errorf("token %q is not kubernetes:<group>/<version>:<Kind>", tok)
	}
	return parts[1][:strings.LastIndex(parts[1], "/")], parts[2], nil
}

// resourcePlural is the lowercase plural RBAC names a custom resource by.
// Definitions may declare any plural; every one the catalog creates follows
// this English rule, and a definition that does not surfaces as a gate
// failure naming the guess, never as a silent pass.
func resourcePlural(kind string) string {
	lower := strings.ToLower(kind)
	switch {
	case strings.HasSuffix(lower, "s"), strings.HasSuffix(lower, "x"), strings.HasSuffix(lower, "z"),
		strings.HasSuffix(lower, "ch"), strings.HasSuffix(lower, "sh"):
		return lower + "es"
	case strings.HasSuffix(lower, "y") && len(lower) > 1 && !strings.ContainsRune("aeiou", rune(lower[len(lower)-2])):
		return lower[:len(lower)-1] + "ies"
	}
	return lower + "s"
}

// The table's tokens are the SDK's own registrations (read from sdk/v4 v4.33.0's
// RegisterResource calls when the table was written); the irregular spellings -- the core
// group's "core", and groups whose SDK directory is a short name -- are pinned here.
func TestPulumiKubernetesKindsNameTheSDKsTokens(t *testing.T) {
	want := map[string]string{
		"NewDeployment":                     "kubernetes:apps/v1:Deployment",
		"NewNamespace":                      "kubernetes:core/v1:Namespace",
		"NewIngress":                        "kubernetes:networking.k8s.io/v1:Ingress",
		"NewClusterRoleBinding":             "kubernetes:rbac.authorization.k8s.io/v1:ClusterRoleBinding",
		"NewHorizontalPodAutoscaler":        "kubernetes:autoscaling/v2:HorizontalPodAutoscaler",
		"NewValidatingWebhookConfiguration": "kubernetes:admissionregistration.k8s.io/v1:ValidatingWebhookConfiguration",
	}
	seen := map[string]bool{}
	for _, kind := range PulumiKubernetesKinds {
		if seen[kind.Token] {
			t.Errorf("%s has two rows", kind.Token)
		}
		seen[kind.Token] = true
		if token, ok := want[kind.Constructor]; ok && token != kind.Token {
			t.Errorf("%s registers %s, the table says %s", kind.Constructor, token, kind.Token)
		}
	}
}

func TestResourcePluralFollowsTheDefinitionsTheCatalogCreates(t *testing.T) {
	for kind, want := range map[string]string{
		"Gateway": "gateways", "GatewayClass": "gatewayclasses", "AuthorizationPolicy": "authorizationpolicies",
		"Telemetry": "telemetries", "KafkaMirrorMaker2": "kafkamirrormaker2s", "EC2NodeClass": "ec2nodeclasses",
	} {
		if got := resourcePlural(kind); got != want {
			t.Errorf("resourcePlural(%q) = %q, want %q", kind, got, want)
		}
	}
}

// The table is read from one provider release; the module's go.mod picks the SDK, and the SDK asks
// the engine for the provider plugin of its own version. A pin that moves without the table being
// re-read would hold manifests to waits the new provider may no longer make, or miss new ones.
func TestPulumiKubernetesTableMatchesTheSDKPin(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("reading go.mod: %v", err)
	}
	pinned := ""
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "github.com/pulumi/pulumi-kubernetes/sdk/v4" {
			pinned = fields[1]
		}
		if len(fields) >= 3 && fields[0] == "require" && fields[1] == "github.com/pulumi/pulumi-kubernetes/sdk/v4" {
			pinned = fields[2]
		}
	}
	if pinned == "" {
		t.Fatal("go.mod requires no github.com/pulumi/pulumi-kubernetes/sdk/v4 -- the table has no provider to describe")
	}
	if pinned != PulumiKubernetesProviderVersion {
		t.Errorf("go.mod pins pulumi-kubernetes %s but pulumikubernetes.go was read from %s -- re-read provider/pkg/await at %s, update the table's rows and citations, then PulumiKubernetesProviderVersion",
			pinned, PulumiKubernetesProviderVersion, pinned)
	}
}

// The scanner fails loudly on what it cannot see, and reads the annotations' values, not their
// names: skipAwait counts only when it is "true", and waitFor adds a readiness wait.
func TestScannerIsLoudAboutWhatItCannotSee(t *testing.T) {
	root := repoRoot(t)
	dir := t.TempDir()
	source := `package module

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	certmanager "github.com/pulumi/pulumi-kubernetes-cert-manager/sdk/go/kubernetescertmanager"
	configmap "github.com/plantonhq/planton/catalog/kubernetes/kubernetesconfigmap/iac/pulumi/module"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func asValue() { constructor := corev1.NewSecret; _ = constructor }

func raw(ctx *pulumi.Context) { ctx.RegisterResource("kubernetes:core/v1:Secret", "s", nil, nil) }

func notSkipped(ctx *pulumi.Context) {
	appsv1.NewDeployment(ctx, "d", &appsv1.DeploymentArgs{Metadata: meta{Annotations: pulumi.StringMap{"pulumi.com/skipAwait": pulumi.String("false")}}})
}

func waited(ctx *pulumi.Context) {
	annotations := map[string]string{}
	annotations["pulumi.com/waitFor"] = "condition=Ready"
	corev1.NewService(ctx, "s", &corev1.ServiceArgs{Metadata: meta{Annotations: pulumi.ToStringMap(annotations)}})
}

var _ = certmanager.NewCertManager
var _ = configmap.Resources
`
	if err := os.WriteFile(filepath.Join(dir, "module.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	scan, err := scanPulumiModule(root, dir)
	if err != nil {
		t.Fatal(err)
	}
	problems := strings.Join(scan.problems, "\n")
	for _, want := range []string{
		"NewSecret is used as a value",
		`RegisterResource("kubernetes:core/v1:Secret")`,
		"pulumi-kubernetes-cert-manager/sdk/go/kubernetescertmanager, a Kubernetes type SDK the gate does not read",
		"kubernetesconfigmap/iac/pulumi/module, which calls",
	} {
		if !strings.Contains(problems, want) {
			t.Errorf("the scanner did not report %q; it reported:\n%s", want, problems)
		}
	}
	for _, object := range scan.objects {
		switch object.kind.Constructor {
		case "NewDeployment":
			if object.skipAwait {
				t.Error("pulumi.com/skipAwait: \"false\" was read as skipping the wait")
			}
		case "NewService":
			if !object.waitFor {
				t.Error("pulumi.com/waitFor set beside the Service's constructor was not read")
			}
		}
	}
	if len(scan.objects) != 2 {
		t.Errorf("want the Deployment and the Service, got %d objects", len(scan.objects))
	}
}

// RBAC's "*" matches any group, resource or verb, so a manifest that grants a wildcard covers the
// reads the table asks for.
func TestWildcardRulesCoverTheWaitsReads(t *testing.T) {
	rules := []*permissionsv1.KubernetesRule{{ApiGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}, ClusterScoped: true}}
	if gaps := uncoveredVerbs(rules, DeletionReads(KubernetesResource{Resource: "namespaces", ClusterScoped: true})); len(gaps) > 0 {
		t.Errorf("a wildcard rule left %v uncovered", gaps)
	}
}
