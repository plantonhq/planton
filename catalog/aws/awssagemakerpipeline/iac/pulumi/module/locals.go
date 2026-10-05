package module

import (
	"strconv"

	awssagemakerpipelinev1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakerpipeline/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awssagemakerpipelinev1alpha1.AwsSagemakerPipeline
	Spec   *awssagemakerpipelinev1alpha1.AwsSagemakerPipelineSpec

	PipelineName string
	DisplayName  string
	AwsTags      map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awssagemakerpipelinev1alpha1.AwsSagemakerPipelineIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// The pipeline's AWS name derives from metadata.name; the display
	// name defaults to it (the provider REQUIRES a display name).
	locals.PipelineName = metadata.Name
	locals.DisplayName = in.Target.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = metadata.Name
	}

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSagemakerPipeline.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
