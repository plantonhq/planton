package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// bucket provisions the R2 bucket (and optional managed domain) and exports outputs.
func bucket(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.R2Bucket, error) {

	// 1. Get the location string directly from the enum.
	// The enum values (auto, WNAM, ENAM, etc.) match Cloudflare's expected strings.
	bucketLocation := locals.CloudflareR2Bucket.Spec.Location.String()

	// 2. Create the bucket.
	createdBucket, err := cloudflare.NewR2Bucket(
		ctx,
		"bucket",
		&cloudflare.R2BucketArgs{
			AccountId: pulumi.String(locals.CloudflareR2Bucket.Spec.AccountId),
			Name:      pulumi.String(locals.CloudflareR2Bucket.Spec.BucketName),
			Location:  pulumi.String(bucketLocation),
		},
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Cloudflare R2 bucket")
	}

	// 4. Handle public access (r2.dev subdomain).
	// Note: The Cloudflare Pulumi provider does not yet expose a direct field for toggling
	// the r2.dev public URL. This currently requires using the Cloudflare API directly
	// or the dashboard. We log a warning if public_access is requested.
	if locals.CloudflareR2Bucket.Spec.PublicAccess {
		ctx.Log.Warn("Public access (r2.dev subdomain) must be enabled manually via Cloudflare Dashboard or API - field noted but not yet implemented in provider.", nil)
	}

	// 5. Handle custom domain configuration.
	if locals.CloudflareR2Bucket.Spec.CustomDomain != nil && locals.CloudflareR2Bucket.Spec.CustomDomain.Enabled {
		customDomain := locals.CloudflareR2Bucket.Spec.CustomDomain
		zoneId := customDomain.ZoneId.GetValue()

		_, err := cloudflare.NewR2CustomDomain(ctx, "custom-domain", &cloudflare.R2CustomDomainArgs{
			AccountId:  pulumi.String(locals.CloudflareR2Bucket.Spec.AccountId),
			BucketName: createdBucket.Name,
			ZoneId:     pulumi.String(zoneId),
			Domain:     pulumi.String(customDomain.Domain),
			Enabled:    pulumi.Bool(true),
		}, pulumi.Provider(cloudflareProvider), pulumi.DependsOn([]pulumi.Resource{createdBucket}))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Cloudflare R2 custom domain")
		}

		// Export custom domain URL
		ctx.Export(OpCustomDomainUrl, pulumi.Sprintf("https://%s", customDomain.Domain))
	}

	// 7. Export stack outputs.
	ctx.Export(OpBucketName, createdBucket.Name)

	return createdBucket, nil
}
