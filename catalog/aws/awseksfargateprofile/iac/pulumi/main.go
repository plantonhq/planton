// Package main provides the Pulumi program entrypoint for AWS EKS Fargate Profile deployment.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awseksfargateprofile/iac/pulumi/module"
	awseksfargateprofilev1alpha1 "github.com/plantonhq/planton/catalog/aws/awseksfargateprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awseksfargateprofilev1alpha1.AwsEksFargateProfileIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
