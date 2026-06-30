package module

import (
	"github.com/pkg/errors"
	gcpcloudsqlv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudsql/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpCloudSql component.
func Resources(ctx *pulumi.Context, stackInput *gcpcloudsqlv1.GcpCloudSqlStackInput) error {
	locals := initializeLocals(stackInput)

	// Set up the GCP provider from the supplied credential.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// Create the Cloud SQL database instance.
	createdInstance, err := databaseInstance(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloud-sql instance")
	}

	// Export stack outputs
	ctx.Export(OpInstanceName, createdInstance.Name)
	ctx.Export(OpConnectionName, createdInstance.ConnectionName)
	ctx.Export(OpSelfLink, createdInstance.SelfLink)

	// Export IP addresses using direct fields
	ctx.Export(OpPublicIp, createdInstance.PublicIpAddress)
	ctx.Export(OpPrivateIp, createdInstance.PrivateIpAddress)

	return nil
}
