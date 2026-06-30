package main

import (
	openstacksecuritygroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstacksecuritygroup/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstacksecuritygroup/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &openstacksecuritygroupv1.OpenStackSecurityGroupStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}

		return module.Resources(ctx, stackInput)
	})
}
