package module

import (
	"strings"

	azurekeyvaultsecretv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurekeyvaultsecret/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureKeyVaultSecret *azurekeyvaultsecretv1alpha1.AzureKeyVaultSecret

	// KeyVaultId and Value are StringValueOrRef fields; the platform
	// middleware resolves valueFrom references (including managed
	// secrets) before IaC modules run, so GetValue() always returns
	// the resolved literal.
	KeyVaultId string
	Value      string

	// AzureTags is the metadata-derived tag map with the spec's user
	// tags merged over it (user tags win on key collision), mirroring
	// the Terraform module's merge order. Key Vault caps a secret at
	// 15 tags -- the spec's own cap leaves room for the derived
	// entries.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurekeyvaultsecretv1alpha1.AzureKeyVaultSecretStackInput) *Locals {
	locals := &Locals{}

	locals.AzureKeyVaultSecret = stackInput.Target
	target := stackInput.Target

	locals.KeyVaultId = target.Spec.KeyVaultId.GetValue()
	locals.Value = target.Spec.Value.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged
	// over them: user tags deliberately win so an org's governance
	// conventions (cost center, owner) can override the derived values
	// where they collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureKeyVaultSecret.String()),
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

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
