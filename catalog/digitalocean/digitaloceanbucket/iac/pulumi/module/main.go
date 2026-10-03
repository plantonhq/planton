package module

import (
	"github.com/pkg/errors"
	digitaloceanbucketv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanbucket/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the pattern used in the VPC module.
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceanbucketv1alpha1.DigitalOceanBucketIacInput,
) error {
	// 1. Prepare locals (metadata, labels, credentials, etc.).
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a DigitalOcean provider from the supplied credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Create the bucket.
	if _, err := bucket(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create bucket")
	}

	return nil
}
