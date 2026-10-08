package module

import (
	"github.com/pkg/errors"
	gcpserviceaccountiammemberv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpserviceaccountiammember/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpserviceaccountiammemberv1alpha1.GcpServiceAccountIamMemberIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// No project of its own: the grant names its service account by path.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, "")
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := iamMember(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create service account IAM member")
	}

	return nil
}
