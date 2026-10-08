package module

import (
	"fmt"

	"github.com/pkg/errors"
	awslbtargetgroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awslbtargetgroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// targetGroup provisions the target group and any static target
// registrations. Name, port, protocol, protocol_version, vpc_id, target_type,
// and ip_address_type are create-only in AWS: the provider replaces the group
// (and re-creates registrations) when they change. Everything else --
// health checks, stickiness, algorithms, drain behavior -- updates in place.
func targetGroup(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) error {
	spec := locals.AwsLbTargetGroup.Spec

	isLambda := spec.TargetType == "lambda"

	// The proto cannot enforce VPC requiredness per target type
	// (message-level CEL on StringValueOrRef fields breaks protovalidate-java),
	// so the module is the enforcement point: fail fast with a clear message
	// instead of letting AWS reject a half-created plan.
	if !isLambda && spec.VpcId.GetValue() == "" {
		return errors.Errorf("vpc_id is required when target_type is %q (only lambda target groups have no VPC)",
			specTargetTypeOrDefault(spec))
	}

	args := &lb.TargetGroupArgs{
		Name: pulumi.String(locals.TargetGroupName),
		Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(locals.AwsTags),
	}

	// Port, protocol, and VPC apply to every target type except lambda -- a
	// Lambda function is invoked directly, never addressed over the network.
	if !isLambda {
		args.Port = pulumi.IntPtr(int(spec.Port))
		args.Protocol = pulumi.StringPtr(spec.Protocol)
		args.VpcId = pulumi.StringPtr(spec.VpcId.GetValue())
	}

	if spec.TargetType != "" {
		args.TargetType = pulumi.StringPtr(spec.TargetType)
	}
	if spec.ProtocolVersion != "" {
		args.ProtocolVersion = pulumi.StringPtr(spec.ProtocolVersion)
	}
	if spec.IpAddressType != "" {
		args.IpAddressType = pulumi.StringPtr(spec.IpAddressType)
	}

	// ALB Target Optimizer: setting the agent port enables per-target
	// readiness routing. Create-only, like the group's other identity fields.
	if spec.TargetControlPort > 0 {
		args.TargetControlPort = pulumi.IntPtr(int(spec.TargetControlPort))
	}

	// 0 means "keep the AWS default" (300s) -- the proto zero value is not
	// distinguishable from unset, so immediate deregistration is expressed as 1.
	if spec.DeregistrationDelaySeconds > 0 {
		args.DeregistrationDelay = pulumi.IntPtr(int(spec.DeregistrationDelaySeconds))
	}
	if spec.SlowStartSeconds > 0 {
		args.SlowStart = pulumi.IntPtr(int(spec.SlowStartSeconds))
	}
	if spec.LoadBalancingAlgorithmType != "" {
		args.LoadBalancingAlgorithmType = pulumi.StringPtr(spec.LoadBalancingAlgorithmType)
	}
	if spec.LoadBalancingAnomalyMitigation != "" {
		args.LoadBalancingAnomalyMitigation = pulumi.StringPtr(spec.LoadBalancingAnomalyMitigation)
	}
	if spec.LoadBalancingCrossZoneEnabled != "" {
		args.LoadBalancingCrossZoneEnabled = pulumi.StringPtr(spec.LoadBalancingCrossZoneEnabled)
	}

	// preserve_client_ip is a nullable tri-state at AWS (the default depends
	// on the target type), which is why the proto models it as optional bool
	// and the provider takes a string.
	if spec.PreserveClientIp != nil {
		args.PreserveClientIp = pulumi.StringPtr(fmt.Sprintf("%t", spec.GetPreserveClientIp()))
	}
	if spec.ProxyProtocolV2 {
		args.ProxyProtocolV2 = pulumi.BoolPtr(true)
	}
	if spec.ConnectionTermination {
		args.ConnectionTermination = pulumi.BoolPtr(true)
	}
	if spec.LambdaMultiValueHeadersEnabled {
		args.LambdaMultiValueHeadersEnabled = pulumi.BoolPtr(true)
	}

	if spec.HealthCheck != nil {
		args.HealthCheck = healthCheckArgs(spec.HealthCheck)
	}
	if spec.Stickiness != nil {
		args.Stickiness = stickinessArgs(spec.Stickiness)
	}
	if spec.TargetGroupHealth != nil {
		args.TargetGroupHealth = targetGroupHealthArgs(spec.TargetGroupHealth)
	}
	if spec.TargetHealthState != nil {
		args.TargetHealthStates = lb.TargetGroupTargetHealthStateArray{
			&lb.TargetGroupTargetHealthStateArgs{
				EnableUnhealthyConnectionTermination: pulumi.Bool(spec.TargetHealthState.EnableUnhealthyConnectionTermination),
				UnhealthyDrainingInterval:            pulumi.IntPtr(int(spec.TargetHealthState.UnhealthyDrainingIntervalSeconds)),
			},
		}
	}

	// The Pulumi resource names are fixed: the AWS name travels only through
	// args.Name, so renaming the component never replaces the group.
	createdTargetGroup, err := lb.NewTargetGroup(ctx, "target-group", args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create target group")
	}

	// Static registrations. Attachments are keyed by index (not target id)
	// because a target id may be a reference resolved only at apply time.
	for i, target := range spec.Targets {
		attachmentArgs := &lb.TargetGroupAttachmentArgs{
			TargetGroupArn: createdTargetGroup.Arn,
			TargetId:       pulumi.String(target.TargetId.GetValue()),
		}
		if target.Port > 0 {
			attachmentArgs.Port = pulumi.IntPtr(int(target.Port))
		}
		if target.AvailabilityZone != "" {
			attachmentArgs.AvailabilityZone = pulumi.StringPtr(target.AvailabilityZone)
		}
		if target.QuicServerId != "" {
			attachmentArgs.QuicServerId = pulumi.StringPtr(target.QuicServerId)
		}
		if _, err := lb.NewTargetGroupAttachment(ctx,
			fmt.Sprintf("target-%d", i),
			attachmentArgs, pulumi.Provider(provider)); err != nil {
			return errors.Wrapf(err, "failed to register target %d", i)
		}
	}

	ctx.Export(OpTargetGroupArn, createdTargetGroup.Arn)
	ctx.Export(OpTargetGroupName, createdTargetGroup.Name)
	ctx.Export(OpArnSuffix, createdTargetGroup.ArnSuffix)

	return nil
}

// healthCheckArgs maps the spec's health check onto provider args. Only
// explicitly set fields are sent, so AWS keeps its protocol-appropriate
// defaults for the rest.
func healthCheckArgs(hc *awslbtargetgroupv1alpha1.AwsLbTargetGroupHealthCheck) *lb.TargetGroupHealthCheckArgs {
	args := &lb.TargetGroupHealthCheckArgs{}

	// enabled is optional bool: nil keeps the AWS default (true); an explicit
	// false disables checks (lambda groups only).
	if hc.Enabled != nil {
		args.Enabled = pulumi.BoolPtr(hc.GetEnabled())
	}
	if hc.Protocol != "" {
		args.Protocol = pulumi.StringPtr(hc.Protocol)
	}
	if hc.Port != "" {
		args.Port = pulumi.StringPtr(hc.Port)
	}
	if hc.Path != "" {
		args.Path = pulumi.StringPtr(hc.Path)
	}
	if hc.HealthyThreshold > 0 {
		args.HealthyThreshold = pulumi.IntPtr(int(hc.HealthyThreshold))
	}
	if hc.UnhealthyThreshold > 0 {
		args.UnhealthyThreshold = pulumi.IntPtr(int(hc.UnhealthyThreshold))
	}
	if hc.IntervalSeconds > 0 {
		args.Interval = pulumi.IntPtr(int(hc.IntervalSeconds))
	}
	if hc.TimeoutSeconds > 0 {
		args.Timeout = pulumi.IntPtr(int(hc.TimeoutSeconds))
	}
	if hc.Matcher != "" {
		args.Matcher = pulumi.StringPtr(hc.Matcher)
	}

	return args
}

// stickinessArgs maps the spec's stickiness onto provider args. Configuring
// the block implies enabling it unless the spec says otherwise -- the same
// semantics AWS applies.
func stickinessArgs(s *awslbtargetgroupv1alpha1.AwsLbTargetGroupStickiness) *lb.TargetGroupStickinessArgs {
	enabled := true
	if s.Enabled != nil {
		enabled = s.GetEnabled()
	}

	args := &lb.TargetGroupStickinessArgs{
		Type:    pulumi.String(s.Type),
		Enabled: pulumi.BoolPtr(enabled),
	}
	if s.CookieDurationSeconds > 0 {
		args.CookieDuration = pulumi.IntPtr(int(s.CookieDurationSeconds))
	}
	if s.CookieName != "" {
		args.CookieName = pulumi.StringPtr(s.CookieName)
	}

	return args
}

// targetGroupHealthArgs maps the group-level health policy. The DNS-failover
// count is a string at AWS (it accepts "off"); the unhealthy-state-routing
// count is a plain integer -- an AWS asymmetry the proto mirrors on purpose.
func targetGroupHealthArgs(h *awslbtargetgroupv1alpha1.AwsLbTargetGroupHealthPolicy) *lb.TargetGroupTargetGroupHealthArgs {
	args := &lb.TargetGroupTargetGroupHealthArgs{}

	if h.DnsFailover != nil {
		failover := &lb.TargetGroupTargetGroupHealthDnsFailoverArgs{}
		if h.DnsFailover.MinimumHealthyTargetsCount != "" {
			failover.MinimumHealthyTargetsCount = pulumi.StringPtr(h.DnsFailover.MinimumHealthyTargetsCount)
		}
		if h.DnsFailover.MinimumHealthyTargetsPercentage != "" {
			failover.MinimumHealthyTargetsPercentage = pulumi.StringPtr(h.DnsFailover.MinimumHealthyTargetsPercentage)
		}
		args.DnsFailover = failover
	}

	if h.UnhealthyStateRouting != nil {
		routing := &lb.TargetGroupTargetGroupHealthUnhealthyStateRoutingArgs{}
		if h.UnhealthyStateRouting.MinimumHealthyTargetsCount > 0 {
			routing.MinimumHealthyTargetsCount = pulumi.IntPtr(int(h.UnhealthyStateRouting.MinimumHealthyTargetsCount))
		}
		if h.UnhealthyStateRouting.MinimumHealthyTargetsPercentage != "" {
			routing.MinimumHealthyTargetsPercentage = pulumi.StringPtr(h.UnhealthyStateRouting.MinimumHealthyTargetsPercentage)
		}
		args.UnhealthyStateRouting = routing
	}

	return args
}

// specTargetTypeOrDefault reports the effective target type for error
// messages ("instance" is the AWS default when unset).
func specTargetTypeOrDefault(spec *awslbtargetgroupv1alpha1.AwsLbTargetGroupSpec) string {
	if spec.TargetType == "" {
		return "instance"
	}
	return spec.TargetType
}

// truncateName enforces AWS's 32-character target group name limit.
func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen]
}
