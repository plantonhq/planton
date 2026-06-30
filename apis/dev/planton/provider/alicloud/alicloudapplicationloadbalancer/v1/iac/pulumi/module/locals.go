package module

import (
	"strings"

	alicloudapplicationloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudapplicationloadbalancer/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudApplicationLoadBalancer *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancer
	Tags                            map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudApplicationLoadBalancer = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudApplicationLoadBalancer.String()),
	}

	if target.Metadata.Id != "" {
		locals.Tags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Tags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Tags["environment"] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.Tags[k] = v
	}

	return locals
}

func addressType(spec *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerSpec) string {
	if spec.AddressType != nil {
		return *spec.AddressType
	}
	return "Internet"
}

func loadBalancerEdition(spec *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerSpec) string {
	if spec.LoadBalancerEdition != nil {
		return *spec.LoadBalancerEdition
	}
	return "Standard"
}

func serverGroupProtocol(sg *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerServerGroup) string {
	if sg.Protocol != nil {
		return *sg.Protocol
	}
	return "HTTP"
}

func serverGroupScheduler(sg *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerServerGroup) string {
	if sg.Scheduler != nil {
		return *sg.Scheduler
	}
	return "Wrr"
}

func healthCheckProtocol(hc *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerHealthCheckConfig) string {
	if hc.HealthCheckProtocol != nil {
		return *hc.HealthCheckProtocol
	}
	return "HTTP"
}

func healthCheckMethod(hc *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerHealthCheckConfig) string {
	if hc.HealthCheckMethod != nil {
		return *hc.HealthCheckMethod
	}
	return "HEAD"
}

func healthCheckConnectPort(hc *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerHealthCheckConfig) int {
	if hc.HealthCheckConnectPort != nil {
		return int(*hc.HealthCheckConnectPort)
	}
	return 0
}

func healthCheckInterval(hc *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerHealthCheckConfig) int {
	if hc.HealthCheckInterval != nil {
		return int(*hc.HealthCheckInterval)
	}
	return 2
}

func healthCheckTimeout(hc *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerHealthCheckConfig) int {
	if hc.HealthCheckTimeout != nil {
		return int(*hc.HealthCheckTimeout)
	}
	return 5
}

func healthyThreshold(hc *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerHealthCheckConfig) int {
	if hc.HealthyThreshold != nil {
		return int(*hc.HealthyThreshold)
	}
	return 3
}

func unhealthyThreshold(hc *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerHealthCheckConfig) int {
	if hc.UnhealthyThreshold != nil {
		return int(*hc.UnhealthyThreshold)
	}
	return 3
}

func listenerGzipEnabled(l *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerListener) bool {
	if l.GzipEnabled != nil {
		return *l.GzipEnabled
	}
	return true
}

func listenerHttp2Enabled(l *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerListener) bool {
	if l.Http2Enabled != nil {
		return *l.Http2Enabled
	}
	return true
}

func listenerIdleTimeout(l *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerListener) int {
	if l.IdleTimeout != nil {
		return int(*l.IdleTimeout)
	}
	return 60
}

func listenerRequestTimeout(l *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerListener) int {
	if l.RequestTimeout != nil {
		return int(*l.RequestTimeout)
	}
	return 60
}

func cookieTimeout(sc *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerStickySessionConfig) int {
	if sc.CookieTimeout != nil {
		return int(*sc.CookieTimeout)
	}
	return 1000
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}
