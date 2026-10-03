package module

import (
	"github.com/pkg/errors"
	gcpgcsbucketv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgcsbucket/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpgcsbucketv1alpha1.GcpGcsBucketIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	createdBucket, err := gcsBucket(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create gcs bucket")
	}

	// Structural companions: folders, managed folders, and Pub/Sub
	// notification configs — all children of the bucket.
	if err := bucketCompanions(ctx, locals, gcpProvider, createdBucket); err != nil {
		return errors.Wrap(err, "failed to create bucket companions")
	}

	return nil
}
