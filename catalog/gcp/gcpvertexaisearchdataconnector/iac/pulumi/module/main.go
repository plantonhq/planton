package module

import (
	"github.com/pkg/errors"
	gcpvertexaisearchdataconnectorv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaisearchdataconnector/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpvertexaisearchdataconnectorv1alpha1.GcpVertexAiSearchDataConnectorStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := dataConnector(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create vertex ai search data connector")
	}

	return nil
}
