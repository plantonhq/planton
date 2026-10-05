package module

import (
	"strings"

	azureplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tokenSecretName is the Container App secret holding the runner token —
// the app's own secret store, referenced by the env's secret_name. A
// fixed name: the app carries exactly one secret, and the name is part of
// the cross-engine contract (and an output).
const tokenSecretName = "runner-token"

// Locals holds computed values derived from the IaC input for use across
// the Pulumi module. Every resolution here has an exact twin in the
// Terraform module's locals.tf — keep them in lockstep.
type Locals struct {
	AzurePlantonRunner *azureplantonrunnerv1alpha1.AzurePlantonRunner

	// AzureTags carries the platform attribution tags applied on the app.
	AzureTags map[string]string

	// ResourceGroupName is the resolved literal resource group.
	ResourceGroupName string

	// RunnerName is the name the runner registers itself under when it
	// joins the control plane: "<env>-<metadata.name>" (metadata.name
	// outside an environment) — the SAME derivation the platform uses for
	// records that reference this runner (its minted token, its managed
	// destroy); changing this formula breaks arrival attribution and
	// managed teardown.
	RunnerName string
}

// initializeLocals pulls values from the IaC input and populates the
// Locals struct. Similar to Terraform's "locals" concept.
func initializeLocals(_ *pulumi.Context, iacInput *azureplantonrunnerv1alpha1.AzurePlantonRunnerIacInput) *Locals {
	target := iacInput.Target

	locals := &Locals{
		AzurePlantonRunner: target,
		ResourceGroupName:  target.Spec.ResourceGroup.GetValue(),
	}

	locals.RunnerName = target.Metadata.Name
	if target.Metadata.Env != "" {
		locals.RunnerName = target.Metadata.Env + "-" + target.Metadata.Name
	}

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzurePlantonRunner.String()),
	}
	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	return locals
}
