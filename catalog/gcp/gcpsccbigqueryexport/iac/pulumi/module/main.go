package module

import (
	"github.com/pkg/errors"
	gcpsccbigqueryexportv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccbigqueryexport/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpsccbigqueryexportv1alpha1.GcpSccBigQueryExportIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := bigQueryExport(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create the Security Command Center BigQuery export")
	}

	return nil
}
