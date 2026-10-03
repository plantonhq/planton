// Package main provides the Pulumi program entrypoint for AWS EKS Node Group deployment.
// Auto-release test: Multi-provider Pulumi change (AWS component).
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awseksnodegroup/iac/pulumi/module"
	awseksnodegroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awseksnodegroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awseksnodegroupv1alpha1.AwsEksNodeGroupIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
