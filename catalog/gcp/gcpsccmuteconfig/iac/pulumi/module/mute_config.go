package module

import (
	"strings"

	"github.com/pkg/errors"
	gcpsccmuteconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccmuteconfig/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/securitycenter"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// muteConfig creates the mute rule at whichever scope the spec selects --
// one kind, three provider resources, exactly one created:
//
//	scope empty / scope.project_id -> securitycenter.V2ProjectMuteConfig
//	scope.folder_id                -> securitycenter.V2FolderMuteConfig
//	scope.organization_id          -> securitycenter.V2OrganizationMuteConfig
func muteConfig(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	target := locals.GcpSccMuteConfig
	spec := target.Spec
	scope := spec.GetScope()

	location := spec.Location
	if location == "" {
		location = "global"
	}
	var description, deletionPolicy pulumi.StringPtrInput
	if spec.Description != "" {
		description = pulumi.StringPtr(spec.Description)
	}
	if spec.DeletionPolicy != "" {
		deletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	switch {
	case scope.GetFolderId().GetValue() != "":
		created, err := securitycenter.NewV2FolderMuteConfig(ctx, target.Metadata.Name, &securitycenter.V2FolderMuteConfigArgs{
			Folder:         pulumi.String(strings.TrimPrefix(scope.GetFolderId().GetValue(), "folders/")),
			MuteConfigId:   pulumi.String(spec.MuteConfigId),
			Location:       pulumi.StringPtr(location),
			Filter:         pulumi.String(spec.Filter),
			Type:           pulumi.String(spec.Type),
			Description:    description,
			DeletionPolicy: deletionPolicy,
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create folder mute config")
		}
		ctx.Export(OpName, created.Name)
	case scope.GetOrganizationId() != "":
		created, err := securitycenter.NewV2OrganizationMuteConfig(ctx, target.Metadata.Name, &securitycenter.V2OrganizationMuteConfigArgs{
			Organization:   pulumi.String(scope.GetOrganizationId()),
			MuteConfigId:   pulumi.String(spec.MuteConfigId),
			Location:       pulumi.StringPtr(location),
			Filter:         pulumi.String(spec.Filter),
			Type:           pulumi.String(spec.Type),
			Description:    description,
			DeletionPolicy: deletionPolicy,
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create organization mute config")
		}
		ctx.Export(OpName, created.Name)
	default:
		project, api, err := projectWithApi(ctx, scope, gcpProvider)
		if err != nil {
			return err
		}
		created, err := securitycenter.NewV2ProjectMuteConfig(ctx, target.Metadata.Name, &securitycenter.V2ProjectMuteConfigArgs{
			Project:        pulumi.StringPtr(project),
			MuteConfigId:   pulumi.String(spec.MuteConfigId),
			Location:       pulumi.StringPtr(location),
			Filter:         pulumi.String(spec.Filter),
			Type:           pulumi.String(spec.Type),
			Description:    description,
			DeletionPolicy: deletionPolicy,
		}, pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{api}))
		if err != nil {
			return errors.Wrap(err, "failed to create project mute config")
		}
		ctx.Export(OpName, created.Name)
	}
	return nil
}

// projectWithApi resolves a project rule's project (an empty scope means
// the provider's default project -- the Terraform module's
// google_client_config twin) and enables the Security Command Center API
// there. disable_on_destroy is false: activation, findings, and other
// configs in the project depend on the API.
func projectWithApi(ctx *pulumi.Context, scope *gcpsccmuteconfigv1alpha1.GcpSccMuteConfigScope, gcpProvider *gcp.Provider) (string, pulumi.Resource, error) {
	project := strings.TrimPrefix(scope.GetProjectId().GetValue(), "projects/")
	if project == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return "", nil, errors.Wrap(err, "failed to resolve the provider's default project for the rule scope")
		}
		if clientConfig.Project == "" {
			return "", nil, errors.New("the rule names no scope and the provider has no default project -- set spec.scope or configure a project")
		}
		project = clientConfig.Project
	}
	api, err := projects.NewService(ctx, "gcpsccm-securitycenter.googleapis.com", &projects.ServiceArgs{
		Project:                  pulumi.String(project),
		Service:                  pulumi.String("securitycenter.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}, pulumi.Provider(gcpProvider))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to enable securitycenter.googleapis.com api")
	}
	return project, api, nil
}
