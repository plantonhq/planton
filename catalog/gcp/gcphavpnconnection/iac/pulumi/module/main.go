package module

import (
	"github.com/pkg/errors"
	gcphavpnconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphavpnconnection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions one site connection on an existing HA VPN gateway:
// the external VPN gateway (when the peer is an external device), then one
// tunnel, one Cloud Router interface, and one BGP peer per tunnels[] entry.
func Resources(ctx *pulumi.Context, stackInput *gcphavpnconnectionv1alpha1.GcpHaVpnConnectionStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	createdExternalGateway, err := externalGateway(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create external vpn gateway")
	}

	if err := tunnels(ctx, locals, gcpProvider, createdExternalGateway); err != nil {
		return errors.Wrap(err, "failed to create vpn tunnels")
	}

	return nil
}
