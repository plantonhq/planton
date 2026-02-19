package module

import (
	"strings"

	alicloudnlbloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudnlbloadbalancer/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudNlbLoadBalancer *alicloudnlbloadbalancerv1.AlicloudNlbLoadBalancer
	Tags                    map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudnlbloadbalancerv1.AlicloudNlbLoadBalancerStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudNlbLoadBalancer = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudNlbLoadBalancer.String()),
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

func addressType(spec *alicloudnlbloadbalancerv1.AlicloudNlbLoadBalancerSpec) string {
	if spec.AddressType != nil {
		return *spec.AddressType
	}
	return "Internet"
}

func crossZoneEnabled(spec *alicloudnlbloadbalancerv1.AlicloudNlbLoadBalancerSpec) bool {
	if spec.CrossZoneEnabled != nil {
		return *spec.CrossZoneEnabled
	}
	return true
}

func serverGroupProtocol(sg *alicloudnlbloadbalancerv1.AlicloudNlbServerGroup) string {
	if sg.Protocol != nil {
		return *sg.Protocol
	}
	return "TCP"
}

func serverGroupScheduler(sg *alicloudnlbloadbalancerv1.AlicloudNlbServerGroup) string {
	if sg.Scheduler != nil {
		return *sg.Scheduler
	}
	return "Wrr"
}

func connectionDrainEnabled(sg *alicloudnlbloadbalancerv1.AlicloudNlbServerGroup) bool {
	if sg.ConnectionDrainEnabled != nil {
		return *sg.ConnectionDrainEnabled
	}
	return false
}

func connectionDrainTimeout(sg *alicloudnlbloadbalancerv1.AlicloudNlbServerGroup) int {
	if sg.ConnectionDrainTimeout != nil {
		return int(*sg.ConnectionDrainTimeout)
	}
	return 10
}

func preserveClientIpEnabled(sg *alicloudnlbloadbalancerv1.AlicloudNlbServerGroup) bool {
	if sg.PreserveClientIpEnabled != nil {
		return *sg.PreserveClientIpEnabled
	}
	return true
}

func healthCheckType(hc *alicloudnlbloadbalancerv1.AlicloudNlbHealthCheckConfig) string {
	if hc.HealthCheckType != nil {
		return *hc.HealthCheckType
	}
	return "TCP"
}

func healthCheckConnectPort(hc *alicloudnlbloadbalancerv1.AlicloudNlbHealthCheckConfig) int {
	if hc.HealthCheckConnectPort != nil {
		return int(*hc.HealthCheckConnectPort)
	}
	return 0
}

func healthCheckConnectTimeout(hc *alicloudnlbloadbalancerv1.AlicloudNlbHealthCheckConfig) int {
	if hc.HealthCheckConnectTimeout != nil {
		return int(*hc.HealthCheckConnectTimeout)
	}
	return 5
}

func healthCheckInterval(hc *alicloudnlbloadbalancerv1.AlicloudNlbHealthCheckConfig) int {
	if hc.HealthCheckInterval != nil {
		return int(*hc.HealthCheckInterval)
	}
	return 10
}

func healthyThreshold(hc *alicloudnlbloadbalancerv1.AlicloudNlbHealthCheckConfig) int {
	if hc.HealthyThreshold != nil {
		return int(*hc.HealthyThreshold)
	}
	return 2
}

func unhealthyThreshold(hc *alicloudnlbloadbalancerv1.AlicloudNlbHealthCheckConfig) int {
	if hc.UnhealthyThreshold != nil {
		return int(*hc.UnhealthyThreshold)
	}
	return 2
}

func httpCheckMethod(hc *alicloudnlbloadbalancerv1.AlicloudNlbHealthCheckConfig) string {
	if hc.HttpCheckMethod != nil {
		return *hc.HttpCheckMethod
	}
	return "GET"
}

func listenerIdleTimeout(l *alicloudnlbloadbalancerv1.AlicloudNlbListener) int {
	if l.IdleTimeout != nil {
		return int(*l.IdleTimeout)
	}
	return 900
}

func listenerProxyProtocolEnabled(l *alicloudnlbloadbalancerv1.AlicloudNlbListener) bool {
	if l.ProxyProtocolEnabled != nil {
		return *l.ProxyProtocolEnabled
	}
	return false
}

func listenerCaEnabled(l *alicloudnlbloadbalancerv1.AlicloudNlbListener) bool {
	if l.CaEnabled != nil {
		return *l.CaEnabled
	}
	return false
}
