package module

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	hetznercloudloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznercloudloadbalancer/v1"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func loadBalancer(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	spec := locals.HetznerCloudLoadBalancer.Spec

	algorithmType := "round_robin"
	if spec.Algorithm != hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_algorithm_unspecified {
		algorithmType = spec.Algorithm.String()
	}

	createdLb, err := hcloud.NewLoadBalancer(
		ctx,
		"load-balancer",
		&hcloud.LoadBalancerArgs{
			Name:             pulumi.String(locals.HetznerCloudLoadBalancer.Metadata.Name),
			LoadBalancerType: pulumi.String(spec.LoadBalancerType),
			Location:         pulumi.StringPtr(spec.Location),
			Labels:           pulumi.ToStringMap(locals.Labels),
			DeleteProtection: pulumi.Bool(spec.DeleteProtection),
			Algorithm: &hcloud.LoadBalancerAlgorithmArgs{
				Type: pulumi.StringPtr(algorithmType),
			},
		},
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud load balancer")
	}

	// The Pulumi hcloud SDK uses StringInput for LoadBalancerService.LoadBalancerId
	// but IntInput for LoadBalancerTarget and LoadBalancerNetwork. Prepare both
	// forms of the load balancer ID.
	lbIdStr := createdLb.ID().ToStringOutput()
	lbIdInt := createdLb.ID().ApplyT(func(id pulumi.ID) (int, error) {
		return strconv.Atoi(string(id))
	}).(pulumi.IntOutput)

	// Optional network attachment (create before targets so private IP routing
	// is available if targets use use_private_ip).
	var createdNetwork *hcloud.LoadBalancerNetwork
	if spec.Network != nil {
		var networkErr error
		createdNetwork, networkErr = createNetworkAttachment(ctx, spec.Network, lbIdInt, hcloudProvider)
		if networkErr != nil {
			return errors.Wrap(networkErr, "failed to create load balancer network attachment")
		}
	}

	if err := createServices(ctx, spec.Services, lbIdStr, hcloudProvider); err != nil {
		return errors.Wrap(err, "failed to create load balancer services")
	}

	if err := createTargets(ctx, spec, lbIdInt, hcloudProvider, createdNetwork); err != nil {
		return errors.Wrap(err, "failed to create load balancer targets")
	}

	ctx.Export(OpLoadBalancerId, createdLb.ID())
	ctx.Export(OpIpv4Address, createdLb.Ipv4)
	ctx.Export(OpIpv6Address, createdLb.Ipv6)

	return nil
}

// createServices creates an hcloud_load_balancer_service for each service
// entry. Services are keyed by listen_port per CG02.
func createServices(
	ctx *pulumi.Context,
	services []*hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_Service,
	lbIdStr pulumi.StringOutput,
	hcloudProvider *hcloud.Provider,
) error {
	for _, svc := range services {
		listenPort := effectiveListenPort(svc)
		destPort := effectiveDestinationPort(svc, listenPort)

		serviceArgs := &hcloud.LoadBalancerServiceArgs{
			LoadBalancerId:  lbIdStr,
			Protocol:        pulumi.String(svc.Protocol.String()),
			ListenPort:      pulumi.IntPtr(int(listenPort)),
			DestinationPort: pulumi.IntPtr(int(destPort)),
			Proxyprotocol:   pulumi.BoolPtr(svc.Proxyprotocol),
		}

		if svc.Http != nil && svc.Protocol != hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_tcp {
			httpArgs, err := buildHttpConfig(svc.Http)
			if err != nil {
				return errors.Wrapf(err, "failed to build http config for service on port %d", listenPort)
			}
			serviceArgs.Http = httpArgs
		}

		if svc.HealthCheck != nil {
			hcArgs := buildHealthCheck(svc.HealthCheck, svc.Protocol, destPort)
			serviceArgs.HealthCheck = hcArgs
		}

		resourceName := fmt.Sprintf("service-%d", listenPort)
		if _, err := hcloud.NewLoadBalancerService(
			ctx,
			resourceName,
			serviceArgs,
			pulumi.Provider(hcloudProvider),
		); err != nil {
			return errors.Wrapf(err, "failed to create service on port %d", listenPort)
		}
	}

	return nil
}

// createTargets creates hcloud_load_balancer_target resources for all three
// target types. Targets that use use_private_ip depend on the network
// attachment if present.
func createTargets(
	ctx *pulumi.Context,
	spec *hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec,
	lbIdInt pulumi.IntOutput,
	hcloudProvider *hcloud.Provider,
	createdNetwork *hcloud.LoadBalancerNetwork,
) error {
	for _, target := range spec.ServerTargets {
		serverId, err := strconv.Atoi(target.ServerId.GetValue())
		if err != nil {
			return errors.Wrapf(err, "failed to parse server_id %q as integer",
				target.ServerId.GetValue())
		}

		resourceName := fmt.Sprintf("target-server-%s", target.ServerId.GetValue())
		opts := []pulumi.ResourceOption{pulumi.Provider(hcloudProvider)}
		if target.UsePrivateIp && createdNetwork != nil {
			opts = append(opts, pulumi.DependsOn([]pulumi.Resource{createdNetwork}))
		}

		if _, err := hcloud.NewLoadBalancerTarget(
			ctx,
			resourceName,
			&hcloud.LoadBalancerTargetArgs{
				LoadBalancerId: lbIdInt,
				Type:           pulumi.String("server"),
				ServerId:       pulumi.IntPtr(serverId),
				UsePrivateIp:   pulumi.BoolPtr(target.UsePrivateIp),
			},
			opts...,
		); err != nil {
			return errors.Wrapf(err, "failed to create server target %s", target.ServerId.GetValue())
		}
	}

	for _, target := range spec.LabelSelectorTargets {
		resourceName := fmt.Sprintf("target-label-%s", sanitizeSelector(target.Selector))
		opts := []pulumi.ResourceOption{pulumi.Provider(hcloudProvider)}
		if target.UsePrivateIp && createdNetwork != nil {
			opts = append(opts, pulumi.DependsOn([]pulumi.Resource{createdNetwork}))
		}

		if _, err := hcloud.NewLoadBalancerTarget(
			ctx,
			resourceName,
			&hcloud.LoadBalancerTargetArgs{
				LoadBalancerId: lbIdInt,
				Type:           pulumi.String("label_selector"),
				LabelSelector:  pulumi.StringPtr(target.Selector),
				UsePrivateIp:   pulumi.BoolPtr(target.UsePrivateIp),
			},
			opts...,
		); err != nil {
			return errors.Wrapf(err, "failed to create label selector target %q", target.Selector)
		}
	}

	for _, target := range spec.IpTargets {
		resourceName := fmt.Sprintf("target-ip-%s", sanitizeIp(target.Ip))

		if _, err := hcloud.NewLoadBalancerTarget(
			ctx,
			resourceName,
			&hcloud.LoadBalancerTargetArgs{
				LoadBalancerId: lbIdInt,
				Type:           pulumi.String("ip"),
				Ip:             pulumi.StringPtr(target.Ip),
			},
			pulumi.Provider(hcloudProvider),
		); err != nil {
			return errors.Wrapf(err, "failed to create ip target %s", target.Ip)
		}
	}

	return nil
}

// createNetworkAttachment attaches the load balancer to a private network.
func createNetworkAttachment(
	ctx *pulumi.Context,
	netCfg *hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_NetworkAttachment,
	lbIdInt pulumi.IntOutput,
	hcloudProvider *hcloud.Provider,
) (*hcloud.LoadBalancerNetwork, error) {
	networkId, err := strconv.Atoi(netCfg.NetworkId.GetValue())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse network_id %q as integer",
			netCfg.NetworkId.GetValue())
	}

	networkArgs := &hcloud.LoadBalancerNetworkArgs{
		LoadBalancerId: lbIdInt,
		NetworkId:      pulumi.IntPtr(networkId),
	}

	if netCfg.Ip != "" {
		networkArgs.Ip = pulumi.StringPtr(netCfg.Ip)
	}

	if netCfg.EnablePublicInterface != nil {
		networkArgs.EnablePublicInterface = pulumi.BoolPtr(*netCfg.EnablePublicInterface)
	} else {
		networkArgs.EnablePublicInterface = pulumi.BoolPtr(true)
	}

	created, err := hcloud.NewLoadBalancerNetwork(
		ctx,
		"network",
		networkArgs,
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create hetzner cloud load balancer network attachment")
	}

	return created, nil
}

// buildHttpConfig converts the proto HttpConfig into Pulumi LoadBalancerServiceHttpArgs.
func buildHttpConfig(
	httpCfg *hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_HttpConfig,
) (*hcloud.LoadBalancerServiceHttpArgs, error) {
	args := &hcloud.LoadBalancerServiceHttpArgs{
		StickySessions: pulumi.BoolPtr(httpCfg.StickySessions),
		RedirectHttp:   pulumi.BoolPtr(httpCfg.RedirectHttp),
	}

	if httpCfg.CookieName != "" {
		args.CookieName = pulumi.StringPtr(httpCfg.CookieName)
	}

	if httpCfg.CookieLifetime > 0 {
		args.CookieLifetime = pulumi.IntPtr(int(httpCfg.CookieLifetime))
	}

	if len(httpCfg.CertificateIds) > 0 {
		certIds := make([]pulumi.IntInput, 0, len(httpCfg.CertificateIds))
		for _, ref := range httpCfg.CertificateIds {
			id, err := strconv.Atoi(ref.GetValue())
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse certificate_id %q as integer",
					ref.GetValue())
			}
			certIds = append(certIds, pulumi.Int(id))
		}
		args.Certificates = pulumi.IntArray(certIds)
	}

	return args, nil
}

// buildHealthCheck converts the proto HealthCheck into Pulumi args, applying
// defaults for unset fields.
func buildHealthCheck(
	hc *hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_HealthCheck,
	svcProtocol hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_ServiceProtocol,
	destPort int32,
) *hcloud.LoadBalancerServiceHealthCheckArgs {
	protocol := defaultHealthCheckProtocol(hc.Protocol, svcProtocol)
	port := int(destPort)
	if hc.Port != nil {
		port = int(*hc.Port)
	}
	interval := 15
	if hc.Interval != nil {
		interval = int(*hc.Interval)
	}
	timeout := 10
	if hc.Timeout != nil {
		timeout = int(*hc.Timeout)
	}
	retries := 3
	if hc.Retries != nil {
		retries = int(*hc.Retries)
	}

	args := &hcloud.LoadBalancerServiceHealthCheckArgs{
		Protocol: pulumi.String(protocol),
		Port:     pulumi.Int(port),
		Interval: pulumi.Int(interval),
		Timeout:  pulumi.Int(timeout),
		Retries:  pulumi.Int(retries),
	}

	if hc.Http != nil {
		httpArgs := &hcloud.LoadBalancerServiceHealthCheckHttpArgs{}
		if hc.Http.Domain != "" {
			httpArgs.Domain = pulumi.StringPtr(hc.Http.Domain)
		}
		if hc.Http.Path != "" {
			httpArgs.Path = pulumi.StringPtr(hc.Http.Path)
		}
		if hc.Http.Response != "" {
			httpArgs.Response = pulumi.StringPtr(hc.Http.Response)
		}
		if hc.Http.Tls {
			httpArgs.Tls = pulumi.BoolPtr(true)
		}
		if len(hc.Http.StatusCodes) > 0 {
			httpArgs.StatusCodes = pulumi.ToStringArray(hc.Http.StatusCodes)
		}
		args.Http = httpArgs
	}

	return args
}

// effectiveListenPort returns the listen port for a service, applying
// protocol-specific defaults when the field is not set.
func effectiveListenPort(svc *hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_Service) int32 {
	if svc.ListenPort != nil {
		return *svc.ListenPort
	}
	switch svc.Protocol {
	case hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_http:
		return 80
	case hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_https:
		return 443
	default:
		return 0
	}
}

// effectiveDestinationPort returns the destination port for a service,
// defaulting to the listen port when not explicitly set.
func effectiveDestinationPort(svc *hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_Service, listenPort int32) int32 {
	if svc.DestinationPort != nil {
		return *svc.DestinationPort
	}
	return listenPort
}

// defaultHealthCheckProtocol returns the health check protocol string,
// defaulting to the service protocol when unspecified. HTTPS services default
// to "http" health checks because the LB terminates TLS and backends
// typically serve plain HTTP.
func defaultHealthCheckProtocol(
	hcProtocol hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_ServiceProtocol,
	svcProtocol hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_ServiceProtocol,
) string {
	if hcProtocol != hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_service_protocol_unspecified {
		return hcProtocol.String()
	}
	if svcProtocol == hetznercloudloadbalancerv1.HetznerCloudLoadBalancerSpec_https {
		return "http"
	}
	return svcProtocol.String()
}

// sanitizeSelector converts a Hetzner Cloud label selector into a Pulumi-safe
// resource name component by replacing special characters with hyphens.
func sanitizeSelector(selector string) string {
	r := strings.NewReplacer("=", "-", ",", "-", " ", "-")
	return r.Replace(selector)
}

// sanitizeIp converts an IP address into a Pulumi-safe resource name component.
func sanitizeIp(ip string) string {
	r := strings.NewReplacer(".", "-", ":", "-")
	return r.Replace(ip)
}
