package verify

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// haVpnConnectionVerifier probes every tunnel of the connection through the
// compute vpnTunnels API and every BGP session through the router status
// API. Tunnels reach ESTABLISHED and sessions reach Established only once
// the other side exists (the fixture chain deploys the reverse connection
// before the instance under test), so both are polled for a bounded time.
// The list outputs flatten to indexed keys (tunnel_names.0, ...), which is
// how the verifier walks them.
type haVpnConnectionVerifier struct{}

func (v *haVpnConnectionVerifier) IDOutputKey() string { return "tunnel_names.0" }

// indexed collects the values of a flattened list output in index order.
func indexed(outputs map[string]string, key string) []string {
	var values []string
	for i := 0; ; i++ {
		value, ok := outputs[fmt.Sprintf("%s.%d", key, i)]
		if !ok {
			return values
		}
		values = append(values, value)
	}
}

func (v *haVpnConnectionVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	tunnelLinks := indexed(outputs, "tunnel_self_links")
	peerNames := indexed(outputs, "bgp_peer_names")
	if len(tunnelLinks) == 0 || len(peerNames) != len(tunnelLinks) {
		return errors.Errorf("expected matching tunnel_self_links and bgp_peer_names outputs after deploy, got %d and %d", len(tunnelLinks), len(peerNames))
	}
	// The router shares the gateway's region; its name is its own output.
	region, _, err := regionalLocator(outputs["gateway_self_link"], "vpnGateways")
	if err != nil {
		return errors.Wrap(err, "after deploy")
	}
	routerName := outputs["router_name"]
	if routerName == "" {
		return errors.New("router_name output missing after deploy")
	}

	deadline := time.Now().Add(5 * time.Minute)
	for {
		pending, err := v.pending(ctx, svc, region, routerName, tunnelLinks, peerNames)
		if err != nil {
			return err
		}
		if pending == "" {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.Errorf("after deploy: %s", pending)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(15 * time.Second):
		}
	}
}

// pending returns a description of the first tunnel or session not yet up,
// or "" when every tunnel is ESTABLISHED and every BGP peer is Established.
func (v *haVpnConnectionVerifier) pending(ctx context.Context, svc *Services, region, routerName string, tunnelLinks, peerNames []string) (string, error) {
	for _, link := range tunnelLinks {
		tunnelRegion, tunnelName, err := regionalLocator(link, "vpnTunnels")
		if err != nil {
			return "", errors.Wrap(err, "after deploy")
		}
		tunnel, err := svc.Compute.VpnTunnels.Get(svc.Project, tunnelRegion, tunnelName).Context(ctx).Do()
		if err != nil {
			return "", errors.Wrapf(err, "vpn tunnel %s in %s not found after deploy", tunnelName, tunnelRegion)
		}
		if tunnel.Status != "ESTABLISHED" {
			return fmt.Sprintf("tunnel %s is %s (%s), want ESTABLISHED", tunnelName, tunnel.Status, tunnel.DetailedStatus), nil
		}
	}

	status, err := svc.Compute.Routers.GetRouterStatus(svc.Project, region, routerName).Context(ctx).Do()
	if err != nil {
		return "", errors.Wrapf(err, "router %s status not readable after deploy", routerName)
	}
	states := map[string]string{}
	if status.Result != nil {
		for _, peer := range status.Result.BgpPeerStatus {
			if peer != nil {
				states[peer.Name] = peer.State
			}
		}
	}
	for _, name := range peerNames {
		state, ok := states[name]
		if !ok {
			return fmt.Sprintf("router %s reports no BGP peer named %s", routerName, name), nil
		}
		if state != "Established" {
			return fmt.Sprintf("BGP peer %s on router %s is %s, want Established", name, routerName, state), nil
		}
	}
	return "", nil
}

func (v *haVpnConnectionVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	for _, link := range indexed(outputs, "tunnel_self_links") {
		region, tunnelName, err := regionalLocator(link, "vpnTunnels")
		if err != nil {
			continue
		}
		if _, err := svc.Compute.VpnTunnels.Get(svc.Project, region, tunnelName).Context(ctx).Do(); err == nil {
			return errors.Errorf("vpn tunnel %s still exists after destroy", tunnelName)
		} else if !isGone(err) {
			return errors.Wrapf(err, "unexpected error probing vpn tunnel %s after destroy", tunnelName)
		}
	}
	// The router outlives the connection (it belongs to the gateway); its
	// interfaces and peers for this connection must be gone from it.
	region, _, err := regionalLocator(outputs["gateway_self_link"], "vpnGateways")
	if err != nil {
		return nil
	}
	routerName := outputs["router_name"]
	router, err := svc.Compute.Routers.Get(svc.Project, region, routerName).Context(ctx).Do()
	if err != nil {
		// The gateway (and its router) may have been torn down after the
		// connection; a router that is gone has no peers left to report.
		return nil
	}
	for _, name := range indexed(outputs, "bgp_peer_names") {
		for _, peer := range router.BgpPeers {
			if peer != nil && peer.Name == name {
				return errors.Errorf("BGP peer %s still exists on router %s after destroy", name, routerName)
			}
		}
	}
	for _, name := range indexed(outputs, "router_interface_names") {
		for _, iface := range router.Interfaces {
			if iface != nil && iface.Name == name {
				return errors.Errorf("router interface %s still exists on router %s after destroy", name, routerName)
			}
		}
	}
	return nil
}
