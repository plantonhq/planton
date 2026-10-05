package module

import (
	"strconv"

	awsapprunnervpcconnectorv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsapprunnervpcconnector/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set applied to the VPC connector.
type Locals struct {
	AwsAppRunnerVpcConnector *awsapprunnervpcconnectorv1alpha1.AwsAppRunnerVpcConnector
	AwsTags                  map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsapprunnervpcconnectorv1alpha1.AwsAppRunnerVpcConnectorIacInput) *Locals {
	locals := &Locals{}
	locals.AwsAppRunnerVpcConnector = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsAppRunnerVpcConnector.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsAppRunnerVpcConnector.Metadata.Org,
		awstagkeys.Environment:  locals.AwsAppRunnerVpcConnector.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsAppRunnerVpcConnector.String(),
		awstagkeys.ResourceId:   locals.AwsAppRunnerVpcConnector.Metadata.Id,
	}

	return locals
}
