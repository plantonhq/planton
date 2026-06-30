package module

import (
	awscloudfrontv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awscloudfront/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudfront"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createDistribution(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*cloudfront.Distribution, error) {
	spec := locals.Spec

	// Build origins from spec and find the default origin
	var origins cloudfront.DistributionOriginArray
	var defaultOriginId pulumi.StringInput

	for i, o := range spec.Origins {
		// Generate a unique origin ID for Pulumi
		originId := pulumi.Sprintf("origin-%d", i+1)

		originArgs := &cloudfront.DistributionOriginArgs{
			DomainName: pulumi.String(o.DomainName),
			OriginId:   originId,
		}
		if o.OriginPath != "" {
			originArgs.OriginPath = pulumi.String(o.OriginPath)
		}
		// 80/20 default: custom origin config with HTTPS
		originArgs.CustomOriginConfig = &cloudfront.DistributionOriginCustomOriginConfigArgs{
			OriginProtocolPolicy: pulumi.String("https-only"),
			HttpPort:             pulumi.Int(80),
			HttpsPort:            pulumi.Int(443),
			OriginSslProtocols: pulumi.StringArray{
				pulumi.String("TLSv1.2"),
			},
		}
		origins = append(origins, originArgs)

		// Set the default origin ID if this origin is marked as default
		if o.IsDefault {
			defaultOriginId = originId
		}
	}

	// Default cache behavior: minimal safe settings
	defaultBehavior := &cloudfront.DistributionDefaultCacheBehaviorArgs{
		TargetOriginId:       defaultOriginId,
		ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
		AllowedMethods: pulumi.StringArray{
			pulumi.String("GET"), pulumi.String("HEAD"),
		},
		CachedMethods: pulumi.StringArray{
			pulumi.String("GET"), pulumi.String("HEAD"),
		},
		ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
			QueryString: pulumi.Bool(false),
			Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
				Forward: pulumi.String("none"),
			},
		},
		MinTtl:     pulumi.Int(0),
		DefaultTtl: pulumi.Int(3600),
		MaxTtl:     pulumi.Int(86400),
	}

	// Viewer certificate
	var viewerCert *cloudfront.DistributionViewerCertificateArgs
	if spec.CertificateArn != nil && spec.CertificateArn.GetValue() != "" {
		viewerCert = &cloudfront.DistributionViewerCertificateArgs{
			AcmCertificateArn:      pulumi.String(spec.CertificateArn.GetValue()),
			SslSupportMethod:       pulumi.String("sni-only"),
			MinimumProtocolVersion: pulumi.String("TLSv1.2_2021"),
		}
	} else {
		viewerCert = &cloudfront.DistributionViewerCertificateArgs{
			CloudfrontDefaultCertificate: pulumi.Bool(true),
		}
	}

	// Price class mapping
	priceClass := pulumi.String("PriceClass_All")
	switch spec.PriceClass {
	case awscloudfrontv1.AwsCloudFrontSpec_PRICE_CLASS_100:
		priceClass = pulumi.String("PriceClass_100")
	case awscloudfrontv1.AwsCloudFrontSpec_PRICE_CLASS_200:
		priceClass = pulumi.String("PriceClass_200")
	case awscloudfrontv1.AwsCloudFrontSpec_PRICE_CLASS_ALL:
		priceClass = pulumi.String("PriceClass_All")
	}

	dist, err := cloudfront.NewDistribution(ctx, locals.Target.Metadata.Name, &cloudfront.DistributionArgs{
		Enabled:              pulumi.Bool(spec.Enabled),
		Aliases:              pulumi.ToStringArray(spec.Aliases),
		PriceClass:           priceClass,
		Origins:              origins,
		DefaultRootObject:    pulumi.String(spec.DefaultRootObject),
		DefaultCacheBehavior: defaultBehavior,
		Restrictions: &cloudfront.DistributionRestrictionsArgs{
			GeoRestriction: &cloudfront.DistributionRestrictionsGeoRestrictionArgs{
				RestrictionType: pulumi.String("none"),
			},
		},
		ViewerCertificate: viewerCert,
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}
	return dist, nil
}
