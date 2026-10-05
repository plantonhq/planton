package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awscodepipelinev1alpha1 "github.com/plantonhq/planton/catalog/aws/awscodepipeline/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsCodePipeline *awscodepipelinev1alpha1.AwsCodePipeline

	// PipelineName is the pipeline's cloud name, derived from metadata.name.
	// Pipeline names are create-time-immutable (a name change replaces the
	// pipeline), which is why the name is not spec surface. Same basis as
	// the Terraform module.
	PipelineName string

	// AwsTags are the resource-identity tags, matching the Terraform module
	// key-for-key. DELIBERATELY five keys, no Name: the pipeline carries its
	// own name argument (the ECR/GA recorded convention). Identity tagging is
	// the only tagging surface this module manages; user-defined custom tags
	// are a platform-wide concern, not per-kind spec surface.
	AwsTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awscodepipelinev1alpha1.AwsCodePipelineIacInput) *Locals {
	locals := &Locals{}
	locals.AwsCodePipeline = in.Target

	locals.PipelineName = locals.AwsCodePipeline.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsCodePipeline.Metadata.Org,
		awstagkeys.Environment:  locals.AwsCodePipeline.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsCodePipeline.String(),
		awstagkeys.ResourceId:   locals.AwsCodePipeline.Metadata.Id,
	}

	return locals
}
