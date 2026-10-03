package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awscloudfrontv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudfront/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds resolved input values and tag metadata for the Pulumi stack.
type Locals struct {
	AwsCloudFront *awscloudfrontv1alpha1.AwsCloudFront
	AwsTags       map[string]string
}

// initializeLocals prepares a Locals object by resolving IaC input and metadata-derived tags.
func initializeLocals(_ *pulumi.Context, iacInput *awscloudfrontv1alpha1.AwsCloudFrontIacInput) *Locals {
	locals := &Locals{}
	locals.AwsCloudFront = iacInput.Target

	metadata := iacInput.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	// CloudFront distributions have no AWS name -- metadata.name drives the
	// Name tag and consumers address the distribution through its
	// ID/ARN/domain outputs.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsCloudFront.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
