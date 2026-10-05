package module

import (
	"fmt"
	"strconv"

	awss3directorybucketv1alpha1 "github.com/plantonhq/planton/catalog/aws/awss3directorybucket/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awss3directorybucketv1alpha1.AwsS3DirectoryBucket
	Spec   *awss3directorybucketv1alpha1.AwsS3DirectoryBucketSpec

	// The FULL bucket name AWS mandates
	// ("{base}--{zone_id}--x-s3") - derived so the name and the
	// location can never disagree; matches the Terraform module.
	BucketName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awss3directorybucketv1alpha1.AwsS3DirectoryBucketIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	locals.BucketName = fmt.Sprintf("%s--%s--x-s3", metadata.Name, locals.Spec.ZoneId)

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsS3DirectoryBucket.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
