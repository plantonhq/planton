package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// sharedVpcServiceProject attaches the service project to its host.
//
// Both projects are immutable (moving to another host is a detach and an
// attach). deletion_policy on this resource is a single opt-in: ABANDON
// leaves the attachment in place on destroy; anything else (the default)
// detaches, which Google refuses while resources in the service project
// still use a host subnetwork.
func sharedVpcServiceProject(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpSharedVpcServiceProject.Spec

	args := &compute.SharedVPCServiceProjectArgs{
		HostProject:    pulumi.String(spec.HostProjectId.GetValue()),
		ServiceProject: pulumi.String(spec.ServiceProjectId.GetValue()),
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := compute.NewSharedVPCServiceProject(ctx, locals.GcpSharedVpcServiceProject.Metadata.Name, args,
		pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to attach shared vpc service project")
	}

	ctx.Export(OpServiceProjectId, created.ServiceProject)
	ctx.Export(OpHostProjectId, created.HostProject)

	return nil
}
