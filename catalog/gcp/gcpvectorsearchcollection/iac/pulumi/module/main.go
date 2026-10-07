package module

import (
	"github.com/pkg/errors"
	gcpvectorsearchcollectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvectorsearchcollection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpvectorsearchcollectionv1alpha1.GcpVectorSearchCollectionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := collection(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create vector search collection")
	}

	return nil
}
