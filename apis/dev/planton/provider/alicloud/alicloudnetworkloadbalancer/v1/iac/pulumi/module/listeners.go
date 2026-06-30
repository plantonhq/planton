package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudnetworkloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudnetworkloadbalancer/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/nlb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func listener(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	lb *nlb.LoadBalancer,
	serverGroupIdByName map[string]pulumi.IDOutput,
	l *alicloudnetworkloadbalancerv1.AliCloudNetworkLoadBalancerListener,
) error {
	sgIdOutput, ok := serverGroupIdByName[l.ServerGroupName]
	if !ok {
		return fmt.Errorf(
			"listener on port %d references server group %q which is not defined in server_groups",
			l.ListenerPort, l.ServerGroupName,
		)
	}

	resourceName := fmt.Sprintf("listener-%d-%s", l.ListenerPort, l.ListenerProtocol)

	args := &nlb.ListenerArgs{
		LoadBalancerId:       lb.ID(),
		ListenerPort:         pulumi.Int(int(l.ListenerPort)),
		ListenerProtocol:     pulumi.String(l.ListenerProtocol),
		ServerGroupId:        sgIdOutput,
		IdleTimeout:          pulumi.Int(listenerIdleTimeout(l)),
		ProxyProtocolEnabled: pulumi.Bool(listenerProxyProtocolEnabled(l)),
	}

	if l.ListenerDescription != "" {
		args.ListenerDescription = pulumi.String(l.ListenerDescription)
	}

	if len(l.CertificateIds) > 0 {
		args.CertificateIds = pulumi.ToStringArray(l.CertificateIds)
	}

	if l.SecurityPolicyId != "" {
		args.SecurityPolicyId = pulumi.String(l.SecurityPolicyId)
	}

	if len(l.CaCertificateIds) > 0 {
		args.CaCertificateIds = pulumi.ToStringArray(l.CaCertificateIds)
	}

	if listenerCaEnabled(l) {
		args.CaEnabled = pulumi.Bool(true)
	}

	_, err := nlb.NewListener(ctx, resourceName, args,
		pulumi.Provider(provider),
		pulumi.Parent(lb),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create NLB listener %s", resourceName)
	}

	return nil
}
