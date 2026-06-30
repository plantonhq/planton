package main

import (
	gcpcloudcomposerenvironmentv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudcomposerenvironment/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudcomposerenvironment/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpcloudcomposerenvironmentv1.GcpCloudComposerEnvironmentStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
