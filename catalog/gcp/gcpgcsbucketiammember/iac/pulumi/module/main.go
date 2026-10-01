package module

import (
	"github.com/pkg/errors"
	gcpgcsbucketiammemberv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgcsbucketiammember/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpgcsbucketiammemberv1alpha1.GcpGcsBucketIamMemberStackInput) error {
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := iamMember(ctx, stackInput.Target, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create gcs bucket IAM member")
	}

	return nil
}
