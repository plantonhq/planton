package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudbuildv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// repository links one code-host repository into Cloud Build through its
// parent connection. No API enablement here: the parent connection's block
// enabled Cloud Build in this project.
func repository(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudBuildRepository.Spec
	resourceName := locals.GcpCloudBuildRepository.Metadata.Name
	parentConnection := spec.GetParentConnection().GetValue()

	project, location, err := parseConnectionName(parentConnection)
	if err != nil {
		return err
	}

	// The repository ID defaults to metadata.name -- identical to the
	// Terraform module.
	repositoryId := spec.RepositoryId
	if repositoryId == "" {
		repositoryId = resourceName
	}

	args := &cloudbuildv2.RepositoryArgs{
		Project:          pulumi.StringPtr(project),
		Location:         pulumi.StringPtr(location),
		ParentConnection: pulumi.String(parentConnection),
		Name:             pulumi.StringPtr(repositoryId),
		RemoteUri:        pulumi.String(spec.RemoteUri),
	}
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringMap(spec.Annotations)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := cloudbuildv2.NewRepository(ctx, resourceName, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create repository link")
	}

	ctx.Export(OpName, created.ID())
	ctx.Export(OpRepositoryId, created.Name)
	ctx.Export(OpRemoteUri, created.RemoteUri)
	return nil
}

// parseConnectionName returns the project and location of a connection's
// full name ("projects/{p}/locations/{l}/connections/{c}") -- the same
// split as the Terraform module.
func parseConnectionName(name string) (string, string, error) {
	segments := strings.Split(name, "/")
	if len(segments) != 6 || segments[0] != "projects" || segments[2] != "locations" || segments[4] != "connections" {
		return "", "", errors.Errorf("parent_connection %q is not projects/{project}/locations/{location}/connections/{connection}", name)
	}
	return segments[1], segments[3], nil
}
