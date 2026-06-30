package module

import (
	"fmt"

	awsnlbv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsnetworkloadbalancer/v1"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// listeners creates a target group and listener for each entry in the spec's
// listener list. It exports map-keyed outputs for listener ARNs and target
// group ARNs, keyed by the listener name.
func listeners(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	nlbResource *lb.LoadBalancer,
) error {
	listenerArnMap := pulumi.StringMap{}
	targetGroupArnMap := pulumi.StringMap{}

	for _, listenerSpec := range locals.Nlb.Spec.Listeners {
		tg, err := targetGroup(ctx, locals, provider, listenerSpec)
		if err != nil {
			return errors.Wrapf(err, "failed to create target group for listener %s", listenerSpec.Name)
		}

		createdListener, err := listener(ctx, locals, provider, nlbResource, listenerSpec, tg)
		if err != nil {
			return errors.Wrapf(err, "failed to create listener %s", listenerSpec.Name)
		}

		listenerArnMap[listenerSpec.Name] = createdListener.Arn
		targetGroupArnMap[listenerSpec.Name] = tg.Arn
	}

	ctx.Export(OpListenerArns, listenerArnMap)
	ctx.Export(OpTargetGroupArns, targetGroupArnMap)

	return nil
}

// targetGroup creates an lb.TargetGroup from a listener's inline target group spec.
func targetGroup(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	listenerSpec *awsnlbv1.AwsNetworkLoadBalancerListener,
) (*lb.TargetGroup, error) {
	tgSpec := listenerSpec.TargetGroup
	tgName := fmt.Sprintf("%s-%s", locals.Nlb.Metadata.Name, listenerSpec.Name)

	// Target type defaults to "instance" when not specified.
	targetType := "instance"
	if tgSpec.TargetType != "" {
		targetType = tgSpec.TargetType
	}

	args := &lb.TargetGroupArgs{
		Name:       pulumi.String(truncateName(tgName, 32)),
		Port:       pulumi.Int(int(tgSpec.Port)),
		Protocol:   pulumi.String(tgSpec.Protocol),
		TargetType: pulumi.String(targetType),
		VpcId:      pulumi.StringPtr(""), // Will be inferred from the NLB's VPC.
		Tags:       pulumi.ToStringMap(locals.AwsTags),
	}

	// Deregistration delay (default 300s handled by AWS when 0).
	if tgSpec.DeregistrationDelaySeconds > 0 {
		args.DeregistrationDelay = pulumi.IntPtr(int(tgSpec.DeregistrationDelaySeconds))
	}

	// Preserve client IP.
	if tgSpec.PreserveClientIp {
		args.PreserveClientIp = pulumi.StringPtr("true")
	}

	// Proxy Protocol v2.
	if tgSpec.ProxyProtocolV2 {
		args.ProxyProtocolV2 = pulumi.BoolPtr(true)
	}

	// Connection termination on deregistration.
	if tgSpec.ConnectionTermination {
		args.ConnectionTermination = pulumi.BoolPtr(true)
	}

	// Source IP stickiness.
	if tgSpec.StickinessEnabled {
		args.Stickiness = &lb.TargetGroupStickinessArgs{
			Enabled: pulumi.Bool(true),
			Type:    pulumi.String("source_ip"),
		}
	}

	// Health check configuration.
	if tgSpec.HealthCheck != nil {
		hc := tgSpec.HealthCheck
		healthCheck := &lb.TargetGroupHealthCheckArgs{}

		if hc.Protocol != "" {
			healthCheck.Protocol = pulumi.StringPtr(hc.Protocol)
		}
		if hc.Port != "" {
			healthCheck.Port = pulumi.StringPtr(hc.Port)
		}
		if hc.Path != "" {
			healthCheck.Path = pulumi.StringPtr(hc.Path)
		}
		if hc.HealthyThreshold > 0 {
			healthCheck.HealthyThreshold = pulumi.IntPtr(int(hc.HealthyThreshold))
		}
		if hc.UnhealthyThreshold > 0 {
			healthCheck.UnhealthyThreshold = pulumi.IntPtr(int(hc.UnhealthyThreshold))
		}
		if hc.IntervalSeconds > 0 {
			healthCheck.Interval = pulumi.IntPtr(int(hc.IntervalSeconds))
		}
		if hc.TimeoutSeconds > 0 {
			healthCheck.Timeout = pulumi.IntPtr(int(hc.TimeoutSeconds))
		}
		if hc.Matcher != "" {
			healthCheck.Matcher = pulumi.StringPtr(hc.Matcher)
		}

		args.HealthCheck = healthCheck
	}

	createdTg, err := lb.NewTargetGroup(ctx, tgName, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create target group %s", tgName)
	}

	return createdTg, nil
}

// listener creates an lb.Listener that forwards to the given target group.
func listener(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	nlbResource *lb.LoadBalancer,
	listenerSpec *awsnlbv1.AwsNetworkLoadBalancerListener,
	tg *lb.TargetGroup,
) (*lb.Listener, error) {
	listenerName := fmt.Sprintf("%s-%s", locals.Nlb.Metadata.Name, listenerSpec.Name)

	args := &lb.ListenerArgs{
		LoadBalancerArn: nlbResource.Arn,
		Port:            pulumi.Int(int(listenerSpec.Port)),
		Protocol:        pulumi.String(listenerSpec.Protocol),
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: tg.Arn,
			},
		},
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// TLS configuration.
	if listenerSpec.Tls != nil {
		args.CertificateArn = pulumi.StringPtr(listenerSpec.Tls.CertificateArn.GetValue())
		if listenerSpec.Tls.SslPolicy != "" {
			args.SslPolicy = pulumi.StringPtr(listenerSpec.Tls.SslPolicy)
		}
	}

	// TCP idle timeout (only for TCP protocol).
	if listenerSpec.TcpIdleTimeoutSeconds > 0 {
		args.TcpIdleTimeoutSeconds = pulumi.IntPtr(int(listenerSpec.TcpIdleTimeoutSeconds))
	}

	// ALPN policy (only for TLS protocol).
	if listenerSpec.AlpnPolicy != "" {
		args.AlpnPolicy = pulumi.StringPtr(listenerSpec.AlpnPolicy)
	}

	createdListener, err := lb.NewListener(ctx, listenerName, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create listener %s", listenerName)
	}

	return createdListener, nil
}

// truncateName ensures a name does not exceed maxLen characters. AWS target
// group names have a 32-character limit.
func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen]
}
