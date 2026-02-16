package module

import (
	awsmwaaenvironmentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsmwaaenvironment/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsMwaaEnvironment *awsmwaaenvironmentv1.AwsMwaaEnvironment
	Labels             map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsmwaaenvironmentv1.AwsMwaaEnvironmentStackInput) *Locals {
	locals := &Locals{}

	locals.AwsMwaaEnvironment = in.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsMwaaEnvironment.Metadata.Org,
		"planton.org/environment":   locals.AwsMwaaEnvironment.Metadata.Env,
		"planton.org/resource-kind": "AwsMwaaEnvironment",
		"planton.org/resource-id":   locals.AwsMwaaEnvironment.Metadata.Id,
	}

	return locals
}
