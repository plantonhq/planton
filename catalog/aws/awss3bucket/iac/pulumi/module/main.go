package module

import (
	"github.com/pkg/errors"
	awss3bucketv1alpha1 "github.com/plantonhq/planton/catalog/aws/awss3bucket/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates bucket creation — the root resource plus one
// satellite per folded spec block — and exports the outputs.
func Resources(ctx *pulumi.Context, iacInput *awss3bucketv1alpha1.AwsS3BucketIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static keys, keyless web
	// identity, or ambient chain). The region is the resource's region.
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdBucket, createdVersioning, err := bucket(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "bucket and security posture")
	}

	if err := lifecycle(ctx, locals, provider, createdBucket, createdVersioning); err != nil {
		return errors.Wrap(err, "lifecycle configuration")
	}

	if err := replication(ctx, locals, provider, createdBucket, createdVersioning); err != nil {
		return errors.Wrap(err, "replication configuration")
	}

	createdWebsite, err := website(ctx, locals, provider, createdBucket)
	if err != nil {
		return errors.Wrap(err, "website configuration")
	}

	if err := logging(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "access logging")
	}

	if err := cors(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "cors configuration")
	}

	if err := notification(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "event notifications")
	}

	if err := objectLock(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "object lock configuration")
	}

	if err := accelerate(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "transfer acceleration")
	}

	if err := requestPayment(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "requester pays")
	}

	if err := intelligentTiering(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "intelligent tiering")
	}

	if err := abac(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "abac")
	}

	if err := analyticsConfigurations(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "analytics configurations")
	}

	if err := inventoryConfigurations(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "inventory configurations")
	}

	if err := metricsConfigurations(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "metrics configurations")
	}

	if err := metadataConfiguration(ctx, locals, provider, createdBucket); err != nil {
		return errors.Wrap(err, "metadata configuration")
	}

	// Export outputs matching AwsS3BucketOutputs. Website outputs are
	// exported as empty strings when hosting is not configured so the
	// output contract stays shape-stable across both engines.
	ctx.Export(OpBucketId, createdBucket.Bucket)
	ctx.Export(OpBucketArn, createdBucket.Arn)
	ctx.Export(OpRegion, createdBucket.Region)
	ctx.Export(OpBucketRegionalDomainName, createdBucket.BucketRegionalDomainName)
	ctx.Export(OpHostedZoneId, createdBucket.HostedZoneId)
	ctx.Export(OpBucketDomainName, createdBucket.BucketDomainName)
	if createdWebsite != nil {
		ctx.Export(OpWebsiteEndpoint, createdWebsite.WebsiteEndpoint)
		ctx.Export(OpWebsiteDomain, createdWebsite.WebsiteDomain)
	} else {
		ctx.Export(OpWebsiteEndpoint, pulumi.String(""))
		ctx.Export(OpWebsiteDomain, pulumi.String(""))
	}

	return nil
}
