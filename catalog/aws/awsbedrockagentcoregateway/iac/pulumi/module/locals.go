package module

import (
	"strconv"

	awsbedrockagentcoregatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockagentcoregateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbedrockagentcoregatewayv1alpha1.AwsBedrockAgentCoreGateway
	Spec   *awsbedrockagentcoregatewayv1alpha1.AwsBedrockAgentCoreGatewaySpec

	// GatewayName is metadata.name -- the naming basis both engines share
	// so a manifest deploys identically on either. AWS allows letters and
	// digits with single hyphens, max 100 characters (no underscores, no
	// consecutive hyphens).
	GatewayName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockagentcoregatewayv1alpha1.AwsBedrockAgentCoreGatewayIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata
	locals.GatewayName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBedrockAgentCoreGateway.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
