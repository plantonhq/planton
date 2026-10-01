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

// autokeyConfig applies the Autokey configuration at the scope the spec
// selects -- a folder (kms.AutokeyConfig) or a project
// (kms.ProjectAutokeyConfig). Exactly one configuration exists per folder
// and per project: create and update are the same PATCH (applying takes
// over an existing configuration), and destroy under DELETE clears it.
func autokeyConfig(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpKmsAutokeyConfig.Spec
	name := locals.GcpKmsAutokeyConfig.Metadata.Name

	// Scope selection. An empty scope means the provider's default
	// project, read from the provider's configuration -- the Terraform
	// module's google_client_config twin.
	folderID := strings.TrimPrefix(spec.GetScope().GetFolderId().GetValue(), "folders/")
	configProject := ""
	if folderID == "" {
		configProject = strings.TrimPrefix(spec.GetScope().GetProjectId().GetValue(), "projects/")
		if configProject == "" {
			clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
			if err != nil {
				return errors.Wrap(err, "failed to resolve the provider's default project for the Autokey scope")
			}
			if clientConfig.Project == "" {
				return errors.New("the configuration names no scope and the provider has no default project -- set spec.scope or configure a project")
			}
			configProject = clientConfig.Project
		}
	}
	keyProject := strings.TrimPrefix(spec.GetKeyProject().GetValue(), "projects/")

	// Autokey creates keys through the Cloud KMS API, so it must be on in
	// the project a project configuration governs and in a folder's key
	// project. disable_on_destroy is false: the keys outlive this block.
	var dependsOn []pulumi.Resource
	for _, project := range []string{configProject, keyProject} {
		if project == "" {
			continue
		}
		api, err := projects.NewService(ctx, "gcpakcfg-cloudkms-"+project, &projects.ServiceArgs{
			Project:                  pulumi.String(project),
			Service:                  pulumi.String("cloudkms.googleapis.com"),
			DisableDependentServices: pulumi.BoolPtr(true),
			DisableOnDestroy:         pulumi.BoolPtr(false),
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to enable cloudkms.googleapis.com on %s", project)
		}
		dependsOn = append(dependsOn, api)
	}

	var resolutionMode, deletionPolicy pulumi.StringPtrInput
	if spec.KeyProjectResolutionMode != "" {
		resolutionMode = pulumi.StringPtr(spec.KeyProjectResolutionMode)
	}
	if spec.DeletionPolicy != "" {
		deletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	if folderID != "" {
		args := &kms.AutokeyConfigArgs{
			Folder:                   pulumi.String(folderID),
			KeyProjectResolutionMode: resolutionMode,
			DeletionPolicy:           deletionPolicy,
		}
		if keyProject != "" {
			args.KeyProject = pulumi.StringPtr("projects/" + keyProject)
		}
		created, err := kms.NewAutokeyConfig(ctx, name, args,
			pulumi.Provider(gcpProvider), pulumi.DependsOn(dependsOn))
		if err != nil {
			return errors.Wrap(err, "failed to apply folder Autokey configuration")
		}
		ctx.Export(OpName, created.ID())
		ctx.Export(OpParent, pulumi.String("folders/"+folderID))
		return nil
	}

	created, err := kms.NewProjectAutokeyConfig(ctx, name, &kms.ProjectAutokeyConfigArgs{
		Project:                  pulumi.StringPtr(configProject),
		KeyProjectResolutionMode: resolutionMode,
		DeletionPolicy:           deletionPolicy,
	}, pulumi.Provider(gcpProvider), pulumi.DependsOn(dependsOn))
	if err != nil {
		return errors.Wrap(err, "failed to apply project Autokey configuration")
	}
	ctx.Export(OpName, created.ID())
	ctx.Export(OpParent, pulumi.String("projects/"+configProject))
	return nil
}
