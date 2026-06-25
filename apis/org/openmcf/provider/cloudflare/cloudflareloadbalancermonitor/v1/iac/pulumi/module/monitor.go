package module

import (
	"github.com/pkg/errors"
	cloudflareloadbalancermonitorv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareloadbalancermonitor/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// monitor provisions the account-scoped Cloudflare Load Balancer monitor and
// exports its outputs.
func monitor(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.LoadBalancerMonitor, error) {
	spec := locals.CloudflareLoadBalancerMonitor.Spec

	// The enum's unspecified zero value maps to the Cloudflare default protocol.
	monitorType := "http"
	if spec.Type != cloudflareloadbalancermonitorv1.CloudflareLoadBalancerMonitorType_monitor_type_unspecified {
		monitorType = spec.Type.String()
	}

	args := &cloudflare.LoadBalancerMonitorArgs{
		AccountId: pulumi.String(spec.AccountId),
		Type:      pulumi.String(monitorType),
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Path != "" {
		args.Path = pulumi.StringPtr(spec.Path)
	}
	if spec.ExpectedCodes != "" {
		args.ExpectedCodes = pulumi.StringPtr(spec.ExpectedCodes)
	}
	if spec.ExpectedBody != "" {
		args.ExpectedBody = pulumi.StringPtr(spec.ExpectedBody)
	}
	if spec.Method != "" {
		args.Method = pulumi.StringPtr(spec.Method)
	}
	if spec.ProbeZone != "" {
		args.ProbeZone = pulumi.StringPtr(spec.ProbeZone)
	}
	if spec.Port > 0 {
		args.Port = pulumi.IntPtr(int(spec.Port))
	}
	// 0 means "use the Cloudflare default" for these tuning knobs.
	if spec.Interval > 0 {
		args.Interval = pulumi.IntPtr(int(spec.Interval))
	}
	if spec.Timeout > 0 {
		args.Timeout = pulumi.IntPtr(int(spec.Timeout))
	}
	if spec.Retries > 0 {
		args.Retries = pulumi.IntPtr(int(spec.Retries))
	}
	if spec.ConsecutiveUp > 0 {
		args.ConsecutiveUp = pulumi.IntPtr(int(spec.ConsecutiveUp))
	}
	if spec.ConsecutiveDown > 0 {
		args.ConsecutiveDown = pulumi.IntPtr(int(spec.ConsecutiveDown))
	}
	args.FollowRedirects = pulumi.BoolPtr(spec.FollowRedirects)
	args.AllowInsecure = pulumi.BoolPtr(spec.AllowInsecure)

	if len(spec.Headers) > 0 {
		headers := pulumi.StringArrayMap{}
		for _, h := range spec.Headers {
			headers[h.Name] = pulumi.ToStringArray(h.Values)
		}
		args.Header = headers
	}

	created, err := cloudflare.NewLoadBalancerMonitor(
		ctx,
		"monitor",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare load balancer monitor")
	}

	ctx.Export(OpMonitorId, created.ID())
	ctx.Export(OpMonitorType, created.Type)

	return created, nil
}
