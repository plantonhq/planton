// Package main provides the Pulumi program entrypoint for the AWS Planton
// Runner appliance: a standing, outbound-only runner on ECS Fargate that
// executes deploy and cloud operations from inside the VPC.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awsplantonrunner/iac/pulumi/module"
	awsplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awsplantonrunnerv1alpha1.AwsPlantonRunnerIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
