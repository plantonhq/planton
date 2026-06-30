package module

import (
	awscodebuildprojectv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awscodebuildproject/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsCodeBuildProject *awscodebuildprojectv1.AwsCodeBuildProject
	Labels              map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awscodebuildprojectv1.AwsCodeBuildProjectStackInput) *Locals {
	locals := &Locals{}
	locals.AwsCodeBuildProject = in.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsCodeBuildProject.Metadata.Org,
		"planton.org/environment":   locals.AwsCodeBuildProject.Metadata.Env,
		"planton.org/resource-kind": "AwsCodeBuildProject",
		"planton.org/resource-id":   locals.AwsCodeBuildProject.Metadata.Id,
	}

	return locals
}
