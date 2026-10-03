package module

import (
	"github.com/pkg/errors"
	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceanbucketv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanbucket/v1alpha1"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// bucket provisions the Spaces bucket, its per-bucket settings satellites
// (CORS configuration, bucket policy, access logging), and exports the
// outputs declared in outputs.proto.
func bucket(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.SpacesBucket, error) {
	spec := locals.DigitalOceanBucket.Spec

	// The spec's two-value access enum renders the provider's free-string
	// canned ACL.
	acl := pulumi.String("private")
	if spec.AccessControl == digitaloceanbucketv1alpha1.DigitalOceanBucketAccessControl_PUBLIC_READ {
		acl = pulumi.String("public-read")
	}

	bucketArgs := &digitalocean.SpacesBucketArgs{
		Name: pulumi.String(spec.BucketName),
		Acl:  acl,
		// DANGER: when true, destroy empties the bucket — every object AND
		// every object version — before deleting it.
		ForceDestroy: pulumi.Bool(spec.ForceDestroy),
	}

	// When unset, the provider applies its own default region (nyc3);
	// sending the unspecified enum's name would be rejected.
	if spec.Region != digitaloceanprovider.DigitalOceanRegion_digital_ocean_region_unspecified {
		bucketArgs.Region = pulumi.StringPtr(spec.Region.String())
	}

	// Versioning: once enabled it can never be removed — flipping
	// versioning_enabled back to false suspends it, keeping existing
	// versions.
	if spec.VersioningEnabled {
		bucketArgs.Versioning = &digitalocean.SpacesBucketVersioningArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	if len(spec.LifecycleRules) > 0 {
		var lifecycleRules digitalocean.SpacesBucketLifecycleRuleArray
		for _, rule := range spec.LifecycleRules {
			ruleArgs := &digitalocean.SpacesBucketLifecycleRuleArgs{
				Enabled: pulumi.Bool(rule.GetEnabled()),
			}
			// When omitted, the provider generates a rule id.
			if rule.Id != "" {
				ruleArgs.Id = pulumi.StringPtr(rule.Id)
			}
			if rule.Prefix != "" {
				ruleArgs.Prefix = pulumi.StringPtr(rule.Prefix)
			}
			if rule.AbortIncompleteMultipartUploadDays > 0 {
				ruleArgs.AbortIncompleteMultipartUploadDays = pulumi.IntPtr(int(rule.AbortIncompleteMultipartUploadDays))
			}
			if rule.Expiration != nil {
				// The spec requires exactly one trigger; only the set one is
				// sent.
				expirationArgs := &digitalocean.SpacesBucketLifecycleRuleExpirationArgs{}
				if rule.Expiration.Date != "" {
					expirationArgs.Date = pulumi.StringPtr(rule.Expiration.Date)
				}
				if rule.Expiration.Days > 0 {
					expirationArgs.Days = pulumi.IntPtr(int(rule.Expiration.Days))
				}
				if rule.Expiration.ExpiredObjectDeleteMarker {
					expirationArgs.ExpiredObjectDeleteMarker = pulumi.BoolPtr(true)
				}
				ruleArgs.Expiration = expirationArgs
			}
			if rule.NoncurrentVersionExpiration != nil {
				ruleArgs.NoncurrentVersionExpiration = &digitalocean.SpacesBucketLifecycleRuleNoncurrentVersionExpirationArgs{
					Days: pulumi.IntPtr(int(rule.NoncurrentVersionExpiration.Days)),
				}
			}
			lifecycleRules = append(lifecycleRules, ruleArgs)
		}
		bucketArgs.LifecycleRules = lifecycleRules
	}

	createdBucket, err := digitalocean.NewSpacesBucket(
		ctx,
		"bucket",
		bucketArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean spaces bucket")
	}

	// --- Per-bucket settings satellites --------------------------------
	// Separate provider resources whose lifecycle is identical to the
	// bucket's. Their region argument is required by the provider, so the
	// spec requires an explicit region whenever any satellite is
	// configured (CEL rule satellites_require_region).

	// CORS through the standalone resource: the bucket's inline cors_rule
	// is deprecated at the pinned provider and performs no drift detection.
	if len(spec.CorsRules) > 0 {
		var corsRules digitalocean.SpacesBucketCorsConfigurationCorsRuleArray
		for _, rule := range spec.CorsRules {
			ruleArgs := &digitalocean.SpacesBucketCorsConfigurationCorsRuleArgs{
				AllowedMethods: pulumi.ToStringArray(rule.AllowedMethods),
				AllowedOrigins: pulumi.ToStringArray(rule.AllowedOrigins),
			}
			if len(rule.AllowedHeaders) > 0 {
				ruleArgs.AllowedHeaders = pulumi.ToStringArray(rule.AllowedHeaders)
			}
			if len(rule.ExposeHeaders) > 0 {
				ruleArgs.ExposeHeaders = pulumi.ToStringArray(rule.ExposeHeaders)
			}
			if rule.Id != "" {
				ruleArgs.Id = pulumi.StringPtr(rule.Id)
			}
			if rule.MaxAgeSeconds > 0 {
				ruleArgs.MaxAgeSeconds = pulumi.IntPtr(int(rule.MaxAgeSeconds))
			}
			corsRules = append(corsRules, ruleArgs)
		}
		if _, err := digitalocean.NewSpacesBucketCorsConfiguration(
			ctx,
			"cors",
			&digitalocean.SpacesBucketCorsConfigurationArgs{
				Bucket:    createdBucket.ID(),
				Region:    pulumi.String(spec.Region.String()),
				CorsRules: corsRules,
			},
			pulumi.Provider(digitalOceanProvider),
			pulumi.Parent(createdBucket),
		); err != nil {
			return nil, errors.Wrap(err, "failed to create spaces bucket cors configuration")
		}
	}

	if spec.Policy != "" {
		if _, err := digitalocean.NewSpacesBucketPolicy(
			ctx,
			"policy",
			&digitalocean.SpacesBucketPolicyArgs{
				Bucket: createdBucket.ID(),
				Region: pulumi.String(spec.Region.String()),
				Policy: pulumi.String(spec.Policy),
			},
			pulumi.Provider(digitalOceanProvider),
			pulumi.Parent(createdBucket),
		); err != nil {
			return nil, errors.Wrap(err, "failed to create spaces bucket policy")
		}
	}

	if spec.Logging != nil {
		if _, err := digitalocean.NewSpacesBucketLogging(
			ctx,
			"logging",
			&digitalocean.SpacesBucketLoggingArgs{
				Bucket: createdBucket.ID(),
				Region: pulumi.String(spec.Region.String()),
				// The FK resolves to the log-receiving bucket's name.
				TargetBucket: pulumi.String(spec.Logging.TargetBucket.GetValue()),
				TargetPrefix: pulumi.String(spec.Logging.TargetPrefix),
			},
			pulumi.Provider(digitalOceanProvider),
			pulumi.Parent(createdBucket),
		); err != nil {
			return nil, errors.Wrap(err, "failed to create spaces bucket logging")
		}
	}

	// Outputs from the SDK's real attribute names. The provider's urn
	// attribute is BucketUrn in the SDK (URN() is Pulumi's own resource
	// URN, a different thing).
	ctx.Export(OpBucketId, createdBucket.ID())
	ctx.Export(OpEndpoint, createdBucket.Endpoint)
	ctx.Export(OpRegion, createdBucket.Region)
	ctx.Export(OpBucketDomainName, createdBucket.BucketDomainName)
	ctx.Export(OpUrn, createdBucket.BucketUrn)

	return createdBucket, nil
}
