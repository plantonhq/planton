// Package main provides the Pulumi program entrypoint for AWS Network Load Balancer deployment.
// Binary releases are gzip-compressed to reduce download size.
package main

import (
	"github.com/pkg/errors"
	awsnlbv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsnetworkloadbalancer/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsnetworkloadbalancer/v1/iac/pulumi/module"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &awsnlbv1.AwsNetworkLoadBalancerStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
