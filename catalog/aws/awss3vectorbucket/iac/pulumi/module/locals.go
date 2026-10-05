package module

import (
	"strconv"

	awss3vectorbucketv1alpha1 "github.com/plantonhq/planton/catalog/aws/awss3vectorbucket/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awss3vectorbucketv1alpha1.AwsS3VectorBucket
	Spec   *awss3vectorbucketv1alpha1.AwsS3VectorBucketSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awss3vectorbucketv1alpha1.AwsS3VectorBucketIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsS3VectorBucket.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
