package module

import (
	awsbatchcomputeenvironmentv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsbatchcomputeenvironment/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsBatchComputeEnvironment *awsbatchcomputeenvironmentv1.AwsBatchComputeEnvironment
	Labels                     map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsbatchcomputeenvironmentv1.AwsBatchComputeEnvironmentStackInput) *Locals {
	locals := &Locals{}
	locals.AwsBatchComputeEnvironment = in.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsBatchComputeEnvironment.Metadata.Org,
		"planton.org/environment":   locals.AwsBatchComputeEnvironment.Metadata.Env,
		"planton.org/resource-kind": "AwsBatchComputeEnvironment",
		"planton.org/resource-id":   locals.AwsBatchComputeEnvironment.Metadata.Id,
	}

	return locals
}
