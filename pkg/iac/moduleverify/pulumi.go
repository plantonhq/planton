//go:build !codegen
// +build !codegen

package moduleverify

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"gopkg.in/yaml.v3"
)

const (
	pulumiProjectFileName = "Pulumi.yaml"
	pulumiEntrypointName  = "main.go"
)

// verifyPulumi runs the Pulumi contract checks: the project file's declared
// runtime, the entrypoint's shape, the Go module context, every secret home
// the schema declares being read, and — when the toolchain is available — a
// real compile.
func verifyPulumi(kind cloudresourcekind.CloudResourceKind, kindName, moduleDir string, in Input, result *Result) {
	checkPulumiProjectFile(moduleDir, result)
	checkPulumiEntrypoint(kindName, moduleDir, result)
	checkEnclosingGoMod(moduleDir, result)
	checkSecretHomesPulumi(declaredSecretHomes(kind), moduleDir, result)

	// The outputs override machinery is engine-neutral: a pulumi module can
	// carry the same output_transform.yaml / transform-outputs override.
	checkOutputsOverride(kind, moduleDir, in.SampleOutputs, result)

	if in.SkipToolchainChecks {
		result.addNotice("compile check skipped on request (go build)")
	} else {
		runGoBuild(moduleDir, result)
	}
}

// checkPulumiProjectFile enforces the project file and its declared runtime.
// Deployments build customized pulumi modules from source with the Go
// toolchain, keyed on the declared `go` runtime — any other runtime cannot
// be executed.
func checkPulumiProjectFile(moduleDir string, result *Result) {
	projectPath := filepath.Join(moduleDir, pulumiProjectFileName)
	content, err := os.ReadFile(projectPath)
	if err != nil {
		result.addError(pulumiProjectFileName, "Pulumi.yaml is missing — a pulumi module must declare its project name and the go runtime (name: <project>, runtime: go)")
		return
	}

	runtimeName, err := parsePulumiRuntime(content)
	if err != nil {
		result.addError(pulumiProjectFileName, fmt.Sprintf("Pulumi.yaml could not be parsed: %v", err))
		return
	}
	if runtimeName != "go" {
		result.addError(pulumiProjectFileName, fmt.Sprintf(
			"the declared runtime is %q, but customized pulumi modules must declare the go runtime — deployments build them from source with the Go toolchain", runtimeName))
	}
}

// parsePulumiRuntime reads the runtime name from a Pulumi project file,
// accepting both forms the format allows (`runtime: go` and
// `runtime: {name: go}`).
func parsePulumiRuntime(content []byte) (string, error) {
	var project struct {
		Runtime yaml.Node `yaml:"runtime"`
	}
	if err := yaml.Unmarshal(content, &project); err != nil {
		return "", err
	}

	switch project.Runtime.Kind {
	case yaml.ScalarNode:
		return project.Runtime.Value, nil
	case yaml.MappingNode:
		var runtime struct {
			Name string `yaml:"name"`
		}
		if err := project.Runtime.Decode(&runtime); err != nil {
			return "", err
		}
		return runtime.Name, nil
	default:
		return "", nil
	}
}

// checkPulumiEntrypoint enforces the entrypoint contract: a `package main`
// at the module root that loads the kind's typed stack input. The typed
// input is how every deployment feeds the module; the loader is how the
// input reaches it.
func checkPulumiEntrypoint(kindName, moduleDir string, result *Result) {
	entrypointPath := filepath.Join(moduleDir, pulumiEntrypointName)
	if _, err := os.Stat(entrypointPath); err != nil {
		result.addError(pulumiEntrypointName, "main.go is missing at the module root — deployments build the module's root package as the Pulumi program")
		return
	}

	parsed, err := parser.ParseFile(token.NewFileSet(), entrypointPath, nil, 0)
	if err != nil {
		result.addError(pulumiEntrypointName, fmt.Sprintf("main.go could not be parsed: %v", err))
		return
	}

	if parsed.Name.Name != "main" {
		result.addError(pulumiEntrypointName, fmt.Sprintf(
			"the module root declares package %q — it must be package main, the Pulumi program deployments build and run", parsed.Name.Name))
	}

	stackInputTypeName := kindName + "StackInput"
	usesStackInput, usesLoader := false, false
	ast.Inspect(parsed, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if !ok {
			return true
		}
		switch ident.Name {
		case stackInputTypeName:
			usesStackInput = true
		case "LoadStackInput":
			usesLoader = true
		}
		return true
	})

	if !usesStackInput {
		result.addError(pulumiEntrypointName, fmt.Sprintf(
			"the entrypoint never references %s — deployments feed the module through that typed stack input; load it and pass it to your resources", stackInputTypeName))
	}
	if !usesLoader {
		result.addWarning(pulumiEntrypointName, fmt.Sprintf(
			"the entrypoint does not call LoadStackInput — the stack-input loader is how deployments deliver the %s value; loading it another way is on you to keep compatible", stackInputTypeName))
	}
}

// checkEnclosingGoMod requires a Go module context: deployments build the
// module from source, which needs a go.mod in the module directory or a
// parent (the module may live inside a larger repository that declares one).
func checkEnclosingGoMod(moduleDir string, result *Result) {
	dir := moduleDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	result.addError("go.mod", "no go.mod found in the module directory or any parent — deployments build the module from source and need a Go module context; create one with 'go mod init'")
}
