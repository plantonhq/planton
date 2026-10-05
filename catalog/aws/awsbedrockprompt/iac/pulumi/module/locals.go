package module

import (
	"strconv"

	awsbedrockpromptv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockprompt/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbedrockpromptv1alpha1.AwsBedrockPrompt
	Spec   *awsbedrockpromptv1alpha1.AwsBedrockPromptSpec

	// PromptName is metadata.name -- the naming basis both engines share
	// so a manifest deploys identically on either.
	PromptName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockpromptv1alpha1.AwsBedrockPromptIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata
	locals.PromptName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBedrockPrompt.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
