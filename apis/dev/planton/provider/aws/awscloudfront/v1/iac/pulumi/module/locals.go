package module

import (
	awscloudfrontv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awscloudfront/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals provides convenient access to commonly used values from the stack input.
type Locals struct {
	Target *awscloudfrontv1.AwsCloudFront
	Spec   *awscloudfrontv1.AwsCloudFrontSpec
}

// initializeLocals constructs the locals struct from the stack input.
func initializeLocals(ctx *pulumi.Context, in *awscloudfrontv1.AwsCloudFrontStackInput) *Locals {
	return &Locals{
		Target: in.Target,
		Spec:   in.Target.Spec,
	}
}
