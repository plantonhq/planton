//go:build !codegen
// +build !codegen

package moduleverify

// An output the schema marks sensitive is a secret the resource generates, and the engine
// sees it before the platform does: exported in the clear, it is printed in the engine's deploy
// logs and kept readable in state. This suite pins that every official module of every kind
// exports exactly its schema's secret outputs as secrets, in both engines, and that each check
// fires on a module that does not -- a gate that cannot fail teaches false confidence.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
)

// secretOutputFindings returns the secret-output violations (both directions) in a result.
func secretOutputFindings(result *Result) []string {
	var found []string
	for _, v := range result.Violations {
		if strings.Contains(v.Summary, "is a secret in") || strings.Contains(v.Summary, "is exported as a secret") {
			found = append(found, string(v.Severity)+": "+v.File+": "+v.Summary)
		}
	}
	return found
}

func TestOutputFields_ReadTheSchemasMarks(t *testing.T) {
	fields, err := outputFields(catalogkindreflect.KindFromString("CloudflareZeroTrustAccessServiceToken"))
	if err != nil {
		t.Fatal(err)
	}
	if !fields["client_secret"] {
		t.Errorf("client_secret is marked sensitive in the schema; read %v", fields)
	}
	if fields["client_id"] {
		t.Errorf("client_id is not a secret; read %v", fields)
	}
}

func TestVerify_SecretOutputs_EveryOfficialModuleExportsExactlyItsSecrets(t *testing.T) {
	root := repoRoot(t)
	kindsWithSecrets := 0
	for _, kind := range catalogkindreflect.KindsList() {
		fields, err := outputFields(kind)
		if err != nil {
			continue
		}
		hasSecret := false
		for _, secret := range fields {
			hasSecret = hasSecret || secret
		}
		if hasSecret {
			kindsWithSecrets++
		}
		kindName := catalogkindreflect.ExtractKindNameByKind(kind)
		iacDir := filepath.Join(root, "catalog", catalogkindreflect.ProviderDirName(catalogkindreflect.GetProvider(kind)),
			strings.ToLower(kind.String()), "iac")
		for _, engine := range []struct {
			dir         string
			provisioner provisioner.ProvisionerType
		}{
			{"tf", provisioner.ProvisionerTypeTofu},
			{"pulumi", provisioner.ProvisionerTypePulumi},
		} {
			moduleDir := filepath.Join(iacDir, engine.dir)
			if _, err := os.Stat(moduleDir); err != nil {
				continue
			}
			t.Run(kindName+"/"+engine.dir, func(t *testing.T) {
				result := mustVerify(t, Input{KindName: kindName, ModuleDir: moduleDir, Provisioner: engine.provisioner, SkipToolchainChecks: true})
				for _, finding := range secretOutputFindings(result) {
					t.Error(finding)
				}
			})
		}
	}
	if kindsWithSecrets < 16 {
		t.Fatalf("found %d kinds with a secret output; the catalog carries at least 16 -- the walk is not reaching the outputs", kindsWithSecrets)
	}
}

func TestVerify_Tofu_SecretOutputs(t *testing.T) {
	variables := generatedVariablesTF(t, "CloudflareZeroTrustAccessServiceToken")
	verify := func(outputs string) *Result {
		moduleDir := writeModule(t, map[string]string{"variables.tf": variables, "main.tf": "", "outputs.tf": outputs})
		return mustVerify(t, Input{KindName: "CloudflareZeroTrustAccessServiceToken", ModuleDir: moduleDir, SkipToolchainChecks: true})
	}

	requireErrorContaining(t, verify(`output "client_secret" { value = "x" }`),
		`the OpenTofu output "client_secret" is a secret in CloudflareZeroTrustAccessServiceToken's schema`)

	if findings := secretOutputFindings(verify(`output "client_secret" {
  value     = "x"
  sensitive = true
}`)); len(findings) > 0 {
		t.Errorf("a secret declared sensitive must not be flagged: %v", findings)
	}

	requireWarningContaining(t, verify(`output "client_id" {
  value     = "x"
  sensitive = true
}`), `the OpenTofu output "client_id" is exported as a secret`)
}

func TestVerify_Pulumi_SecretOutputs(t *testing.T) {
	verify := func(module string) *Result {
		moduleDir := writeModule(t, map[string]string{
			"Pulumi.yaml":      "name: x\nruntime: go\n",
			"main.go":          validPulumiMain,
			"module/module.go": module,
		})
		return mustVerify(t, Input{KindName: "CloudflareZeroTrustAccessServiceToken", ModuleDir: moduleDir,
			Provisioner: provisioner.ProvisionerTypePulumi, SkipToolchainChecks: true})
	}

	clear := verify(`package module

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

const OpClientSecret = "client_secret"

func export(ctx *pulumi.Context, secret pulumi.StringOutput) {
	ctx.Export(OpClientSecret, secret)
}
`)
	requireErrorContaining(t, clear, `the Pulumi output "client_secret" is a secret in CloudflareZeroTrustAccessServiceToken's schema`)
	requireErrorContaining(t, clear, `ctx.Export("client_secret", pulumi.ToSecret(value))`)

	// An aliased import and a literal name resolve the same way.
	if findings := secretOutputFindings(verify(`package module

import sdk "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

func export(ctx *sdk.Context, secret sdk.StringOutput) {
	ctx.Export("client_secret", sdk.ToSecret(secret))
}
`)); len(findings) > 0 {
		t.Errorf("a secret exported through ToSecret must not be flagged: %v", findings)
	}

	requireWarningContaining(t, verify(`package module

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

func export(ctx *pulumi.Context, id pulumi.StringOutput) {
	ctx.Export("client_id", pulumi.ToSecret(id))
}
`), `the Pulumi output "client_id" is exported as a secret`)

	computed := verify(`package module

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func export(ctx *pulumi.Context, id pulumi.StringOutput) {
	ctx.Export(fmt.Sprintf("%s_id", "client"), id)
	ctx.Export("not_a_schema_field", id)
}
`)
	requireWarningContaining(t, computed, "not_a_schema_field")
	noticed := false
	for _, notice := range computed.Notices {
		noticed = noticed || strings.Contains(notice, "compute their output name")
	}
	if !noticed {
		t.Errorf("a computed export name must be reported as unchecked; notices: %v", computed.Notices)
	}
}
