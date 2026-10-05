package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awsecsservicev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecsservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals defines local variables used throughout our Pulumi module.
// This includes the AWS ECS Service resource definition (AwsEcsService)
// and a map of AWS tags to apply to resources.
type Locals struct {
	AwsEcsService *awsecsservicev1alpha1.AwsEcsService
	AwsTags       map[string]string
}

// initializeLocals pulls values from the IaC input (AwsEcsServiceIacInput)
// and populates the Locals struct. Similar to Terraform "locals" concept.
func initializeLocals(ctx *pulumi.Context, iacInput *awsecsservicev1alpha1.AwsEcsServiceIacInput) *Locals {
	locals := &Locals{
		AwsEcsService: iacInput.Target,
	}

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsEcsService.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsEcsService.Metadata.Org,
		awstagkeys.Environment:  locals.AwsEcsService.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEcsService.String(),
		awstagkeys.ResourceId:   locals.AwsEcsService.Metadata.Id,
	}

	return locals
}
