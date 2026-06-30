package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	gcpprojectv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpproject/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func project(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*organizations.Project, error) {
	var projectId pulumi.StringInput

	// Check if add_suffix is enabled (defaults to false)
	if locals.GcpProject.Spec.GetAddSuffix() {
		// Create a random 3-char suffix when add_suffix is true
		createdRand, err := random.NewRandomString(ctx,
			fmt.Sprintf("%s-suffix", locals.GcpProject.Spec.ProjectId),
			&random.RandomStringArgs{
				Length:  pulumi.Int(3),
				Special: pulumi.Bool(false),
				Numeric: pulumi.Bool(false),
				Upper:   pulumi.Bool(false),
				Lower:   pulumi.Bool(true),
			},
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate random suffix for projectId")
		}

		// Append suffix to the project_id from spec
		projectId = pulumi.All(createdRand.Result).ApplyT(func(args []interface{}) (string, error) {
			suffix := args[0].(string)
			finalId := fmt.Sprintf("%s-%s", locals.GcpProject.Spec.ProjectId, suffix)
			// Ensure we don't exceed 30 chars
			if len(finalId) > 30 {
				finalId = finalId[:30]
			}
			// Remove trailing hyphens if any (GCP disallows ending with '-')
			finalId = strings.TrimRight(finalId, "-")
			return finalId, nil
		}).(pulumi.StringOutput)
	} else {
		// Use project_id directly from spec without suffix
		projectId = pulumi.String(locals.GcpProject.Spec.ProjectId)
	}

	// Determine deletion policy based on delete_protection flag
	// When delete_protection is true, set to "PREVENT" to block project deletion
	deletionPolicy := "DELETE"
	if locals.GcpProject.Spec.GetDeleteProtection() {
		deletionPolicy = "PREVENT"
	}

	projectArgs := &organizations.ProjectArgs{
		Name:              pulumi.String(locals.GcpProject.Metadata.Name),
		ProjectId:         projectId,
		BillingAccount:    pulumi.String(locals.GcpProject.Spec.BillingAccountId),
		Labels:            pulumi.ToStringMap(locals.GcpLabels),
		AutoCreateNetwork: pulumi.Bool(!locals.GcpProject.Spec.GetDisableDefaultNetwork()),
		DeletionPolicy:    pulumi.String(deletionPolicy),
	}

	if locals.GcpProject.Spec.ParentType == gcpprojectv1.GcpProjectParentType_organization {
		projectArgs.OrgId = pulumi.String(locals.GcpProject.Spec.ParentId)
	}
	if locals.GcpProject.Spec.ParentType == gcpprojectv1.GcpProjectParentType_folder {
		projectArgs.FolderId = pulumi.String(locals.GcpProject.Spec.ParentId)
	}

	// Create the GCP project
	// Use the base project_id from spec as the Pulumi resource name for consistency
	createdProject, err := organizations.NewProject(ctx, locals.GcpProject.Spec.ProjectId, projectArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCP project")
	}

	return createdProject, nil
}
