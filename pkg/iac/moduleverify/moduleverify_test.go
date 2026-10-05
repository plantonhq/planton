//go:build !codegen
// +build !codegen

package moduleverify

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
)

// writeModule materializes a module directory in a temp dir.
func writeModule(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func mustVerify(t *testing.T, in Input) *Result {
	t.Helper()
	in.SkipToolchainChecks = true
	result, err := Verify(in)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}
	return result
}

func errorSummaries(result *Result) []string {
	var summaries []string
	for _, v := range result.Violations {
		if v.Severity == SeverityError {
			summaries = append(summaries, v.Summary)
		}
	}
	return summaries
}

func requireErrorContaining(t *testing.T, result *Result, fragment string) {
	t.Helper()
	for _, summary := range errorSummaries(result) {
		if strings.Contains(summary, fragment) {
			return
		}
	}
	t.Errorf("expected an error containing %q, got errors: %v", fragment, errorSummaries(result))
}

func requireWarningContaining(t *testing.T, result *Result, fragment string) {
	t.Helper()
	for _, v := range result.Violations {
		if v.Severity == SeverityWarning && strings.Contains(v.Summary, fragment) {
			return
		}
	}
	t.Errorf("expected a warning containing %q, got: %+v", fragment, result.Violations)
}

// generatedVariablesTF renders the kind's canonical variables.tf — a
// schema-conformant input surface by construction.
func generatedVariablesTF(t *testing.T, kindName string) string {
	t.Helper()
	instance, err := catalogkindreflect.NewInstance(catalogkindreflect.KindFromString(kindName))
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := generators.ProtoToVariablesTF(instance)
	if err != nil {
		t.Fatal(err)
	}
	return rendered
}

// --- The green invariant ---------------------------------------------------

// A schema-generated input surface must verify clean: verify's checks are
// anchored to the same schema, so a module that declares exactly the
// generated surface can never be in violation.
func TestVerify_GeneratedSurfaceIsGreen(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"variables.tf": generatedVariablesTF(t, "AwsS3Bucket"),
		"main.tf":      "",
		"outputs.tf":   `output "bucket_id" { value = "x" }`,
	})

	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir})
	if result.HasErrors() {
		t.Errorf("a generated input surface must verify without errors, got: %v", errorSummaries(result))
	}
}

// An UNMODIFIED ejected official module must verify green (no errors) — the
// one outcome verify must never produce is failing the module it just
// handed the user. Runs against the real catalog: a generator-owned kind, a
// hand-curated kind, and a pulumi module.
func TestVerify_UnmodifiedOfficialModulesAreGreen(t *testing.T) {
	root := repoRoot(t)

	cases := []struct {
		kindName    string
		moduleDir   string
		provisioner provisioner.ProvisionerType
	}{
		// Generator-owned variables.tf.
		{"AwsS3Bucket", "catalog/aws/awss3bucket/iac/tf", provisioner.ProvisionerTypeTofu},
		// Hand-curated variables.tf (diverges from the generator on the
		// optional surface — allowed to warn, never to error).
		{"GcpGkeCluster", "catalog/gcp/gcpgkecluster/iac/tf", provisioner.ProvisionerTypeTofu},
		// Pulumi shape.
		{"AwsS3Bucket", "catalog/aws/awss3bucket/iac/pulumi", provisioner.ProvisionerTypePulumi},
	}

	for _, tc := range cases {
		t.Run(tc.moduleDir, func(t *testing.T) {
			result := mustVerify(t, Input{
				KindName:    tc.kindName,
				ModuleDir:   filepath.Join(root, tc.moduleDir),
				Provisioner: tc.provisioner,
			})
			if result.HasErrors() {
				t.Errorf("an unmodified official module must verify green, got: %v", errorSummaries(result))
			}
		})
	}
}

// repoRoot locates the repository checkout; skips under Bazel where the
// catalog tree is not part of the test's inputs (same posture as the
// sibling repo-reading gates).
func repoRoot(t *testing.T) string {
	t.Helper()
	if os.Getenv("TEST_WORKSPACE") != "" {
		t.Skip("repo-reading test; skipped under Bazel")
	}
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog tree not present; skipping repo-reading test")
	}
	return root
}

// --- Hermetic negative fixtures ---------------------------------------------
// A gate that cannot fail teaches false confidence: every violation class
// fires against a synthetic module built to break exactly that rule.

func TestVerify_Tofu_MissingVariablesFile(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{"main.tf": ""})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir})
	requireErrorContaining(t, result, "variables.tf is missing")
}

func TestVerify_Tofu_MissingSpecVariable(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"main.tf":      "",
		"variables.tf": `variable "metadata" { type = object({ name = string }) }`,
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir})
	requireErrorContaining(t, result, `no "spec" variable`)
}

func TestVerify_Tofu_ExtraRequiredVariable(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"main.tf": "",
		"variables.tf": generatedVariablesTF(t, "AwsS3Bucket") + `
variable "team_tag" {
  type = string
}
`,
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir})
	requireErrorContaining(t, result, `"team_tag" has no default`)
}

func TestVerify_Tofu_ExtraVariableWithDefaultIsFine(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"main.tf": "",
		"variables.tf": generatedVariablesTF(t, "AwsS3Bucket") + `
variable "team_tag" {
  type    = string
  default = "platform"
}
`,
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir})
	if result.HasErrors() {
		t.Errorf("an extra variable with a default must not error, got: %v", errorSummaries(result))
	}
}

func TestVerify_Tofu_SpecMissingRequiredAttribute(t *testing.T) {
	// AwsS3Bucket's spec requires region; a spec type without it fails the
	// value conversion on every deployment.
	moduleDir := writeModule(t, map[string]string{
		"main.tf": "",
		"variables.tf": `
variable "metadata" {
  type = object({ name = string })
}
variable "spec" {
  type = object({
    force_destroy = optional(bool, false)
  })
}
`,
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir})
	requireErrorContaining(t, result, "spec.region")
}

func TestVerify_Tofu_SpecExtraRequiredAttribute(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"main.tf": "",
		"variables.tf": `
variable "metadata" {
  type = object({ name = string })
}
variable "spec" {
  type = object({
    region        = string
    mandatory_tag = string
  })
}
`,
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir})
	requireErrorContaining(t, result, "spec.mandatory_tag")
}

func TestVerify_Tofu_SpecExtraOptionalAttributeWarns(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"main.tf": "",
		"variables.tf": `
variable "metadata" {
  type = object({ name = string })
}
variable "spec" {
  type = object({
    region     = string
    extra_knob = optional(string, "")
  })
}
`,
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir})
	requireWarningContaining(t, result, "spec.extra_knob")
	for _, summary := range errorSummaries(result) {
		if strings.Contains(summary, "extra_knob") {
			t.Errorf("an unused optional attribute must warn, not error: %s", summary)
		}
	}
}

func TestVerify_Tofu_RequiredWhereSchemaOptionalWarns(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"main.tf": "",
		"variables.tf": `
variable "metadata" {
  type = object({ name = string })
}
variable "spec" {
  type = object({
    region        = string
    force_destroy = bool
  })
}
`,
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir})
	requireWarningContaining(t, result, "spec.force_destroy")
}

func TestVerify_Tofu_UnknownOutputWarns(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"main.tf":      "",
		"variables.tf": generatedVariablesTF(t, "AwsS3Bucket"),
		"outputs.tf":   `output "not_a_schema_field" { value = "x" }`,
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir})
	requireWarningContaining(t, result, "not_a_schema_field")
}

func TestVerify_Pulumi_MissingProjectFile(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{"main.go": "package main\n"})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir, Provisioner: provisioner.ProvisionerTypePulumi})
	requireErrorContaining(t, result, "Pulumi.yaml is missing")
}

func TestVerify_Pulumi_WrongRuntime(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"Pulumi.yaml": "name: x\nruntime: nodejs\n",
		"main.go":     "package main\n",
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir, Provisioner: provisioner.ProvisionerTypePulumi})
	requireErrorContaining(t, result, `runtime is "nodejs"`)
}

func TestVerify_Pulumi_MappingFormRuntimeAccepted(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"Pulumi.yaml": "name: x\nruntime:\n  name: go\n",
		"main.go":     validPulumiMain,
		"go.mod":      "module example.com/x\n\ngo 1.26\n",
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir, Provisioner: provisioner.ProvisionerTypePulumi})
	for _, summary := range errorSummaries(result) {
		if strings.Contains(summary, "runtime") {
			t.Errorf("the mapping form of the go runtime must be accepted: %s", summary)
		}
	}
}

const validPulumiMain = `package main

import "github.com/example/x/module"

type AwsS3BucketIacInput struct{}

func LoadIacInput(v interface{}) error { return nil }

func main() {
	iacInput := &AwsS3BucketIacInput{}
	_ = LoadIacInput(iacInput)
	_ = module.Resources
}
`

func TestVerify_Pulumi_MissingEntrypoint(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"Pulumi.yaml": "name: x\nruntime: go\n",
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir, Provisioner: provisioner.ProvisionerTypePulumi})
	requireErrorContaining(t, result, "main.go is missing")
}

func TestVerify_Pulumi_WrongPackageName(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"Pulumi.yaml": "name: x\nruntime: go\n",
		"main.go":     "package notmain\n",
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir, Provisioner: provisioner.ProvisionerTypePulumi})
	requireErrorContaining(t, result, "package main")
}

func TestVerify_Pulumi_MissingIacInputReference(t *testing.T) {
	moduleDir := writeModule(t, map[string]string{
		"Pulumi.yaml": "name: x\nruntime: go\n",
		"main.go":     "package main\n\nfunc main() {}\n",
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir, Provisioner: provisioner.ProvisionerTypePulumi})
	requireErrorContaining(t, result, "AwsS3BucketIacInput")
}

func TestVerify_Pulumi_MissingGoModContext(t *testing.T) {
	// t.TempDir lives outside any Go module, so the walk up finds nothing.
	moduleDir := writeModule(t, map[string]string{
		"Pulumi.yaml": "name: x\nruntime: go\n",
		"main.go":     validPulumiMain,
	})
	result := mustVerify(t, Input{KindName: "AwsS3Bucket", ModuleDir: moduleDir, Provisioner: provisioner.ProvisionerTypePulumi})
	requireErrorContaining(t, result, "no go.mod found")
}

// --- Invocation-level behavior ----------------------------------------------

func TestVerify_UnknownKind(t *testing.T) {
	if _, err := Verify(Input{KindName: "NoSuchKind", ModuleDir: t.TempDir()}); err == nil ||
		!strings.Contains(err.Error(), "unknown catalog kind") {
		t.Fatalf("expected the unknown-kind error, got: %v", err)
	}
}

func TestInferProvisioner(t *testing.T) {
	tofuDir := writeModule(t, map[string]string{"main.tf": ""})
	if prov, err := inferProvisioner(tofuDir); err != nil || prov != provisioner.ProvisionerTypeTofu {
		t.Errorf("tf files must infer tofu, got %v, %v", prov, err)
	}

	pulumiDir := writeModule(t, map[string]string{"Pulumi.yaml": ""})
	if prov, err := inferProvisioner(pulumiDir); err != nil || prov != provisioner.ProvisionerTypePulumi {
		t.Errorf("Pulumi.yaml must infer pulumi, got %v, %v", prov, err)
	}

	ambiguousDir := writeModule(t, map[string]string{"main.tf": "", "main.go": ""})
	if _, err := inferProvisioner(ambiguousDir); err == nil || !strings.Contains(err.Error(), "--provisioner") {
		t.Errorf("an ambiguous directory must demand --provisioner, got: %v", err)
	}

	emptyDir := t.TempDir()
	if _, err := inferProvisioner(emptyDir); err == nil {
		t.Error("an empty directory must not infer an engine")
	}
}
