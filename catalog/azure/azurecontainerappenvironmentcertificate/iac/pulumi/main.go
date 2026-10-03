package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/azure/azurecontainerappenvironmentcertificate/iac/pulumi/module"
	azurecontainerappenvironmentcertificatev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappenvironmentcertificate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &azurecontainerappenvironmentcertificatev1alpha1.AzureContainerAppEnvironmentCertificateIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
