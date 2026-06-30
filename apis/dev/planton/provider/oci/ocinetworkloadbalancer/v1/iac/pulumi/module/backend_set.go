package module

import (
	"fmt"

	ocinetworkloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocinetworkloadbalancer/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/networkloadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var policyMap = map[ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_BackendSet_Policy]string{
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_BackendSet_five_tuple:  "FIVE_TUPLE",
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_BackendSet_three_tuple: "THREE_TUPLE",
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_BackendSet_two_tuple:   "TWO_TUPLE",
}

var healthCheckerProtocolMap = map[ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_HealthChecker_Protocol]string{
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_HealthChecker_http:  "HTTP",
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_HealthChecker_https: "HTTPS",
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_HealthChecker_tcp:   "TCP",
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_HealthChecker_udp:   "UDP",
	ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_HealthChecker_dns:   "DNS",
}

func createBackendSets(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	nlb *networkloadbalancer.NetworkLoadBalancer,
) ([]*networkloadbalancer.BackendSet, error) {
	spec := locals.OciNetworkLoadBalancer.Spec
	var createdSets []*networkloadbalancer.BackendSet

	for _, bsSpec := range spec.BackendSets {
		createdSet, err := createBackendSet(ctx, provider, nlb, bsSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to create backend set %s: %w", bsSpec.Name, err)
		}
		createdSets = append(createdSets, createdSet)

		for _, beSpec := range bsSpec.Backends {
			if err := createBackend(ctx, provider, nlb, createdSet, bsSpec.Name, beSpec); err != nil {
				return nil, fmt.Errorf("failed to create backend in set %s: %w", bsSpec.Name, err)
			}
		}
	}

	return createdSets, nil
}

func createBackendSet(
	ctx *pulumi.Context,
	provider *oci.Provider,
	nlb *networkloadbalancer.NetworkLoadBalancer,
	bsSpec *ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_BackendSet,
) (*networkloadbalancer.BackendSet, error) {
	args := &networkloadbalancer.BackendSetArgs{
		NetworkLoadBalancerId: nlb.ID(),
		Name:                  pulumi.StringPtr(bsSpec.Name),
		Policy:                pulumi.String(policyMap[bsSpec.Policy]),
		HealthChecker:         buildHealthChecker(bsSpec.HealthChecker),
	}

	if bsSpec.IsPreserveSource {
		args.IsPreserveSource = pulumi.BoolPtr(true)
	}
	if bsSpec.IsFailOpen {
		args.IsFailOpen = pulumi.BoolPtr(true)
	}
	if bsSpec.IsInstantFailoverEnabled {
		args.IsInstantFailoverEnabled = pulumi.BoolPtr(true)
	}
	if bsSpec.IsInstantFailoverTcpResetEnabled {
		args.IsInstantFailoverTcpResetEnabled = pulumi.BoolPtr(true)
	}
	if bsSpec.AreOperationallyActiveBackendsPreferred {
		args.AreOperationallyActiveBackendsPreferred = pulumi.BoolPtr(true)
	}
	if bsSpec.IpVersion != "" {
		args.IpVersion = pulumi.StringPtr(bsSpec.IpVersion)
	}

	return networkloadbalancer.NewBackendSet(ctx, bsSpec.Name, args,
		pulumiOciOpt(provider), pulumi.Parent(nlb))
}

func createBackend(
	ctx *pulumi.Context,
	provider *oci.Provider,
	nlb *networkloadbalancer.NetworkLoadBalancer,
	bs *networkloadbalancer.BackendSet,
	bsName string,
	beSpec *ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_Backend,
) error {
	resourceName := fmt.Sprintf("%s-%d", bsName, beSpec.Port)
	if beSpec.IpAddress != "" {
		resourceName = fmt.Sprintf("%s-%s-%d", bsName, beSpec.IpAddress, beSpec.Port)
	} else if beSpec.Name != "" {
		resourceName = fmt.Sprintf("%s-%s", bsName, beSpec.Name)
	}

	args := &networkloadbalancer.BackendArgs{
		NetworkLoadBalancerId: nlb.ID(),
		BackendSetName:        bs.Name,
		Port:                  pulumi.Int(int(beSpec.Port)),
	}

	if beSpec.IpAddress != "" {
		args.IpAddress = pulumi.StringPtr(beSpec.IpAddress)
	}
	if beSpec.TargetId != "" {
		args.TargetId = pulumi.StringPtr(beSpec.TargetId)
	}
	if beSpec.Weight > 0 {
		args.Weight = pulumi.IntPtr(int(beSpec.Weight))
	}
	if beSpec.IsBackup {
		args.IsBackup = pulumi.BoolPtr(true)
	}
	if beSpec.IsDrain {
		args.IsDrain = pulumi.BoolPtr(true)
	}
	if beSpec.IsOffline {
		args.IsOffline = pulumi.BoolPtr(true)
	}
	if beSpec.Name != "" {
		args.Name = pulumi.StringPtr(beSpec.Name)
	}

	_, err := networkloadbalancer.NewBackend(ctx, resourceName, args,
		pulumiOciOpt(provider), pulumi.Parent(bs))
	return err
}

func buildHealthChecker(hc *ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_HealthChecker) networkloadbalancer.BackendSetHealthCheckerArgs {
	args := networkloadbalancer.BackendSetHealthCheckerArgs{
		Protocol: pulumi.String(healthCheckerProtocolMap[hc.Protocol]),
	}

	if hc.Port > 0 {
		args.Port = pulumi.IntPtr(int(hc.Port))
	}
	if hc.UrlPath != "" {
		args.UrlPath = pulumi.StringPtr(hc.UrlPath)
	}
	if hc.ReturnCode > 0 {
		args.ReturnCode = pulumi.IntPtr(int(hc.ReturnCode))
	}
	if hc.ResponseBodyRegex != "" {
		args.ResponseBodyRegex = pulumi.StringPtr(hc.ResponseBodyRegex)
	}
	if hc.IntervalInMillis > 0 {
		args.IntervalInMillis = pulumi.IntPtr(int(hc.IntervalInMillis))
	}
	if hc.TimeoutInMillis > 0 {
		args.TimeoutInMillis = pulumi.IntPtr(int(hc.TimeoutInMillis))
	}
	if hc.Retries > 0 {
		args.Retries = pulumi.IntPtr(int(hc.Retries))
	}
	if hc.RequestData != "" {
		args.RequestData = pulumi.StringPtr(hc.RequestData)
	}
	if hc.ResponseData != "" {
		args.ResponseData = pulumi.StringPtr(hc.ResponseData)
	}

	if hc.DnsHealthCheck != nil {
		args.Dns = buildDnsHealthCheck(hc.DnsHealthCheck)
	}

	return args
}

func buildDnsHealthCheck(dns *ocinetworkloadbalancerv1.OciNetworkLoadBalancerSpec_DnsHealthCheck) *networkloadbalancer.BackendSetHealthCheckerDnsArgs {
	args := &networkloadbalancer.BackendSetHealthCheckerDnsArgs{
		DomainName: pulumi.String(dns.DomainName),
	}
	if dns.QueryClass != "" {
		args.QueryClass = pulumi.StringPtr(dns.QueryClass)
	}
	if dns.QueryType != "" {
		args.QueryType = pulumi.StringPtr(dns.QueryType)
	}
	if len(dns.Rcodes) > 0 {
		args.Rcodes = pulumi.ToStringArray(dns.Rcodes)
	}
	if dns.TransportProtocol != "" {
		args.TransportProtocol = pulumi.StringPtr(dns.TransportProtocol)
	}
	return args
}
