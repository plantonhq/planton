package module

import (
	"strconv"

	awsapprunnerobservabilityconfigurationv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsapprunnerobservabilityconfiguration/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set applied to the configuration.
type Locals struct {
	AwsAppRunnerObservabilityConfiguration *awsapprunnerobservabilityconfigurationv1alpha1.AwsAppRunnerObservabilityConfiguration
	AwsTags                                map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsapprunnerobservabilityconfigurationv1alpha1.AwsAppRunnerObservabilityConfigurationIacInput) *Locals {
	locals := &Locals{}
	locals.AwsAppRunnerObservabilityConfiguration = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsAppRunnerObservabilityConfiguration.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsAppRunnerObservabilityConfiguration.Metadata.Org,
		awstagkeys.Environment:  locals.AwsAppRunnerObservabilityConfiguration.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsAppRunnerObservabilityConfiguration.String(),
		awstagkeys.ResourceId:   locals.AwsAppRunnerObservabilityConfiguration.Metadata.Id,
	}

	return locals
}
