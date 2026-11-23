package aa_e2e

import (
	"context"
	"net"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// HostAddress is how a pod on a kind cluster reaches a server running on the
// machine that runs the cluster (a fake model the e2e process serves, for
// example). Pods dial IP, an IPv4 address: kind clusters are IPv4 by
// default, and Docker Desktop answers host.docker.internal with an IPv6
// address first, which such a pod cannot route to.
type HostAddress struct {
	// Source says how IP was found: host.docker.internal, or the kind
	// network's gateway.
	Source string
	IP     string
}

// commandOutput runs a command and returns its stdout; a variable so the
// fallback order is testable without Docker.
var commandOutput = func(ctx context.Context, name string, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, name, args...).Output()
	return string(out), err
}

// ResolveHostAddress finds the host address from inside kind cluster
// clusterName's control-plane node, in the order Stigmer's own kind smokes
// use: the IPv4 address of host.docker.internal when the node resolves it
// (Docker Desktop, where the kind network's gateway is inside Docker's VM,
// not on this machine), else the IPv4 gateway of the `kind` Docker network
// (Linux, where host.docker.internal is not defined and that gateway is
// this machine). A server reached this way must listen
// on every interface, never only on loopback.
func ResolveHostAddress(ctx context.Context, clusterName string) (HostAddress, error) {
	node := clusterName + "-control-plane"
	resolved, resolveErr := commandOutput(ctx, "docker", "exec", node, "getent", "ahostsv4", "host.docker.internal")
	if resolveErr == nil {
		if fields := strings.Fields(resolved); len(fields) > 0 && net.ParseIP(fields[0]).To4() != nil {
			return HostAddress{Source: "host.docker.internal", IP: fields[0]}, nil
		}
	}
	gateways, gatewayErr := commandOutput(ctx, "docker", "network", "inspect", "kind",
		"--format", "{{range .IPAM.Config}}{{.Gateway}} {{end}}")
	if gatewayErr == nil {
		for _, gateway := range strings.Fields(gateways) {
			if ip := net.ParseIP(gateway); ip != nil && ip.To4() != nil {
				return HostAddress{Source: "the kind network's gateway", IP: gateway}, nil
			}
		}
	}
	return HostAddress{}, errors.Errorf(
		"no way to reach this machine from kind cluster %q: host.docker.internal does not resolve on node %s (%v), and the kind Docker network has no IPv4 gateway (%v)",
		clusterName, node, errOrOutput(resolveErr, resolved), errOrOutput(gatewayErr, gateways))
}

func errOrOutput(err error, out string) string {
	if err != nil {
		return err.Error()
	}
	if strings.TrimSpace(out) == "" {
		return "empty answer"
	}
	return "answered " + strings.TrimSpace(out)
}
