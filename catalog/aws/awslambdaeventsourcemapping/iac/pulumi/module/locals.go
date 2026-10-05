package module

import (
	"strconv"

	awslambdaeventsourcemappingv1alpha1 "github.com/plantonhq/planton/catalog/aws/awslambdaeventsourcemapping/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsLambdaEventSourceMapping *awslambdaeventsourcemappingv1alpha1.AwsLambdaEventSourceMapping

	// MappingName is metadata.name -- the Planton identity for this node.
	// AWS assigns the runtime UUID separately (exported as uuid).
	MappingName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awslambdaeventsourcemappingv1alpha1.AwsLambdaEventSourceMappingIacInput) *Locals {
	locals := &Locals{}
	locals.AwsLambdaEventSourceMapping = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.MappingName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsLambdaEventSourceMapping.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
