package module

import (
	awscognitouserpoolclientv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscognitouserpoolclient/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// App clients carry no AwsTags map: the aws_cognito_user_pool_client resource
// is not taggable (identity tagging lives on the pool).
type Locals struct {
	Target *awscognitouserpoolclientv1alpha1.AwsCognitoUserPoolClient
	Spec   *awscognitouserpoolclientv1alpha1.AwsCognitoUserPoolClientSpec
}

func initializeLocals(_ *pulumi.Context, in *awscognitouserpoolclientv1alpha1.AwsCognitoUserPoolClientIacInput) *Locals {
	return &Locals{
		Target: in.Target,
		Spec:   in.Target.Spec,
	}
}
