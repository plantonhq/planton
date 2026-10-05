package module

import (
	"github.com/pkg/errors"
	gcphavpnconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphavpnconnection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions one site connection on an existing HA VPN gateway:
// the keys it mints when the spec declares none, the external VPN gateway
// (when the peer is an external device), then one tunnel, one Cloud Router
// interface, and one BGP peer per tunnels[] entry.
func Resources(ctx *pulumi.Context, iacInput *gcphavpnconnectionv1alpha1.GcpHaVpnConnectionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	keys, err := mintKeys(ctx, locals)
	if err != nil {
		return errors.Wrap(err, "failed to resolve the connection keys")
	}

	createdExternalGateway, err := externalGateway(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create external vpn gateway")
	}

	if err := tunnels(ctx, locals, gcpProvider, keys, createdExternalGateway); err != nil {
		return errors.Wrap(err, "failed to create vpn tunnels")
	}

	exportMintedKeys(ctx, keys)

	return nil
}
