package module

import (
	"github.com/pkg/errors"
	azureplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The runner's listening ports: the gRPC/CloudOps server and the health
// server the startup probe reads. The app exposes NO ingress at all — the
// runner initiates every connection it uses.
const (
	grpcPort   = 50051
	healthPort = 8093
)

// The runner's Consumption-plan sizing defaults — the spec's documented
// defaults, applied when the platform's defaulting middleware did not
// materialize them into the IaC input.
const (
	defaultCpu    = 0.5
	defaultMemory = "1Gi"
)

// Resources provisions the runner appliance: a single-revision Container
// App inside the referenced environment, pinned to exactly one replica,
// with the runner token in the app's own secret store. The resource group
// and the environment are referenced resources — the module never creates
// or mutates them (the environment decides the network boundary; a
// VNet-integrated one gives the runner reach into private endpoints).
//
// ENROLLMENT IS TOKEN-FIRST: the app ships the runner TOKEN, never an
// identity. The runner joins the control plane on first boot, registers
// itself under RunnerName, and receives its own individually revocable
// identity; replica replacement re-joins with the same token (its lineage
// re-admits the runner it originally admitted).
func Resources(ctx *pulumi.Context, iacInput *azureplantonrunnerv1alpha1.AzurePlantonRunnerIacInput) error {
	locals := initializeLocals(ctx, iacInput)
	spec := locals.AzurePlantonRunner.Spec
	runnerName := locals.AzurePlantonRunner.Metadata.Name

	// Build the Azure provider from the IaC input via the shared
	// builder, which resolves the right credential mechanism (static
	// client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	cpu := defaultCpu
	if spec.Cpu != nil {
		cpu = spec.GetCpu()
	}
	memory := defaultMemory
	if spec.Memory != nil {
		memory = spec.GetMemory()
	}

	// The runner's environment contract: the token via the app's own
	// secret store (referenced by name — never a plain env value; reading
	// the app definition reveals nothing), the name it registers itself
	// under, and the control-plane endpoint when one is declared
	// (omitted, the runner's built-in hosted default applies). No
	// EXECUTION_MODE: only the control plane knows whether its work queue
	// is reachable, so the runner derives its mode from the identity the
	// join returns — a mode knob here would silently strip capability.
	envs := containerapp.AppTemplateContainerEnvArray{
		&containerapp.AppTemplateContainerEnvArgs{
			Name:       pulumi.String("PLANTON_RUNNER_TOKEN"),
			SecretName: pulumi.StringPtr(tokenSecretName),
		},
		&containerapp.AppTemplateContainerEnvArgs{
			Name:  pulumi.String("PLANTON_RUNNER_NAME"),
			Value: pulumi.StringPtr(locals.RunnerName),
		},
	}
	if spec.GetControlPlaneEndpoint() != "" {
		envs = append(envs, &containerapp.AppTemplateContainerEnvArgs{
			Name:  pulumi.String("PLANTON_RUNNER_ENDPOINT"),
			Value: pulumi.StringPtr(spec.GetControlPlaneEndpoint()),
		})
	}
	envs = append(envs,
		&containerapp.AppTemplateContainerEnvArgs{
			Name:  pulumi.String("PORT"),
			Value: pulumi.StringPtr("50051"),
		},
		&containerapp.AppTemplateContainerEnvArgs{
			Name:  pulumi.String("LOG_LEVEL"),
			Value: pulumi.StringPtr("info"),
		},
	)

	container := &containerapp.AppTemplateContainerArgs{
		Name:   pulumi.String("planton-runner"),
		Image:  pulumi.Sprintf("%s:%s", spec.GetImageRepository(), spec.GetRunnerVersion()),
		Cpu:    pulumi.Float64(cpu),
		Memory: pulumi.String(memory),
		// The runner binary's own start command — the image's entrypoint
		// takes the subcommand as args.
		Args: pulumi.ToStringArray([]string{"start"}),
		Envs: envs,
		// The health server answers independently of control-plane
		// reachability, so a runner whose control plane is momentarily
		// unreachable still starts — its readiness contract is the work
		// queue, not the probe.
		StartupProbes: containerapp.AppTemplateContainerStartupProbeArray{
			&containerapp.AppTemplateContainerStartupProbeArgs{
				Transport:             pulumi.String("HTTP"),
				Port:                  pulumi.Int(healthPort),
				Path:                  pulumi.StringPtr("/healthz"),
				InitialDelay:          pulumi.IntPtr(5),
				IntervalSeconds:       pulumi.IntPtr(10),
				FailureCountThreshold: pulumi.IntPtr(3),
				Timeout:               pulumi.IntPtr(5),
			},
		},
	}

	appArgs := &containerapp.AppArgs{
		Name:                      pulumi.String(runnerName),
		ResourceGroupName:         pulumi.String(locals.ResourceGroupName),
		ContainerAppEnvironmentId: pulumi.String(spec.ContainerAppEnvironmentId.GetValue()),
		// Single revision mode: new revisions replace the old one — the
		// only sane model for a workload whose identity contract forbids
		// two live copies. (Revision rollover briefly overlaps two
		// replicas; the draining one's revoked key dies with it, and a
		// rollback self-heals by re-joining.)
		RevisionMode: pulumi.String("Single"),
		// The token lives in the app's own secret store; the env above
		// references it by name. CreateOrUpdate semantics mean the value
		// always rides the app configuration — which is why it is a
		// secret, never a plain env var.
		Secrets: containerapp.AppSecretArray{
			&containerapp.AppSecretArgs{
				Name:  pulumi.String(tokenSecretName),
				Value: pulumi.String(spec.GetToken()),
			},
		},
		Template: &containerapp.AppTemplateArgs{
			Containers: containerapp.AppTemplateContainerArray{container},
			// Exactly one runner per enrollment: a runner's identity is
			// minted for ONE live replica — a second replica joining
			// under the same name would revoke the first's key (token
			// lineage: re-admission re-mints and revokes). Never enable
			// scaling here without redesigning enrollment for fleets.
			MinReplicas: pulumi.IntPtr(1),
			MaxReplicas: pulumi.IntPtr(1),
		},
		Tags: pulumi.ToStringMap(locals.AzureTags),
		// NO ingress block at all: the runner accepts no inbound traffic —
		// it initiates every connection it uses (control plane, work
		// queue, image pulls).
	}

	createdApp, err := containerapp.NewApp(ctx,
		runnerName,
		appArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Container App %s", runnerName)
	}

	ctx.Export(OpContainerAppId, createdApp.ID())
	ctx.Export(OpContainerAppName, createdApp.Name)
	ctx.Export(OpTokenSecretName, pulumi.String(tokenSecretName))
	ctx.Export(OpRunnerName, pulumi.String(locals.RunnerName))
	ctx.Export(OpResourceGroupName, pulumi.String(locals.ResourceGroupName))

	return nil
}
