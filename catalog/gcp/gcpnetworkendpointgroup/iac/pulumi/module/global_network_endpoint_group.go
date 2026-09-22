package module

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// globalNetworkEndpointGroup builds the GLOBAL internet group (spec.zone
// empty): INTERNET_IP_PORT or INTERNET_FQDN_PORT endpoints outside Google
// Cloud, the backend of a global external Application Load Balancer.
// Global endpoints have no bulk twin: each is its own resource, keyed by
// its position in the list, and every argument is immutable.
func globalNetworkEndpointGroup(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpNetworkEndpointGroup.Spec

	args := &compute.GlobalNetworkEndpointGroupArgs{
		Name:                pulumi.String(locals.NegName),
		NetworkEndpointType: pulumi.String(spec.NetworkEndpointType),
	}
	if locals.ProjectId != "" {
		args.Project = pulumi.StringPtr(locals.ProjectId)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.DefaultPort != nil {
		args.DefaultPort = pulumi.IntPtr(int(spec.GetDefaultPort()))
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	group, err := compute.NewGlobalNetworkEndpointGroup(ctx, "global-network-endpoint-group", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create global network endpoint group")
	}

	for i, e := range spec.Endpoints {
		// Global endpoints require a port; the group's default_port is the
		// fallback the spec allows.
		port := e.GetPort()
		if e.Port == nil {
			port = spec.GetDefaultPort()
		}
		epArgs := &compute.GlobalNetworkEndpointArgs{
			GlobalNetworkEndpointGroup: group.Name,
			Port:                       pulumi.Int(int(port)),
		}
		if locals.ProjectId != "" {
			epArgs.Project = pulumi.StringPtr(locals.ProjectId)
		}
		if e.IpAddress != "" {
			epArgs.IpAddress = pulumi.StringPtr(e.IpAddress)
		}
		if e.Fqdn != "" {
			epArgs.Fqdn = pulumi.StringPtr(e.Fqdn)
		}
		if spec.DeletionPolicy != "" {
			epArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		if _, err := compute.NewGlobalNetworkEndpoint(ctx, fmt.Sprintf("global-network-endpoint-%d", i), epArgs,
			pulumi.Provider(gcpProvider), pulumi.Parent(group)); err != nil {
			return errors.Wrapf(err, "failed to create global network endpoint %d", i)
		}
	}

	ctx.Export(OpSelfLink, group.SelfLink)
	ctx.Export(OpNegName, group.Name)
	// The global collection exposes no server-generated numeric id through
	// the provider.
	ctx.Export(OpNegId, pulumi.String(""))
	ctx.Export(OpZone, pulumi.String(""))
	ctx.Export(OpSize, pulumi.String(strconv.Itoa(len(spec.Endpoints))))
	return nil
}
