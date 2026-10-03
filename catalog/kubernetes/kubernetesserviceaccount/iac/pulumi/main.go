// Pulumi entrypoint for the KubernetesServiceAccount component.
// Loads the IaC input and delegates all resource creation to the module package.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/kubernetes/kubernetesserviceaccount/iac/pulumi/module"
	kubernetesserviceaccountv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesserviceaccount/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &kubernetesserviceaccountv1alpha1.KubernetesServiceAccountIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
