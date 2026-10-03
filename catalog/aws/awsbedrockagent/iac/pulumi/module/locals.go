package module

import (
	"strconv"

	awsbedrockagentv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockagent/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbedrockagentv1alpha1.AwsBedrockAgent
	Spec   *awsbedrockagentv1alpha1.AwsBedrockAgentSpec

	// AgentName is metadata.name -- the naming basis both engines share so
	// a manifest deploys identically on either. AWS allows up to 100
	// characters of alphanumeric plus - and _ (no spaces or dots).
	AgentName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockagentv1alpha1.AwsBedrockAgentIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata
	locals.AgentName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBedrockAgent.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
