package module

import (
	"strconv"
	"strings"

	gcphavpnconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphavpnconnection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the derivations both engines share -- the resolved gateway
// trio, the external gateway's name, per-tunnel session and key names, and
// the platform label set every labeled companion carries.
type Locals struct {
	GcpHaVpnConnection *gcphavpnconnectionv1alpha1.GcpHaVpnConnection

	// Gateway, Router, and Region are the resolved references to the
	// GcpHaVpnGateway this connection rides: the gateway's self link, the
	// router's name, and their shared region.
	Gateway string
	Router  string
	Region  string

	// Project is the spec's project_id, empty for the provider's default.
	Project string

	// IsExternalPeer is true when peer.external_gateway is set: this
	// connection creates the external VPN gateway resource and every tunnel
	// names one of its interfaces. False means peer.gcp_gateway.
	IsExternalPeer bool

	// ExternalGatewayName is peer.external_gateway.name, or metadata.name
	// when the spec leaves it empty.
	ExternalGatewayName string

	// SessionName returns the router interface and BGP peer name for a
	// tunnel: bgp_session.name, or the tunnel's name when empty.
	SessionNames []string

	// Md5KeyNames returns the MD5 key name for a tunnel's session: the
	// declared name, or `<tunnel>-md5` when the key is set without one.
	// Empty when the session has no MD5 key.
	Md5KeyNames []string

	// GenerateSharedSecret is true exactly when the module mints the
	// pre-shared key: spec.shared_secret is empty and at least one tunnel
	// declares no key of its own. GenerateMd5AuthenticationKey is the same
	// for the MD5 key over the sessions that declare an
	// md5_authentication_key block. The Terraform twin computes the same
	// predicates in locals.tf.
	GenerateSharedSecret         bool
	GenerateMd5AuthenticationKey bool

	// PlatformLabels is the attribution label set merged into every
	// labeled companion (tunnels, external gateway) after the user's own
	// labels -- identical merge order to the Terraform module.
	PlatformLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcphavpnconnectionv1alpha1.GcpHaVpnConnectionIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	locals := &Locals{
		GcpHaVpnConnection: target,
		Gateway:            spec.Gateway.GetValue(),
		Router:             spec.Router.GetValue(),
		Region:             spec.Region.GetValue(),
		Project:            spec.ProjectId.GetValue(),
		IsExternalPeer:     spec.Peer.GetExternalGateway() != nil,
	}

	if locals.IsExternalPeer {
		locals.ExternalGatewayName = spec.Peer.ExternalGateway.Name
		if locals.ExternalGatewayName == "" {
			locals.ExternalGatewayName = target.Metadata.Name
		}
	}

	for _, tunnel := range spec.Tunnels {
		sessionName := tunnel.BgpSession.GetName()
		if sessionName == "" {
			sessionName = tunnel.Name
		}
		locals.SessionNames = append(locals.SessionNames, sessionName)

		md5KeyName := ""
		if key := tunnel.BgpSession.GetMd5AuthenticationKey(); key != nil {
			md5KeyName = key.Name
			if md5KeyName == "" {
				md5KeyName = tunnel.Name + "-md5"
			}
		}
		locals.Md5KeyNames = append(locals.Md5KeyNames, md5KeyName)

		if tunnel.SharedSecret == "" {
			locals.GenerateSharedSecret = true
		}
		if key := tunnel.BgpSession.GetMd5AuthenticationKey(); key != nil && key.Key == "" {
			locals.GenerateMd5AuthenticationKey = true
		}
	}
	// A declared connection-level key serves every tunnel (session) without
	// its own, so nothing is minted.
	if spec.SharedSecret.GetValue() != "" {
		locals.GenerateSharedSecret = false
	}
	if spec.Md5AuthenticationKey.GetValue() != "" {
		locals.GenerateMd5AuthenticationKey = false
	}

	locals.PlatformLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: target.Metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_GcpHaVpnConnection.String()),
	}
	if target.Metadata.Org != "" {
		locals.PlatformLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.PlatformLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}
	if target.Metadata.Id != "" {
		locals.PlatformLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}

	return locals
}

// mergeLabels returns the user's labels with the platform labels layered on
// top (platform wins on key conflicts).
func (l *Locals) mergeLabels(userLabels map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range userLabels {
		merged[key] = value
	}
	for key, value := range l.PlatformLabels {
		merged[key] = value
	}
	return merged
}
