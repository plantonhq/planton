package main

import (
	"github.com/pkg/errors"
	ociapplicationloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociapplicationloadbalancer/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociapplicationloadbalancer/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &ociapplicationloadbalancerv1.OciApplicationLoadBalancerStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
