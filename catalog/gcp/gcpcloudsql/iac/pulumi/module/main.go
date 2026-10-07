package module

import (
	"github.com/pkg/errors"
	gcpcloudsqlv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudsql/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpCloudSql kind.
func Resources(ctx *pulumi.Context, iacInput *gcpcloudsqlv1alpha1.GcpCloudSqlIacInput) error {
	locals := initializeLocals(iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if _, err := databaseInstance(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create cloud sql instance")
	}

	return nil
}
