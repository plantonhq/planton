package module

import (
	"github.com/pkg/errors"
	scalewayobjectbucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayobjectbucket/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions a Scaleway Object
// Storage bucket with optional versioning, lifecycle rules, and CORS.
//
// This is a standalone resource (not composite): it wraps a single
// scaleway_object_bucket resource. All configuration (versioning,
// lifecycle, CORS) is inline on the bucket resource itself.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewayobjectbucketv1.ScalewayObjectBucketStackInput,
) error {
	// 1. Prepare locals (metadata, labels, resolved references).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create the bucket and export outputs.
	if err := bucket(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create object bucket")
	}

	return nil
}
