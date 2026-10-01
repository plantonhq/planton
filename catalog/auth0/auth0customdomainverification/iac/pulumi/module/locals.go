package module

import (
	auth0customdomainverificationv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0customdomainverification/v1alpha1"
)

// Locals holds the values the module computes from the stack input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	// CustomDomainId is the custom domain to verify. A reference is resolved to
	// its value before the module runs.
	CustomDomainId string
}

func initializeLocals(stackInput *auth0customdomainverificationv1alpha1.Auth0CustomDomainVerificationStackInput) *Locals {
	target := stackInput.Target
	return &Locals{
		ResourceName:   target.Metadata.Name,
		CustomDomainId: target.Spec.CustomDomainId.GetValue(),
	}
}
