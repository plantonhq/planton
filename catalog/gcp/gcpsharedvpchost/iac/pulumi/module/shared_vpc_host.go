package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// sharedVpcHost enables the project as a Shared VPC host.
//
// The provider resource is a flag on the project: its only argument is
// the project, which is immutable (a different project is a different
// host). Destroying it disables the host role, which Google refuses while
// any service project is still attached -- the registry places every
// GcpSharedVpcServiceProject downstream of the host it references, so a
// chart's teardown detaches first.
func sharedVpcHost(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpSharedVpcHost.Spec

	args := &compute.SharedVPCHostProjectArgs{}

	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project. Unlike most GCP resources this one
	// REQUIRES an explicit project argument, so the fallback is made
	// concrete by reading the provider's resolved project from the client
	// config (the Pulumi counterpart of the Terraform module's
	// google_client_config data source).
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	} else {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to read provider client config for the default project")
		}
		if clientConfig.Project == "" {
			return errors.New("project_id is empty and the provider has no default project configured")
		}
		args.Project = pulumi.String(clientConfig.Project)
	}

	// Empty defers to the provider default (DELETE: disable the host role).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := compute.NewSharedVPCHostProject(ctx, locals.GcpSharedVpcHost.Metadata.Name, args,
		pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable shared vpc host project")
	}

	ctx.Export(OpHostProjectId, created.Project)

	return nil
}
