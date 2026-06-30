package main

import (
	"github.com/pkg/errors"
	kubernetesopenfgav1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesopenfga/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesopenfga/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &kubernetesopenfgav1.KubernetesOpenFgaStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
