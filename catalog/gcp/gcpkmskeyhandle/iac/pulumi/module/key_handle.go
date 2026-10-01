package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/kms"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// keyHandle requests a key from Autokey (on for the project or a folder
// above it) for one resource type and location. Every argument is
// immutable; destroy only removes the handle from state -- Google keeps
// it, and the key keeps protecting its resources.
func keyHandle(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpKmsKeyHandle.Spec

	// An empty project means the provider's default project -- the
	// Terraform module's google_client_config twin.
	project := strings.TrimPrefix(spec.GetProjectId().GetValue(), "projects/")
	if project == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to resolve the provider's default project for the key handle")
		}
		if clientConfig.Project == "" {
			return errors.New("the key handle names no project and the provider has no default project -- set spec.project_id or configure a project")
		}
		project = clientConfig.Project
	}

	// Same-project key storage creates the key in the handle's own project
	// through the Cloud KMS API, so the API must be on there even when
	// Autokey was switched on at the folder. disable_on_destroy is false:
	// the key outlives the handle.
	api, err := projects.NewService(ctx, "gcpkmsh-cloudkms.googleapis.com", &projects.ServiceArgs{
		Project:                  pulumi.String(project),
		Service:                  pulumi.String("cloudkms.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable cloudkms.googleapis.com")
	}

	keyHandleName := spec.KeyHandleName
	if keyHandleName == "" {
		keyHandleName = locals.GcpKmsKeyHandle.Metadata.Name
	}

	created, err := kms.NewKeyHandle(ctx, locals.GcpKmsKeyHandle.Metadata.Name, &kms.KeyHandleArgs{
		Project:              pulumi.StringPtr(project),
		Location:             pulumi.String(spec.Location),
		Name:                 pulumi.StringPtr(keyHandleName),
		ResourceTypeSelector: pulumi.String(spec.ResourceTypeSelector),
	}, pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{api}))
	if err != nil {
		return errors.Wrap(err, "failed to create key handle")
	}

	ctx.Export(OpName, created.ID())
	ctx.Export(OpKmsKey, created.KmsKey)
	return nil
}
