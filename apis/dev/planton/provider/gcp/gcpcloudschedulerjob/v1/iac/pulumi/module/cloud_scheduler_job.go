package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudscheduler"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func cloudSchedulerJob(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudSchedulerJob.Spec

	// Determine the job name: explicit job_name or fall back to metadata.name.
	jobName := spec.JobName
	if jobName == "" {
		jobName = locals.GcpCloudSchedulerJob.Metadata.Name
	}

	args := &cloudscheduler.JobArgs{
		Name:     pulumi.StringPtr(jobName),
		Region:   pulumi.StringPtr(spec.Location),
		Project:  pulumi.StringPtr(spec.ProjectId.GetValue()),
		Schedule: pulumi.StringPtr(spec.Schedule),
	}

	// Time zone (defaults to Etc/UTC if not set by GCP).
	if spec.TimeZone != "" {
		args.TimeZone = pulumi.StringPtr(spec.TimeZone)
	}

	// Description.
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// Attempt deadline.
	if spec.AttemptDeadline != "" {
		args.AttemptDeadline = pulumi.StringPtr(spec.AttemptDeadline)
	}

	// Paused state.
	if spec.Paused {
		args.Paused = pulumi.BoolPtr(true)
	}

	// HTTP target.
	if spec.HttpTarget != nil {
		httpArgs := &cloudscheduler.JobHttpTargetArgs{
			Uri: pulumi.String(spec.HttpTarget.Uri),
		}

		if spec.HttpTarget.HttpMethod != "" {
			httpArgs.HttpMethod = pulumi.StringPtr(spec.HttpTarget.HttpMethod)
		}

		if spec.HttpTarget.Body != "" {
			httpArgs.Body = pulumi.StringPtr(spec.HttpTarget.Body)
		}

		if len(spec.HttpTarget.Headers) > 0 {
			headers := pulumi.StringMap{}
			for k, v := range spec.HttpTarget.Headers {
				headers[k] = pulumi.String(v)
			}
			httpArgs.Headers = headers
		}

		// OAuth token (mutually exclusive with OIDC).
		if spec.HttpTarget.OauthToken != nil {
			oauthArgs := &cloudscheduler.JobHttpTargetOauthTokenArgs{
				ServiceAccountEmail: pulumi.String(spec.HttpTarget.OauthToken.ServiceAccountEmail.GetValue()),
			}
			if spec.HttpTarget.OauthToken.Scope != "" {
				oauthArgs.Scope = pulumi.StringPtr(spec.HttpTarget.OauthToken.Scope)
			}
			httpArgs.OauthToken = oauthArgs
		}

		// OIDC token (mutually exclusive with OAuth).
		if spec.HttpTarget.OidcToken != nil {
			oidcArgs := &cloudscheduler.JobHttpTargetOidcTokenArgs{
				ServiceAccountEmail: pulumi.String(spec.HttpTarget.OidcToken.ServiceAccountEmail.GetValue()),
			}
			if spec.HttpTarget.OidcToken.Audience != "" {
				oidcArgs.Audience = pulumi.StringPtr(spec.HttpTarget.OidcToken.Audience)
			}
			httpArgs.OidcToken = oidcArgs
		}

		args.HttpTarget = httpArgs
	}

	// Pub/Sub target.
	if spec.PubsubTarget != nil {
		pubsubArgs := &cloudscheduler.JobPubsubTargetArgs{
			TopicName: pulumi.String(spec.PubsubTarget.TopicName.GetValue()),
		}

		if spec.PubsubTarget.Data != "" {
			pubsubArgs.Data = pulumi.StringPtr(spec.PubsubTarget.Data)
		}

		if len(spec.PubsubTarget.Attributes) > 0 {
			attrs := pulumi.StringMap{}
			for k, v := range spec.PubsubTarget.Attributes {
				attrs[k] = pulumi.String(v)
			}
			pubsubArgs.Attributes = attrs
		}

		args.PubsubTarget = pubsubArgs
	}

	// App Engine HTTP target.
	if spec.AppEngineHttpTarget != nil {
		aeArgs := &cloudscheduler.JobAppEngineHttpTargetArgs{
			RelativeUri: pulumi.String(spec.AppEngineHttpTarget.RelativeUri),
		}

		if spec.AppEngineHttpTarget.HttpMethod != "" {
			aeArgs.HttpMethod = pulumi.StringPtr(spec.AppEngineHttpTarget.HttpMethod)
		}

		if spec.AppEngineHttpTarget.Body != "" {
			aeArgs.Body = pulumi.StringPtr(spec.AppEngineHttpTarget.Body)
		}

		if len(spec.AppEngineHttpTarget.Headers) > 0 {
			headers := pulumi.StringMap{}
			for k, v := range spec.AppEngineHttpTarget.Headers {
				headers[k] = pulumi.String(v)
			}
			aeArgs.Headers = headers
		}

		if spec.AppEngineHttpTarget.AppEngineRouting != nil {
			routingArgs := &cloudscheduler.JobAppEngineHttpTargetAppEngineRoutingArgs{}
			if spec.AppEngineHttpTarget.AppEngineRouting.Service != "" {
				routingArgs.Service = pulumi.StringPtr(spec.AppEngineHttpTarget.AppEngineRouting.Service)
			}
			if spec.AppEngineHttpTarget.AppEngineRouting.Version != "" {
				routingArgs.Version = pulumi.StringPtr(spec.AppEngineHttpTarget.AppEngineRouting.Version)
			}
			if spec.AppEngineHttpTarget.AppEngineRouting.Instance != "" {
				routingArgs.Instance = pulumi.StringPtr(spec.AppEngineHttpTarget.AppEngineRouting.Instance)
			}
			aeArgs.AppEngineRouting = routingArgs
		}

		args.AppEngineHttpTarget = aeArgs
	}

	// Retry config.
	if spec.RetryConfig != nil {
		retryArgs := &cloudscheduler.JobRetryConfigArgs{}
		if spec.RetryConfig.RetryCount != 0 {
			retryArgs.RetryCount = pulumi.IntPtr(int(spec.RetryConfig.RetryCount))
		}
		if spec.RetryConfig.MaxRetryDuration != "" {
			retryArgs.MaxRetryDuration = pulumi.StringPtr(spec.RetryConfig.MaxRetryDuration)
		}
		if spec.RetryConfig.MinBackoffDuration != "" {
			retryArgs.MinBackoffDuration = pulumi.StringPtr(spec.RetryConfig.MinBackoffDuration)
		}
		if spec.RetryConfig.MaxBackoffDuration != "" {
			retryArgs.MaxBackoffDuration = pulumi.StringPtr(spec.RetryConfig.MaxBackoffDuration)
		}
		if spec.RetryConfig.MaxDoublings != 0 {
			retryArgs.MaxDoublings = pulumi.IntPtr(int(spec.RetryConfig.MaxDoublings))
		}
		args.RetryConfig = retryArgs
	}

	createdJob, err := cloudscheduler.NewJob(ctx, "cloud-scheduler-job", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create cloud scheduler job")
	}

	ctx.Export(OpJobId, createdJob.ID())
	ctx.Export(OpJobName, createdJob.Name)
	ctx.Export(OpState, createdJob.State)

	return nil
}
