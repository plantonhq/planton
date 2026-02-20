package module

import (
	"fmt"
	"strings"

	ociapplicationloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ociapplicationloadbalancer/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createBackendSets(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, lb *loadbalancer.LoadBalancer) ([]*loadbalancer.BackendSet, error) {
	spec := locals.OciApplicationLoadBalancer.Spec
	var createdSets []*loadbalancer.BackendSet

	for _, bsSpec := range spec.BackendSets {
		createdSet, err := createBackendSet(ctx, locals, provider, lb, bsSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to create backend set %s: %w", bsSpec.Name, err)
		}
		createdSets = append(createdSets, createdSet)

		for _, beSpec := range bsSpec.Backends {
			if err := createBackend(ctx, provider, lb, createdSet, bsSpec.Name, beSpec); err != nil {
				return nil, fmt.Errorf("failed to create backend %s:%d in set %s: %w",
					beSpec.IpAddress, beSpec.Port, bsSpec.Name, err)
			}
		}
	}

	return createdSets, nil
}

func createBackendSet(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, lb *loadbalancer.LoadBalancer, bsSpec *ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_BackendSet) (*loadbalancer.BackendSet, error) {
	policyMap := map[ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_BackendSet_Policy]string{
		ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_BackendSet_round_robin:       "ROUND_ROBIN",
		ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_BackendSet_least_connections: "LEAST_CONNECTIONS",
		ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_BackendSet_ip_hash:           "IP_HASH",
	}

	args := &loadbalancer.BackendSetArgs{
		LoadBalancerId: lb.ID(),
		Name:           pulumi.String(bsSpec.Name),
		Policy:         pulumi.String(policyMap[bsSpec.Policy]),
		HealthChecker:  buildHealthChecker(bsSpec.HealthChecker),
	}

	if bsSpec.BackendMaxConnections > 0 {
		args.BackendMaxConnections = pulumi.IntPtr(int(bsSpec.BackendMaxConnections))
	}

	if bsSpec.SslConfiguration != nil {
		args.SslConfiguration = buildBackendSetSslConfiguration(bsSpec.SslConfiguration)
	}

	if lbCookie := bsSpec.GetLbCookieSessionPersistence(); lbCookie != nil {
		args.LbCookieSessionPersistenceConfiguration = buildLbCookieSessionPersistence(lbCookie)
	}
	if appCookie := bsSpec.GetAppCookieSessionPersistence(); appCookie != nil {
		args.SessionPersistenceConfiguration = buildSessionPersistence(appCookie)
	}

	return loadbalancer.NewBackendSet(ctx, bsSpec.Name, args,
		pulumiOciOpt(provider), pulumi.Parent(lb))
}

func createBackend(ctx *pulumi.Context, provider *oci.Provider, lb *loadbalancer.LoadBalancer, bs *loadbalancer.BackendSet, bsName string, beSpec *ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_Backend) error {
	resourceName := fmt.Sprintf("%s-%s-%d", bsName, beSpec.IpAddress, beSpec.Port)

	args := &loadbalancer.BackendArgs{
		LoadBalancerId: lb.ID(),
		BackendsetName: bs.Name,
		IpAddress:      pulumi.String(beSpec.IpAddress),
		Port:           pulumi.Int(int(beSpec.Port)),
	}

	if beSpec.Weight > 0 {
		args.Weight = pulumi.IntPtr(int(beSpec.Weight))
	}
	if beSpec.Backup {
		args.Backup = pulumi.BoolPtr(true)
	}
	if beSpec.Drain {
		args.Drain = pulumi.BoolPtr(true)
	}
	if beSpec.Offline {
		args.Offline = pulumi.BoolPtr(true)
	}
	if beSpec.MaxConnections > 0 {
		args.MaxConnections = pulumi.IntPtr(int(beSpec.MaxConnections))
	}

	_, err := loadbalancer.NewBackend(ctx, resourceName, args,
		pulumiOciOpt(provider), pulumi.Parent(bs))
	return err
}

func buildHealthChecker(hc *ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_HealthChecker) loadbalancer.BackendSetHealthCheckerArgs {
	protocolMap := map[ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_HealthChecker_Protocol]string{
		ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_HealthChecker_http: "HTTP",
		ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_HealthChecker_tcp:  "TCP",
	}

	args := loadbalancer.BackendSetHealthCheckerArgs{
		Protocol: pulumi.String(protocolMap[hc.Protocol]),
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
	if hc.IntervalMs > 0 {
		args.IntervalMs = pulumi.IntPtr(int(hc.IntervalMs))
	}
	if hc.TimeoutInMillis > 0 {
		args.TimeoutInMillis = pulumi.IntPtr(int(hc.TimeoutInMillis))
	}
	if hc.Retries > 0 {
		args.Retries = pulumi.IntPtr(int(hc.Retries))
	}
	if hc.IsForcePlainText {
		args.IsForcePlainText = pulumi.BoolPtr(true)
	}
	return args
}

func buildBackendSetSslConfiguration(ssl *ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_SslConfiguration) *loadbalancer.BackendSetSslConfigurationArgs {
	args := &loadbalancer.BackendSetSslConfigurationArgs{}
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
	return args
}

func buildLbCookieSessionPersistence(cfg *ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_LbCookieSessionPersistenceConfig) *loadbalancer.BackendSetLbCookieSessionPersistenceConfigurationArgs {
	args := &loadbalancer.BackendSetLbCookieSessionPersistenceConfigurationArgs{}
	if cfg.CookieName != "" {
		args.CookieName = pulumi.StringPtr(cfg.CookieName)
	}
	if cfg.DisableFallback {
		args.DisableFallback = pulumi.BoolPtr(true)
	}
	if cfg.Domain != "" {
		args.Domain = pulumi.StringPtr(cfg.Domain)
	}
	if cfg.IsHttpOnly {
		args.IsHttpOnly = pulumi.BoolPtr(true)
	}
	if cfg.IsSecure {
		args.IsSecure = pulumi.BoolPtr(true)
	}
	if cfg.MaxAgeInSeconds > 0 {
		args.MaxAgeInSeconds = pulumi.IntPtr(int(cfg.MaxAgeInSeconds))
	}
	if cfg.Path != "" {
		args.Path = pulumi.StringPtr(cfg.Path)
	}
	return args
}

func buildSessionPersistence(cfg *ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_SessionPersistenceConfig) *loadbalancer.BackendSetSessionPersistenceConfigurationArgs {
	args := &loadbalancer.BackendSetSessionPersistenceConfigurationArgs{
		CookieName: pulumi.String(cfg.CookieName),
	}
	if cfg.DisableFallback {
		args.DisableFallback = pulumi.BoolPtr(true)
	}
	return args
}

func toUpperPolicy(p string) string {
	return strings.ToUpper(p)
}
