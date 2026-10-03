package module

import (
	"strings"

	gcpplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tokenSecretSuffix names the module-created Secret Manager secret
// (`<name>-token`) holding the runner token.
const tokenSecretSuffix = "-token"

// Locals holds computed values derived from the IaC input for use across
// the Pulumi module. Every resolution here has an exact twin in the
// Terraform module's locals.tf — keep them in lockstep.
type Locals struct {
	GcpPlantonRunner *gcpplantonrunnerv1alpha1.GcpPlantonRunner

	// GcpLabels carries the platform attribution labels applied on every
	// module-created resource.
	GcpLabels map[string]string

	// RunnerName is the name the runner registers itself under when it
	// joins the control plane: "<env>-<metadata.name>" (metadata.name
	// outside an environment) — the SAME derivation the platform uses for
	// records that reference this runner (its minted token, its managed
	// destroy); changing this formula breaks arrival attribution and
	// managed teardown.
	RunnerName string

	// TokenSecretId is the Secret Manager secret id (`<name>-token`)
	// holding the runner token.
	TokenSecretId string

	// ProjectId is the resolved literal project ("" = the provider's
	// default project, the ambient-project contract every GCP kind
	// honors).
	ProjectId string
}

// initializeLocals pulls values from the IaC input and populates the
// Locals struct. Similar to Terraform's "locals" concept.
func initializeLocals(_ *pulumi.Context, iacInput *gcpplantonrunnerv1alpha1.GcpPlantonRunnerIacInput) *Locals {
	target := iacInput.Target

	locals := &Locals{
		GcpPlantonRunner: target,
		TokenSecretId:    target.Metadata.Name + tokenSecretSuffix,
		ProjectId:        target.Spec.ProjectId.GetValue(),
	}

	locals.RunnerName = target.Metadata.Name
	if target.Metadata.Env != "" {
		locals.RunnerName = target.Metadata.Env + "-" + target.Metadata.Name
	}

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: target.Metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_GcpPlantonRunner.String()),
	}
	if target.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}
	if target.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}

	return locals
}
