package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudtasks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func cloudTasksQueue(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudTasksQueue.Spec

	args := &cloudtasks.QueueArgs{
		Name:     pulumi.String(spec.QueueName),
		Location: pulumi.String(spec.Location),
		Project:  pulumi.StringPtr(spec.ProjectId.GetValue()),
	}

	// Desired state (RUNNING / PAUSED).
	if spec.DesiredState != "" {
		args.DesiredState = pulumi.StringPtr(spec.DesiredState)
	}

	// HTTP target (queue-level auth, headers, URI override).
	if spec.HttpTarget != nil {
		httpTargetArgs := &cloudtasks.QueueHttpTargetArgs{}

		if spec.HttpTarget.HttpMethod != "" {
			httpTargetArgs.HttpMethod = pulumi.StringPtr(spec.HttpTarget.HttpMethod)
		}

		// Header overrides.
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

		// OAuth token (mutually exclusive with OIDC).
		if spec.HttpTarget.OauthToken != nil {
			oauthArgs := &cloudtasks.QueueHttpTargetOauthTokenArgs{
				ServiceAccountEmail: pulumi.String(spec.HttpTarget.OauthToken.ServiceAccountEmail.GetValue()),
			}
			if spec.HttpTarget.OauthToken.Scope != "" {
				oauthArgs.Scope = pulumi.StringPtr(spec.HttpTarget.OauthToken.Scope)
			}
			httpTargetArgs.OauthToken = oauthArgs
		}

		// OIDC token (mutually exclusive with OAuth).
		if spec.HttpTarget.OidcToken != nil {
			oidcArgs := &cloudtasks.QueueHttpTargetOidcTokenArgs{
				ServiceAccountEmail: pulumi.String(spec.HttpTarget.OidcToken.ServiceAccountEmail.GetValue()),
			}
			if spec.HttpTarget.OidcToken.Audience != "" {
				oidcArgs.Audience = pulumi.StringPtr(spec.HttpTarget.OidcToken.Audience)
			}
			httpTargetArgs.OidcToken = oidcArgs
		}

		// URI override.
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

			// Path override (flattened from nested path_override.path).
			if spec.HttpTarget.UriOverride.Path != "" {
				uriArgs.PathOverride = &cloudtasks.QueueHttpTargetUriOverridePathOverrideArgs{
					Path: pulumi.StringPtr(spec.HttpTarget.UriOverride.Path),
				}
			}

			// Query override (flattened from nested query_override.query_params).
			if spec.HttpTarget.UriOverride.QueryParams != "" {
				uriArgs.QueryOverride = &cloudtasks.QueueHttpTargetUriOverrideQueryOverrideArgs{
					QueryParams: pulumi.StringPtr(spec.HttpTarget.UriOverride.QueryParams),
				}
			}

			httpTargetArgs.UriOverride = uriArgs
		}

		args.HttpTarget = httpTargetArgs
	}

	// Rate limits.
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

	// Retry config.
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

	// Stackdriver logging config.
	if spec.StackdriverLoggingConfig != nil {
		args.StackdriverLoggingConfig = &cloudtasks.QueueStackdriverLoggingConfigArgs{
			SamplingRatio: pulumi.Float64(spec.StackdriverLoggingConfig.SamplingRatio),
		}
	}

	createdQueue, err := cloudtasks.NewQueue(ctx, "cloud-tasks-queue", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create cloud tasks queue")
	}

	ctx.Export(OpQueueId, createdQueue.ID())
	ctx.Export(OpQueueName, createdQueue.Name)
	ctx.Export(OpState, createdQueue.State)

	return nil
}
