package module

import (
	awss3objectsetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awss3objectset/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsS3ObjectSet *awss3objectsetv1.AwsS3ObjectSet
	Labels         map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awss3objectsetv1.AwsS3ObjectSetStackInput) *Locals {
	locals := &Locals{}

	locals.AwsS3ObjectSet = stackInput.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsS3ObjectSet.Metadata.Org,
		"planton.org/environment":   locals.AwsS3ObjectSet.Metadata.Env,
		"planton.org/resource-kind": "AwsS3ObjectSet",
		"planton.org/resource-id":   locals.AwsS3ObjectSet.Metadata.Id,
	}

	return locals
}
