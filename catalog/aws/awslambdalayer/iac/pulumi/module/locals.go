package module

import (
	awslambdalayerv1alpha1 "github.com/plantonhq/planton/catalog/aws/awslambdalayer/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// No AwsTags map here: neither layer versions nor layer-version
// permissions are taggable at AWS (the Terraform module carries the
// same absence).
type Locals struct {
	Target *awslambdalayerv1alpha1.AwsLambdaLayer
	Spec   *awslambdalayerv1alpha1.AwsLambdaLayerSpec
}

func initializeLocals(_ *pulumi.Context, in *awslambdalayerv1alpha1.AwsLambdaLayerIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	return locals
}
