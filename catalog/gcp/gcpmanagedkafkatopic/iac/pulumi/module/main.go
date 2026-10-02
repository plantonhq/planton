package module

import (
	"github.com/pkg/errors"
	gcpmanagedkafkatopicv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmanagedkafkatopic/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpmanagedkafkatopicv1alpha1.GcpManagedKafkaTopicStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := topic(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create managed kafka topic")
	}

	return nil
}
