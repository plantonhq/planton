// Package main provides the Pulumi program entrypoint for the Azure
// Planton Runner appliance: a standing, outbound-only runner on Azure
// Container Apps that executes deploy and cloud operations from inside
// your network perimeter.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/azure/azureplantonrunner/iac/pulumi/module"
	azureplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &azureplantonrunnerv1alpha1.AzurePlantonRunnerIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
