package module

import (
	awscognitoresourceserverv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscognitoresourceserver/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// Resource servers carry no AwsTags map: the aws_cognito_resource_server
// resource is not taggable (identity tagging lives on the pool).
type Locals struct {
	Target *awscognitoresourceserverv1alpha1.AwsCognitoResourceServer
	Spec   *awscognitoresourceserverv1alpha1.AwsCognitoResourceServerSpec
}

func initializeLocals(_ *pulumi.Context, in *awscognitoresourceserverv1alpha1.AwsCognitoResourceServerIacInput) *Locals {
	return &Locals{
		Target: in.Target,
		Spec:   in.Target.Spec,
	}
}
