package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	gcpsccbigqueryexportv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccbigqueryexport/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/securitycenter"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// bigQueryExport creates the findings export at whichever scope the spec
// selects -- one kind, three provider resources, exactly one created:
//
//	scope empty / scope.project_id -> securitycenter.V2ProjectSccBigQueryExport
//	scope.folder_id                -> securitycenter.V2FolderSccBigQueryExport
//	scope.organization_id          -> securitycenter.V2OrganizationSccBigQueryExport
//
// The organization resource carries name as an argument (Optional, not
// Computed); the module always composes Google's own value so a refresh
// never shows it drifting -- identical to the Terraform module.
func bigQueryExport(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	target := locals.GcpSccBigQueryExport
	spec := target.Spec
	scope := spec.GetScope()

	location := spec.Location
	if location == "" {
		location = "global"
	}
	// Google takes projects/{project}/datasets/{dataset}; a
	// GcpBigQueryDataset reference resolves to the dataset's self link.
	dataset := strings.TrimPrefix(spec.GetDataset().GetValue(), "https://bigquery.googleapis.com/bigquery/v2/")
	var filter, description, deletionPolicy pulumi.StringPtrInput
	if spec.Filter != "" {
		filter = pulumi.StringPtr(spec.Filter)
	}
	if spec.Description != "" {
		description = pulumi.StringPtr(spec.Description)
	}
	if spec.DeletionPolicy != "" {
		deletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	switch {
	case scope.GetFolderId().GetValue() != "":
		created, err := securitycenter.NewV2FolderSccBigQueryExport(ctx, target.Metadata.Name, &securitycenter.V2FolderSccBigQueryExportArgs{
			Folder:           pulumi.String(strings.TrimPrefix(scope.GetFolderId().GetValue(), "folders/")),
			BigQueryExportId: pulumi.String(spec.BigQueryExportId),
			Location:         pulumi.StringPtr(location),
			Dataset:          pulumi.StringPtr(dataset),
			Filter:           filter,
			Description:      description,
			DeletionPolicy:   deletionPolicy,
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create folder BigQuery export")
		}
		ctx.Export(OpName, created.Name)
		ctx.Export(OpPrincipal, created.Principal)
	case scope.GetOrganizationId() != "":
		created, err := securitycenter.NewV2OrganizationSccBigQueryExport(ctx, target.Metadata.Name, &securitycenter.V2OrganizationSccBigQueryExportArgs{
			Name:             pulumi.StringPtr(fmt.Sprintf("organizations/%s/locations/%s/bigQueryExports/%s", scope.GetOrganizationId(), location, spec.BigQueryExportId)),
			Organization:     pulumi.String(scope.GetOrganizationId()),
			BigQueryExportId: pulumi.String(spec.BigQueryExportId),
			Location:         pulumi.StringPtr(location),
			Dataset:          pulumi.StringPtr(dataset),
			Filter:           filter,
			Description:      description,
			DeletionPolicy:   deletionPolicy,
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create organization BigQuery export")
		}
		ctx.Export(OpName, created.Name)
		ctx.Export(OpPrincipal, created.Principal)
	default:
		project, api, err := projectWithApi(ctx, scope, gcpProvider)
		if err != nil {
			return err
		}
		created, err := securitycenter.NewV2ProjectSccBigQueryExport(ctx, target.Metadata.Name, &securitycenter.V2ProjectSccBigQueryExportArgs{
			Project:          pulumi.StringPtr(project),
			BigQueryExportId: pulumi.String(spec.BigQueryExportId),
			Location:         pulumi.StringPtr(location),
			Dataset:          pulumi.StringPtr(dataset),
			Filter:           filter,
			Description:      description,
			DeletionPolicy:   deletionPolicy,
		}, pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{api}))
		if err != nil {
			return errors.Wrap(err, "failed to create project BigQuery export")
		}
		ctx.Export(OpName, created.Name)
		ctx.Export(OpPrincipal, created.Principal)
	}
	return nil
}

// projectWithApi resolves a project export's project (an empty scope means
// the provider's default project -- the Terraform module's
// google_client_config twin) and enables the Security Command Center API
// there. disable_on_destroy is false: activation, findings, and other
// configs in the project depend on the API.
func projectWithApi(ctx *pulumi.Context, scope *gcpsccbigqueryexportv1alpha1.GcpSccBigQueryExportScope, gcpProvider *gcp.Provider) (string, pulumi.Resource, error) {
	project := strings.TrimPrefix(scope.GetProjectId().GetValue(), "projects/")
	if project == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return "", nil, errors.Wrap(err, "failed to resolve the provider's default project for the export scope")
		}
		if clientConfig.Project == "" {
			return "", nil, errors.New("the export names no scope and the provider has no default project -- set spec.scope or configure a project")
		}
		project = clientConfig.Project
	}
	api, err := projects.NewService(ctx, "gcpsccx-securitycenter.googleapis.com", &projects.ServiceArgs{
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
