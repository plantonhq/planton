package module

import (
	"strconv"

	awsbedrockguardrailv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockguardrail/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbedrockguardrailv1alpha1.AwsBedrockGuardrail
	Spec   *awsbedrockguardrailv1alpha1.AwsBedrockGuardrailSpec

	// GuardrailName is metadata.name -- the naming basis both engines
	// share so a manifest deploys identically on either. AWS allows 1-50
	// characters of alphanumeric plus - and _ (no spaces or dots).
	GuardrailName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockguardrailv1alpha1.AwsBedrockGuardrailIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata
	locals.GuardrailName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBedrockGuardrail.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
