package module

import (
	awscodepipelinev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscodepipeline/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsCodePipeline *awscodepipelinev1.AwsCodePipeline
	Labels          map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awscodepipelinev1.AwsCodePipelineStackInput) *Locals {
	locals := &Locals{}
	locals.AwsCodePipeline = in.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsCodePipeline.Metadata.Org,
		"planton.org/environment":   locals.AwsCodePipeline.Metadata.Env,
		"planton.org/resource-kind": "AwsCodePipeline",
		"planton.org/resource-id":   locals.AwsCodePipeline.Metadata.Id,
	}

	return locals
}
