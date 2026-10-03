package module

import (
	cogidpv1 "github.com/plantonhq/planton/catalog/aws/awscognitoidentityprovider/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// Identity providers carry no AwsTags map: the aws_cognito_identity_provider
// resource is not taggable (identity tagging lives on the pool).
type Locals struct {
	Target *cogidpv1.AwsCognitoIdentityProvider
	Spec   *cogidpv1.AwsCognitoIdentityProviderSpec
}

func initializeLocals(ctx *pulumi.Context, iacInput *cogidpv1.AwsCognitoIdentityProviderIacInput) *Locals {
	locals := &Locals{}
	locals.Target = iacInput.Target
	locals.Spec = iacInput.Target.Spec

	return locals
}
