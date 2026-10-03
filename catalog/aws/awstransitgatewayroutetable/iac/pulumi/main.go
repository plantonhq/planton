// Package main provides the Pulumi program entrypoint for AWS Transit
// Gateway route table deployment.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awstransitgatewayroutetable/iac/pulumi/module"
	awstgwrtv1 "github.com/plantonhq/planton/catalog/aws/awstransitgatewayroutetable/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awstgwrtv1.AwsTransitGatewayRouteTableIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
