package module

import (
	"fmt"

	"github.com/pkg/errors"
	azuremssqlfailovergroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremssqlfailovergroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mssql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremssqlfailovergroupv1alpha1.AzureMssqlFailoverGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMssqlFailoverGroup.Spec

	partnerServers := mssql.FailoverGroupPartnerServerArray{}
	for _, id := range locals.PartnerServerIds {
		partnerServers = append(partnerServers, mssql.FailoverGroupPartnerServerArgs{
			Id: pulumi.String(id),
		})
	}

	rwPolicy := mssql.FailoverGroupReadWriteEndpointFailoverPolicyArgs{
		Mode: pulumi.String(locals.FailoverMode),
	}
	// grace_minutes is required for Automatic and rejected for Manual; the
	// spec CEL guarantees the pairing, so send it only for Automatic.
	if spec.ReadWriteEndpointFailoverPolicy.Mode == azuremssqlfailovergroupv1alpha1.AzureMssqlFailoverGroupFailoverMode_AUTOMATIC {
		rwPolicy.GraceMinutes = pulumi.Int(int(spec.ReadWriteEndpointFailoverPolicy.GraceMinutes))
	}

	fogArgs := &mssql.FailoverGroupArgs{
		Name:                            pulumi.String(spec.Name),
		ServerId:                        pulumi.String(locals.ServerId),
		PartnerServers:                  partnerServers,
		ReadWriteEndpointFailoverPolicy: rwPolicy,
		Tags:                            pulumi.ToStringMap(locals.AzureTags),
	}

	if len(locals.DatabaseIds) > 0 {
		fogArgs.Databases = pulumi.ToStringArray(locals.DatabaseIds)
	}
	// The provider sends Disabled for the read-only endpoint when this is
	// unset; only an explicit choice is sent so an unspecified spec deploys
	// identically on both engines.
	if spec.ReadonlyEndpointFailoverPolicyEnabled != nil {
		fogArgs.ReadonlyEndpointFailoverPolicyEnabled = pulumi.Bool(spec.GetReadonlyEndpointFailoverPolicyEnabled())
	}

	createdFog, err := mssql.NewFailoverGroup(ctx,
		spec.Name,
		fogArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create failover group %s", spec.Name)
	}

	ctx.Export(OpFailoverGroupId, createdFog.ID())
	ctx.Export(OpFailoverGroupName, createdFog.Name)

	// The listener FQDNs are DNS-derived from the group name (Azure does not
	// return them as resource attributes); compose them so downstream
	// connection strings can reference a single failover-following endpoint.
	ctx.Export(OpReadWriteListenerEndpoint, pulumi.String(fmt.Sprintf("%s.database.windows.net", spec.Name)))
	ctx.Export(OpReadOnlyListenerEndpoint, pulumi.String(fmt.Sprintf("%s.secondary.database.windows.net", spec.Name)))

	return nil
}
