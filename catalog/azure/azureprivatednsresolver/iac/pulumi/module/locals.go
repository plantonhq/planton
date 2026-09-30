package module

import (
	"strings"

	azureprivatednsresolverv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatednsresolver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzurePrivateDnsResolver *azureprivatednsresolverv1alpha1.AzurePrivateDnsResolver

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order. Applied to the resolver AND its
	// endpoints.
	AzureTags map[string]string
}

// allocationMethodWire maps the spec's allocation-method enum to the ARM
// wire value. Unspecified applies "Dynamic", the provider's own default,
// kept explicit so both engines send identical wire shapes.
var allocationMethodWire = map[azureprivatednsresolverv1alpha1.AzurePrivateDnsResolverIpAllocationMethod]string{
	azureprivatednsresolverv1alpha1.AzurePrivateDnsResolverIpAllocationMethod_DYNAMIC: "Dynamic",
	azureprivatednsresolverv1alpha1.AzurePrivateDnsResolverIpAllocationMethod_STATIC:  "Static",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureprivatednsresolverv1alpha1.AzurePrivateDnsResolverStackInput) *Locals {
	locals := &Locals{}

	locals.AzurePrivateDnsResolver = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzurePrivateDnsResolver.String()),
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
