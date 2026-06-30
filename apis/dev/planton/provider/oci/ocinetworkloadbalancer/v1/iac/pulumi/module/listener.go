package module

import (
	"fmt"

	ocinetworkloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocinetworkloadbalancer/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/networkloadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var listenerProtocolMap = map[ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_Listener_Protocol]string{
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_Listener_tcp:         "TCP",
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_Listener_udp:         "UDP",
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_Listener_tcp_and_udp: "TCP_AND_UDP",
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_Listener_any:         "ANY",
}

func createListeners(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	nlb *networkloadbalancer.NetworkLoadBalancer,
	backendSets []*networkloadbalancer.BackendSet,
) error {
	spec := locals.OciNetworkLoadBalancer.Spec

	var deps []pulumi.Resource
	for _, bs := range backendSets {
		deps = append(deps, bs)
	}

	for _, lnSpec := range spec.Listeners {
		if err := createListener(ctx, provider, nlb, lnSpec, deps); err != nil {
			return fmt.Errorf("failed to create listener %s: %w", lnSpec.Name, err)
		}
	}
	return nil
}

func createListener(
	ctx *pulumi.Context,
	provider *oci.Provider,
	nlb *networkloadbalancer.NetworkLoadBalancer,
	lnSpec *ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_Listener,
	deps []pulumi.Resource,
) error {
	args := &networkloadbalancer.ListenerArgs{
		NetworkLoadBalancerId: nlb.ID(),
		Name:                  pulumi.StringPtr(lnSpec.Name),
		Port:                  pulumi.Int(int(lnSpec.Port)),
		Protocol:              pulumi.String(listenerProtocolMap[lnSpec.Protocol]),
		DefaultBackendSetName: pulumi.String(lnSpec.DefaultBackendSetName),
	}

	if lnSpec.IpVersion != "" {
		args.IpVersion = pulumi.StringPtr(lnSpec.IpVersion)
	}
	if lnSpec.IsPpv2Enabled {
		args.IsPpv2enabled = pulumi.BoolPtr(true)
	}
	if lnSpec.TcpIdleTimeout > 0 {
		args.TcpIdleTimeout = pulumi.IntPtr(int(lnSpec.TcpIdleTimeout))
	}
	if lnSpec.UdpIdleTimeout > 0 {
		args.UdpIdleTimeout = pulumi.IntPtr(int(lnSpec.UdpIdleTimeout))
	}
	if lnSpec.L3IpIdleTimeout > 0 {
		args.L3ipIdleTimeout = pulumi.IntPtr(int(lnSpec.L3IpIdleTimeout))
	}

	opts := []pulumi.ResourceOption{
		pulumiOciOpt(provider),
		pulumi.Parent(nlb),
	}
	if len(deps) > 0 {
		opts = append(opts, pulumi.DependsOn(deps))
	}

	_, err := networkloadbalancer.NewListener(ctx, lnSpec.Name, args, opts...)
	return err
}
