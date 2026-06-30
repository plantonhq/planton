package module

import (
	awss3bucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awss3bucket/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsS3Bucket *awss3bucketv1.AwsS3Bucket
	Labels      map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awss3bucketv1.AwsS3BucketStackInput) *Locals {
	locals := &Locals{}

	locals.AwsS3Bucket = stackInput.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsS3Bucket.Metadata.Org,
		"planton.org/environment":   locals.AwsS3Bucket.Metadata.Env,
		"planton.org/resource-kind": "AwsS3Bucket",
		"planton.org/resource-id":   locals.AwsS3Bucket.Metadata.Id,
	}

	return locals
}
