package module

import (
	awsbedrockmodelaccessv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockmodelaccess/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// Model access creates no taggable resources (the agreement and the
// use-case form carry no tag surface at AWS), so the shared identity-tag
// map is deliberately absent here.
type Locals struct {
	Target *awsbedrockmodelaccessv1alpha1.AwsBedrockModelAccess
	Spec   *awsbedrockmodelaccessv1alpha1.AwsBedrockModelAccessSpec
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockmodelaccessv1alpha1.AwsBedrockModelAccessIacInput) *Locals {
	return &Locals{
		Target: in.Target,
		Spec:   in.Target.Spec,
	}
}
