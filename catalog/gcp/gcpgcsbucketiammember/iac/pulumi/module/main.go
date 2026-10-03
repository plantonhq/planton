package module

import (
	"github.com/pkg/errors"
	gcpgcsbucketiammemberv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgcsbucketiammember/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpgcsbucketiammemberv1alpha1.GcpGcsBucketIamMemberIacInput) error {
	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := iamMember(ctx, iacInput.Target, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create gcs bucket IAM member")
	}

	return nil
}
