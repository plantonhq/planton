package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cognito"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func domain(ctx *pulumi.Context, locals *Locals, createdPool *cognito.UserPool, provider *aws.Provider) error {
	if locals.Spec.Domain == nil || locals.Spec.Domain.Domain == "" {
		// No domain configured — export empty values.
		ctx.Export(OpUserPoolDomain, pulumi.String(""))
		ctx.Export(OpCloudfrontDistributionArn, pulumi.String(""))
		return nil
	}

	domainSpec := locals.Spec.Domain
	resourceName := locals.Target.Metadata.Name + "-domain"

	args := &cognito.UserPoolDomainArgs{
		Domain:     pulumi.String(domainSpec.Domain),
		UserPoolId: createdPool.ID(),
	}

	// Custom domains (containing a dot) require an ACM certificate.
	isCustomDomain := strings.Contains(domainSpec.Domain, ".")
	if isCustomDomain && domainSpec.CertificateArn.GetValue() != "" {
		args.CertificateArn = pulumi.StringPtr(domainSpec.CertificateArn.GetValue())
	}

	created, err := cognito.NewUserPoolDomain(ctx, resourceName, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create Cognito user pool domain")
	}

	// Export the full domain URL and CloudFront distribution ARN.
	if isCustomDomain {
		ctx.Export(OpUserPoolDomain, pulumi.Sprintf("https://%s", domainSpec.Domain))
		ctx.Export(OpCloudfrontDistributionArn, created.CloudfrontDistributionArn)
	} else {
		// Cognito-hosted prefix domain: build the URL from domain + region.
		// The region comes from the provider. We use the domain resource's computed
		// CloudFront distribution to confirm creation, but the URL follows a known pattern.
		ctx.Export(OpUserPoolDomain, created.CloudfrontDistribution.ApplyT(func(cf string) string {
			// For prefix domains, the URL is: https://{domain}.auth.{region}.amazoncognito.com
			// We can derive region from the pool ID (format: {region}_{id}).
			// However, the simplest approach is to use the domain + the cloudfront distribution.
			// Since cloudfront_distribution for prefix domains returns the CF domain, we use
			// the known URL pattern instead.
			return "https://" + domainSpec.Domain + ".auth." + cf
		}).(pulumi.StringOutput))

		// For prefix domains, the CF distribution ARN is not meaningful for DNS aliasing.
		ctx.Export(OpCloudfrontDistributionArn, pulumi.String(""))
	}

	return nil
}
