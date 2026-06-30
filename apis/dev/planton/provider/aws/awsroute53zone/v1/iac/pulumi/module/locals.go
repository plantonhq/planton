package module

import (
	awsroute53zonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsroute53zone/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsRoute53Zone *awsroute53zonev1.AwsRoute53Zone
}

func initializeLocals(ctx *pulumi.Context, stackInput *awsroute53zonev1.AwsRoute53ZoneStackInput) *Locals {
	locals := &Locals{}
	locals.AwsRoute53Zone = stackInput.Target
	return locals
}
