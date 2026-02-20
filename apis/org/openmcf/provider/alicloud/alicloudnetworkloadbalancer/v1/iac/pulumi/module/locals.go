package module

import (
	"strings"

	alicloudnetworkloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudnetworkloadbalancer/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudNetworkLoadBalancer *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancer
	Tags                        map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudNetworkLoadBalancer = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudNetworkLoadBalancer.String()),
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

func addressType(spec *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerSpec) string {
	if spec.AddressType != nil {
		return *spec.AddressType
	}
	return "Internet"
}

func crossZoneEnabled(spec *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerSpec) bool {
	if spec.CrossZoneEnabled != nil {
		return *spec.CrossZoneEnabled
	}
	return true
}

func serverGroupProtocol(sg *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerServerGroup) string {
	if sg.Protocol != nil {
		return *sg.Protocol
	}
	return "TCP"
}

func serverGroupScheduler(sg *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerServerGroup) string {
	if sg.Scheduler != nil {
		return *sg.Scheduler
	}
	return "Wrr"
}

func connectionDrainEnabled(sg *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerServerGroup) bool {
	if sg.ConnectionDrainEnabled != nil {
		return *sg.ConnectionDrainEnabled
	}
	return false
}

func connectionDrainTimeout(sg *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerServerGroup) int {
	if sg.ConnectionDrainTimeout != nil {
		return int(*sg.ConnectionDrainTimeout)
	}
	return 10
}

func preserveClientIpEnabled(sg *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerServerGroup) bool {
	if sg.PreserveClientIpEnabled != nil {
		return *sg.PreserveClientIpEnabled
	}
	return true
}

func healthCheckType(hc *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerHealthCheckConfig) string {
	if hc.HealthCheckType != nil {
		return *hc.HealthCheckType
	}
	return "TCP"
}

func healthCheckConnectPort(hc *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerHealthCheckConfig) int {
	if hc.HealthCheckConnectPort != nil {
		return int(*hc.HealthCheckConnectPort)
	}
	return 0
}

func healthCheckConnectTimeout(hc *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerHealthCheckConfig) int {
	if hc.HealthCheckConnectTimeout != nil {
		return int(*hc.HealthCheckConnectTimeout)
	}
	return 5
}

func healthCheckInterval(hc *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerHealthCheckConfig) int {
	if hc.HealthCheckInterval != nil {
		return int(*hc.HealthCheckInterval)
	}
	return 10
}

func healthyThreshold(hc *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerHealthCheckConfig) int {
	if hc.HealthyThreshold != nil {
		return int(*hc.HealthyThreshold)
	}
	return 2
}

func unhealthyThreshold(hc *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerHealthCheckConfig) int {
	if hc.UnhealthyThreshold != nil {
		return int(*hc.UnhealthyThreshold)
	}
	return 2
}

func httpCheckMethod(hc *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerHealthCheckConfig) string {
	if hc.HttpCheckMethod != nil {
		return *hc.HttpCheckMethod
	}
	return "GET"
}

func listenerIdleTimeout(l *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerListener) int {
	if l.IdleTimeout != nil {
		return int(*l.IdleTimeout)
	}
	return 900
}

func listenerProxyProtocolEnabled(l *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerListener) bool {
	if l.ProxyProtocolEnabled != nil {
		return *l.ProxyProtocolEnabled
	}
	return false
}

func listenerCaEnabled(l *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerListener) bool {
	if l.CaEnabled != nil {
		return *l.CaEnabled
	}
	return false
}
