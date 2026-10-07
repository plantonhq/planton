package module

import (
	"github.com/pkg/errors"
	gcppubsubtopiciammemberv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcppubsubtopiciammember/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcppubsubtopiciammemberv1alpha1.GcpPubSubTopicIamMemberIacInput) error {
	// No project of its own: the grant names its topic by path.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, "")
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := iamMember(ctx, iacInput.Target, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create pubsub topic IAM member")
	}

	return nil
}
