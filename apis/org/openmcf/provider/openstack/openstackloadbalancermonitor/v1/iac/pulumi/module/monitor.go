package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// monitor provisions the OpenStack Octavia health monitor and exports outputs.
func monitor(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackLoadBalancerMonitor.Spec
	monitorName := locals.OpenStackLoadBalancerMonitor.Metadata.Name

	monitorArgs := &loadbalancer.MonitorArgs{
		Name:       pulumi.String(monitorName),
		PoolId:     pulumi.String(locals.PoolId),
		Type:       pulumi.String(spec.Type),
		Delay:      pulumi.Int(int(spec.Delay)),
		Timeout:    pulumi.Int(int(spec.Timeout)),
		MaxRetries: pulumi.Int(int(spec.MaxRetries)),
	}

	// Set max_retries_down if explicitly provided.
	if spec.MaxRetriesDown != nil {
		monitorArgs.MaxRetriesDown = pulumi.IntPtr(int(spec.GetMaxRetriesDown()))
	}

	// Set url_path if provided (HTTP/HTTPS monitors only).
	if spec.UrlPath != "" {
		monitorArgs.UrlPath = pulumi.StringPtr(spec.UrlPath)
	}

	// Set http_method if provided (HTTP/HTTPS monitors only).
	if spec.HttpMethod != "" {
		monitorArgs.HttpMethod = pulumi.StringPtr(spec.HttpMethod)
	}

	// Set expected_codes if provided (HTTP/HTTPS monitors only).
	if spec.ExpectedCodes != "" {
		monitorArgs.ExpectedCodes = pulumi.StringPtr(spec.ExpectedCodes)
	}

	// Set admin_state_up if explicitly provided.
	if spec.AdminStateUp != nil {
		monitorArgs.AdminStateUp = pulumi.BoolPtr(spec.GetAdminStateUp())
	}

	// Note: Health monitors do NOT support tags in the Terraform provider.

	// Set region override if provided.
	if spec.Region != "" {
		monitorArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdMonitor, err := loadbalancer.NewMonitor(
		ctx,
		strings.ToLower(monitorName),
		monitorArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack load balancer monitor")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpMonitorId, createdMonitor.ID())
	ctx.Export(OpName, createdMonitor.Name)
	ctx.Export(OpType, createdMonitor.Type)
	ctx.Export(OpPoolId, createdMonitor.PoolId)
	ctx.Export(OpRegion, createdMonitor.Region)

	return nil
}
