package module

import (
	"github.com/pkg/errors"
	ocidynamicroutinggatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocidynamicroutinggateway/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates the creation of the DRG and all sub-resources.
//
// Creation order reflects the dependency chain:
//  1. DRG (primary resource)
//  2. Route distributions (depend on DRG only)
//  3. Route tables (may reference distributions for import)
//  4. Attachments (may reference route tables and distributions)
//  5. Distribution statements (may reference attachments via match criteria)
//  6. Static route rules (reference attachments as next hop)
func Resources(ctx *pulumi.Context, stackInput *ocidynamicroutinggatewayv1.OciDynamicRoutingGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	createdDrg, err := createDrg(ctx, locals, ociProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create drg")
	}

	distMap, err := createRouteDistributions(ctx, locals, ociProvider, createdDrg)
	if err != nil {
		return errors.Wrap(err, "failed to create route distributions")
	}

	rtMap, err := createRouteTables(ctx, locals, ociProvider, createdDrg, distMap)
	if err != nil {
		return errors.Wrap(err, "failed to create route tables")
	}

	attachmentIdMap, err := createAttachments(ctx, locals, ociProvider, createdDrg, rtMap, distMap)
	if err != nil {
		return errors.Wrap(err, "failed to create drg attachments")
	}

	if err := createDistributionStatements(ctx, locals, ociProvider, distMap, attachmentIdMap); err != nil {
		return errors.Wrap(err, "failed to create distribution statements")
	}

	if err := createStaticRouteRules(ctx, locals, ociProvider, rtMap, attachmentIdMap); err != nil {
		return errors.Wrap(err, "failed to create static route rules")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
