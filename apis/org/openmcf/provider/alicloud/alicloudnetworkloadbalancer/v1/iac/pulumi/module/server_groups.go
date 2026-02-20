package module

import (
	"github.com/pkg/errors"
	alicloudnetworkloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudnetworkloadbalancer/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/nlb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func serverGroup(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	vpcId string,
	sg *alicloudnetworkloadbalancerv1.AlicloudNetworkLoadBalancerServerGroup,
) (*nlb.ServerGroup, error) {
	hc := sg.HealthCheck

	hcArgs := nlb.ServerGroupHealthCheckArgs{
		HealthCheckEnabled:        pulumi.Bool(hc.HealthCheckEnabled),
		HealthCheckType:           pulumi.String(healthCheckType(hc)),
		HealthCheckConnectPort:    pulumi.Int(healthCheckConnectPort(hc)),
		HealthCheckConnectTimeout: pulumi.Int(healthCheckConnectTimeout(hc)),
		HealthCheckInterval:       pulumi.Int(healthCheckInterval(hc)),
		HealthyThreshold:          pulumi.Int(healthyThreshold(hc)),
		UnhealthyThreshold:        pulumi.Int(unhealthyThreshold(hc)),
	}

	if hc.HealthCheckUrl != "" {
		hcArgs.HealthCheckUrl = pulumi.String(hc.HealthCheckUrl)
	}

	if hc.HealthCheckDomain != "" {
		hcArgs.HealthCheckDomain = pulumi.String(hc.HealthCheckDomain)
	}

	if hc.HttpCheckMethod != nil || healthCheckType(hc) == "HTTP" {
		hcArgs.HttpCheckMethod = pulumi.String(httpCheckMethod(hc))
	}

	if len(hc.HealthCheckHttpCodes) > 0 {
		hcArgs.HealthCheckHttpCodes = pulumi.ToStringArray(hc.HealthCheckHttpCodes)
	}

	args := &nlb.ServerGroupArgs{
		ServerGroupName:         pulumi.String(sg.Name),
		VpcId:                   pulumi.String(vpcId),
		Protocol:                pulumi.String(serverGroupProtocol(sg)),
		Scheduler:               pulumi.String(serverGroupScheduler(sg)),
		ConnectionDrainEnabled:  pulumi.Bool(connectionDrainEnabled(sg)),
		ConnectionDrainTimeout:  pulumi.Int(connectionDrainTimeout(sg)),
		PreserveClientIpEnabled: pulumi.Bool(preserveClientIpEnabled(sg)),
		HealthCheck:             hcArgs,
	}

	created, err := nlb.NewServerGroup(ctx, sg.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create NLB server group %s", sg.Name)
	}

	return created, nil
}
