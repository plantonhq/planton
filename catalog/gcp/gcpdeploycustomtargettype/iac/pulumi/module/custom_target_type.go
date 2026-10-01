package module

import (
	"github.com/pkg/errors"
	gcpdeploycustomtargettypev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdeploycustomtargettype/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/clouddeploy"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// customTargetType enables the Cloud Deploy API and creates the custom
// target type with its custom actions or tasks.
func customTargetType(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDeployCustomTargetType.Spec
	resourceName := locals.GcpDeployCustomTargetType.Metadata.Name
	project := spec.GetProjectId().GetValue()

	// The Cloud Deploy API. DisableOnDestroy stays false: tearing down one
	// custom target type must never disable the API for every pipeline in
	// the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("clouddeploy.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if project != "" {
		apiArgs.Project = pulumi.String(project)
	}
	createdApi, err := projects.NewService(ctx, "gcpcdctt-clouddeploy.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable clouddeploy.googleapis.com")
	}

	// The type ID defaults to metadata.name -- identical to the Terraform
	// module.
	customTargetTypeId := spec.CustomTargetTypeId
	if customTargetTypeId == "" {
		customTargetTypeId = resourceName
	}

	args := &clouddeploy.CustomTargetTypeArgs{
		Project:        optionalString(project),
		Location:       pulumi.String(spec.Location),
		Name:           pulumi.String(customTargetTypeId),
		Description:    optionalString(spec.Description),
		Labels:         pulumi.ToStringMap(locals.GcpLabels),
		DeletionPolicy: optionalString(spec.DeletionPolicy),
		CustomActions:  customActions(spec.CustomActions),
		Tasks:          tasks(spec.Tasks),
	}
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringMap(spec.Annotations)
	}

	createdType, err := clouddeploy.NewCustomTargetType(ctx, resourceName, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create custom target type")
	}

	ctx.Export(OpName, createdType.ID())
	ctx.Export(OpCustomTargetTypeId, createdType.Name)
	ctx.Export(OpUid, createdType.Uid)
	return nil
}

// customActions builds the Skaffold custom-action definition; each module
// carries exactly the one source it declares (the spec enforces it).
func customActions(c *gcpdeploycustomtargettypev1alpha1.GcpDeployCustomTargetTypeCustomActions) clouddeploy.CustomTargetTypeCustomActionsPtrInput {
	if c == nil {
		return nil
	}
	args := &clouddeploy.CustomTargetTypeCustomActionsArgs{
		DeployAction: pulumi.String(c.DeployAction),
		RenderAction: optionalString(c.RenderAction),
	}
	if len(c.IncludeSkaffoldModules) > 0 {
		modules := clouddeploy.CustomTargetTypeCustomActionsIncludeSkaffoldModuleArray{}
		for _, module := range c.IncludeSkaffoldModules {
			moduleArgs := &clouddeploy.CustomTargetTypeCustomActionsIncludeSkaffoldModuleArgs{
				Configs: optionalStrings(module.Configs),
			}
			if g := module.Git; g != nil {
				moduleArgs.Git = &clouddeploy.CustomTargetTypeCustomActionsIncludeSkaffoldModuleGitArgs{
					Repo: pulumi.String(g.Repo),
					Path: optionalString(g.Path),
					Ref:  optionalString(g.Ref),
				}
			}
			if r := module.GoogleCloudBuildRepo; r != nil {
				moduleArgs.GoogleCloudBuildRepo = &clouddeploy.CustomTargetTypeCustomActionsIncludeSkaffoldModuleGoogleCloudBuildRepoArgs{
					Repository: pulumi.String(r.GetRepository().GetValue()),
					Path:       optionalString(r.Path),
					Ref:        optionalString(r.Ref),
				}
			}
			if s := module.GoogleCloudStorage; s != nil {
				moduleArgs.GoogleCloudStorage = &clouddeploy.CustomTargetTypeCustomActionsIncludeSkaffoldModuleGoogleCloudStorageArgs{
					Source: pulumi.String(s.Source),
					Path:   optionalString(s.Path),
				}
			}
			modules = append(modules, moduleArgs)
		}
		args.IncludeSkaffoldModules = modules
	}
	return args
}

// tasks builds the container-task definition: the deploy task always, the
// render task and each container only when declared.
func tasks(t *gcpdeploycustomtargettypev1alpha1.GcpDeployCustomTargetTypeTasks) clouddeploy.CustomTargetTypeTasksPtrInput {
	if t == nil {
		return nil
	}
	deployArgs := &clouddeploy.CustomTargetTypeTasksDeployArgs{}
	if c := t.GetDeploy().GetContainer(); c != nil {
		deployArgs.Container = &clouddeploy.CustomTargetTypeTasksDeployContainerArgs{
			Image:    pulumi.String(c.Image),
			Commands: optionalStrings(c.Command),
			Args:     optionalStrings(c.Args),
			Env:      optionalStringMap(c.Env),
		}
	}
	args := &clouddeploy.CustomTargetTypeTasksArgs{Deploy: deployArgs}
	if r := t.Render; r != nil {
		renderArgs := &clouddeploy.CustomTargetTypeTasksRenderArgs{}
		if c := r.Container; c != nil {
			renderArgs.Container = &clouddeploy.CustomTargetTypeTasksRenderContainerArgs{
				Image:    pulumi.String(c.Image),
				Commands: optionalStrings(c.Command),
				Args:     optionalStrings(c.Args),
				Env:      optionalStringMap(c.Env),
			}
		}
		args.Render = renderArgs
	}
	return args
}

// optionalString returns nil for an empty value so the provider default
// applies -- the Terraform module's `!= "" ? value : null`.
func optionalString(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.StringPtr(value)
}

// optionalStrings returns nil for an empty list -- the Terraform module's
// `length(...) > 0 ? value : null`.
func optionalStrings(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringArray(values)
}

// optionalStringMap returns nil for an empty map -- the Terraform module's
// `length(...) > 0 ? value : null`.
func optionalStringMap(values map[string]string) pulumi.StringMapInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringMap(values)
}
