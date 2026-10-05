package module

import (
	"strconv"

	awsbedrockflowv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockflow/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbedrockflowv1alpha1.AwsBedrockFlow
	Spec   *awsbedrockflowv1alpha1.AwsBedrockFlowSpec

	// FlowName is metadata.name -- the naming basis both engines share so
	// a manifest deploys identically on either.
	FlowName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockflowv1alpha1.AwsBedrockFlowIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata
	locals.FlowName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBedrockFlow.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
