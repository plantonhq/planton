package module

import (
	"github.com/pkg/errors"
	cloudflarequeuev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarequeue/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// queue provisions the Cloudflare Queue and (when configured) its single consumer.
func queue(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareQueue.Spec

	queueArgs := &cloudflare.QueueArgs{
		AccountId: pulumi.String(spec.AccountId),
		QueueName: pulumi.String(spec.QueueName),
	}

	if s := spec.Settings; s != nil {
		settingsArgs := &cloudflare.QueueSettingsArgs{
			DeliveryPaused: pulumi.Bool(s.DeliveryPaused),
		}
		if s.DeliveryDelay > 0 {
			settingsArgs.DeliveryDelay = pulumi.Float64(float64(s.DeliveryDelay))
		}
		if s.MessageRetentionPeriod > 0 {
			settingsArgs.MessageRetentionPeriod = pulumi.Float64(float64(s.MessageRetentionPeriod))
		}
		queueArgs.Settings = settingsArgs
	}

	createdQueue, err := cloudflare.NewQueue(
		ctx,
		"queue",
		queueArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare queue")
	}

	if c := spec.Consumer; c != nil {
		consumerType := c.Type.String()
		consumerArgs := &cloudflare.QueueConsumerArgs{
			AccountId: pulumi.String(spec.AccountId),
			QueueId:   createdQueue.QueueId,
			Type:      pulumi.String(consumerType),
		}
		if consumerType == "worker" && c.ScriptName != nil && c.ScriptName.GetValue() != "" {
			consumerArgs.ScriptName = pulumi.String(c.ScriptName.GetValue())
		}
		if c.DeadLetterQueue != nil && c.DeadLetterQueue.GetValue() != "" {
			consumerArgs.DeadLetterQueue = pulumi.String(c.DeadLetterQueue.GetValue())
		}
		if cs := c.Settings; cs != nil {
			consumerArgs.Settings = buildConsumerSettings(cs, consumerType)
		}

		if _, err := cloudflare.NewQueueConsumer(
			ctx,
			"queue-consumer",
			consumerArgs,
			pulumi.Provider(cloudflareProvider),
			pulumi.Parent(createdQueue),
		); err != nil {
			return errors.Wrap(err, "failed to create cloudflare queue consumer")
		}
	}

	ctx.Export(OpQueueId, createdQueue.QueueId)
	ctx.Export(OpQueueName, createdQueue.QueueName)
	ctx.Export(OpCreatedOn, createdQueue.CreatedOn)
	ctx.Export(OpModifiedOn, createdQueue.ModifiedOn)

	return nil
}

// buildConsumerSettings maps consumer settings, gating type-restricted fields:
// max_concurrency / max_wait_time_ms apply only to worker consumers, while
// visibility_timeout_ms applies only to http_pull consumers.
func buildConsumerSettings(
	cs *cloudflarequeuev1.CloudflareQueueConsumerSettings,
	consumerType string,
) *cloudflare.QueueConsumerSettingsArgs {
	args := &cloudflare.QueueConsumerSettingsArgs{}
	if cs.BatchSize > 0 {
		args.BatchSize = pulumi.Float64(float64(cs.BatchSize))
	}
	if cs.MaxRetries > 0 {
		args.MaxRetries = pulumi.Float64(float64(cs.MaxRetries))
	}
	if cs.RetryDelay > 0 {
		args.RetryDelay = pulumi.Float64(float64(cs.RetryDelay))
	}
	if consumerType == "worker" {
		if cs.MaxConcurrency > 0 {
			args.MaxConcurrency = pulumi.Float64(float64(cs.MaxConcurrency))
		}
		if cs.MaxWaitTimeMs > 0 {
			args.MaxWaitTimeMs = pulumi.Float64(float64(cs.MaxWaitTimeMs))
		}
	}
	if consumerType == "http_pull" && cs.VisibilityTimeoutMs > 0 {
		args.VisibilityTimeoutMs = pulumi.Float64(float64(cs.VisibilityTimeoutMs))
	}
	return args
}
