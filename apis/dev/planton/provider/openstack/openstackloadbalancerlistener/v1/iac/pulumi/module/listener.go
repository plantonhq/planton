package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// listener provisions the OpenStack Octavia listener and exports outputs.
func listener(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackLoadBalancerListener.Spec
	listenerName := locals.OpenStackLoadBalancerListener.Metadata.Name

	listenerArgs := &loadbalancer.ListenerArgs{
		Name:           pulumi.String(listenerName),
		LoadbalancerId: pulumi.String(locals.LoadBalancerId),
		Protocol:       pulumi.String(spec.Protocol),
		ProtocolPort:   pulumi.Int(int(spec.ProtocolPort)),
	}

	// Set description if provided.
	if spec.Description != "" {
		listenerArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set connection_limit if explicitly provided.
	if spec.ConnectionLimit != nil {
		listenerArgs.ConnectionLimit = pulumi.IntPtr(int(spec.GetConnectionLimit()))
	}

	// Set default_tls_container_ref if provided (required for TERMINATED_HTTPS).
	if spec.DefaultTlsContainerRef != "" {
		listenerArgs.DefaultTlsContainerRef = pulumi.StringPtr(spec.DefaultTlsContainerRef)
	}

	// Set insert_headers if provided.
	if len(spec.InsertHeaders) > 0 {
		headers := make(pulumi.StringMap, len(spec.InsertHeaders))
		for k, v := range spec.InsertHeaders {
			headers[k] = pulumi.String(v)
		}
		listenerArgs.InsertHeaders = headers
	}

	// Set allowed_cidrs if provided.
	if len(spec.AllowedCidrs) > 0 {
		cidrs := make(pulumi.StringArray, len(spec.AllowedCidrs))
		for i, cidr := range spec.AllowedCidrs {
			cidrs[i] = pulumi.String(cidr)
		}
		listenerArgs.AllowedCidrs = cidrs
	}

	// Set admin_state_up if explicitly provided.
	if spec.AdminStateUp != nil {
		listenerArgs.AdminStateUp = pulumi.BoolPtr(spec.GetAdminStateUp())
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, len(spec.Tags))
		for i, tag := range spec.Tags {
			tags[i] = pulumi.String(tag)
		}
		listenerArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		listenerArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdListener, err := loadbalancer.NewListener(
		ctx,
		strings.ToLower(listenerName),
		listenerArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack load balancer listener")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpListenerId, createdListener.ID())
	ctx.Export(OpName, createdListener.Name)
	ctx.Export(OpProtocol, createdListener.Protocol)
	ctx.Export(OpProtocolPort, createdListener.ProtocolPort)
	ctx.Export(OpRegion, createdListener.Region)

	return nil
}
