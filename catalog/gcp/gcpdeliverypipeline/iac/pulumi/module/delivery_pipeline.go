package module

import (
	"github.com/pkg/errors"
	v1 "github.com/plantonhq/planton/catalog/gcp/gcpdeliverypipeline/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	cd "github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/clouddeploy"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// deliveryPipeline enables the Cloud Deploy API and creates the pipeline,
// then its automations.
func deliveryPipeline(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDeliveryPipeline.Spec
	resourceName := locals.GcpDeliveryPipeline.Metadata.Name
	project := spec.GetProjectId().GetValue()

	// The Cloud Deploy API. DisableOnDestroy stays false: tearing down one
	// pipeline must never disable the API for every other pipeline in the
	// project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("clouddeploy.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if project != "" {
		apiArgs.Project = pulumi.String(project)
	}
	createdApi, err := projects.NewService(ctx, "gcpcdpip-clouddeploy.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable clouddeploy.googleapis.com")
	}

	// The pipeline ID defaults to metadata.name -- identical to the
	// Terraform module.
	pipelineId := spec.DeliveryPipelineId
	if pipelineId == "" {
		pipelineId = resourceName
	}

	args := &cd.DeliveryPipelineArgs{
		Project:        optionalString(project),
		Location:       pulumi.String(spec.Location),
		Name:           pulumi.String(pipelineId),
		Description:    optionalString(spec.Description),
		Labels:         pulumi.ToStringMap(locals.GcpLabels),
		Annotations:    stringMap(spec.Annotations),
		Suspended:      optionalTrue(spec.Suspended),
		DeletionPolicy: optionalString(spec.DeletionPolicy),
		SerialPipeline: serialPipelineArgs(spec.SerialPipeline),
	}

	// The provider always deletes with force=true, removing the pipeline's
	// releases, rollouts, and automations with it.
	createdPipeline, err := cd.NewDeliveryPipeline(ctx, resourceName, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create delivery pipeline")
	}

	if err := automations(ctx, locals, gcpProvider, createdPipeline); err != nil {
		return err
	}

	ctx.Export(OpName, createdPipeline.ID())
	ctx.Export(OpDeliveryPipelineId, createdPipeline.Name)
	ctx.Export(OpUid, createdPipeline.Uid)
	return nil
}

// optionalString sends a string only when it is set, so the provider's
// default applies otherwise.
func optionalString(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.StringPtr(value)
}

// optionalTrue sends a boolean only when it is true -- the Terraform
// module's `x ? true : null`.
func optionalTrue(value bool) pulumi.BoolPtrInput {
	if !value {
		return nil
	}
	return pulumi.BoolPtr(true)
}

// stringArray sends a list only when it has entries.
func stringArray(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringArray(values)
}

// stringMap sends a map only when it has entries.
func stringMap(values map[string]string) pulumi.StringMapInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringMap(values)
}

func serialPipelineArgs(serial *v1.GcpDeliveryPipelineSerialPipeline) cd.DeliveryPipelineSerialPipelinePtrInput {
	if serial == nil {
		return nil
	}
	stages := cd.DeliveryPipelineSerialPipelineStageArray{}
	for _, stage := range serial.Stages {
		stageArgs := cd.DeliveryPipelineSerialPipelineStageArgs{
			TargetId: optionalString(stage.GetTargetId().GetValue()),
			Profiles: stringArray(stage.Profiles),
			Strategy: strategyArgs(stage.Strategy),
		}
		if len(stage.DeployParameters) > 0 {
			parameters := cd.DeliveryPipelineSerialPipelineStageDeployParameterArray{}
			for _, p := range stage.DeployParameters {
				parameters = append(parameters, cd.DeliveryPipelineSerialPipelineStageDeployParameterArgs{
					Values:            pulumi.ToStringMap(p.Values),
					MatchTargetLabels: stringMap(p.MatchTargetLabels),
				})
			}
			stageArgs.DeployParameters = parameters
		}
		stages = append(stages, stageArgs)
	}
	pipeline := &cd.DeliveryPipelineSerialPipelineArgs{}
	if len(stages) > 0 {
		pipeline.Stages = stages
	}
	return pipeline
}

func strategyArgs(strategy *v1.GcpDeliveryPipelineStrategy) cd.DeliveryPipelineSerialPipelineStageStrategyPtrInput {
	if strategy == nil {
		return nil
	}
	return &cd.DeliveryPipelineSerialPipelineStageStrategyArgs{
		Standard: standardArgs(strategy.Standard),
		Canary:   canaryArgs(strategy.Canary),
	}
}

// containerTask is one task's container, read once and shaped into each
// path's own Pulumi type below (the provider's schema repeats the same
// container block under every job, and the SDK gives each its own type).
type containerTask struct {
	image   pulumi.StringInput
	command pulumi.StringArrayInput
	args    pulumi.StringArrayInput
	env     pulumi.StringMapInput
}

func readContainer(task *v1.GcpDeliveryPipelineTask) *containerTask {
	c := task.GetContainer()
	if c == nil {
		return nil
	}
	return &containerTask{
		image:   pulumi.String(c.Image),
		command: stringArray(c.Command),
		args:    stringArray(c.Args),
		env:     stringMap(c.Env),
	}
}

// --- standard strategy -----------------------------------------------------

func standardArgs(standard *v1.GcpDeliveryPipelineStandard) cd.DeliveryPipelineSerialPipelineStageStrategyStandardPtrInput {
	if standard == nil {
		return nil
	}
	args := &cd.DeliveryPipelineSerialPipelineStageStrategyStandardArgs{
		Verify: optionalTrue(standard.Verify),
	}
	if p := standard.Predeploy; p != nil {
		hook := &cd.DeliveryPipelineSerialPipelineStageStrategyStandardPredeployArgs{Actions: stringArray(p.Actions)}
		if len(p.Tasks) > 0 {
			tasks := cd.DeliveryPipelineSerialPipelineStageStrategyStandardPredeployTaskArray{}
			for _, task := range p.Tasks {
				taskArgs := cd.DeliveryPipelineSerialPipelineStageStrategyStandardPredeployTaskArgs{}
				if c := readContainer(task); c != nil {
					taskArgs.Container = &cd.DeliveryPipelineSerialPipelineStageStrategyStandardPredeployTaskContainerArgs{
						Image: c.image, Commands: c.command, Args: c.args, Env: c.env,
					}
				}
				tasks = append(tasks, taskArgs)
			}
			hook.Tasks = tasks
		}
		args.Predeploy = hook
	}
	if p := standard.Postdeploy; p != nil {
		hook := &cd.DeliveryPipelineSerialPipelineStageStrategyStandardPostdeployArgs{Actions: stringArray(p.Actions)}
		if len(p.Tasks) > 0 {
			tasks := cd.DeliveryPipelineSerialPipelineStageStrategyStandardPostdeployTaskArray{}
			for _, task := range p.Tasks {
				taskArgs := cd.DeliveryPipelineSerialPipelineStageStrategyStandardPostdeployTaskArgs{}
				if c := readContainer(task); c != nil {
					taskArgs.Container = &cd.DeliveryPipelineSerialPipelineStageStrategyStandardPostdeployTaskContainerArgs{
						Image: c.image, Commands: c.command, Args: c.args, Env: c.env,
					}
				}
				tasks = append(tasks, taskArgs)
			}
			hook.Tasks = tasks
		}
		args.Postdeploy = hook
	}
	if v := standard.VerifyConfig; v != nil {
		verify := &cd.DeliveryPipelineSerialPipelineStageStrategyStandardVerifyConfigArgs{}
		if len(v.Tasks) > 0 {
			tasks := cd.DeliveryPipelineSerialPipelineStageStrategyStandardVerifyConfigTaskArray{}
			for _, task := range v.Tasks {
				taskArgs := cd.DeliveryPipelineSerialPipelineStageStrategyStandardVerifyConfigTaskArgs{}
				if c := readContainer(task); c != nil {
					taskArgs.Container = &cd.DeliveryPipelineSerialPipelineStageStrategyStandardVerifyConfigTaskContainerArgs{
						Image: c.image, Commands: c.command, Args: c.args, Env: c.env,
					}
				}
				tasks = append(tasks, taskArgs)
			}
			verify.Tasks = tasks
		}
		args.VerifyConfig = verify
	}
	if a := standard.Analysis; a != nil {
		analysis := &cd.DeliveryPipelineSerialPipelineStageStrategyStandardAnalysisArgs{Duration: pulumi.String(a.Duration)}
		if g := a.GoogleCloud; g != nil {
			googleCloud := &cd.DeliveryPipelineSerialPipelineStageStrategyStandardAnalysisGoogleCloudArgs{}
			if len(g.AlertPolicyChecks) > 0 {
				checks := cd.DeliveryPipelineSerialPipelineStageStrategyStandardAnalysisGoogleCloudAlertPolicyCheckArray{}
				for _, check := range g.AlertPolicyChecks {
					checks = append(checks, cd.DeliveryPipelineSerialPipelineStageStrategyStandardAnalysisGoogleCloudAlertPolicyCheckArgs{
						Id: pulumi.String(check.Id), AlertPolicies: alertPolicies(check), Labels: stringMap(check.Labels),
					})
				}
				googleCloud.AlertPolicyChecks = checks
			}
			analysis.GoogleCloud = googleCloud
		}
		if len(a.CustomChecks) > 0 {
			checks := cd.DeliveryPipelineSerialPipelineStageStrategyStandardAnalysisCustomCheckArray{}
			for _, check := range a.CustomChecks {
				checkArgs := cd.DeliveryPipelineSerialPipelineStageStrategyStandardAnalysisCustomCheckArgs{
					Id: pulumi.String(check.Id), Frequency: optionalString(check.Frequency),
				}
				if check.Task != nil {
					taskArgs := &cd.DeliveryPipelineSerialPipelineStageStrategyStandardAnalysisCustomCheckTaskArgs{}
					if c := readContainer(check.Task); c != nil {
						taskArgs.Container = &cd.DeliveryPipelineSerialPipelineStageStrategyStandardAnalysisCustomCheckTaskContainerArgs{
							Image: c.image, Commands: c.command, Args: c.args, Env: c.env,
						}
					}
					checkArgs.Task = taskArgs
				}
				checks = append(checks, checkArgs)
			}
			analysis.CustomChecks = checks
		}
		args.Analysis = analysis
	}
	return args
}

// alertPolicies flattens an alert-policy check's policy names.
func alertPolicies(check *v1.GcpDeliveryPipelineAlertPolicyCheck) pulumi.StringArrayInput {
	names := make([]string, 0, len(check.AlertPolicies))
	for _, policy := range check.AlertPolicies {
		names = append(names, policy.GetValue())
	}
	return pulumi.ToStringArray(names)
}

// --- canary strategy -------------------------------------------------------

func canaryArgs(canary *v1.GcpDeliveryPipelineCanary) cd.DeliveryPipelineSerialPipelineStageStrategyCanaryPtrInput {
	if canary == nil {
		return nil
	}
	return &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryArgs{
		CanaryDeployment:       canaryDeploymentArgs(canary.CanaryDeployment),
		CustomCanaryDeployment: customCanaryDeploymentArgs(canary.CustomCanaryDeployment),
		RuntimeConfig:          runtimeConfigArgs(canary.RuntimeConfig),
	}
}

// canaryDeploymentArgs shapes the same-every-phase canary. Its predeploy
// and postdeploy take actions only: the provider declares no tasks on the
// canary paths.
func canaryDeploymentArgs(deployment *v1.GcpDeliveryPipelineCanaryDeployment) cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentPtrInput {
	if deployment == nil {
		return nil
	}
	percentages := make([]int, 0, len(deployment.Percentages))
	for _, p := range deployment.Percentages {
		percentages = append(percentages, int(p))
	}
	args := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentArgs{
		Percentages: pulumi.ToIntArray(percentages),
		Verify:      optionalTrue(deployment.Verify),
	}
	if p := deployment.Predeploy; p != nil {
		args.Predeploy = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentPredeployArgs{Actions: stringArray(p.Actions)}
	}
	if p := deployment.Postdeploy; p != nil {
		args.Postdeploy = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentPostdeployArgs{Actions: stringArray(p.Actions)}
	}
	if v := deployment.VerifyConfig; v != nil {
		verify := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentVerifyConfigArgs{}
		if len(v.Tasks) > 0 {
			tasks := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentVerifyConfigTaskArray{}
			for _, task := range v.Tasks {
				taskArgs := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentVerifyConfigTaskArgs{}
				if c := readContainer(task); c != nil {
					taskArgs.Container = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentVerifyConfigTaskContainerArgs{
						Image: c.image, Commands: c.command, Args: c.args, Env: c.env,
					}
				}
				tasks = append(tasks, taskArgs)
			}
			verify.Tasks = tasks
		}
		args.VerifyConfig = verify
	}
	if a := deployment.Analysis; a != nil {
		analysis := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentAnalysisArgs{Duration: pulumi.String(a.Duration)}
		if g := a.GoogleCloud; g != nil {
			googleCloud := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentAnalysisGoogleCloudArgs{}
			if len(g.AlertPolicyChecks) > 0 {
				checks := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentAnalysisGoogleCloudAlertPolicyCheckArray{}
				for _, check := range g.AlertPolicyChecks {
					checks = append(checks, cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentAnalysisGoogleCloudAlertPolicyCheckArgs{
						Id: pulumi.String(check.Id), AlertPolicies: alertPolicies(check), Labels: stringMap(check.Labels),
					})
				}
				googleCloud.AlertPolicyChecks = checks
			}
			analysis.GoogleCloud = googleCloud
		}
		if len(a.CustomChecks) > 0 {
			checks := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentAnalysisCustomCheckArray{}
			for _, check := range a.CustomChecks {
				checkArgs := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentAnalysisCustomCheckArgs{
					Id: pulumi.String(check.Id), Frequency: optionalString(check.Frequency),
				}
				if check.Task != nil {
					taskArgs := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentAnalysisCustomCheckTaskArgs{}
					if c := readContainer(check.Task); c != nil {
						taskArgs.Container = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCanaryDeploymentAnalysisCustomCheckTaskContainerArgs{
							Image: c.image, Commands: c.command, Args: c.args, Env: c.env,
						}
					}
					checkArgs.Task = taskArgs
				}
				checks = append(checks, checkArgs)
			}
			analysis.CustomChecks = checks
		}
		args.Analysis = analysis
	}
	return args
}

// customCanaryDeploymentArgs shapes the phase-by-phase canary; each
// phase's percentage is always sent.
func customCanaryDeploymentArgs(deployment *v1.GcpDeliveryPipelineCustomCanaryDeployment) cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPtrInput {
	if deployment == nil {
		return nil
	}
	phases := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigArray{}
	for _, phase := range deployment.PhaseConfigs {
		phaseArgs := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigArgs{
			PhaseId:    pulumi.String(phase.PhaseId),
			Percentage: pulumi.Int(int(phase.Percentage)),
			Profiles:   stringArray(phase.Profiles),
			Verify:     optionalTrue(phase.Verify),
		}
		if p := phase.Predeploy; p != nil {
			phaseArgs.Predeploy = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigPredeployArgs{Actions: stringArray(p.Actions)}
		}
		if p := phase.Postdeploy; p != nil {
			phaseArgs.Postdeploy = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigPostdeployArgs{Actions: stringArray(p.Actions)}
		}
		if v := phase.VerifyConfig; v != nil {
			verify := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigVerifyConfigArgs{}
			if len(v.Tasks) > 0 {
				tasks := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigVerifyConfigTaskArray{}
				for _, task := range v.Tasks {
					taskArgs := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigVerifyConfigTaskArgs{}
					if c := readContainer(task); c != nil {
						taskArgs.Container = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigVerifyConfigTaskContainerArgs{
							Image: c.image, Commands: c.command, Args: c.args, Env: c.env,
						}
					}
					tasks = append(tasks, taskArgs)
				}
				verify.Tasks = tasks
			}
			phaseArgs.VerifyConfig = verify
		}
		if a := phase.Analysis; a != nil {
			analysis := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigAnalysisArgs{Duration: pulumi.String(a.Duration)}
			if g := a.GoogleCloud; g != nil {
				googleCloud := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigAnalysisGoogleCloudArgs{}
				if len(g.AlertPolicyChecks) > 0 {
					checks := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigAnalysisGoogleCloudAlertPolicyCheckArray{}
					for _, check := range g.AlertPolicyChecks {
						checks = append(checks, cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigAnalysisGoogleCloudAlertPolicyCheckArgs{
							Id: pulumi.String(check.Id), AlertPolicies: alertPolicies(check), Labels: stringMap(check.Labels),
						})
					}
					googleCloud.AlertPolicyChecks = checks
				}
				analysis.GoogleCloud = googleCloud
			}
			if len(a.CustomChecks) > 0 {
				checks := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigAnalysisCustomCheckArray{}
				for _, check := range a.CustomChecks {
					checkArgs := cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigAnalysisCustomCheckArgs{
						Id: pulumi.String(check.Id), Frequency: optionalString(check.Frequency),
					}
					if check.Task != nil {
						taskArgs := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigAnalysisCustomCheckTaskArgs{}
						if c := readContainer(check.Task); c != nil {
							taskArgs.Container = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentPhaseConfigAnalysisCustomCheckTaskContainerArgs{
								Image: c.image, Commands: c.command, Args: c.args, Env: c.env,
							}
						}
						checkArgs.Task = taskArgs
					}
					checks = append(checks, checkArgs)
				}
				analysis.CustomChecks = checks
			}
			phaseArgs.Analysis = analysis
		}
		phases = append(phases, phaseArgs)
	}
	return &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryCustomCanaryDeploymentArgs{PhaseConfigs: phases}
}

func runtimeConfigArgs(runtime *v1.GcpDeliveryPipelineRuntimeConfig) cd.DeliveryPipelineSerialPipelineStageStrategyCanaryRuntimeConfigPtrInput {
	if runtime == nil {
		return nil
	}
	args := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryRuntimeConfigArgs{}
	if r := runtime.CloudRun; r != nil {
		args.CloudRun = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryRuntimeConfigCloudRunArgs{
			AutomaticTrafficControl: optionalTrue(r.AutomaticTrafficControl),
			CanaryRevisionTags:      stringArray(r.CanaryRevisionTags),
			PriorRevisionTags:       stringArray(r.PriorRevisionTags),
			StableRevisionTags:      stringArray(r.StableRevisionTags),
		}
	}
	if k := runtime.Kubernetes; k != nil {
		kubernetes := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryRuntimeConfigKubernetesArgs{}
		if g := k.GatewayServiceMesh; g != nil {
			mesh := &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryRuntimeConfigKubernetesGatewayServiceMeshArgs{
				HttpRoute:             pulumi.String(g.HttpRoute),
				Service:               pulumi.String(g.Service),
				Deployment:            pulumi.String(g.Deployment),
				RouteUpdateWaitTime:   optionalString(g.RouteUpdateWaitTime),
				StableCutbackDuration: optionalString(g.StableCutbackDuration),
				PodSelectorLabel:      optionalString(g.PodSelectorLabel),
			}
			if d := g.RouteDestinations; d != nil {
				mesh.RouteDestinations = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryRuntimeConfigKubernetesGatewayServiceMeshRouteDestinationsArgs{
					DestinationIds:   pulumi.ToStringArray(d.DestinationIds),
					PropagateService: optionalTrue(d.PropagateService),
				}
			}
			kubernetes.GatewayServiceMesh = mesh
		}
		if s := k.ServiceNetworking; s != nil {
			kubernetes.ServiceNetworking = &cd.DeliveryPipelineSerialPipelineStageStrategyCanaryRuntimeConfigKubernetesServiceNetworkingArgs{
				Service:                    pulumi.String(s.Service),
				Deployment:                 pulumi.String(s.Deployment),
				DisablePodOverprovisioning: optionalTrue(s.DisablePodOverprovisioning),
				PodSelectorLabel:           optionalString(s.PodSelectorLabel),
			}
		}
		args.Kubernetes = kubernetes
	}
	return args
}
