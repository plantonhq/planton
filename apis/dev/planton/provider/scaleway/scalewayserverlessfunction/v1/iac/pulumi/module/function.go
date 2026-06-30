package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scalewayv2 "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/functions"
)

// privacyEnumToString maps the proto privacy enum to the string
// values expected by the Scaleway API.
var privacyEnumToString = map[string]string{
	"privacy_public":  "public",
	"privacy_private": "private",
}

// httpOptionEnumToString maps the proto HTTP option enum to the string
// values expected by the Scaleway API.
var httpOptionEnumToString = map[string]string{
	"enabled":    "enabled",
	"redirected": "redirected",
}

// serverlessFunction provisions a Scaleway function namespace, the
// function itself, and optional cron triggers, then exports stack
// outputs.
//
// Uses the functions.NewNamespace, functions.NewFunction, and
// functions.NewCron from the scaleway/functions subpackage
// (the preferred API path in the pulumiverse SDK v1.43.0).
func serverlessFunction(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scalewayv2.Provider,
) error {
	spec := locals.ScalewayServerlessFunction.Spec
	metadata := locals.ScalewayServerlessFunction.Metadata

	// ── 1. Create the function namespace ──────────────────────────────

	nsArgs := &functions.NamespaceArgs{
		Name:        pulumi.String(metadata.Name),
		Description: pulumi.String(spec.Description),
		Region:      pulumi.String(spec.Region),
		Tags:        toPulumiStringArray(locals.ScalewayTags),
	}

	createdNamespace, err := functions.NewNamespace(
		ctx,
		"namespace",
		nsArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create function namespace")
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

	// ── 3. Resolve enum values ────────────────────────────────────────

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

	// ── 4. Create the function ────────────────────────────────────────

	fnArgs := &functions.FunctionArgs{
		NamespaceId:                createdNamespace.ID(),
		Name:                       pulumi.String(metadata.Name),
		Runtime:                    pulumi.String(spec.Runtime),
		Handler:                    pulumi.String(spec.Handler),
		Privacy:                    pulumi.String(privacy),
		EnvironmentVariables:       envVars,
		SecretEnvironmentVariables: secretEnvVars,
		HttpOption:                 pulumi.String(httpOption),
		Tags:                       toPulumiStringArray(locals.ScalewayTags),
	}

	// Optional: description.
	if spec.Description != "" {
		fnArgs.Description = pulumi.String(spec.Description)
	}

	// Optional: memory limit.
	if spec.MemoryLimitMb > 0 {
		fnArgs.MemoryLimit = pulumi.IntPtr(int(spec.MemoryLimitMb))
	}

	// Optional: scaling.
	if spec.MinScale > 0 {
		fnArgs.MinScale = pulumi.IntPtr(int(spec.MinScale))
	}
	if spec.MaxScale > 0 {
		fnArgs.MaxScale = pulumi.IntPtr(int(spec.MaxScale))
	}

	// Optional: timeout.
	if spec.TimeoutSeconds > 0 {
		fnArgs.Timeout = pulumi.IntPtr(int(spec.TimeoutSeconds))
	}

	// Optional: sandbox.
	if spec.Sandbox != "" {
		fnArgs.Sandbox = pulumi.String(spec.Sandbox)
	}

	// Optional: zip-based code deployment.
	if spec.ZipFile != "" {
		fnArgs.ZipFile = pulumi.String(spec.ZipFile)
		fnArgs.Deploy = pulumi.Bool(true)
		if spec.ZipHash != "" {
			fnArgs.ZipHash = pulumi.String(spec.ZipHash)
		}
	}

	// Optional: Private Network.
	if spec.PrivateNetworkId != nil && spec.PrivateNetworkId.GetValue() != "" {
		fnArgs.PrivateNetworkId = pulumi.String(spec.PrivateNetworkId.GetValue())
	}

	createdFunction, err := functions.NewFunction(
		ctx,
		"function",
		fnArgs,
		pulumi.Provider(scalewayProvider),
		pulumi.DependsOn([]pulumi.Resource{createdNamespace}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create serverless function")
	}

	// ── 5. Create optional cron triggers ──────────────────────────────

	for idx, trigger := range spec.CronTriggers {
		triggerName := trigger.Name
		if triggerName == "" {
			triggerName = fmt.Sprintf("cron-%d", idx)
		}

		// Build a unique Pulumi resource name from the trigger name.
		resourceName := fmt.Sprintf("cron-%s", strings.ReplaceAll(triggerName, " ", "-"))

		cronArgs := &functions.CronArgs{
			FunctionId: createdFunction.ID(),
			Schedule:   pulumi.String(trigger.Schedule),
			Args:       pulumi.String(trigger.Args),
		}

		if trigger.Name != "" {
			cronArgs.Name = pulumi.String(trigger.Name)
		}

		_, err := functions.NewCron(
			ctx,
			resourceName,
			cronArgs,
			pulumi.Provider(scalewayProvider),
			pulumi.DependsOn([]pulumi.Resource{createdFunction}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create cron trigger %s", resourceName)
		}
	}

	// ── 6. Export stack outputs ────────────────────────────────────────

	ctx.Export(OpFunctionId, createdFunction.ID())
	ctx.Export(OpNamespaceId, createdNamespace.ID())
	ctx.Export(OpDomainName, createdFunction.DomainName)

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
