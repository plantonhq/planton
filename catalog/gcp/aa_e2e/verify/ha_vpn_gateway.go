package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// haVpnGatewayVerifier probes the HA VPN gateway and its Cloud Router
// through the compute API by the gateway_self_link and router_self_link
// outputs, and asserts the two public interface addresses the modules
// exported are the ones Google assigned. 404 on both is the destroyed shape.
type haVpnGatewayVerifier struct{}

func (v *haVpnGatewayVerifier) IDOutputKey() string { return "gateway_self_link" }

// regionalLocator extracts (region, name) from a regional self link of the
// form .../projects/{p}/regions/{region}/{collection}/{name}.
func regionalLocator(selfLink, collection string) (string, string, error) {
	parts := strings.Split(selfLink, "/")
	for i := 0; i < len(parts)-3; i++ {
		if parts[i] == "regions" && parts[i+2] == collection {
			return parts[i+1], parts[i+3], nil
		}
	}
	return "", "", errors.Errorf("cannot parse region/%s from self link %q", collection, selfLink)
}

func (v *haVpnGatewayVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	region, gatewayName, err := regionalLocator(outputs["gateway_self_link"], "vpnGateways")
	if err != nil {
		return errors.Wrap(err, "after deploy")
	}
	gateway, err := svc.Compute.VpnGateways.Get(svc.Project, region, gatewayName).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "ha vpn gateway %s in %s not found after deploy", gatewayName, region)
	}
	if len(gateway.VpnInterfaces) != 2 {
		return errors.Errorf("ha vpn gateway %s has %d interfaces after deploy, want 2", gatewayName, len(gateway.VpnInterfaces))
	}
	// The exported addresses must be the ones Google assigned, picked by
	// interface id on both engines.
	for _, iface := range gateway.VpnInterfaces {
		key := "interface_0_ip_address"
		if iface.Id == 1 {
			key = "interface_1_ip_address"
		}
		if want := outputs[key]; want != "" && want != iface.IpAddress {
			return errors.Errorf("ha vpn gateway %s interface %d has address %q, want the %s output %q", gatewayName, iface.Id, iface.IpAddress, key, want)
		}
	}

	routerRegion, routerName, err := routerLocator(outputs["router_self_link"])
	if err != nil {
		return errors.Wrap(err, "after deploy")
	}
	router, err := svc.Compute.Routers.Get(svc.Project, routerRegion, routerName).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "router %s in %s not found after deploy", routerName, routerRegion)
	}
	if name := outputs["router_name"]; name != "" && router.Name != name {
		return errors.Errorf("router resolved to name %q, want the router_name output %q", router.Name, name)
	}
	if router.Bgp == nil {
		return errors.Errorf("router %s has no bgp block after deploy; an HA VPN router must speak BGP", routerName)
	}
	return nil
}

func (v *haVpnGatewayVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	region, gatewayName, err := regionalLocator(outputs["gateway_self_link"], "vpnGateways")
	if err != nil {
		return nil
	}
	if _, err := svc.Compute.VpnGateways.Get(svc.Project, region, gatewayName).Context(ctx).Do(); err == nil {
		return errors.Errorf("ha vpn gateway %s still exists after destroy", gatewayName)
	} else if !isGone(err) {
		return errors.Wrapf(err, "unexpected error probing ha vpn gateway %s after destroy", gatewayName)
	}
	routerRegion, routerName, err := routerLocator(outputs["router_self_link"])
	if err != nil {
		return nil
	}
	if _, err := svc.Compute.Routers.Get(svc.Project, routerRegion, routerName).Context(ctx).Do(); err == nil {
		return errors.Errorf("router %s still exists after destroy", routerName)
	} else if !isGone(err) {
		return errors.Wrapf(err, "unexpected error probing router %s after destroy", routerName)
	}
	return nil
}

// isGone reports whether a compute API error is the destroyed shape: 404, or
// 403 when the identity loses visibility of a deleted resource's parent.
func isGone(err error) bool {
	var apiErr *googleapi.Error
	return errors.As(err, &apiErr) && (apiErr.Code == 404 || apiErr.Code == 403)
}
