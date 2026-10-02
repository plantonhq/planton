package module

import (
	"strings"

	"github.com/pkg/errors"
	gcpsccnotificationconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccnotificationconfig/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/securitycenter"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// notificationConfig streams findings to Pub/Sub at whichever scope the
// spec selects -- one kind, three provider resources, exactly one created:
//
//	scope empty / scope.project_id -> securitycenter.V2ProjectNotificationConfig
//	scope.folder_id                -> securitycenter.V2FolderNotificationConfig
//	scope.organization_id          -> securitycenter.V2OrganizationNotificationConfig
//
// streaming_config is always sent: Google requires it, and an empty filter
// streams every finding.
func notificationConfig(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	target := locals.GcpSccNotificationConfig
	spec := target.Spec
	scope := spec.GetScope()

	location := spec.Location
	if location == "" {
		location = "global"
	}
	var description, pubsubTopic, deletionPolicy pulumi.StringPtrInput
	if spec.Description != "" {
		description = pulumi.StringPtr(spec.Description)
	}
	if topic := spec.GetPubsubTopic().GetValue(); topic != "" {
		pubsubTopic = pulumi.StringPtr(topic)
	}
	if spec.DeletionPolicy != "" {
		deletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	switch {
	case scope.GetFolderId().GetValue() != "":
		created, err := securitycenter.NewV2FolderNotificationConfig(ctx, target.Metadata.Name, &securitycenter.V2FolderNotificationConfigArgs{
			Folder:          pulumi.String(strings.TrimPrefix(scope.GetFolderId().GetValue(), "folders/")),
			ConfigId:        pulumi.String(spec.ConfigId),
			Location:        pulumi.StringPtr(location),
			Description:     description,
			PubsubTopic:     pulumi.String(spec.GetPubsubTopic().GetValue()),
			DeletionPolicy:  deletionPolicy,
			StreamingConfig: &securitycenter.V2FolderNotificationConfigStreamingConfigArgs{Filter: pulumi.String(spec.Filter)},
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create folder notification config")
		}
		exportOutputs(ctx, created.Name, created.ServiceAccount)
	case scope.GetOrganizationId() != "":
		created, err := securitycenter.NewV2OrganizationNotificationConfig(ctx, target.Metadata.Name, &securitycenter.V2OrganizationNotificationConfigArgs{
			Organization:    pulumi.String(scope.GetOrganizationId()),
			ConfigId:        pulumi.String(spec.ConfigId),
			Location:        pulumi.StringPtr(location),
			Description:     description,
			PubsubTopic:     pulumi.String(spec.GetPubsubTopic().GetValue()),
			DeletionPolicy:  deletionPolicy,
			StreamingConfig: &securitycenter.V2OrganizationNotificationConfigStreamingConfigArgs{Filter: pulumi.String(spec.Filter)},
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create organization notification config")
		}
		exportOutputs(ctx, created.Name, created.ServiceAccount)
	default:
		project, api, err := projectWithApi(ctx, scope, gcpProvider)
		if err != nil {
			return err
		}
		created, err := securitycenter.NewV2ProjectNotificationConfig(ctx, target.Metadata.Name, &securitycenter.V2ProjectNotificationConfigArgs{
			Project:         pulumi.StringPtr(project),
			ConfigId:        pulumi.String(spec.ConfigId),
			Location:        pulumi.StringPtr(location),
			Description:     description,
			PubsubTopic:     pubsubTopic,
			DeletionPolicy:  deletionPolicy,
			StreamingConfig: &securitycenter.V2ProjectNotificationConfigStreamingConfigArgs{Filter: pulumi.String(spec.Filter)},
		}, pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{api}))
		if err != nil {
			return errors.Wrap(err, "failed to create project notification config")
		}
		exportOutputs(ctx, created.Name, created.ServiceAccount)
	}
	return nil
}

// projectWithApi resolves a project config's project (an empty scope means
// the provider's default project -- the Terraform module's
// google_client_config twin) and enables the Security Command Center API
// there. disable_on_destroy is false: activation, findings, and other
// configs in the project depend on the API.
func projectWithApi(ctx *pulumi.Context, scope *gcpsccnotificationconfigv1alpha1.GcpSccNotificationConfigScope, gcpProvider *gcp.Provider) (string, pulumi.Resource, error) {
	project := strings.TrimPrefix(scope.GetProjectId().GetValue(), "projects/")
	if project == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return "", nil, errors.Wrap(err, "failed to resolve the provider's default project for the config scope")
		}
		if clientConfig.Project == "" {
			return "", nil, errors.New("the config names no scope and the provider has no default project -- set spec.scope or configure a project")
		}
		project = clientConfig.Project
	}
	api, err := projects.NewService(ctx, "gcpsccn-securitycenter.googleapis.com", &projects.ServiceArgs{
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
