package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcppscserviceattachment/iac/pulumi/module"
	gcppscserviceattachmentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcppscserviceattachment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcppscserviceattachmentv1alpha1.GcpPscServiceAttachmentStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
