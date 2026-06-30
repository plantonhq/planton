package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func iam(ctx *pulumi.Context, locals *Locals, createdProject *organizations.Project, gcpProvider *gcp.Provider) error {
	// Create IAM roles and bindings
	// Optionally assign an owner IAM member
	if locals.GcpProject.Spec.OwnerMember != "" {
		_, iamErr := projects.NewIAMMember(ctx,
			fmt.Sprintf("%s-owner-binding", locals.GcpProject.Spec.ProjectId),
			&projects.IAMMemberArgs{
				Project: createdProject.ProjectId,
				Role:    pulumi.String("roles/owner"),
				Member:  pulumi.String(locals.GcpProject.Spec.OwnerMember),
			},
			pulumi.Provider(gcpProvider),
			pulumi.DependsOn([]pulumi.Resource{createdProject}),
		)
		if iamErr != nil {
			return errors.Wrap(iamErr, "failed to create owner IAM binding")
		}
	}
	return nil
}
