package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func eventBus(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	// "default" names the account's built-in bus; AWS rejects creating
	// another. Failing before the API call beats an opaque CreateEventBus
	// error (mirrors the Terraform module's plan-time precondition).
	if locals.Target.Metadata.Name == "default" {
		return errors.New("the bus name (metadata.name) must not be \"default\" — every AWS account already has a default event bus; rules can target it via event_bus_name without creating a bus resource")
	}

	args := &cloudwatch.EventBusArgs{
		Name: pulumi.StringPtr(locals.Target.Metadata.Name),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Description
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// KMS encryption
	if spec.KmsKeyIdentifier.GetValue() != "" {
		args.KmsKeyIdentifier = pulumi.StringPtr(spec.KmsKeyIdentifier.GetValue())
	}

	// Partner event source
	if spec.EventSourceName != "" {
		args.EventSourceName = pulumi.StringPtr(spec.EventSourceName)
	}

	// Dead letter config
	if spec.DeadLetterConfig != nil && spec.DeadLetterConfig.Arn.GetValue() != "" {
		args.DeadLetterConfig = &cloudwatch.EventBusDeadLetterConfigArgs{
			Arn: pulumi.StringPtr(spec.DeadLetterConfig.Arn.GetValue()),
		}
	}

	// Logging config
	if spec.LogConfig != nil && spec.LogConfig.Level != "" {
		logArgs := &cloudwatch.EventBusLogConfigArgs{
			Level: pulumi.StringPtr(spec.LogConfig.Level),
		}
		if spec.LogConfig.IncludeDetail != "" {
			logArgs.IncludeDetail = pulumi.StringPtr(spec.LogConfig.IncludeDetail)
		}
		args.LogConfig = logArgs
	}

	bus, err := cloudwatch.NewEventBus(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create EventBridge bus")
	}

	// The bus's resource-based policy is a single-per-bus setting
	// (cross-account / cross-org PutEvents grants) keyed by the bus name. AWS
	// models it as its own PutPermission API call, so it materializes as its
	// own resource; deleting it removes all cross-account grants.
	if spec.ResourcePolicy != nil {
		policyJSON, err := json.Marshal(spec.ResourcePolicy.AsMap())
		if err != nil {
			return errors.Wrap(err, "failed to serialize resource policy")
		}
		_, err = cloudwatch.NewEventBusPolicy(ctx, locals.Target.Metadata.Name, &cloudwatch.EventBusPolicyArgs{
			EventBusName: bus.Name,
			Policy:       pulumi.String(string(policyJSON)),
		}, pulumi.Provider(provider), pulumi.Parent(bus))
		if err != nil {
			return errors.Wrap(err, "failed to create event bus policy")
		}
	}

	// Event archives recorded from THIS bus — one archive per spec entry,
	// keyed by the archive's own name (names are identity: reordering entries
	// is a no-op and a rename is honestly a new archive). Replay is an
	// on-demand EventBridge operation (StartReplay), never declarative
	// configuration. Optional fields are sent only when set — omission keeps
	// AWS's defaults (retain indefinitely, archive every event, AWS-owned key).
	createdArchives := make([]*cloudwatch.EventArchive, 0, len(spec.Archives))
	for _, archive := range spec.Archives {
		archiveArgs := &cloudwatch.EventArchiveArgs{
			Name:           pulumi.StringPtr(archive.Name),
			EventSourceArn: bus.Arn,
		}
		if archive.Description != "" {
			archiveArgs.Description = pulumi.StringPtr(archive.Description)
		}
		if archive.RetentionDays != 0 {
			archiveArgs.RetentionDays = pulumi.IntPtr(int(archive.RetentionDays))
		}
		if archive.EventPattern != nil {
			patternJSON, err := json.Marshal(archive.EventPattern.AsMap())
			if err != nil {
				return errors.Wrapf(err, "failed to serialize event pattern for archive %s", archive.Name)
			}
			archiveArgs.EventPattern = pulumi.StringPtr(string(patternJSON))
		}
		if archive.KmsKeyIdentifier.GetValue() != "" {
			archiveArgs.KmsKeyIdentifier = pulumi.StringPtr(archive.KmsKeyIdentifier.GetValue())
		}

		createdArchive, err := cloudwatch.NewEventArchive(ctx, archive.Name, archiveArgs,
			pulumi.Provider(provider), pulumi.Parent(bus))
		if err != nil {
			return errors.Wrapf(err, "failed to create event archive %s", archive.Name)
		}
		createdArchives = append(createdArchives, createdArchive)
	}

	// Export outputs matching AwsEventBridgeBusOutputs.
	ctx.Export(OpBusName, bus.Name)
	ctx.Export(OpBusArn, bus.Arn)

	// The archives in spec order — each entry's name and the ARN AWS assigned
	// it, for replay operations (StartReplay) and IAM policies.
	archiveOutputs := make(pulumi.Array, 0, len(createdArchives))
	for i, createdArchive := range createdArchives {
		archiveOutputs = append(archiveOutputs, pulumi.Map{
			"name": pulumi.String(spec.Archives[i].Name),
			"arn":  createdArchive.Arn,
		})
	}
	ctx.Export(OpArchives, archiveOutputs)

	return nil
}
