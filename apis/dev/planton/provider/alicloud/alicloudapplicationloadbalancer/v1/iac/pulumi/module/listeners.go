package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudapplicationloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudapplicationloadbalancer/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/alb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func listener(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	lb *alb.LoadBalancer,
	serverGroupIdByName map[string]pulumi.IDOutput,
	l *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerListener,
) error {
	sgIdOutput, ok := serverGroupIdByName[l.DefaultActionServerGroupName]
	if !ok {
		return fmt.Errorf(
			"listener on port %d references server group %q which is not defined in server_groups",
			l.ListenerPort, l.DefaultActionServerGroupName,
		)
	}

	resourceName := fmt.Sprintf("listener-%d-%s", l.ListenerPort, l.ListenerProtocol)

	args := &alb.ListenerArgs{
		LoadBalancerId:   lb.ID(),
		ListenerPort:     pulumi.Int(int(l.ListenerPort)),
		ListenerProtocol: pulumi.String(l.ListenerProtocol),
		DefaultActions: alb.ListenerDefaultActionArray{
			alb.ListenerDefaultActionArgs{
				Type: pulumi.String("ForwardGroup"),
				ForwardGroupConfig: alb.ListenerDefaultActionForwardGroupConfigArgs{
					ServerGroupTuples: alb.ListenerDefaultActionForwardGroupConfigServerGroupTupleArray{
						alb.ListenerDefaultActionForwardGroupConfigServerGroupTupleArgs{
							ServerGroupId: sgIdOutput,
						},
					},
				},
			},
		},
		GzipEnabled:    pulumi.Bool(listenerGzipEnabled(l)),
		Http2Enabled:   pulumi.Bool(listenerHttp2Enabled(l)),
		IdleTimeout:    pulumi.Int(listenerIdleTimeout(l)),
		RequestTimeout: pulumi.Int(listenerRequestTimeout(l)),
	}

	if l.ListenerDescription != "" {
		args.ListenerDescription = pulumi.String(l.ListenerDescription)
	}

	if l.CertificateId != "" {
		args.Certificates = alb.ListenerCertificatesArgs{
			CertificateId: pulumi.String(l.CertificateId),
		}
	}

	if l.SecurityPolicyId != "" {
		args.SecurityPolicyId = pulumi.String(l.SecurityPolicyId)
	}

	_, err := alb.NewListener(ctx, resourceName, args,
		pulumi.Provider(provider),
		pulumi.Parent(lb),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create ALB listener %s", resourceName)
	}

	return nil
}
