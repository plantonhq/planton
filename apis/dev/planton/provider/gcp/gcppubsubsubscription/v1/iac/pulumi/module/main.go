package module

import (
	"github.com/pkg/errors"
	gcppubsubsubscriptionv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcppubsubsubscription/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcppubsubsubscriptionv1.GcpPubSubSubscriptionStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := subscription(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create pubsub subscription")
	}

	return nil
}
