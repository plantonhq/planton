package module

import (
	"github.com/pkg/errors"
	gcpcloudtasksqueuev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudtasksqueue/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudtasks"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func cloudTasksQueue(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudTasksQueue.Spec

	// Enable the Cloud Tasks API — the control plane that owns queues.
	// DisableOnDestroy stays false: tearing down one queue must never
	// disable the API for everything else in the project (other queues
	// keep dispatching).
	cloudtasksApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("cloudtasks.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		cloudtasksApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdCloudtasksApi, err := projects.NewService(ctx,
		"gcptq-cloudtasks.googleapis.com", cloudtasksApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable cloudtasks.googleapis.com api")
	}

	// The Cloud Tasks queue — the dispatch-rate, retry, and routing contract
	// for every task enqueued into it. Name and location are immutable, and
	// a deleted queue's ID stays reserved by the API for up to 7 days, so
	// renames both replace the queue and burn the old identifier for that
	// window.
	args := &cloudtasks.QueueArgs{
		Name:     pulumi.String(spec.QueueName),
		Location: pulumi.String(spec.Location),

		// Declarative pause/resume: sent explicitly so a PAUSED -> RUNNING
		// spec edit resumes dispatch — mirrors the Terraform module.
		DesiredState: pulumi.StringPtr(desiredState(spec)),
	}

	// DELETE (provider default) removes the queue and its backlog on
	// destroy; PREVENT fails the destroy; ABANDON leaves the queue
	// running. Sent only when set — mirrors the Terraform module.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	// Queue-level HTTP task settings. These OVERRIDE task-level
	// configuration at dispatch time — the pattern that lets producers
	// enqueue bare payloads while the queue owns auth and routing.
	if spec.HttpTarget != nil {
		httpTargetArgs := &cloudtasks.QueueHttpTargetArgs{}

		if spec.HttpTarget.HttpMethod != "" {
			httpTargetArgs.HttpMethod = pulumi.StringPtr(spec.HttpTarget.HttpMethod)
		}

		if len(spec.HttpTarget.HeaderOverrides) > 0 {
			overrides := cloudtasks.QueueHttpTargetHeaderOverrideArray{}
			for _, ho := range spec.HttpTarget.HeaderOverrides {
				overrides = append(overrides, &cloudtasks.QueueHttpTargetHeaderOverrideArgs{
					Header: &cloudtasks.QueueHttpTargetHeaderOverrideHeaderArgs{
						Key:   pulumi.String(ho.Key),
						Value: pulumi.String(ho.Value),
					},
				})
			}
			httpTargetArgs.HeaderOverrides = overrides
		}

		// OAuth (for *.googleapis.com targets) and OIDC (for Cloud Run /
		// Cloud Functions / custom endpoints) are mutually exclusive —
		// enforced pre-deploy by the spec's CEL rule.
		if spec.HttpTarget.OauthToken != nil {
			oauthArgs := &cloudtasks.QueueHttpTargetOauthTokenArgs{
				ServiceAccountEmail: pulumi.String(spec.HttpTarget.OauthToken.ServiceAccountEmail.GetValue()),
			}
			if spec.HttpTarget.OauthToken.Scope != "" {
				oauthArgs.Scope = pulumi.StringPtr(spec.HttpTarget.OauthToken.Scope)
			}
			httpTargetArgs.OauthToken = oauthArgs
		}

		if spec.HttpTarget.OidcToken != nil {
			oidcArgs := &cloudtasks.QueueHttpTargetOidcTokenArgs{
				ServiceAccountEmail: pulumi.String(spec.HttpTarget.OidcToken.ServiceAccountEmail.GetValue()),
			}
			if spec.HttpTarget.OidcToken.Audience != "" {
				oidcArgs.Audience = pulumi.StringPtr(spec.HttpTarget.OidcToken.Audience)
			}
			httpTargetArgs.OidcToken = oidcArgs
		}

		// The spec flattens the provider's nested path_override /
		// query_override single-field blocks into plain path/query_params
		// strings; the blocks are only sent when set, which also avoids the
		// provider's query_override perpetual-diff on the 6.x line when the
		// block would otherwise be sent empty.
		if spec.HttpTarget.UriOverride != nil {
			uriArgs := &cloudtasks.QueueHttpTargetUriOverrideArgs{}

			if spec.HttpTarget.UriOverride.Scheme != "" {
				uriArgs.Scheme = pulumi.StringPtr(spec.HttpTarget.UriOverride.Scheme)
			}
			if spec.HttpTarget.UriOverride.Host != "" {
				uriArgs.Host = pulumi.StringPtr(spec.HttpTarget.UriOverride.Host)
			}
			if spec.HttpTarget.UriOverride.Port != "" {
				uriArgs.Port = pulumi.StringPtr(spec.HttpTarget.UriOverride.Port)
			}
			if spec.HttpTarget.UriOverride.EnforceMode != "" {
				uriArgs.UriOverrideEnforceMode = pulumi.StringPtr(spec.HttpTarget.UriOverride.EnforceMode)
			}

			if spec.HttpTarget.UriOverride.Path != "" {
				uriArgs.PathOverride = &cloudtasks.QueueHttpTargetUriOverridePathOverrideArgs{
					Path: pulumi.StringPtr(spec.HttpTarget.UriOverride.Path),
				}
			}

			if spec.HttpTarget.UriOverride.QueryParams != "" {
				uriArgs.QueryOverride = &cloudtasks.QueueHttpTargetUriOverrideQueryOverrideArgs{
					QueryParams: pulumi.StringPtr(spec.HttpTarget.UriOverride.QueryParams),
				}
			}

			httpTargetArgs.UriOverride = uriArgs
		}

		args.HttpTarget = httpTargetArgs
	}

	// Routing override for App Engine tasks: pins the whole queue's App
	// Engine tasks to one service/version/instance instead of per-task
	// routing. Ignored for HTTP tasks.
	if spec.AppEngineRoutingOverride != nil {
		routingArgs := &cloudtasks.QueueAppEngineRoutingOverrideArgs{}
		if spec.AppEngineRoutingOverride.Service != "" {
			routingArgs.Service = pulumi.StringPtr(spec.AppEngineRoutingOverride.Service)
		}
		if spec.AppEngineRoutingOverride.Version != "" {
			routingArgs.Version = pulumi.StringPtr(spec.AppEngineRoutingOverride.Version)
		}
		if spec.AppEngineRoutingOverride.Instance != "" {
			routingArgs.Instance = pulumi.StringPtr(spec.AppEngineRoutingOverride.Instance)
		}
		args.AppEngineRoutingOverride = routingArgs
	}

	// Zero means "not set" for these dials — Cloud Tasks then applies its
	// own defaults (which it also does for a wholly omitted block).
	if spec.RateLimits != nil {
		rateLimitsArgs := &cloudtasks.QueueRateLimitsArgs{}
		if spec.RateLimits.MaxDispatchesPerSecond > 0 {
			rateLimitsArgs.MaxDispatchesPerSecond = pulumi.Float64Ptr(spec.RateLimits.MaxDispatchesPerSecond)
		}
		if spec.RateLimits.MaxConcurrentDispatches > 0 {
			rateLimitsArgs.MaxConcurrentDispatches = pulumi.IntPtr(int(spec.RateLimits.MaxConcurrentDispatches))
		}
		args.RateLimits = rateLimitsArgs
	}

	// max_attempts genuinely distinguishes -1 (unlimited) from unset, so
	// the not-set sentinel is 0, never -1.
	if spec.RetryConfig != nil {
		retryArgs := &cloudtasks.QueueRetryConfigArgs{}
		if spec.RetryConfig.MaxAttempts != 0 {
			retryArgs.MaxAttempts = pulumi.IntPtr(int(spec.RetryConfig.MaxAttempts))
		}
		if spec.RetryConfig.MaxRetryDuration != "" {
			retryArgs.MaxRetryDuration = pulumi.StringPtr(spec.RetryConfig.MaxRetryDuration)
		}
		if spec.RetryConfig.MinBackoff != "" {
			retryArgs.MinBackoff = pulumi.StringPtr(spec.RetryConfig.MinBackoff)
		}
		if spec.RetryConfig.MaxBackoff != "" {
			retryArgs.MaxBackoff = pulumi.StringPtr(spec.RetryConfig.MaxBackoff)
		}
		if spec.RetryConfig.MaxDoublings != 0 {
			retryArgs.MaxDoublings = pulumi.IntPtr(int(spec.RetryConfig.MaxDoublings))
		}
		args.RetryConfig = retryArgs
	}

	// sampling_ratio 0.0 is a meaningful value (log nothing), so the block
	// is driven purely by presence.
	if spec.StackdriverLoggingConfig != nil {
		args.StackdriverLoggingConfig = &cloudtasks.QueueStackdriverLoggingConfigArgs{
			SamplingRatio: pulumi.Float64(spec.StackdriverLoggingConfig.SamplingRatio),
		}
	}

	createdQueue, err := cloudtasks.NewQueue(ctx, "cloud-tasks-queue", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdCloudtasksApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create cloud tasks queue")
	}

	ctx.Export(OpQueueId, createdQueue.ID())
	ctx.Export(OpQueueName, createdQueue.Name)
	// GCP computes max_burst_size from max_dispatches_per_second; the
	// rate_limits block is populated by the API even when the manifest
	// omits rate limits entirely, so the effective value is always present.
	ctx.Export(OpMaxBurstSize, createdQueue.RateLimits.MaxBurstSize())

	return nil
}

// desiredState resolves the queue's dispatch state, defaulting to RUNNING
// (the provider default) so the field is always reconciled explicitly and
// an out-of-band pause is reverted on the next apply.
func desiredState(spec *gcpcloudtasksqueuev1alpha1.GcpCloudTasksQueueSpec) string {
	if spec.DesiredState != "" {
		return spec.DesiredState
	}
	return "RUNNING"
}
