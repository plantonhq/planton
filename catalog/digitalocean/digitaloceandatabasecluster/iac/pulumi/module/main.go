package module

import (
	"github.com/pkg/errors"
	digitaloceandatabaseclusterv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandatabasecluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the DigitalOcean VPC module style.
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceandatabaseclusterv1alpha1.DigitalOceanDatabaseClusterIacInput,
) error {
	// 1. Prepare locals (metadata, labels, credentials, etc.).
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a DigitalOcean provider from the supplied credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Create the database cluster.
	if _, err := cluster(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create database cluster")
	}

	return nil
}
