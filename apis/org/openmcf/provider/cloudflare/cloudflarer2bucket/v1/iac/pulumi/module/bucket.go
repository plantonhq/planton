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

	// 1. Assemble the bucket arguments. The location hint is omitted when "auto"
	// (the enum zero value) so Cloudflare selects the region. The enum value
	// name matches the string the provider expects, so .String() is used directly.
	bucketArgs := &cloudflare.R2BucketArgs{
		AccountId: pulumi.String(locals.CloudflareR2Bucket.Spec.AccountId),
		Name:      pulumi.String(locals.CloudflareR2Bucket.Spec.BucketName),
	}
	if locals.CloudflareR2Bucket.Spec.Location != 0 {
		bucketArgs.Location = pulumi.String(locals.CloudflareR2Bucket.Spec.Location.String())
	}

	// 2. Create the bucket.
	createdBucket, err := cloudflare.NewR2Bucket(
		ctx,
		"bucket",
		bucketArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Cloudflare R2 bucket")
	}

	// 3. Public access via the managed r2.dev URL has its own lifecycle and is
	// configured outside this module; the custom domain below is the production path.

	// 4. Handle custom domain configuration.
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

	// 5. Export stack outputs.
	ctx.Export(OpBucketName, createdBucket.Name)
	ctx.Export(OpBucketUrl, pulumi.Sprintf(
		"https://%s.r2.cloudflarestorage.com/%s",
		locals.CloudflareR2Bucket.Spec.AccountId,
		locals.CloudflareR2Bucket.Spec.BucketName,
	))

	return createdBucket, nil
}
