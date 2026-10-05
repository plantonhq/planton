package module

import (
	"strconv"

	awsapprunnerautoscalingconfigurationv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsapprunnerautoscalingconfiguration/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set applied to the configuration.
type Locals struct {
	AwsAppRunnerAutoScalingConfiguration *awsapprunnerautoscalingconfigurationv1alpha1.AwsAppRunnerAutoScalingConfiguration
	AwsTags                              map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsapprunnerautoscalingconfigurationv1alpha1.AwsAppRunnerAutoScalingConfigurationIacInput) *Locals {
	locals := &Locals{}
	locals.AwsAppRunnerAutoScalingConfiguration = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsAppRunnerAutoScalingConfiguration.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsAppRunnerAutoScalingConfiguration.Metadata.Org,
		awstagkeys.Environment:  locals.AwsAppRunnerAutoScalingConfiguration.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsAppRunnerAutoScalingConfiguration.String(),
		awstagkeys.ResourceId:   locals.AwsAppRunnerAutoScalingConfiguration.Metadata.Id,
	}

	return locals
}
