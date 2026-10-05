package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The runner's listening ports: the gRPC/CloudOps server (the container
// port Cloud Run routes to) and the health server the startup probe reads.
// Nothing meaningful is reachable through the service's URL — the runner
// initiates every connection it uses, and with no run.invoker grants the
// default authenticated-only posture means nothing can call it anyway.
const (
	grpcPort   = 50051
	healthPort = 8093
)

// runnerService provisions the Cloud Run v2 service that keeps exactly one
// runner running, and exports the component's outputs.
func runnerService(
	ctx *pulumi.Context,
	locals *Locals,
	provider *gcp.Provider,
	serviceAccountEmail pulumi.StringOutput,
	createdServiceAccount pulumi.Resource,
	createdSecretVersion pulumi.Resource,
	createdAccessorGrant pulumi.Resource,
) error {
	spec := locals.GcpPlantonRunner.Spec
	runnerName := locals.GcpPlantonRunner.Metadata.Name

	// Enable the Cloud Run Admin API first so a fresh project works on the
	// first deploy. disable_on_destroy stays false: tearing down one runner
	// must never disable the API for everything else in the project.
	runApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("run.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	// An empty project falls back to the provider's default project — the
	// ambient-project contract every GCP kind honors.
	if locals.ProjectId != "" {
		runApiArgs.Project = pulumi.String(locals.ProjectId)
	}
	createdRunApi, err := projects.NewService(ctx,
		"runner-run.googleapis.com", runApiArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to enable run.googleapis.com api")
	}

	// The runner's environment contract: the token via a Secret Manager
	// reference (resolved at instance start through the runtime service
	// account — never plaintext here), the name it registers itself under,
	// and the control-plane endpoint when one is declared (omitted, the
	// runner's built-in hosted default applies). No EXECUTION_MODE: only
	// the control plane knows whether its work queue is reachable, so the
	// runner derives its mode from the identity the join returns — a mode
	// knob here would silently strip capability.
	envs := cloudrunv2.ServiceTemplateContainerEnvArray{
		&cloudrunv2.ServiceTemplateContainerEnvArgs{
			Name: pulumi.String("PLANTON_RUNNER_TOKEN"),
			ValueSource: &cloudrunv2.ServiceTemplateContainerEnvValueSourceArgs{
				SecretKeyRef: &cloudrunv2.ServiceTemplateContainerEnvValueSourceSecretKeyRefArgs{
					Secret: pulumi.String(locals.TokenSecretId),
					// "latest" deliberately: token rotation needs no
					// service update (see secret.go).
					Version: pulumi.String("latest"),
				},
			},
		},
		&cloudrunv2.ServiceTemplateContainerEnvArgs{
			Name:  pulumi.String("PLANTON_RUNNER_NAME"),
			Value: pulumi.String(locals.RunnerName),
		},
	}
	if spec.GetControlPlaneEndpoint() != "" {
		envs = append(envs, &cloudrunv2.ServiceTemplateContainerEnvArgs{
			Name:  pulumi.String("PLANTON_RUNNER_ENDPOINT"),
			Value: pulumi.String(spec.GetControlPlaneEndpoint()),
		})
	}
	// NO explicit PORT env: Cloud Run reserves the name and rejects the whole
	// create with a 400 when a template sets it (proven live). The platform
	// injects PORT itself from the ports block's ContainerPort, which is how
	// the runner's server learns its port here.
	envs = append(envs,
		&cloudrunv2.ServiceTemplateContainerEnvArgs{
			Name:  pulumi.String("LOG_LEVEL"),
			Value: pulumi.String("info"),
		},
	)

	container := &cloudrunv2.ServiceTemplateContainerArgs{
		Image: pulumi.Sprintf("%s:%s", spec.GetImageRepository(), spec.GetRunnerVersion()),
		// Args, NEVER Commands: Cloud Run's command field REPLACES the image
		// entrypoint (the runner binary), so Commands=["start"] makes the
		// platform exec a binary literally named "start" -- the instance dies
		// with "Application exec likely failed" before one log line (proven
		// live). The image's entrypoint is the bare runner binary; the
		// subcommand rides Args.
		Args: pulumi.ToStringArray([]string{"start"}),
		// h2c: the runner's server speaks plaintext HTTP/2 (gRPC) behind
		// Cloud Run's TLS edge.
		Ports: &cloudrunv2.ServiceTemplateContainerPortsArgs{
			Name:          pulumi.String("h2c"),
			ContainerPort: pulumi.Int(grpcPort),
		},
		Envs: envs,
		Resources: &cloudrunv2.ServiceTemplateContainerResourcesArgs{
			Limits: pulumi.StringMap{
				"cpu":    pulumi.String(spec.GetCpu()),
				"memory": pulumi.String(spec.GetMemory()),
			},
			// CPU stays allocated between requests: the runner is a
			// PULL-based worker — it polls its work queue and executes
			// long-running IaC operations with no inbound request in
			// flight, so throttled-between-requests CPU (the Cloud Run
			// default) would freeze it mid-operation.
			CpuIdle: pulumi.Bool(false),
			// Faster cold start for the one instance replacement ever in
			// flight.
			StartupCpuBoost: pulumi.Bool(true),
		},
		// The health server answers independently of control-plane
		// reachability, so a runner whose control plane is momentarily
		// unreachable still starts — its readiness contract is the work
		// queue, not the probe.
		StartupProbe: &cloudrunv2.ServiceTemplateContainerStartupProbeArgs{
			InitialDelaySeconds: pulumi.Int(5),
			PeriodSeconds:       pulumi.Int(10),
			FailureThreshold:    pulumi.Int(3),
			TimeoutSeconds:      pulumi.Int(5),
			HttpGet: &cloudrunv2.ServiceTemplateContainerStartupProbeHttpGetArgs{
				Path: pulumi.String("/healthz"),
				Port: pulumi.Int(healthPort),
			},
		},
	}

	template := &cloudrunv2.ServiceTemplateArgs{
		Containers: cloudrunv2.ServiceTemplateContainerArray{container},
		// Exactly one runner per enrollment: a runner's identity is minted
		// for ONE live instance — a second instance joining under the same
		// name would revoke the first's key (token lineage: re-admission
		// re-mints and revokes). Never enable scaling here without
		// redesigning enrollment for fleets. (Revision rollover briefly
		// overlaps two instances; the draining one's revoked key dies with
		// it, and a rollback self-heals by re-joining.)
		Scaling: &cloudrunv2.ServiceTemplateScalingArgs{
			MinInstanceCount: pulumi.Int(1),
			MaxInstanceCount: pulumi.Int(1),
		},
		ServiceAccount: serviceAccountEmail,
		// GEN2 for full Linux compatibility — IaC toolchains the runner
		// executes (tofu/pulumi child processes) assume a complete syscall
		// surface.
		ExecutionEnvironment: pulumi.String("EXECUTION_ENVIRONMENT_GEN2"),
	}

	// Direct VPC egress: only private-range traffic rides the VPC (the
	// route to private endpoints); the runner's control-plane dial-out
	// keeps its normal internet path — so a misconfigured VPC can never
	// sever the runner from the control plane that manages it.
	if spec.VpcAccess != nil {
		networkInterface := &cloudrunv2.ServiceTemplateVpcAccessNetworkInterfaceArgs{
			Network:    pulumi.String(spec.VpcAccess.Network.GetValue()),
			Subnetwork: pulumi.String(spec.VpcAccess.Subnetwork.GetValue()),
		}
		if len(spec.VpcAccess.Tags) > 0 {
			networkInterface.Tags = pulumi.ToStringArray(spec.VpcAccess.Tags)
		}
		template.VpcAccess = &cloudrunv2.ServiceTemplateVpcAccessArgs{
			NetworkInterfaces: cloudrunv2.ServiceTemplateVpcAccessNetworkInterfaceArray{networkInterface},
			Egress:            pulumi.String("PRIVATE_RANGES_ONLY"),
		}
	}

	serviceArgs := &cloudrunv2.ServiceArgs{
		Name:     pulumi.String(runnerName),
		Location: pulumi.String(spec.Region),
		Template: template,
		Labels:   pulumi.ToStringMap(locals.GcpLabels),
		// The appliance's managed teardown IS the destroy path, and the
		// token makes it re-mintable standing infrastructure — the
		// provider's default deletion protection would turn every teardown
		// into a two-step dance for no protective gain.
		DeletionProtection: pulumi.Bool(false),
		// Ingress setting is required by the API; nothing meaningful is
		// reachable either way — the default authenticated-only posture
		// (no run.invoker grants exist) already refuses every caller.
		Ingress: pulumi.String("INGRESS_TRAFFIC_ALL"),
	}
	if locals.ProjectId != "" {
		serviceArgs.Project = pulumi.String(locals.ProjectId)
	}

	// The first instance start resolves the token reference, so the
	// version and the accessor grant must exist before the service — a
	// missing edge here fails at instance START, not at plan.
	dependencies := []pulumi.Resource{createdRunApi, createdSecretVersion, createdAccessorGrant}
	if createdServiceAccount != nil {
		dependencies = append(dependencies, createdServiceAccount)
	}

	createdService, err := cloudrunv2.NewService(ctx,
		"runner-service",
		serviceArgs,
		pulumi.Provider(provider),
		pulumi.DependsOn(dependencies),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create Cloud Run v2 service")
	}

	ctx.Export(OpServiceName, createdService.ID())
	ctx.Export(OpServiceShortName, createdService.Name)
	ctx.Export(OpServiceAccountEmail, serviceAccountEmail)
	ctx.Export(OpTokenSecretId, pulumi.Sprintf("projects/%s/secrets/%s",
		createdService.Project, locals.TokenSecretId))
	ctx.Export(OpRunnerName, pulumi.String(locals.RunnerName))
	ctx.Export(OpProjectId, createdService.Project)
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return nil
}
