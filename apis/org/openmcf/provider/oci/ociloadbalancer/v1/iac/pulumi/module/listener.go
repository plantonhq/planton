package module

import (
	"fmt"

	ociloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ociloadbalancer/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createListeners(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	lb *loadbalancer.LoadBalancer,
	backendSets []*loadbalancer.BackendSet,
	hostnames []*loadbalancer.Hostname,
	ruleSets []*loadbalancer.RuleSet,
	certificates []*loadbalancer.Certificate,
) error {
	spec := locals.OciLoadBalancer.Spec
	deps := collectDeps(backendSets, hostnames, ruleSets, certificates)

	for _, lnSpec := range spec.Listeners {
		if err := createListener(ctx, provider, lb, lnSpec, deps); err != nil {
			return fmt.Errorf("failed to create listener %s: %w", lnSpec.Name, err)
		}
	}
	return nil
}

func createListener(
	ctx *pulumi.Context,
	provider *oci.Provider,
	lb *loadbalancer.LoadBalancer,
	lnSpec *ociloadbalancerv1.OciLoadBalancerSpec_Listener,
	deps []pulumi.Resource,
) error {
	protocolMap := map[ociloadbalancerv1.OciLoadBalancerSpec_Listener_Protocol]string{
		ociloadbalancerv1.OciLoadBalancerSpec_Listener_http:  "HTTP",
		ociloadbalancerv1.OciLoadBalancerSpec_Listener_http2: "HTTP2",
		ociloadbalancerv1.OciLoadBalancerSpec_Listener_tcp:   "TCP",
		ociloadbalancerv1.OciLoadBalancerSpec_Listener_grpc:  "GRPC",
	}

	args := &loadbalancer.ListenerArgs{
		LoadBalancerId:        lb.ID(),
		Name:                  pulumi.String(lnSpec.Name),
		Port:                  pulumi.Int(int(lnSpec.Port)),
		Protocol:              pulumi.String(protocolMap[lnSpec.Protocol]),
		DefaultBackendSetName: pulumi.String(lnSpec.DefaultBackendSetName),
	}

	if lnSpec.SslConfiguration != nil {
		args.SslConfiguration = buildListenerSslConfiguration(lnSpec.SslConfiguration)
	}

	if lnSpec.ConnectionConfiguration != nil {
		args.ConnectionConfiguration = buildConnectionConfiguration(lnSpec.ConnectionConfiguration)
	}

	if len(lnSpec.HostnameNames) > 0 {
		args.HostnameNames = pulumi.ToStringArray(lnSpec.HostnameNames)
	}

	if len(lnSpec.RuleSetNames) > 0 {
		args.RuleSetNames = pulumi.ToStringArray(lnSpec.RuleSetNames)
	}

	if lnSpec.RoutingPolicyName != "" {
		args.RoutingPolicyName = pulumi.StringPtr(lnSpec.RoutingPolicyName)
	}

	opts := []pulumi.ResourceOption{
		pulumiOciOpt(provider),
		pulumi.Parent(lb),
	}
	if len(deps) > 0 {
		opts = append(opts, pulumiDependsOn(deps...))
	}

	_, err := loadbalancer.NewListener(ctx, lnSpec.Name, args, opts...)
	return err
}

func buildListenerSslConfiguration(ssl *ociloadbalancerv1.OciLoadBalancerSpec_SslConfiguration) *loadbalancer.ListenerSslConfigurationArgs {
	args := &loadbalancer.ListenerSslConfigurationArgs{}
	if len(ssl.CertificateIds) > 0 {
		args.CertificateIds = pulumi.ToStringArray(ssl.CertificateIds)
	}
	if ssl.CertificateName != "" {
		args.CertificateName = pulumi.StringPtr(ssl.CertificateName)
	}
	if ssl.CipherSuiteName != "" {
		args.CipherSuiteName = pulumi.StringPtr(ssl.CipherSuiteName)
	}
	if len(ssl.Protocols) > 0 {
		args.Protocols = pulumi.ToStringArray(ssl.Protocols)
	}
	if ssl.ServerOrderPreference != "" {
		args.ServerOrderPreference = pulumi.StringPtr(ssl.ServerOrderPreference)
	}
	if len(ssl.TrustedCertificateAuthorityIds) > 0 {
		args.TrustedCertificateAuthorityIds = pulumi.ToStringArray(ssl.TrustedCertificateAuthorityIds)
	}
	if ssl.VerifyDepth > 0 {
		args.VerifyDepth = pulumi.IntPtr(int(ssl.VerifyDepth))
	}
	if ssl.VerifyPeerCertificate {
		args.VerifyPeerCertificate = pulumi.BoolPtr(true)
	}
	if ssl.HasSessionResumption {
		args.HasSessionResumption = pulumi.BoolPtr(true)
	}
	return args
}

func buildConnectionConfiguration(cc *ociloadbalancerv1.OciLoadBalancerSpec_ConnectionConfiguration) *loadbalancer.ListenerConnectionConfigurationArgs {
	args := &loadbalancer.ListenerConnectionConfigurationArgs{
		IdleTimeoutInSeconds: pulumi.String(fmt.Sprintf("%d", cc.IdleTimeoutInSeconds)),
	}
	if cc.BackendTcpProxyProtocolVersion > 0 {
		args.BackendTcpProxyProtocolVersion = pulumi.IntPtr(int(cc.BackendTcpProxyProtocolVersion))
	}
	return args
}
