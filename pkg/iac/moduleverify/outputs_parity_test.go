//go:build !codegen
// +build !codegen

package moduleverify

// A kind's outputs schema is the contract every deployment of it keeps: the platform reads
// those fields to find, judge and wire what was deployed (a Cloud Run service's project and
// region, an ECS service's cluster). A module that never emits a field leaves it empty on every
// resource it deploys, and whatever reads it fails far from the cause -- a Cloud Run service that
// ran but read "cannot be located on the provider". Both engines feed one transformer, so each
// must emit every field. This suite pins that every official module of every kind, in each
// engine it ships, populates its kind's whole contract, and that the check fires per engine.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
)

// outputsPending names the official modules that do not yet emit their kind's whole contract.
// The ledger only shrinks: a listed module that now emits everything fails the suite until its
// line is removed, so a fix is never left unrecorded, and a module not listed may never regress.
var outputsPending = map[string]string{
	"Auth0Client/tf":          "callback_url_template and global are not yet emitted",
	"Auth0Client/pulumi":      "callback_url_template and global are not yet emitted",
	"Auth0EventStream/pulumi": "aws_partner_event_source is not yet exported",
	"KubernetesJenkins/tf":    "the module declares none of the kind's service coordinates yet",
}

// unpopulatedOutputWarnings are the verifier's findings that an outputs field is never emitted.
func unpopulatedOutputWarnings(result *Result) []string {
	var found []string
	for _, v := range result.Violations {
		if strings.Contains(v.Summary, "outputs fields, so they stay empty") ||
			strings.Contains(v.Summary, "outputs field will stay empty") {
			found = append(found, v.Summary)
		}
	}
	return found
}

func TestVerify_Outputs_EveryOfficialModulePopulatesItsKindsContract(t *testing.T) {
	root := repoRoot(t)
	modules := 0
	for _, kind := range catalogkindreflect.KindsList() {
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
			modules++
			id := kindName + "/" + engine.dir
			t.Run(id, func(t *testing.T) {
				result := mustVerify(t, Input{KindName: kindName, ModuleDir: moduleDir, Provisioner: engine.provisioner, SkipToolchainChecks: true})
				warnings := unpopulatedOutputWarnings(result)
				if _, pending := outputsPending[id]; pending {
					if len(warnings) == 0 {
						t.Errorf("%s now emits its whole contract -- remove it from outputsPending", id)
					}
					return
				}
				for _, warning := range warnings {
					t.Error(warning)
				}
			})
		}
	}
	if modules < 100 {
		t.Fatalf("found %d official modules; the catalog carries hundreds -- the walk is not reaching catalog/", modules)
	}
}

// A nested field is exported by its path; the transformer flattens "password_secret.name" onto
// password_secret, so the check must count the path for the top-level field it names -- neither
// calling the export unknown nor the field empty.
func TestVerify_Pulumi_ADottedExportPopulatesItsTopLevelField(t *testing.T) {
	module := writeModule(t, map[string]string{
		"Pulumi.yaml": "name: x\nruntime: go\n",
		"main.go":     validPulumiMain,
		"module/outputs.go": `package module

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

func export(ctx *pulumi.Context) {
	ctx.Export("password_secret.name", pulumi.String("session-cache-auth"))
	ctx.Export("password_secret.key", pulumi.String("default"))
}
`,
	})
	result := mustVerify(t, Input{KindName: "KubernetesValkey", ModuleDir: module, Provisioner: provisioner.ProvisionerTypePulumi, SkipToolchainChecks: true})
	for _, v := range result.Violations {
		if strings.Contains(v.Summary, "password_secret") {
			t.Errorf("a dotted export of password_secret must count for it: %s", v.Summary)
		}
	}
}
