// Package main provides the Pulumi program entrypoint for Azure AKS Cluster deployment.
// Binary releases are gzip-compressed to reduce download size.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/azure/azureakscluster/iac/pulumi/module"
	azureaksclusterv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureakscluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &azureaksclusterv1alpha1.AzureAksClusterIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
