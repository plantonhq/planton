package module

import (
	"github.com/pkg/errors"
	gcpbigqueryreservationgroupv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbigqueryreservationgroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpbigqueryreservationgroupv1alpha1.GcpBigQueryReservationGroupStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := reservationGroup(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create bigquery reservation group")
	}

	return nil
}
