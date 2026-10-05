package module

import (
	azureprivatednsresolvervirtualnetworklinkv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatednsresolvervirtualnetworklink/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The virtual network link carries NO tags argument on the provider
// (ARM gives ruleset links a free-form metadata map instead, modeled
// as spec.metadata), so no tag map is derived here. Reference fields
// arrive pre-resolved (the platform middleware resolves valueFrom
// references before IaC modules run), so GetValue() always returns
// the resolved literal ARM ids.
type Locals struct {
	AzurePrivateDnsResolverVirtualNetworkLink *azureprivatednsresolvervirtualnetworklinkv1alpha1.AzurePrivateDnsResolverVirtualNetworkLink
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureprivatednsresolvervirtualnetworklinkv1alpha1.AzurePrivateDnsResolverVirtualNetworkLinkIacInput) *Locals {
	locals := &Locals{}

	locals.AzurePrivateDnsResolverVirtualNetworkLink = iacInput.Target

	return locals
}
