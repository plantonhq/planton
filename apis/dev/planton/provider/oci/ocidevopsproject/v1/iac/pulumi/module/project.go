package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/devops"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func devopsProject(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciDevopsProject.Spec

	args := &devops.ProjectArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		Name:          pulumi.String(locals.ProjectName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
		NotificationConfig: &devops.ProjectNotificationConfigArgs{
			TopicId: pulumi.String(spec.NotificationTopicId.GetValue()),
		},
	}

	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	project, err := devops.NewProject(ctx, locals.ProjectName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create devops project")
	}

	ctx.Export(OpProjectId, project.ID())
	ctx.Export(OpNamespace, project.Namespace)

	return nil
}
