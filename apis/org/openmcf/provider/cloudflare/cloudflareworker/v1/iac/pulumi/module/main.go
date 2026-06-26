package module

import (
	"fmt"

	"github.com/pkg/errors"
	cloudflareworkerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareworker/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry-point expected by the OpenMCF CLI.
func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflareworkerv1.CloudflareWorkerStackInput,
) error {
	locals := initializeLocals(ctx, stackInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(ctx, locals.CloudflareProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to set up cloudflare provider")
	}

	// An AWS (S3) provider aimed at the R2 endpoint is only needed when the
	// script source is an R2 bundle.
	var r2Provider *aws.Provider
	if locals.CloudflareWorker.Spec.GetR2Bundle() != nil {
		r2Provider, err = newR2Provider(ctx, locals)
		if err != nil {
			return errors.Wrap(err, "failed to create R2 provider")
		}
	}

	if err := worker(ctx, locals, cloudflareProvider, r2Provider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare worker")
	}

	return nil
}

// newR2Provider builds an S3-compatible provider pointed at the account's R2
// endpoint, using explicit credentials when present, otherwise the environment.
func newR2Provider(ctx *pulumi.Context, locals *Locals) (*aws.Provider, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", locals.CloudflareWorker.Spec.AccountId)

	args := &aws.ProviderArgs{
		Region:                    pulumi.String("auto"),
		S3UsePathStyle:            pulumi.Bool(true),
		SkipCredentialsValidation: pulumi.Bool(true),
		SkipMetadataApiCheck:      pulumi.Bool(true),
		SkipRegionValidation:      pulumi.Bool(true),
		SkipRequestingAccountId:   pulumi.Bool(true),
	}
	if r2 := locals.CloudflareProviderConfig.GetR2(); r2 != nil {
		if r2.Endpoint != "" {
			endpoint = r2.Endpoint
		}
		args.AccessKey = pulumi.String(r2.AccessKeyId)
		args.SecretKey = pulumi.String(r2.SecretAccessKey)
	}
	args.Endpoints = aws.ProviderEndpointArray{
		aws.ProviderEndpointArgs{S3: pulumi.String(endpoint)},
	}

	return aws.NewProvider(ctx, "r2-provider", args)
}
