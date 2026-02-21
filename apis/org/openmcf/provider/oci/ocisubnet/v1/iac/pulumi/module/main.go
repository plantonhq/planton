package module

import (
	"github.com/pkg/errors"
	ocisubnetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocisubnet/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *ocisubnetv1.OciSubnetStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	var routeTableId pulumi.StringOutput

	if len(locals.OciSubnet.Spec.RouteRules) > 0 {
		createdRouteTable, err := routeTable(ctx, locals, ociProvider)
		if err != nil {
			return errors.Wrap(err, "failed to create route table")
		}
		routeTableId = createdRouteTable.ID().ToStringOutput()
	}

	if err := subnet(ctx, locals, ociProvider, routeTableId); err != nil {
		return errors.Wrap(err, "failed to create subnet")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
