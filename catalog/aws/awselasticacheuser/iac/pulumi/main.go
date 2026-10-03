// Package main provides the Pulumi program entrypoint for AWS ElastiCache user deployment.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awselasticacheuser/iac/pulumi/module"
	awselasticacheuserv1alpha1 "github.com/plantonhq/planton/catalog/aws/awselasticacheuser/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awselasticacheuserv1alpha1.AwsElasticacheUserIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
