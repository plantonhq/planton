package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scalewayv2 "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/object"
)

// bucket provisions the Scaleway Object Storage bucket with optional
// versioning, lifecycle rules, and CORS configuration.
//
// This function creates a single object.Bucket resource (the new
// subpackage path, replacing the deprecated top-level ObjectBucket).
// All configuration is inline on the bucket itself.
func bucket(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scalewayv2.Provider,
) error {
	spec := locals.ScalewayObjectBucket.Spec

	// ── Build tag map ───────────────────────────────────────────────
	// Object Storage uses key-value map tags (S3-compatible), not flat
	// string tags like other Scaleway resources.
	tags := pulumi.StringMap{}
	for k, v := range locals.ScalewayTags {
		tags[k] = pulumi.String(v)
	}

	// ── Build bucket arguments ──────────────────────────────────────
	bucketArgs := &object.BucketArgs{
		Name:              pulumi.String(locals.ScalewayObjectBucket.Metadata.Name),
		Region:            pulumi.String(spec.Region),
		Tags:              tags,
		ForceDestroy:      pulumi.Bool(spec.ForceDestroy),
		ObjectLockEnabled: pulumi.Bool(spec.ObjectLockEnabled),
	}

	// ── Versioning ──────────────────────────────────────────────────
	if spec.VersioningEnabled {
		bucketArgs.Versioning = &object.BucketVersioningArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	// ── Lifecycle Rules ─────────────────────────────────────────────
	if len(spec.LifecycleRules) > 0 {
		lifecycleRules := object.BucketLifecycleRuleArray{}
		for _, rule := range spec.LifecycleRules {
			lifecycleRule := &object.BucketLifecycleRuleArgs{
				Id:      pulumi.String(rule.Id),
				Enabled: pulumi.Bool(rule.Enabled),
			}

			if rule.Prefix != "" {
				lifecycleRule.Prefix = pulumi.String(rule.Prefix)
			}

			// Tag-based filter.
			if len(rule.Tags) > 0 {
				ruleTags := pulumi.StringMap{}
				for k, v := range rule.Tags {
					ruleTags[k] = pulumi.String(v)
				}
				lifecycleRule.Tags = ruleTags
			}

			// Expiration.
			if rule.ExpirationDays > 0 {
				lifecycleRule.Expiration = &object.BucketLifecycleRuleExpirationArgs{
					Days: pulumi.Int(rule.ExpirationDays),
				}
			}

			// Storage class transitions.
			if len(rule.Transitions) > 0 {
				transitions := object.BucketLifecycleRuleTransitionArray{}
				for _, t := range rule.Transitions {
					transitions = append(transitions, &object.BucketLifecycleRuleTransitionArgs{
						Days:         pulumi.Int(t.Days),
						StorageClass: pulumi.String(t.StorageClass),
					})
				}
				lifecycleRule.Transitions = transitions
			}

			// Abort incomplete multipart uploads.
			if rule.AbortIncompleteMultipartUploadDays > 0 {
				lifecycleRule.AbortIncompleteMultipartUploadDays = pulumi.Int(rule.AbortIncompleteMultipartUploadDays)
			}

			lifecycleRules = append(lifecycleRules, lifecycleRule)
		}
		bucketArgs.LifecycleRules = lifecycleRules
	}

	// ── CORS Rules ──────────────────────────────────────────────────
	if len(spec.CorsRules) > 0 {
		corsRules := object.BucketCorsRuleArray{}
		for _, rule := range spec.CorsRules {
			corsRule := &object.BucketCorsRuleArgs{}

			// Allowed methods (required).
			allowedMethods := pulumi.StringArray{}
			for _, m := range rule.AllowedMethods {
				allowedMethods = append(allowedMethods, pulumi.String(m))
			}
			corsRule.AllowedMethods = allowedMethods

			// Allowed origins (required).
			allowedOrigins := pulumi.StringArray{}
			for _, o := range rule.AllowedOrigins {
				allowedOrigins = append(allowedOrigins, pulumi.String(o))
			}
			corsRule.AllowedOrigins = allowedOrigins

			// Allowed headers (optional).
			if len(rule.AllowedHeaders) > 0 {
				allowedHeaders := pulumi.StringArray{}
				for _, h := range rule.AllowedHeaders {
					allowedHeaders = append(allowedHeaders, pulumi.String(h))
				}
				corsRule.AllowedHeaders = allowedHeaders
			}

			// Expose headers (optional).
			if len(rule.ExposeHeaders) > 0 {
				exposeHeaders := pulumi.StringArray{}
				for _, h := range rule.ExposeHeaders {
					exposeHeaders = append(exposeHeaders, pulumi.String(h))
				}
				corsRule.ExposeHeaders = exposeHeaders
			}

			// Max age (optional).
			if rule.MaxAgeSeconds > 0 {
				corsRule.MaxAgeSeconds = pulumi.Int(rule.MaxAgeSeconds)
			}

			corsRules = append(corsRules, corsRule)
		}
		bucketArgs.CorsRules = corsRules
	}

	// ── Create Bucket ───────────────────────────────────────────────
	createdBucket, err := object.NewBucket(
		ctx,
		"bucket",
		bucketArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create scaleway object bucket")
	}

	// ── Export Outputs ───────────────────────────────────────────────
	ctx.Export(OpBucketId, createdBucket.ID())
	ctx.Export(OpEndpoint, createdBucket.Endpoint)
	ctx.Export(OpApiEndpoint, createdBucket.ApiEndpoint)
	ctx.Export(OpBucketName, createdBucket.Name)
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return nil
}
