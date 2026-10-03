package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/azure/azuremachinelearningonlinedeployment/iac/pulumi/module"
	azuremachinelearningonlinedeploymentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningonlinedeployment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &azuremachinelearningonlinedeploymentv1alpha1.AzureMachineLearningOnlineDeploymentIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
