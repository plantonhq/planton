package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// zonalNetworkEndpointGroup builds the ZONAL group (spec.zone set): VM,
// hybrid, or internet endpoints inside a VPC, the backend of regional and
// passthrough load balancers. Everything but the endpoint list is
// immutable. The membership is written as ONE set through Google's bulk
// endpoint operation, so the manifest's list is the group's whole
// membership; a group declared ahead of its members carries no endpoint
// resource at all.
func zonalNetworkEndpointGroup(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpNetworkEndpointGroup.Spec

	args := &compute.NetworkEndpointGroupArgs{
		Name:    pulumi.String(locals.NegName),
		Zone:    pulumi.String(spec.Zone),
		Network: pulumi.String(spec.Network.GetValue()),
	}

	// Empty project falls back to the provider's default project — the same
	// ambient contract the Terraform module honors.
	if locals.ProjectId != "" {
		args.Project = pulumi.StringPtr(locals.ProjectId)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if v := spec.Subnetwork.GetValue(); v != "" {
		args.Subnetwork = pulumi.StringPtr(v)
	}
	// The zonal resource defaults the type to GCE_VM_IP_PORT itself, so an
	// empty spec value is left nil and Google's default stands.
	if spec.NetworkEndpointType != "" {
		args.NetworkEndpointType = pulumi.StringPtr(spec.NetworkEndpointType)
	}
	if spec.DefaultPort != nil {
		args.DefaultPort = pulumi.IntPtr(int(spec.GetDefaultPort()))
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	group, err := compute.NewNetworkEndpointGroup(ctx, "network-endpoint-group", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create network endpoint group")
	}

	if len(spec.Endpoints) > 0 {
		var members compute.NetworkEndpointListNetworkEndpointArray
		for _, e := range spec.Endpoints {
			m := &compute.NetworkEndpointListNetworkEndpointArgs{}
			if v := e.Instance.GetValue(); v != "" {
				m.Instance = pulumi.StringPtr(v)
			}
			if e.IpAddress != "" {
				m.IpAddress = pulumi.StringPtr(e.IpAddress)
			}
			if e.Port != nil {
				m.Port = pulumi.IntPtr(int(e.GetPort()))
			}
			members = append(members, m)
		}
		endpointsArgs := &compute.NetworkEndpointListArgs{
			NetworkEndpointGroup: group.Name,
			Zone:                 pulumi.String(spec.Zone),
			NetworkEndpoints:     members,
		}
		if locals.ProjectId != "" {
			endpointsArgs.Project = pulumi.StringPtr(locals.ProjectId)
		}
		if spec.DeletionPolicy != "" {
			endpointsArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		if _, err := compute.NewNetworkEndpointList(ctx, "network-endpoints", endpointsArgs,
			pulumi.Provider(gcpProvider), pulumi.Parent(group)); err != nil {
			return errors.Wrap(err, "failed to write the network endpoint set")
		}
	}

	ctx.Export(OpSelfLink, group.SelfLink)
	ctx.Export(OpNegName, group.Name)
	ctx.Export(OpNegId, group.GeneratedId.ApplyT(func(id int) string { return strconv.Itoa(id) }).(pulumi.StringOutput))
	ctx.Export(OpZone, pulumi.String(spec.Zone))
	// The declared membership: the manifest's list is the whole membership
	// on both scopes (PARITY with the Terraform module).
	ctx.Export(OpSize, pulumi.String(strconv.Itoa(len(spec.Endpoints))))
	return nil
}
