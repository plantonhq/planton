package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// vpc provisions the Civo network and exports its outputs.
func vpc(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.Network, error) {

	// 1. Build resource arguments from proto fields.
	networkArgs := &civo.NetworkArgs{
		Label:  pulumi.String(locals.CivoVpc.Spec.NetworkName),
		Region: pulumi.String(locals.CivoVpc.Spec.Region),
	}

	// Optional: IPv4 CIDR block (if not specified, Civo auto-allocates)
	if locals.CivoVpc.Spec.IpRangeCidr != "" {
		networkArgs.CidrV4 = pulumi.String(locals.CivoVpc.Spec.IpRangeCidr)
	}

	// Note: The 'is_default_for_region' field is not supported by Pulumi Civo SDK v2.
	// The NetworkArgs struct doesn't have a 'Default' field as of v2.4.8.
	// This feature may need to be set via Civo API directly or wait for provider support.
	if locals.CivoVpc.Spec.IsDefaultForRegion {
		ctx.Log.Warn(fmt.Sprintf(
			"Network '%s' has 'is_default_for_region' set to true, but this is not supported by "+
				"Pulumi Civo SDK v2.4.8. The network will be created without being set as default. "+
				"To set a network as default, use the Civo CLI: 'civo network default <network-id>'",
			locals.CivoVpc.Spec.NetworkName,
		), nil)
	}

	// Optional: Description
	// Note: The Civo provider's Network resource doesn't currently expose a description field.
	// This field is recorded in metadata but not applied to the Civo resource.
	if locals.CivoVpc.Spec.Description != "" {
		ctx.Log.Info(fmt.Sprintf(
			"Description specified for network '%s': '%s'. "+
				"Note: The Civo Network provider doesn't currently support description field. "+
				"This is recorded in Planton metadata only.",
			locals.CivoVpc.Spec.NetworkName,
			locals.CivoVpc.Spec.Description,
		), nil)
	}

	// 2. Create the network.
	createdNetwork, err := civo.NewNetwork(
		ctx,
		"network",
		networkArgs,
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create civo network")
	}

	// 3. Export stack outputs.
	ctx.Export(OpNetworkId, createdNetwork.ID())
	ctx.Export(OpCidrBlock, createdNetwork.CidrV4)
	// Note: created_at_rfc3339 is not exposed by the Civo provider as an attribute.
	// The network exists after creation, but the timestamp is not available via the provider.
	// This can be added manually if needed using time.Now() at creation time.

	return createdNetwork, nil
}
