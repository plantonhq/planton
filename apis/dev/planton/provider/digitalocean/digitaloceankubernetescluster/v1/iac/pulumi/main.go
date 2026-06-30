package main

import (
	"github.com/pkg/errors"
	digitaloceankubernetesclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceankubernetescluster/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceankubernetescluster/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &digitaloceankubernetesclusterv1.DigitalOceanKubernetesClusterStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
