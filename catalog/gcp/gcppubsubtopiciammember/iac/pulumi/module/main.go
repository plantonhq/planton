package module

import (
	"github.com/pkg/errors"
	gcppubsubtopiciammemberv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcppubsubtopiciammember/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcppubsubtopiciammemberv1alpha1.GcpPubSubTopicIamMemberStackInput) error {
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := iamMember(ctx, stackInput.Target, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create pubsub topic IAM member")
	}

	return nil
}
