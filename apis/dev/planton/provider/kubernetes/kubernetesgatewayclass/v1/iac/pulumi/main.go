package main

import (
	"github.com/pkg/errors"
	kubernetesgatewayclassv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgatewayclass/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgatewayclass/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &kubernetesgatewayclassv1.KubernetesGatewayClassStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
