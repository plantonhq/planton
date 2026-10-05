package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awss3vectorbucket/iac/pulumi/module"
	awss3vectorbucketv1alpha1 "github.com/plantonhq/planton/catalog/aws/awss3vectorbucket/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awss3vectorbucketv1alpha1.AwsS3VectorBucketIacInput{}
		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}
		return module.Resources(ctx, iacInput)
	})
}
