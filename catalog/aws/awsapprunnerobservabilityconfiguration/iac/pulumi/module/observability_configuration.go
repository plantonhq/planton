package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apprunner"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// observabilityConfiguration registers the App Runner observability
// configuration version. AWS versions this resource, but the provider's
// update path is TAGS-ONLY and does not replace on trace changes: adding,
// removing, or changing the trace block on an existing configuration plans
// an in-place update that silently changes nothing at AWS (an upstream
// provider gap -- it briefly made trace settings ForceNew in 2023 and
// reverted the same day). To actually change tracing posture, register a
// NEW configuration (a new metadata.name) and repoint referencing services.
// With X-Ray as the only vendor AWS supports, the block's practical states
// are present or absent at creation time.
func observabilityConfiguration(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.AwsAppRunnerObservabilityConfiguration.Spec

	args := &apprunner.ObservabilityConfigurationArgs{
		// The cloud name is metadata.name -- the same basis the Terraform
		// module uses, so both engines register the same physical identity.
		ObservabilityConfigurationName: pulumi.String(locals.AwsAppRunnerObservabilityConfiguration.Metadata.Name),
		Tags:                           pulumi.ToStringMap(locals.AwsTags),
	}

	// The trace block is emitted only when the spec configures tracing; a
	// configuration without it is valid but inert. X-Ray is the only vendor
	// App Runner supports today -- applications must emit spans through the
	// AWS Distro for OpenTelemetry SDK for the collector to forward anything.
	if spec.TraceConfiguration != nil {
		traceArgs := &apprunner.ObservabilityConfigurationTraceConfigurationArgs{}
		if spec.TraceConfiguration.Vendor != nil {
			traceArgs.Vendor = pulumi.StringPtr(*spec.TraceConfiguration.Vendor)
		}
		args.TraceConfiguration = traceArgs
	}

	createdConfiguration, err := apprunner.NewObservabilityConfiguration(ctx,
		locals.AwsAppRunnerObservabilityConfiguration.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create observability configuration")
	}

	// Export outputs matching AwsAppRunnerObservabilityConfigurationOutputs.
	ctx.Export(OpConfigurationArn, createdConfiguration.Arn)
	ctx.Export(OpConfigurationRevision, createdConfiguration.ObservabilityConfigurationRevision)
	ctx.Export(OpLatest, createdConfiguration.Latest)

	return nil
}
