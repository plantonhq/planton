package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/loadbalancers"
)

// backends creates all backend server pools defined in the spec.
//
// Each backend is a named group of servers with its own health check,
// load-balancing algorithm, and connection settings. The returned map
// is keyed by backend name for frontend→backend resolution.
func backends(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
	lb *loadbalancers.LoadBalancer,
) (map[string]*loadbalancers.Backend, error) {
	backendMap := make(map[string]*loadbalancers.Backend, len(locals.ScalewayLoadBalancer.Spec.Backends))

	for _, backendSpec := range locals.ScalewayLoadBalancer.Spec.Backends {
		args := &loadbalancers.BackendArgs{
			LbId:            lb.ID(),
			Name:            pulumi.String(backendSpec.Name),
			ForwardPort:     pulumi.Int(int(backendSpec.ForwardPort)),
			ForwardProtocol: pulumi.String(backendSpec.ForwardProtocol),
			ServerIps:       pulumi.ToStringArray(backendSpec.ServerIps),
		}

		// Load-balancing algorithm.
		if backendSpec.ForwardPortAlgorithm != "" {
			args.ForwardPortAlgorithm = pulumi.String(backendSpec.ForwardPortAlgorithm)
		}

		// Sticky sessions.
		if backendSpec.StickySessions != "" {
			args.StickySessions = pulumi.String(backendSpec.StickySessions)
		}
		if backendSpec.StickySessionsCookieName != "" {
			args.StickySessionsCookieName = pulumi.String(backendSpec.StickySessionsCookieName)
		}

		// Timeouts.
		if backendSpec.TimeoutConnect != "" {
			args.TimeoutConnect = pulumi.String(backendSpec.TimeoutConnect)
		}
		if backendSpec.TimeoutServer != "" {
			args.TimeoutServer = pulumi.String(backendSpec.TimeoutServer)
		}

		// Marked-down action.
		if backendSpec.OnMarkedDownAction != "" {
			args.OnMarkedDownAction = pulumi.String(backendSpec.OnMarkedDownAction)
		}

		// SSL bridging.
		if backendSpec.SslBridging {
			args.SslBridging = pulumi.Bool(true)
		}

		// PROXY protocol.
		if backendSpec.ProxyProtocol != "" {
			args.ProxyProtocol = pulumi.String(backendSpec.ProxyProtocol)
		}

		// Health check configuration.
		if hc := backendSpec.HealthCheck; hc != nil {
			if hc.CheckDelay != "" {
				args.HealthCheckDelay = pulumi.String(hc.CheckDelay)
			}
			if hc.CheckTimeout != "" {
				args.HealthCheckTimeout = pulumi.String(hc.CheckTimeout)
			}
			if hc.CheckMaxRetries > 0 {
				args.HealthCheckMaxRetries = pulumi.Int(int(hc.CheckMaxRetries))
			}
			if hc.Port > 0 {
				args.HealthCheckPort = pulumi.Int(int(hc.Port))
			}

			switch hc.Type {
			case "http":
				uri := hc.Uri
				if uri == "" {
					uri = "/"
				}
				code := hc.ExpectedCode
				if code == 0 {
					code = 200
				}
				args.HealthCheckHttp = &loadbalancers.BackendHealthCheckHttpArgs{
					Uri:  pulumi.String(uri),
					Code: pulumi.Int(int(code)),
				}
			case "https":
				uri := hc.Uri
				if uri == "" {
					uri = "/"
				}
				code := hc.ExpectedCode
				if code == 0 {
					code = 200
				}
				args.HealthCheckHttps = &loadbalancers.BackendHealthCheckHttpsArgs{
					Uri:  pulumi.String(uri),
					Code: pulumi.Int(int(code)),
				}
			default:
				// "tcp" or unspecified -- use TCP health check (the default).
				args.HealthCheckTcp = &loadbalancers.BackendHealthCheckTcpArgs{}
			}
		}

		resourceName := fmt.Sprintf("backend-%s", backendSpec.Name)
		createdBackend, err := loadbalancers.NewBackend(
			ctx,
			resourceName,
			args,
			pulumi.Provider(scalewayProvider),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create backend %q", backendSpec.Name)
		}

		backendMap[backendSpec.Name] = createdBackend
	}

	return backendMap, nil
}
