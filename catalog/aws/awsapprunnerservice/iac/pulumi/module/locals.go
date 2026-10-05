package module

import (
	"strconv"

	awsapprunnerservicev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsapprunnerservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set applied to the service.
type Locals struct {
	AwsAppRunnerService *awsapprunnerservicev1alpha1.AwsAppRunnerService
	AwsTags             map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsapprunnerservicev1alpha1.AwsAppRunnerServiceIacInput) *Locals {
	locals := &Locals{}
	locals.AwsAppRunnerService = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsAppRunnerService.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsAppRunnerService.Metadata.Org,
		awstagkeys.Environment:  locals.AwsAppRunnerService.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsAppRunnerService.String(),
		awstagkeys.ResourceId:   locals.AwsAppRunnerService.Metadata.Id,
	}

	return locals
}
