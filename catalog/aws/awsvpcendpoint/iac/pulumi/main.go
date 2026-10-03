// Package main provides the Pulumi program entrypoint for AWS VPC Endpoint deployment.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awsvpcendpoint/iac/pulumi/module"
	awsvpcendpointv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsvpcendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awsvpcendpointv1alpha1.AwsVpcEndpointIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
