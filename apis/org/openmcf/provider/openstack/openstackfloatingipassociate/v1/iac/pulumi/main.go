package main

import (
	openstackfloatingipassociatev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackfloatingipassociate/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackfloatingipassociate/v1/iac/pulumi/module"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &openstackfloatingipassociatev1.OpenStackFloatingIpAssociateStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}

		return module.Resources(ctx, stackInput)
	})
}
