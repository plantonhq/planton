package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scalewayv2 "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/containers"
)

// privacyEnumToString maps the proto privacy enum to the string
// values expected by the Scaleway API.
var privacyEnumToString = map[string]string{
	"public":  "public",
	"private": "private",
}

// httpOptionEnumToString maps the proto HTTP option enum to the string
// values expected by the Scaleway API.
var httpOptionEnumToString = map[string]string{
	"enabled":    "enabled",
	"redirected": "redirected",
}

// protocolEnumToString maps the proto protocol enum to the string
// values expected by the Scaleway API.
var protocolEnumToString = map[string]string{
	"http1": "http1",
	"h2c":   "h2c",
}

// serverlessContainer provisions a Scaleway container namespace, the
// container itself, and optional cron triggers, then exports stack
// outputs.
//
// Uses containers.NewNamespace, containers.NewContainer, and
// containers.NewCron from the scaleway/containers subpackage
// (the preferred API path in the pulumiverse SDK v1.43.0).
func serverlessContainer(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scalewayv2.Provider,
) error {
	spec := locals.ScalewayServerlessContainer.Spec
	metadata := locals.ScalewayServerlessContainer.Metadata

	// ── 1. Create the container namespace ─────────────────────────────

	nsArgs := &containers.NamespaceArgs{
		Name:        pulumi.String(metadata.Name),
		Description: pulumi.String(spec.Description),
		Region:      pulumi.String(spec.Region),
		Tags:        toPulumiStringArray(locals.ScalewayTags),
	}

	createdNamespace, err := containers.NewNamespace(
		ctx,
		"namespace",
		nsArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create container namespace")
	}

	// ── 2. Build environment variable maps ────────────────────────────

	envVars := pulumi.StringMap{}
	secretEnvVars := pulumi.StringMap{}

	if spec.Env != nil {
		for _, ev := range spec.Env.Variables {
			envVars[ev.Name] = pulumi.String(ev.Value)
		}
		for _, ev := range spec.Env.Secrets {
			secretEnvVars[ev.Name] = pulumi.String(ev.Value)
		}
	}

	// ── 3. Compose the registry image URL ─────────────────────────────

	registryEndpoint := ""
	if spec.Image != nil && spec.Image.RegistryEndpoint != nil {
		registryEndpoint = spec.Image.RegistryEndpoint.GetValue()
	}
	registryImage := fmt.Sprintf("%s/%s:%s", registryEndpoint, spec.Image.Name, spec.Image.Tag)

	// ── 4. Resolve enum values ────────────────────────────────────────

	privacy := "public"
	if mapped, ok := privacyEnumToString[spec.Privacy.String()]; ok {
		privacy = mapped
	}

	httpOption := "enabled"
	if spec.HttpOption != 0 {
		if mapped, ok := httpOptionEnumToString[spec.HttpOption.String()]; ok {
			httpOption = mapped
		}
	}

	protocol := "http1"
	if spec.Protocol != 0 {
		if mapped, ok := protocolEnumToString[spec.Protocol.String()]; ok {
			protocol = mapped
		}
	}

	// ── 5. Create the container ───────────────────────────────────────

	ctArgs := &containers.ContainerArgs{
		NamespaceId:                createdNamespace.ID(),
		Name:                       pulumi.String(metadata.Name),
		RegistryImage:              pulumi.String(registryImage),
		Privacy:                    pulumi.String(privacy),
		EnvironmentVariables:       envVars,
		SecretEnvironmentVariables: secretEnvVars,
		HttpOption:                 pulumi.String(httpOption),
		Protocol:                   pulumi.String(protocol),
		Tags:                       toPulumiStringArray(locals.ScalewayTags),
	}

	// Optional: description.
	if spec.Description != "" {
		ctArgs.Description = pulumi.String(spec.Description)
	}

	// Optional: port.
	if spec.Port > 0 {
		ctArgs.Port = pulumi.IntPtr(int(spec.Port))
	}

	// Optional: memory limit.
	if spec.MemoryLimitMb > 0 {
		ctArgs.MemoryLimit = pulumi.IntPtr(int(spec.MemoryLimitMb))
	}

	// Optional: CPU limit.
	if spec.CpuLimit > 0 {
		ctArgs.CpuLimit = pulumi.IntPtr(int(spec.CpuLimit))
	}

	// Optional: scaling.
	if spec.MinScale > 0 {
		ctArgs.MinScale = pulumi.IntPtr(int(spec.MinScale))
	}
	if spec.MaxScale > 0 {
		ctArgs.MaxScale = pulumi.IntPtr(int(spec.MaxScale))
	}

	// Optional: timeout.
	if spec.TimeoutSeconds > 0 {
		ctArgs.Timeout = pulumi.IntPtr(int(spec.TimeoutSeconds))
	}

	// Optional: sandbox.
	if spec.Sandbox != "" {
		ctArgs.Sandbox = pulumi.String(spec.Sandbox)
	}

	// Optional: registry SHA256 (deployment trigger).
	if spec.RegistrySha256 != "" {
		ctArgs.RegistrySha256 = pulumi.String(spec.RegistrySha256)
	}

	// Optional: deploy flag.
	if spec.Deploy {
		ctArgs.Deploy = pulumi.Bool(true)
	}

	// Optional: command override.
	if len(spec.Commands) > 0 {
		ctArgs.Commands = toPulumiStringArray(spec.Commands)
	}

	// Optional: args override.
	if len(spec.Args) > 0 {
		ctArgs.Args = toPulumiStringArray(spec.Args)
	}

	// Optional: local storage limit.
	if spec.LocalStorageLimitMb > 0 {
		ctArgs.LocalStorageLimit = pulumi.IntPtr(int(spec.LocalStorageLimitMb))
	}

	// Optional: Private Network.
	if spec.PrivateNetworkId != nil && spec.PrivateNetworkId.GetValue() != "" {
		ctArgs.PrivateNetworkId = pulumi.String(spec.PrivateNetworkId.GetValue())
	}

	// Optional: health check.
	if spec.HealthCheck != nil && spec.HealthCheck.Path != "" {
		ctArgs.HealthChecks = containers.ContainerHealthCheckArray{
			containers.ContainerHealthCheckArgs{
				FailureThreshold: pulumi.Int(int(spec.HealthCheck.FailureThreshold)),
				Interval:         pulumi.String(fmt.Sprintf("%ds", spec.HealthCheck.IntervalSeconds)),
				Https: containers.ContainerHealthCheckHttpArray{
					containers.ContainerHealthCheckHttpArgs{
						Path: pulumi.String(spec.HealthCheck.Path),
					},
				},
			},
		}
	}

	// Optional: scaling options.
	if spec.ScalingOption != nil {
		scalingOpts := containers.ContainerScalingOptionArgs{}
		hasScalingOpt := false
		if spec.ScalingOption.ConcurrentRequestsThreshold > 0 {
			scalingOpts.ConcurrentRequestsThreshold = pulumi.IntPtr(int(spec.ScalingOption.ConcurrentRequestsThreshold))
			hasScalingOpt = true
		}
		if spec.ScalingOption.CpuUsageThreshold > 0 {
			scalingOpts.CpuUsageThreshold = pulumi.IntPtr(int(spec.ScalingOption.CpuUsageThreshold))
			hasScalingOpt = true
		}
		if spec.ScalingOption.MemoryUsageThreshold > 0 {
			scalingOpts.MemoryUsageThreshold = pulumi.IntPtr(int(spec.ScalingOption.MemoryUsageThreshold))
			hasScalingOpt = true
		}
		if hasScalingOpt {
			ctArgs.ScalingOptions = containers.ContainerScalingOptionArray{scalingOpts}
		}
	}

	createdContainer, err := containers.NewContainer(
		ctx,
		"container",
		ctArgs,
		pulumi.Provider(scalewayProvider),
		pulumi.DependsOn([]pulumi.Resource{createdNamespace}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create serverless container")
	}

	// ── 6. Create optional cron triggers ──────────────────────────────

	for idx, trigger := range spec.CronTriggers {
		triggerName := trigger.Name
		if triggerName == "" {
			triggerName = fmt.Sprintf("cron-%d", idx)
		}

		// Build a unique Pulumi resource name from the trigger name.
		resourceName := fmt.Sprintf("cron-%s", strings.ReplaceAll(triggerName, " ", "-"))

		cronArgs := &containers.CronArgs{
			ContainerId: createdContainer.ID(),
			Schedule:    pulumi.String(trigger.Schedule),
			Args:        pulumi.String(trigger.Args),
		}

		if trigger.Name != "" {
			cronArgs.Name = pulumi.String(trigger.Name)
		}

		_, err := containers.NewCron(
			ctx,
			resourceName,
			cronArgs,
			pulumi.Provider(scalewayProvider),
			pulumi.DependsOn([]pulumi.Resource{createdContainer}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create cron trigger %s", resourceName)
		}
	}

	// ── 7. Export stack outputs ────────────────────────────────────────

	ctx.Export(OpContainerId, createdContainer.ID())
	ctx.Export(OpNamespaceId, createdNamespace.ID())
	ctx.Export(OpDomainName, createdContainer.DomainName)

	return nil
}

// toPulumiStringArray converts a Go string slice to a Pulumi
// StringArray for use in resource arguments.
func toPulumiStringArray(tags []string) pulumi.StringArray {
	arr := make(pulumi.StringArray, len(tags))
	for i, t := range tags {
		arr[i] = pulumi.String(t)
	}
	return arr
}
