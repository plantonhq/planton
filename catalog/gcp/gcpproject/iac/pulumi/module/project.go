package module

import (
	"strings"

	"github.com/pkg/errors"
	gcpprojectv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpproject/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// project provisions the Google Cloud project — the Layer-0 container
// every other GCP resource lives in. IAM grants are deliberately NOT
// bundled here; model them as first-class GcpProjectIamMember resources.
func project(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*organizations.Project, error) {
	spec := locals.GcpProject.Spec

	projectArgs := &organizations.ProjectArgs{
		Name:      pulumi.String(locals.DisplayName),
		ProjectId: pulumi.String(spec.ProjectId),
		Labels:    pulumi.ToStringMap(locals.GcpLabels),
		// False by default: deleting the auto-created "default" network is
		// a standard hardening step, and explicit GcpVpcNetwork resources
		// are the composable path.
		AutoCreateNetwork: pulumi.Bool(spec.GetAutoCreateNetwork()),
		DeletionPolicy:    pulumi.String(locals.DeletionPolicy),
	}

	if spec.BillingAccountId != "" {
		projectArgs.BillingAccount = pulumi.StringPtr(spec.BillingAccountId)
	}

	// Resource Manager tags bind at create time only; changing them
	// afterwards forces recreation (bind tag values to an existing project
	// with GcpTagBinding instead).
	if len(spec.Tags) > 0 {
		projectArgs.Tags = pulumi.ToStringMap(spec.Tags)
	}

	// Exactly one of org_id / folder_id is sent. A folder_id reference (a
	// GcpFolder's folder_id output, or a literal folder id) IS the parent
	// and wins; otherwise parent_type selects which argument parent_id
	// fills. The spec's CEL keeps the two forms from being combined. The
	// same rule lives in the Terraform module's locals.tf.
	if folder := strings.TrimPrefix(spec.FolderId.GetValue(), "folders/"); folder != "" {
		projectArgs.FolderId = pulumi.String(folder)
	} else if spec.ParentType == gcpprojectv1alpha1.GcpProjectParentType_organization {
		projectArgs.OrgId = pulumi.String(spec.ParentId)
	} else if spec.ParentType == gcpprojectv1alpha1.GcpProjectParentType_folder {
		projectArgs.FolderId = pulumi.String(spec.ParentId)
	}

	createdProject, err := organizations.NewProject(ctx, spec.ProjectId, projectArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCP project")
	}

	return createdProject, nil
}
