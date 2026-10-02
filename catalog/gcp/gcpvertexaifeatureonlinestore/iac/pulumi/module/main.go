package module

import (
	"github.com/pkg/errors"
	gcpvertexaifeatureonlinestorev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaifeatureonlinestore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpvertexaifeatureonlinestorev1alpha1.GcpVertexAiFeatureOnlineStoreStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := featureOnlineStore(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create vertex ai feature online store")
	}

	return nil
}
