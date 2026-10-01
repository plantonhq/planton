//go:build !codegen
// +build !codegen

package moduleverify

// A secret home is part of a module's contract: the platform refuses a secret in a field every
// viewer reads and sends the author to the home, so a module that never reads the home deploys
// without the value the author was told to put there. This suite pins that every official module
// of every kind declaring a home reads it in both engines, and that the check fires, per engine,
// on a module that does not -- a gate that cannot fail teaches false confidence.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
)

func secretHomeWarnings(result *Result) []string {
	var found []string
	for _, v := range result.Violations {
		if strings.Contains(v.Summary, "never reads spec.") {
			found = append(found, v.Summary)
		}
	}
	return found
}

func TestDeclaredSecretHomes_ReadFromTheKindsSchema(t *testing.T) {
	homes := declaredSecretHomes(crkreflect.KindFromString("GcpCloudRun"))
	for _, home := range homes {
		if home.HomePath == "containers.env.secret_value" && home.Refused == "containers.env.value" {
			return
		}
	}
	t.Fatalf("GcpCloudRun declares env[].value's home as env[].secret_value; found %+v", homes)
}

func TestVerify_SecretHomes_EveryOfficialModuleReadsItsKindsHomes(t *testing.T) {
	root := repoRoot(t)
	kindsWithHomes := 0
	for _, kind := range crkreflect.KindsList() {
		homes := declaredSecretHomes(kind)
		if len(homes) == 0 {
			continue
		}
		kindsWithHomes++
		kindName := crkreflect.ExtractKindNameByKind(kind)
		iacDir := filepath.Join(root, "catalog", crkreflect.ProviderDirName(crkreflect.GetProvider(kind)),
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
				result := mustVerify(t, Input{KindName: kindName, ModuleDir: moduleDir, Provisioner: engine.provisioner})
				for _, warning := range secretHomeWarnings(result) {
					t.Error(warning)
				}
			})
		}
	}
	if kindsWithHomes < 3 {
		t.Fatalf("found %d kinds declaring a secret home; the catalog carries at least 3 -- the walk is not reaching the specs", kindsWithHomes)
	}
}

func TestVerify_Tofu_AnUnreadSecretHomeWarns(t *testing.T) {
	unread := writeModule(t, map[string]string{
		"variables.tf": generatedVariablesTF(t, "GcpCloudRun"),
		"main.tf":      `locals { names = [for c in var.spec.containers : [for e in c.env : e.value]] }`,
	})
	requireWarningContaining(t, mustVerify(t, Input{KindName: "GcpCloudRun", ModuleDir: unread}),
		"never reads spec.containers.env.secret_value")

	read := writeModule(t, map[string]string{
		"variables.tf": generatedVariablesTF(t, "GcpCloudRun"),
		"main.tf":      `locals { secrets = [for c in var.spec.containers : [for e in c.env : e.secret_value]] }`,
	})
	if warnings := secretHomeWarnings(mustVerify(t, Input{KindName: "GcpCloudRun", ModuleDir: read})); len(warnings) > 0 {
		t.Errorf("a module that reads the home must not be warned: %v", warnings)
	}
}

func TestVerify_Pulumi_AnUnreadSecretHomeWarns(t *testing.T) {
	const project = "name: x\nruntime: go\n"
	unread := writeModule(t, map[string]string{
		"Pulumi.yaml":      project,
		"main.go":          validPulumiMain,
		"module/module.go": "package module\n\nfunc value(e struct{ Value string }) string { return e.Value }\n",
	})
	requireWarningContaining(t, mustVerify(t, Input{KindName: "GcpCloudRun", ModuleDir: unread, Provisioner: provisioner.ProvisionerTypePulumi}),
		"never reads spec.containers.env.secret_value")

	read := writeModule(t, map[string]string{
		"Pulumi.yaml":      project,
		"main.go":          validPulumiMain,
		"module/module.go": "package module\n\nfunc secret(e struct{ SecretValue string }) string { return e.SecretValue }\n",
	})
	if warnings := secretHomeWarnings(mustVerify(t, Input{KindName: "GcpCloudRun", ModuleDir: read, Provisioner: provisioner.ProvisionerTypePulumi})); len(warnings) > 0 {
		t.Errorf("a module that reads the home must not be warned: %v", warnings)
	}
}

func TestVerify_Pulumi_ASecretHomeReadThroughUnreachableCatalogHelpersIsANoticeNotAWarning(t *testing.T) {
	// An ejected module outside the catalog's Go module that delegates to the catalog's shared
	// helpers: the check cannot read those helpers, so it says what it could not confirm instead
	// of warning about code it never saw.
	delegating := writeModule(t, map[string]string{
		"Pulumi.yaml": "name: x\nruntime: go\n",
		"main.go":     validPulumiMain,
		"module/module.go": "package module\n\nimport _ \"" + catalogModulePath +
			"/pkg/iac/pulumi/pulumimodule/provider/gcp/cloudrunenv\"\n",
	})
	result := mustVerify(t, Input{KindName: "GcpCloudRun", ModuleDir: delegating, Provisioner: provisioner.ProvisionerTypePulumi})
	if warnings := secretHomeWarnings(result); len(warnings) > 0 {
		t.Errorf("a home read through unreachable helpers must not be warned: %v", warnings)
	}
	confirmed := false
	for _, notice := range result.Notices {
		confirmed = confirmed || strings.Contains(notice, "could not confirm spec.containers.env.secret_value")
	}
	if !confirmed {
		t.Errorf("the check must say what it could not confirm; notices: %v", result.Notices)
	}
}

func TestGoFieldName(t *testing.T) {
	for proto, want := range map[string]string{
		"secret_value":       "SecretValue",
		"secret_environment": "SecretEnvironment",
		"value":              "Value",
		"s3_key":             "S3Key",
	} {
		if got := goFieldName(proto); got != want {
			t.Errorf("goFieldName(%q) = %q, want %q", proto, got, want)
		}
	}
}
