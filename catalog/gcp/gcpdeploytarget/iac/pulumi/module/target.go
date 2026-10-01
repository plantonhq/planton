package module

import (
	"github.com/pkg/errors"
	gcpdeploytargetv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdeploytarget/v1alpha1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/clouddeploy"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// target enables the Cloud Deploy API and creates the deployment target.
// Optional fields are sent only when set and booleans only when true --
// identical to the Terraform module -- so Cloud Deploy's defaults apply.
func target(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDeployTarget.Spec
	resourceName := locals.GcpDeployTarget.Metadata.Name
	project := spec.GetProjectId().GetValue()

	// The Cloud Deploy API. DisableOnDestroy stays false: tearing down one
	// target must never disable the API for every pipeline in the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("clouddeploy.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if project != "" {
		apiArgs.Project = pulumi.String(project)
	}
	createdApi, err := projects.NewService(ctx, "gcpcdtgt-clouddeploy.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable clouddeploy.googleapis.com")
	}

	// The target ID defaults to metadata.name -- identical to the Terraform
	// module.
	targetId := spec.TargetId
	if targetId == "" {
		targetId = resourceName
	}

	args := &clouddeploy.TargetArgs{
		Location:           pulumi.String(spec.Location),
		Name:               pulumi.String(targetId),
		Project:            optionalString(project),
		Description:        optionalString(spec.Description),
		Labels:             pulumi.ToStringMap(locals.GcpLabels),
		RequireApproval:    optionalTrue(spec.RequireApproval),
		DeletionPolicy:     optionalString(spec.DeletionPolicy),
		Gke:                gkeArgs(spec.Gke),
		AnthosCluster:      anthosClusterArgs(spec.AnthosCluster),
		Run:                runArgs(spec.Run),
		MultiTarget:        multiTargetArgs(spec.MultiTarget),
		CustomTarget:       customTargetArgs(spec.CustomTarget),
		AssociatedEntities: associatedEntitiesArgs(spec.AssociatedEntities),
		ExecutionConfigs:   executionConfigsArgs(spec.ExecutionConfigs),
	}
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringMap(spec.Annotations)
	}
	if len(spec.DeployParameters) > 0 {
		args.DeployParameters = pulumi.ToStringMap(spec.DeployParameters)
	}

	createdTarget, err := clouddeploy.NewTarget(ctx, resourceName, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create deploy target")
	}

	ctx.Export(OpName, createdTarget.ID())
	ctx.Export(OpTargetId, createdTarget.TargetId)
	ctx.Export(OpUid, createdTarget.Uid)
	return nil
}

func gkeArgs(gke *gcpdeploytargetv1alpha1.GcpDeployTargetGke) clouddeploy.TargetGkePtrInput {
	if gke == nil {
		return nil
	}
	return &clouddeploy.TargetGkeArgs{
		Cluster:     optionalString(gke.GetCluster().GetValue()),
		InternalIp:  optionalTrue(gke.InternalIp),
		DnsEndpoint: optionalTrue(gke.DnsEndpoint),
		ProxyUrl:    optionalString(gke.ProxyUrl),
	}
}

func anthosClusterArgs(anthos *gcpdeploytargetv1alpha1.GcpDeployTargetAnthosCluster) clouddeploy.TargetAnthosClusterPtrInput {
	if anthos == nil {
		return nil
	}
	return &clouddeploy.TargetAnthosClusterArgs{
		Membership: optionalString(anthos.GetMembership().GetValue()),
	}
}

func runArgs(run *gcpdeploytargetv1alpha1.GcpDeployTargetRun) clouddeploy.TargetRunPtrInput {
	if run == nil {
		return nil
	}
	return &clouddeploy.TargetRunArgs{Location: pulumi.String(run.Location)}
}

func multiTargetArgs(multi *gcpdeploytargetv1alpha1.GcpDeployTargetMultiTarget) clouddeploy.TargetMultiTargetPtrInput {
	if multi == nil {
		return nil
	}
	return &clouddeploy.TargetMultiTargetArgs{TargetIds: stringValues(multi.TargetIds)}
}

func customTargetArgs(custom *gcpdeploytargetv1alpha1.GcpDeployTargetCustomTarget) clouddeploy.TargetCustomTargetPtrInput {
	if custom == nil {
		return nil
	}
	return &clouddeploy.TargetCustomTargetArgs{
		CustomTargetType: pulumi.String(custom.GetCustomTargetType().GetValue()),
	}
}

// associatedEntitiesArgs keeps the entities keyed by entity_id; Google's
// API holds them as a map, so their order never matters.
func associatedEntitiesArgs(entities []*gcpdeploytargetv1alpha1.GcpDeployTargetAssociatedEntity) clouddeploy.TargetAssociatedEntityArrayInput {
	if len(entities) == 0 {
		return nil
	}
	out := clouddeploy.TargetAssociatedEntityArray{}
	for _, entity := range entities {
		gkeClusters := clouddeploy.TargetAssociatedEntityGkeClusterArray{}
		for _, cluster := range entity.GkeClusters {
			gkeClusters = append(gkeClusters, &clouddeploy.TargetAssociatedEntityGkeClusterArgs{
				Cluster:    optionalString(cluster.GetCluster().GetValue()),
				InternalIp: optionalTrue(cluster.InternalIp),
				ProxyUrl:   optionalString(cluster.ProxyUrl),
			})
		}
		anthosClusters := clouddeploy.TargetAssociatedEntityAnthosClusterArray{}
		for _, cluster := range entity.AnthosClusters {
			anthosClusters = append(anthosClusters, &clouddeploy.TargetAssociatedEntityAnthosClusterArgs{
				Membership: optionalString(cluster.GetMembership().GetValue()),
			})
		}
		entityArgs := &clouddeploy.TargetAssociatedEntityArgs{EntityId: pulumi.String(entity.EntityId)}
		if len(gkeClusters) > 0 {
			entityArgs.GkeClusters = gkeClusters
		}
		if len(anthosClusters) > 0 {
			entityArgs.AnthosClusters = anthosClusters
		}
		out = append(out, entityArgs)
	}
	return out
}

func executionConfigsArgs(configs []*gcpdeploytargetv1alpha1.GcpDeployTargetExecutionConfig) clouddeploy.TargetExecutionConfigArrayInput {
	if len(configs) == 0 {
		return nil
	}
	out := clouddeploy.TargetExecutionConfigArray{}
	for _, config := range configs {
		configArgs := &clouddeploy.TargetExecutionConfigArgs{
			Usages:           pulumi.ToStringArray(config.Usages),
			WorkerPool:       optionalString(config.GetWorkerPool().GetValue()),
			ServiceAccount:   optionalString(config.GetServiceAccount().GetValue()),
			ArtifactStorage:  optionalString(config.GetArtifactStorage().GetValue()),
			ExecutionTimeout: optionalString(config.ExecutionTimeout),
			Verbose:          optionalTrue(config.Verbose),
		}
		if pool := config.DefaultPool; pool != nil {
			configArgs.DefaultPool = &clouddeploy.TargetExecutionConfigDefaultPoolArgs{
				ServiceAccount:  optionalString(pool.GetServiceAccount().GetValue()),
				ArtifactStorage: optionalString(pool.GetArtifactStorage().GetValue()),
			}
		}
		if pool := config.PrivatePool; pool != nil {
			configArgs.PrivatePool = &clouddeploy.TargetExecutionConfigPrivatePoolArgs{
				WorkerPool:      pulumi.String(pool.GetWorkerPool().GetValue()),
				ServiceAccount:  optionalString(pool.GetServiceAccount().GetValue()),
				ArtifactStorage: optionalString(pool.GetArtifactStorage().GetValue()),
			}
		}
		out = append(out, configArgs)
	}
	return out
}

// stringValues flattens resolved references to their literal values.
func stringValues(refs []*foreignkeyv1.StringValueOrRef) pulumi.StringArray {
	values := pulumi.StringArray{}
	for _, ref := range refs {
		values = append(values, pulumi.String(ref.GetValue()))
	}
	return values
}

// optionalString sends a string only when it is set -- the Terraform
// module's `x != "" ? x : null`.
func optionalString(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.StringPtr(value)
}

// optionalTrue sends a boolean only when it is true -- the Terraform
// module's `x ? true : null`; false is the provider's default.
func optionalTrue(value bool) pulumi.BoolPtrInput {
	if !value {
		return nil
	}
	return pulumi.BoolPtr(true)
}
