package module

import (
	"github.com/pkg/errors"
	alicloudalbloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudalbloadbalancer/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/alb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func serverGroup(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	vpcId string,
	sg *alicloudalbloadbalancerv1.AlicloudAlbServerGroup,
) (*alb.ServerGroup, error) {
	hc := sg.HealthCheckConfig

	hcArgs := alb.ServerGroupHealthCheckConfigArgs{
		HealthCheckEnabled:     pulumi.Bool(hc.HealthCheckEnabled),
		HealthCheckProtocol:    pulumi.String(healthCheckProtocol(hc)),
		HealthCheckMethod:      pulumi.String(healthCheckMethod(hc)),
		HealthCheckConnectPort: pulumi.Int(healthCheckConnectPort(hc)),
		HealthCheckInterval:    pulumi.Int(healthCheckInterval(hc)),
		HealthCheckTimeout:     pulumi.Int(healthCheckTimeout(hc)),
		HealthyThreshold:       pulumi.Int(healthyThreshold(hc)),
		UnhealthyThreshold:     pulumi.Int(unhealthyThreshold(hc)),
	}

	if hc.HealthCheckPath != "" {
		hcArgs.HealthCheckPath = pulumi.String(hc.HealthCheckPath)
	}

	if hc.HealthCheckHost != "" {
		hcArgs.HealthCheckHost = pulumi.String(hc.HealthCheckHost)
	}

	if len(hc.HealthCheckCodes) > 0 {
		hcArgs.HealthCheckCodes = pulumi.ToStringArray(hc.HealthCheckCodes)
	}

	args := &alb.ServerGroupArgs{
		ServerGroupName:   pulumi.String(sg.Name),
		VpcId:             pulumi.String(vpcId),
		Protocol:          pulumi.String(serverGroupProtocol(sg)),
		Scheduler:         pulumi.String(serverGroupScheduler(sg)),
		HealthCheckConfig: hcArgs,
	}

	if sg.StickySessionConfig != nil {
		sc := sg.StickySessionConfig
		stickyArgs := alb.ServerGroupStickySessionConfigArgs{
			StickySessionEnabled: pulumi.Bool(sc.StickySessionEnabled),
		}
		if sc.StickySessionType != nil {
			stickyArgs.StickySessionType = pulumi.String(*sc.StickySessionType)
		}
		if sc.Cookie != "" {
			stickyArgs.Cookie = pulumi.String(sc.Cookie)
		}
		if sc.StickySessionEnabled && (sc.StickySessionType == nil || *sc.StickySessionType == "Insert") {
			stickyArgs.CookieTimeout = pulumi.Int(cookieTimeout(sc))
		}
		args.StickySessionConfig = stickyArgs
	}

	created, err := alb.NewServerGroup(ctx, sg.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ALB server group %s", sg.Name)
	}

	return created, nil
}
