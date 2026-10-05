package module

import (
	azuretrafficmanagerendpointv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuretrafficmanagerendpoint/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureTrafficManagerEndpoint *azuretrafficmanagerendpointv1alpha1.AzureTrafficManagerEndpoint

	// ProfileId is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal ARM id.
	ProfileId string
}

// Traffic Manager endpoints carry NO ARM tags on any engine (the
// provider exposes none) -- the platform's derived tags land on the
// owning profile instead, so these locals derive no tag map.
func initializeLocals(ctx *pulumi.Context, iacInput *azuretrafficmanagerendpointv1alpha1.AzureTrafficManagerEndpointIacInput) *Locals {
	locals := &Locals{}

	locals.AzureTrafficManagerEndpoint = iacInput.Target
	locals.ProfileId = iacInput.Target.Spec.ProfileId.GetValue()

	return locals
}
