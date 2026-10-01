package eject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/iac/provisioner"
)

// The full eject pipeline runs against a hermetic fake source (the real
// resolution downloads artifacts or clones the staging repo). What these
// tests pin is everything eject GUARANTEES about its output: the
// never-overwrite rule, the copy filter, the pulumi self-import rewrite and
// go.mod synthesis, the LICENSE/NOTICE invariant, and the contract notes.

// fakeSource stages a module directory (and optionally a repo root with
// LICENSE/NOTICE) and points the resolution seam at it.
func fakeSource(t *testing.T, files map[string]string, withRepoRoot bool) {
	t.Helper()

	sourceRoot := t.TempDir()
	moduleDir := filepath.Join(sourceRoot, "module")
	for name, content := range files {
		path := filepath.Join(moduleDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	repoRoot := ""
	if withRepoRoot {
		repoRoot = sourceRoot
		for _, name := range licenseFileNames {
			if err := os.WriteFile(filepath.Join(sourceRoot, name), []byte(name+" content"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	original := resolveSourceFn
	resolveSourceFn = func(kindName string, prov provisioner.ProvisionerType) (*resolvedSource, error) {
		return &resolvedSource{dir: moduleDir, version: "v0.9.9", repoRoot: repoRoot}, nil
	}
	t.Cleanup(func() { resolveSourceFn = original })
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected file %s: %v", path, err)
	}
	return string(content)
}

func TestEject_TofuModule(t *testing.T) {
	fakeSource(t, map[string]string{
		"main.tf":             `resource "x" "y" {}`,
		"variables.tf":        `variable "spec" {}`,
		"BUILD.bazel":         "bazel scaffolding",
		"plan.tfplan":         "run leftover",
		".terraform/env":      "engine working dir",
		"README.md":           "module readme",
		".terraform.lock.hcl": "provider pins",
	}, true)

	outputDir := filepath.Join(t.TempDir(), "ejected")
	result, err := Eject(Input{
		KindName:    "AwsS3Bucket",
		Provisioner: provisioner.ProvisionerTypeTofu,
		OutputDir:   outputDir,
	})
	if err != nil {
		t.Fatalf("Eject() error: %v", err)
	}

	if result.KindName != "AwsS3Bucket" {
		t.Errorf("KindName = %q, want AwsS3Bucket", result.KindName)
	}
	if result.SourceVersion != "v0.9.9" {
		t.Errorf("SourceVersion = %q, want v0.9.9", result.SourceVersion)
	}

	// Module content travels; provider pins included.
	for _, name := range []string{"main.tf", "variables.tf", "README.md", ".terraform.lock.hcl"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Errorf("expected %s in the ejected copy: %v", name, err)
		}
	}
	// Repo scaffolding and run leftovers do not.
	for _, name := range []string{"BUILD.bazel", "plan.tfplan", ".terraform"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); !os.IsNotExist(err) {
			t.Errorf("%s must not be in the ejected copy", name)
		}
	}

	// The licensing invariant: attribution travels with every eject.
	for _, name := range licenseFileNames {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Errorf("expected %s in the ejected copy (backfilled from the repo root): %v", name, err)
		}
	}

	notes := mustReadFile(t, result.NotesPath)
	for _, want := range []string{"AwsS3Bucket", "v0.9.9", "planton module verify", "--module-dir", "LICENSE"} {
		if !strings.Contains(notes, want) {
			t.Errorf("contract notes missing %q", want)
		}
	}
}

func TestEject_PulumiModule_RewritesImportsAndSynthesizesGoMod(t *testing.T) {
	fakeSource(t, map[string]string{
		"main.go": `package main

import (
	"github.com/plantonhq/planton/catalog/aws/awss3bucket/iac/pulumi/module"
	awss3bucketv1 "github.com/plantonhq/planton/catalog/aws/awss3bucket/v1alpha1"
)

var _ = module.Resources
var _ = awss3bucketv1.AwsS3BucketStackInput{}
`,
		"module/main.go": "package module\n\nfunc Resources() {}\n",
		"Pulumi.yaml":    "name: awss3bucket\nruntime: go\n",
	}, true)

	outputDir := filepath.Join(t.TempDir(), "ejected")
	result, err := Eject(Input{
		KindName:      "AwsS3Bucket",
		Provisioner:   provisioner.ProvisionerTypePulumi,
		OutputDir:     outputDir,
		GoModulePath:  "github.com/example/s3-module",
		SkipGoModTidy: true,
	})
	if err != nil {
		t.Fatalf("Eject() error: %v", err)
	}

	mainGo := mustReadFile(t, filepath.Join(outputDir, "main.go"))
	if strings.Contains(mainGo, "github.com/plantonhq/planton/catalog/aws/awss3bucket/iac/pulumi") {
		t.Error("the self-import must be rewritten onto the user's module path")
	}
	if !strings.Contains(mainGo, `"github.com/example/s3-module/module"`) {
		t.Error("the rewritten self-import must point at the user's module path")
	}
	// Imports of the published engine module (stubs) stay untouched.
	if !strings.Contains(mainGo, "github.com/plantonhq/planton/catalog/aws/awss3bucket/v1alpha1") {
		t.Error("stub imports must keep resolving from the published engine module")
	}

	goMod := mustReadFile(t, filepath.Join(outputDir, "go.mod"))
	if !strings.Contains(goMod, "module github.com/example/s3-module") {
		t.Errorf("go.mod must declare the user's module path, got:\n%s", goMod)
	}
	if !strings.Contains(goMod, "require github.com/plantonhq/planton v0.9.9") {
		t.Errorf("go.mod must pin the engine module to the ejected release, got:\n%s", goMod)
	}

	if result.GoModTidyRan {
		t.Error("tidy must not run when skipped")
	}
}

func TestEject_PulumiModule_SynthesizesMissingPulumiYaml(t *testing.T) {
	// A handful of catalog modules ship without Pulumi.yaml; the ejected
	// copy must always carry one declaring the go runtime.
	fakeSource(t, map[string]string{
		"main.go": "package main\n",
	}, true)

	outputDir := filepath.Join(t.TempDir(), "ejected")
	if _, err := Eject(Input{
		KindName:      "AwsS3Bucket",
		Provisioner:   provisioner.ProvisionerTypePulumi,
		OutputDir:     outputDir,
		GoModulePath:  "github.com/example/s3-module",
		SkipGoModTidy: true,
	}); err != nil {
		t.Fatalf("Eject() error: %v", err)
	}

	pulumiYaml := mustReadFile(t, filepath.Join(outputDir, "Pulumi.yaml"))
	if !strings.Contains(pulumiYaml, "runtime: go") {
		t.Errorf("synthesized Pulumi.yaml must declare the go runtime, got:\n%s", pulumiYaml)
	}
}

func TestEject_GoModOmitsRequireForUnreleasedSource(t *testing.T) {
	// An unreleased source version (a staging checkout of a branch) cannot
	// be a go.mod require version: the require line is left for go mod tidy.
	moduleDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	original := resolveSourceFn
	resolveSourceFn = func(kindName string, prov provisioner.ProvisionerType) (*resolvedSource, error) {
		return &resolvedSource{dir: moduleDir, version: "main"}, nil
	}
	t.Cleanup(func() { resolveSourceFn = original })

	outputDir := filepath.Join(t.TempDir(), "ejected")
	if _, err := Eject(Input{
		KindName:      "AwsS3Bucket",
		Provisioner:   provisioner.ProvisionerTypePulumi,
		OutputDir:     outputDir,
		GoModulePath:  "github.com/example/s3-module",
		SkipGoModTidy: true,
	}); err != nil {
		t.Fatalf("Eject() error: %v", err)
	}

	goMod := mustReadFile(t, filepath.Join(outputDir, "go.mod"))
	if strings.Contains(goMod, "require") {
		t.Errorf("go.mod must not pin a non-release version, got:\n%s", goMod)
	}
}

func TestEject_RefusesNonEmptyOutputDir(t *testing.T) {
	fakeSource(t, map[string]string{"main.tf": ""}, true)

	outputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outputDir, "precious.txt"), []byte("user content"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Eject(Input{
		KindName:    "AwsS3Bucket",
		Provisioner: provisioner.ProvisionerTypeTofu,
		OutputDir:   outputDir,
	})
	if err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("expected the never-overwrite refusal, got: %v", err)
	}
}

func TestEject_RefusesAnEngineTheKindDoesNotRunOn(t *testing.T) {
	fakeSource(t, map[string]string{"main.go": "package main"}, true)

	outputDir := filepath.Join(t.TempDir(), "out")
	_, err := Eject(Input{
		KindName:    "OpenFgaStore",
		Provisioner: provisioner.ProvisionerTypePulumi,
		OutputDir:   outputDir,
	})
	if err == nil || !strings.Contains(err.Error(), "OpenFgaStore runs on OpenTofu or Terraform only, so Pulumi cannot deploy it") {
		t.Fatalf("expected the kind's engine refusal, got: %v", err)
	}
	if _, statErr := os.Stat(outputDir); statErr == nil {
		t.Error("a refused eject must not create its output directory")
	}
}

func TestEject_UnknownKindFailsPlainly(t *testing.T) {
	_, err := Eject(Input{
		KindName:    "NoSuchKind",
		Provisioner: provisioner.ProvisionerTypeTofu,
		OutputDir:   filepath.Join(t.TempDir(), "out"),
	})
	if err == nil || !strings.Contains(err.Error(), "unknown cloud resource kind") {
		t.Fatalf("expected the unknown-kind error, got: %v", err)
	}
}
