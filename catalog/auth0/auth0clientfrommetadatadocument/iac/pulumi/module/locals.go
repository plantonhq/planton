package module

import (
	auth0clientfrommetadatadocumentv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0clientfrommetadatadocument/v1alpha1"
)

// Locals holds the values the module computes from the stack input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	// Spec is the declaration. The module reads presence straight from it:
	// every optional scalar is a pointer, every block a message, and an empty
	// list or map means "not managed" (see clientArgs).
	Spec *auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentSpec
}

func initializeLocals(stackInput *auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentStackInput) *Locals {
	target := stackInput.Target
	return &Locals{
		ResourceName: target.Metadata.Name,
		Spec:         target.Spec,
	}
}
