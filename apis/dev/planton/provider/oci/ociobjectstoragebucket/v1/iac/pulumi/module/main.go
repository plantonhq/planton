package module

import (
	"github.com/pkg/errors"
	ociobjectstoragebucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociobjectstoragebucket/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *ociobjectstoragebucketv1.OciObjectStorageBucketStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	createdBucket, err := bucket(ctx, locals, ociProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create object storage bucket")
	}

	if err := lifecyclePolicy(ctx, locals, ociProvider, createdBucket); err != nil {
		return errors.Wrap(err, "failed to create lifecycle policy")
	}

	if err := replicationPolicies(ctx, locals, ociProvider, createdBucket); err != nil {
		return errors.Wrap(err, "failed to create replication policies")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
