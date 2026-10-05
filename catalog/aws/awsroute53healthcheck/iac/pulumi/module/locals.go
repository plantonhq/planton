package module

import (
	"strconv"

	awsroute53healthcheckv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsroute53healthcheck/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set (the Name tag is what the Route 53 console displays as the health
// check's name).
type Locals struct {
	AwsRoute53HealthCheck *awsroute53healthcheckv1alpha1.AwsRoute53HealthCheck
	AwsTags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsroute53healthcheckv1alpha1.AwsRoute53HealthCheckIacInput) *Locals {
	locals := &Locals{}
	locals.AwsRoute53HealthCheck = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsRoute53HealthCheck.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsRoute53HealthCheck.Metadata.Org,
		awstagkeys.Environment:  locals.AwsRoute53HealthCheck.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsRoute53HealthCheck.String(),
		awstagkeys.ResourceId:   locals.AwsRoute53HealthCheck.Metadata.Id,
	}

	return locals
}
