package module

import (
	"fmt"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/decimalfloat"

	"github.com/pkg/errors"
	awscloudfrontv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudfront/v1alpha1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudfront"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createDistribution provisions the CloudFront distribution plus its folded
// satellites: per-origin Origin Access Controls (for S3 origins that asked
// for one) and the CloudWatch additional-metrics monitoring subscription.
// The rendering mirrors the Terraform module decision-for-decision so both
// engines produce the same distribution from the same manifest.
func createDistribution(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*cloudfront.Distribution, error) {
	spec := locals.AwsCloudFront.Spec
	meta := locals.AwsCloudFront.Metadata

	// Origin Access Controls created for S3 origins that asked for one
	// (s3_origin.create_origin_access_control) -- the modern private-S3
	// shape: CloudFront signs origin requests with SigV4 so the bucket stays
	// fully private, gated by a bucket policy allowing this distribution's
	// ARN. One OAC per requesting origin, named after the origin so
	// multi-origin distributions stay legible in the console.
	oacIds := map[string]pulumi.StringInput{}
	for _, o := range spec.Origins {
		if o.S3Origin == nil || !o.S3Origin.CreateOriginAccessControl {
			continue
		}
		oac, err := cloudfront.NewOriginAccessControl(ctx,
			fmt.Sprintf("%s-%s-oac", meta.Name, o.OriginId),
			&cloudfront.OriginAccessControlArgs{
				Name:                          pulumi.Sprintf("%s-%s", meta.Name, o.OriginId),
				Description:                   pulumi.Sprintf("Origin access control for the %s origin of the %s distribution", o.OriginId, meta.Name),
				OriginAccessControlOriginType: pulumi.String("s3"),
				SigningBehavior:               pulumi.String("always"),
				SigningProtocol:               pulumi.String("sigv4"),
			}, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create origin access control for origin %q", o.OriginId)
		}
		oacIds[o.OriginId] = oac.ID()
	}

	// The continuous-deployment policy this (primary) distribution owns:
	// routes a weighted or header-selected slice of production traffic to a
	// STAGING distribution (one deployed with staging: true, addressed by
	// its CloudFront domain name). CEL guarantees exactly one routing arm,
	// so the provider's required type discriminator is derived from arm
	// presence.
	var cdPolicy *cloudfront.ContinuousDeploymentPolicy
	if cd := spec.ContinuousDeployment; cd != nil {
		stagingDnsNames := make([]string, 0, len(cd.StagingDistributionDnsNames))
		for _, ref := range cd.StagingDistributionDnsNames {
			stagingDnsNames = append(stagingDnsNames, ref.GetValue())
		}

		trafficConfig := &cloudfront.ContinuousDeploymentPolicyTrafficConfigArgs{}
		if cd.SingleWeight != nil {
			trafficConfig.Type = pulumi.String("SingleWeight")
			weightArgs := &cloudfront.ContinuousDeploymentPolicyTrafficConfigSingleWeightConfigArgs{
				Weight: pulumi.Float64(decimalfloat.Float64(cd.SingleWeight.Weight)),
			}
			if ss := cd.SingleWeight.SessionStickiness; ss != nil {
				weightArgs.SessionStickinessConfig = &cloudfront.ContinuousDeploymentPolicyTrafficConfigSingleWeightConfigSessionStickinessConfigArgs{
					IdleTtl:    pulumi.Int(int(ss.IdleTtlSeconds)),
					MaximumTtl: pulumi.Int(int(ss.MaximumTtlSeconds)),
				}
			}
			trafficConfig.SingleWeightConfig = weightArgs
		} else {
			trafficConfig.Type = pulumi.String("SingleHeader")
			trafficConfig.SingleHeaderConfig = &cloudfront.ContinuousDeploymentPolicyTrafficConfigSingleHeaderConfigArgs{
				Header: pulumi.String(cd.SingleHeader.Header),
				Value:  pulumi.String(cd.SingleHeader.Value),
			}
		}

		var err error
		cdPolicy, err = cloudfront.NewContinuousDeploymentPolicy(ctx, meta.Name,
			&cloudfront.ContinuousDeploymentPolicyArgs{
				Enabled: pulumi.Bool(cd.GetEnabled()),
				StagingDistributionDnsNames: &cloudfront.ContinuousDeploymentPolicyStagingDistributionDnsNamesArgs{
					Items:    pulumi.ToStringArray(stagingDnsNames),
					Quantity: pulumi.Int(len(stagingDnsNames)),
				},
				TrafficConfig: trafficConfig,
			}, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create continuous deployment policy")
		}
	}

	// --- Origins ----------------------------------------------------------
	// Exactly one origin-type block renders per origin (CEL enforces the arm
	// exclusivity). An origin with no arm at all is a plain public S3 REST
	// origin, which the provider expresses as an empty s3_origin_config.
	var origins cloudfront.DistributionOriginArray
	for _, o := range spec.Origins {
		originArgs := &cloudfront.DistributionOriginArgs{
			OriginId:   pulumi.String(o.OriginId),
			DomainName: pulumi.String(o.DomainName),
		}
		if o.OriginPath != "" {
			originArgs.OriginPath = pulumi.String(o.OriginPath)
		}
		// Zero keeps the AWS defaults (3 attempts / 10 seconds).
		if o.ConnectionAttempts != 0 {
			originArgs.ConnectionAttempts = pulumi.Int(int(o.ConnectionAttempts))
		}
		if o.ConnectionTimeoutSeconds != 0 {
			originArgs.ConnectionTimeout = pulumi.Int(int(o.ConnectionTimeoutSeconds))
		}
		// Zero leaves the completion cap unenforced (the AWS default). AWS
		// requires the cap to be >= the origin read timeout.
		if o.ResponseCompletionTimeoutSeconds != 0 {
			originArgs.ResponseCompletionTimeout = pulumi.Int(int(o.ResponseCompletionTimeoutSeconds))
		}

		if len(o.CustomHeaders) > 0 {
			var headers cloudfront.DistributionOriginCustomHeaderArray
			for _, h := range o.CustomHeaders {
				headers = append(headers, &cloudfront.DistributionOriginCustomHeaderArgs{
					Name:  pulumi.String(h.Name),
					Value: pulumi.String(h.Value),
				})
			}
			originArgs.CustomHeaders = headers
		}

		if o.OriginShield != nil {
			originArgs.OriginShield = &cloudfront.DistributionOriginOriginShieldArgs{
				Enabled:            pulumi.Bool(true),
				OriginShieldRegion: pulumi.String(o.OriginShield.OriginShieldRegion),
			}
		}

		// The OAC attaches at the origin level and works with any origin
		// type (S3, Lambda URL, MediaPackage v2, MediaStore): either the
		// s3-type one this module created for the origin, or an existing
		// one attached by ID (CEL keeps the two exclusive).
		if oacId, created := oacIds[o.OriginId]; created {
			originArgs.OriginAccessControlId = oacId
		} else if o.OriginAccessControlId != "" {
			originArgs.OriginAccessControlId = pulumi.String(o.OriginAccessControlId)
		}

		switch {
		case o.S3Origin != nil:
			// The legacy OAI path is only for buckets already wired to an
			// existing identity -- this module never creates one (OAC
			// supersedes it). A bare s3_origin with no access arm renders
			// the empty identity, the provider's public-bucket shape. An
			// origin served through an OAC skips the block entirely.
			if _, created := oacIds[o.OriginId]; !created && o.OriginAccessControlId == "" {
				originArgs.S3OriginConfig = &cloudfront.DistributionOriginS3OriginConfigArgs{
					OriginAccessIdentity: pulumi.String(o.S3Origin.OriginAccessIdentity),
				}
			}

		case o.CustomOrigin != nil:
			customArgs := &cloudfront.DistributionOriginCustomOriginConfigArgs{
				OriginProtocolPolicy: pulumi.String(o.CustomOrigin.ProtocolPolicy),
				HttpPort:             pulumi.Int(80),
				HttpsPort:            pulumi.Int(443),
			}
			if o.CustomOrigin.HttpPort != 0 {
				customArgs.HttpPort = pulumi.Int(int(o.CustomOrigin.HttpPort))
			}
			if o.CustomOrigin.HttpsPort != 0 {
				customArgs.HttpsPort = pulumi.Int(int(o.CustomOrigin.HttpsPort))
			}
			// Empty keeps the safe modern floor.
			sslProtocols := o.CustomOrigin.SslProtocols
			if len(sslProtocols) == 0 {
				sslProtocols = []string{"TLSv1.2"}
			}
			customArgs.OriginSslProtocols = pulumi.ToStringArray(sslProtocols)
			if o.CustomOrigin.KeepaliveTimeoutSeconds != 0 {
				customArgs.OriginKeepaliveTimeout = pulumi.Int(int(o.CustomOrigin.KeepaliveTimeoutSeconds))
			}
			if o.CustomOrigin.ReadTimeoutSeconds != 0 {
				customArgs.OriginReadTimeout = pulumi.Int(int(o.CustomOrigin.ReadTimeoutSeconds))
			}
			// Empty keeps ipv4 (the AWS default).
			if o.CustomOrigin.IpAddressType != "" {
				customArgs.IpAddressType = pulumi.String(o.CustomOrigin.IpAddressType)
			}
			// Origin-side mTLS: CloudFront presents this ACM client
			// certificate (us-east-1) to the origin, so only CloudFront
			// can reach the backend.
			if o.CustomOrigin.GetMtlsClientCertificateArn().GetValue() != "" {
				customArgs.OriginMtlsConfig = &cloudfront.DistributionOriginCustomOriginConfigOriginMtlsConfigArgs{
					ClientCertificateArn: pulumi.String(o.CustomOrigin.GetMtlsClientCertificateArn().GetValue()),
				}
			}
			originArgs.CustomOriginConfig = customArgs

		case o.VpcOrigin != nil:
			vpcArgs := &cloudfront.DistributionOriginVpcOriginConfigArgs{
				VpcOriginId: pulumi.String(o.VpcOrigin.VpcOriginId),
			}
			if o.VpcOrigin.KeepaliveTimeoutSeconds != 0 {
				vpcArgs.OriginKeepaliveTimeout = pulumi.Int(int(o.VpcOrigin.KeepaliveTimeoutSeconds))
			}
			if o.VpcOrigin.ReadTimeoutSeconds != 0 {
				vpcArgs.OriginReadTimeout = pulumi.Int(int(o.VpcOrigin.ReadTimeoutSeconds))
			}
			// Empty means the VPC origin belongs to this account.
			if o.VpcOrigin.OwnerAccountId != "" {
				vpcArgs.OwnerAccountId = pulumi.String(o.VpcOrigin.OwnerAccountId)
			}
			originArgs.VpcOriginConfig = vpcArgs

		default:
			// No arm: a plain public S3 REST origin (unless an existing OAC
			// was attached above).
			if o.OriginAccessControlId == "" {
				originArgs.S3OriginConfig = &cloudfront.DistributionOriginS3OriginConfigArgs{
					OriginAccessIdentity: pulumi.String(""),
				}
			}
		}

		origins = append(origins, originArgs)
	}

	// --- Origin groups (primary/failover pairs) ----------------------------
	var originGroups cloudfront.DistributionOriginGroupArray
	for _, g := range spec.OriginGroups {
		var members cloudfront.DistributionOriginGroupMemberArray
		for _, m := range g.MemberOriginIds {
			members = append(members, &cloudfront.DistributionOriginGroupMemberArgs{
				OriginId: pulumi.String(m),
			})
		}
		var statusCodes pulumi.IntArray
		for _, code := range g.FailoverStatusCodes {
			statusCodes = append(statusCodes, pulumi.Int(int(code)))
		}
		originGroups = append(originGroups, &cloudfront.DistributionOriginGroupArgs{
			OriginId: pulumi.String(g.OriginGroupId),
			FailoverCriteria: &cloudfront.DistributionOriginGroupFailoverCriteriaArgs{
				StatusCodes: statusCodes,
			},
			Members: members,
		})
	}

	args := &cloudfront.DistributionArgs{
		// enabled/wait_for_deployment resolve to their annotated defaults
		// (true) at manifest load. A disabled distribution stays
		// deployed-but-dark -- also the state AWS requires before deletion.
		Enabled:              pulumi.Bool(spec.GetEnabled()),
		WaitForDeployment:    pulumi.Bool(spec.GetWaitForDeployment()),
		RetainOnDelete:       pulumi.Bool(spec.RetainOnDelete),
		IsIpv6Enabled:        pulumi.Bool(spec.IsIpv6Enabled),
		Origins:              origins,
		DefaultCacheBehavior: buildDefaultCacheBehavior(spec.DefaultCacheBehavior),
		Restrictions:         buildRestrictions(spec.GeoRestriction),
		ViewerCertificate:    buildViewerCertificate(spec.ViewerCertificate),
		Tags:                 pulumi.ToStringMap(locals.AwsTags),
	}
	if len(spec.Aliases) > 0 {
		args.Aliases = pulumi.ToStringArray(spec.Aliases)
	}
	if spec.Comment != "" {
		args.Comment = pulumi.String(spec.Comment)
	}
	if spec.DefaultRootObject != "" {
		args.DefaultRootObject = pulumi.String(spec.DefaultRootObject)
	}
	if spec.HttpVersion != "" {
		args.HttpVersion = pulumi.String(spec.HttpVersion)
	}
	if spec.PriceClass != "" {
		args.PriceClass = pulumi.String(spec.PriceClass)
	}
	// The provider argument is named WebAclId, but for WAFv2 it takes the
	// web ACL's ARN (the bare ID form is only for retired WAF Classic).
	if spec.GetWebAclArn().GetValue() != "" {
		args.WebAclId = pulumi.String(spec.GetWebAclArn().GetValue())
	}
	if len(originGroups) > 0 {
		args.OriginGroups = originGroups
	}

	// Changing the staging flag REPLACES the distribution (AWS makes it
	// immutable after create).
	args.Staging = pulumi.Bool(spec.Staging)

	// Either the policy this resource created (the continuous_deployment
	// block) or an externally managed one by ID -- CEL keeps the arms
	// exclusive.
	if cdPolicy != nil {
		args.ContinuousDeploymentPolicyId = cdPolicy.ID()
	} else if spec.ContinuousDeploymentPolicyId != "" {
		args.ContinuousDeploymentPolicyId = pulumi.String(spec.ContinuousDeploymentPolicyId)
	}

	// Dedicated static edge IPs -- the list must already be provisioned in
	// the account (a paid feature).
	if spec.AnycastIpListId != "" {
		args.AnycastIpListId = pulumi.String(spec.AnycastIpListId)
	}

	// Tag-based invalidation: origin responses label objects via this
	// header; invalidations by tag purge every object carrying the label.
	// Verbatim pass-through: AWS stores the name LOWERCASED and the
	// provider does not case-suppress, so a mixed-case value would
	// re-plan forever -- the spec's format rule enforces lowercase
	// upstream.
	if spec.CacheTagHeaderName != "" {
		args.CacheTagConfig = &cloudfront.DistributionCacheTagConfigArgs{
			HeaderName: pulumi.String(spec.CacheTagHeaderName),
		}
	}

	// A connection function runs at TCP establishment, before any HTTP
	// parsing -- the earliest programmable point in the request path.
	if spec.ConnectionFunctionId != "" {
		args.ConnectionFunctionAssociation = &cloudfront.DistributionConnectionFunctionAssociationArgs{
			Id: pulumi.String(spec.ConnectionFunctionId),
		}
	}

	// Viewer-side mutual TLS: client certificates validated against a
	// CloudFront trust store ("passthrough" forwards them unvalidated, so
	// it renders without the trust-store wrapper).
	if vm := spec.ViewerMtls; vm != nil {
		mtlsArgs := &cloudfront.DistributionViewerMtlsConfigArgs{}
		if vm.Mode != "" {
			mtlsArgs.Mode = pulumi.String(vm.Mode)
		}
		if vm.TrustStoreId != "" {
			mtlsArgs.TrustStoreConfig = &cloudfront.DistributionViewerMtlsConfigTrustStoreConfigArgs{
				TrustStoreId:               pulumi.String(vm.TrustStoreId),
				AdvertiseTrustStoreCaNames: pulumi.Bool(vm.AdvertiseTrustStoreCaNames),
				IgnoreCertificateExpiry:    pulumi.Bool(vm.IgnoreCertificateExpiry),
			}
		}
		args.ViewerMtlsConfig = mtlsArgs
	}

	// Ordered behaviors are evaluated in list order before the default --
	// first match wins.
	if len(spec.OrderedCacheBehaviors) > 0 {
		var ordered cloudfront.DistributionOrderedCacheBehaviorArray
		for _, ob := range spec.OrderedCacheBehaviors {
			ordered = append(ordered, buildOrderedCacheBehavior(ob))
		}
		args.OrderedCacheBehaviors = ordered
	}

	if len(spec.CustomErrorResponses) > 0 {
		var errorResponses cloudfront.DistributionCustomErrorResponseArray
		for _, er := range spec.CustomErrorResponses {
			erArgs := &cloudfront.DistributionCustomErrorResponseArgs{
				ErrorCode: pulumi.Int(int(er.ErrorCode)),
			}
			if er.ResponseCode != 0 {
				erArgs.ResponseCode = pulumi.Int(int(er.ResponseCode))
			}
			if er.ResponsePagePath != "" {
				erArgs.ResponsePagePath = pulumi.String(er.ResponsePagePath)
			}
			if er.ErrorCachingMinTtlSeconds != 0 {
				erArgs.ErrorCachingMinTtl = pulumi.Int(int(er.ErrorCachingMinTtlSeconds))
			}
			errorResponses = append(errorResponses, erArgs)
		}
		args.CustomErrorResponses = errorResponses
	}

	// An empty bucket keeps v1 log DELIVERY off while still recording the
	// cookie preference -- the transition shape for distributions moving to
	// v2 (CloudWatch-delivered) access logs.
	if spec.Logging != nil {
		loggingArgs := &cloudfront.DistributionLoggingConfigArgs{
			IncludeCookies: pulumi.Bool(spec.Logging.IncludeCookies),
		}
		if spec.Logging.Bucket != "" {
			loggingArgs.Bucket = pulumi.String(spec.Logging.Bucket)
		}
		if spec.Logging.Prefix != "" {
			loggingArgs.Prefix = pulumi.String(spec.Logging.Prefix)
		}
		args.LoggingConfig = loggingArgs
	}

	dist, err := cloudfront.NewDistribution(ctx, meta.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudfront distribution")
	}

	// CloudWatch additional metrics (cache hit rate, origin latency,
	// per-status error rates) -- a distribution-keyed setting materialized as
	// its own provider resource, folded into the spec as one honest toggle.
	if spec.EnableAdditionalMetrics {
		_, err = cloudfront.NewMonitoringSubscription(ctx, meta.Name, &cloudfront.MonitoringSubscriptionArgs{
			DistributionId: dist.ID(),
			MonitoringSubscription: &cloudfront.MonitoringSubscriptionMonitoringSubscriptionArgs{
				RealtimeMetricsSubscriptionConfig: &cloudfront.MonitoringSubscriptionMonitoringSubscriptionRealtimeMetricsSubscriptionConfigArgs{
					RealtimeMetricsSubscriptionStatus: pulumi.String("Enabled"),
				},
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create monitoring subscription")
		}
	}

	return dist, nil
}

// buildRestrictions renders the mandatory restrictions block; an absent spec
// geo_restriction means no geographic restriction.
func buildRestrictions(geo *awscloudfrontv1alpha1.AwsCloudFrontGeoRestriction) *cloudfront.DistributionRestrictionsArgs {
	geoArgs := &cloudfront.DistributionRestrictionsGeoRestrictionArgs{
		RestrictionType: pulumi.String("none"),
	}
	if geo != nil {
		geoArgs.RestrictionType = pulumi.String(geo.RestrictionType)
		geoArgs.Locations = pulumi.ToStringArray(geo.Locations)
	}
	return &cloudfront.DistributionRestrictionsArgs{GeoRestriction: geoArgs}
}

// buildViewerCertificate renders the viewer certificate with the provider's
// arm precedence: IAM certificate, then ACM, then the default
// *.cloudfront.net certificate. Custom certificates default to SNI-only
// ("vip" costs dedicated IPs) and the TLSv1.2_2021 floor.
func buildViewerCertificate(vc *awscloudfrontv1alpha1.AwsCloudFrontViewerCertificate) *cloudfront.DistributionViewerCertificateArgs {
	hasCustom := vc != nil && (vc.GetAcmCertificateArn().GetValue() != "" || vc.IamCertificateId != "")
	if !hasCustom {
		return &cloudfront.DistributionViewerCertificateArgs{
			CloudfrontDefaultCertificate: pulumi.Bool(true),
		}
	}

	args := &cloudfront.DistributionViewerCertificateArgs{}
	if vc.IamCertificateId != "" {
		args.IamCertificateId = pulumi.String(vc.IamCertificateId)
	} else {
		args.AcmCertificateArn = pulumi.String(vc.GetAcmCertificateArn().GetValue())
	}

	sslSupportMethod := vc.SslSupportMethod
	if sslSupportMethod == "" {
		sslSupportMethod = "sni-only"
	}
	args.SslSupportMethod = pulumi.String(sslSupportMethod)

	minimumProtocolVersion := vc.MinimumProtocolVersion
	if minimumProtocolVersion == "" {
		minimumProtocolVersion = "TLSv1.2_2021"
	}
	args.MinimumProtocolVersion = pulumi.String(minimumProtocolVersion)

	return args
}

// behaviorParts carries the field set shared by the default and ordered
// cache behaviors, resolved from the spec with the same defaulting the
// Terraform module applies (empty method lists keep CloudFront's
// static-content defaults; TTLs only apply to the legacy forwarded_values
// generation).
type behaviorParts struct {
	allowedMethods pulumi.StringArrayInput
	cachedMethods  pulumi.StringArrayInput
	minTtl         pulumi.IntPtrInput
	defaultTtl     pulumi.IntPtrInput
	maxTtl         pulumi.IntPtrInput
}

func resolveBehaviorParts(b *awscloudfrontv1alpha1.AwsCloudFrontCacheBehavior) behaviorParts {
	parts := behaviorParts{}

	allowedMethods := b.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = []string{"GET", "HEAD"}
	}
	parts.allowedMethods = pulumi.ToStringArray(allowedMethods)

	cachedMethods := b.CachedMethods
	if len(cachedMethods) == 0 {
		cachedMethods = []string{"GET", "HEAD"}
	}
	parts.cachedMethods = pulumi.ToStringArray(cachedMethods)

	// TTLs only apply to the legacy generation -- with a cache policy the
	// policy owns them and the provider rejects the combination.
	if b.ForwardedValues != nil {
		parts.minTtl = pulumi.Int(int(b.MinTtlSeconds))
		if b.DefaultTtlSeconds != 0 {
			parts.defaultTtl = pulumi.Int(int(b.DefaultTtlSeconds))
		}
		if b.MaxTtlSeconds != 0 {
			parts.maxTtl = pulumi.Int(int(b.MaxTtlSeconds))
		}
	}

	return parts
}

// buildDefaultCacheBehavior maps the shared behavior message onto the
// default-behavior args type. The ordered twin below is identical except for
// the SDK's distinct per-position types and the path pattern -- keep the two
// in lockstep when the surface grows.
func buildDefaultCacheBehavior(b *awscloudfrontv1alpha1.AwsCloudFrontCacheBehavior) *cloudfront.DistributionDefaultCacheBehaviorArgs {
	parts := resolveBehaviorParts(b)

	args := &cloudfront.DistributionDefaultCacheBehaviorArgs{
		TargetOriginId:       pulumi.String(b.TargetOriginId),
		ViewerProtocolPolicy: pulumi.String(b.ViewerProtocolPolicy),
		AllowedMethods:       parts.allowedMethods,
		CachedMethods:        parts.cachedMethods,
		Compress:             pulumi.Bool(b.Compress),
		SmoothStreaming:      pulumi.Bool(b.SmoothStreaming),
		MinTtl:               parts.minTtl,
		DefaultTtl:           parts.defaultTtl,
		MaxTtl:               parts.maxTtl,
	}

	if b.CachePolicyId != "" {
		args.CachePolicyId = pulumi.String(b.CachePolicyId)
	}
	if b.OriginRequestPolicyId != "" {
		args.OriginRequestPolicyId = pulumi.String(b.OriginRequestPolicyId)
	}
	if b.ResponseHeadersPolicyId != "" {
		args.ResponseHeadersPolicyId = pulumi.String(b.ResponseHeadersPolicyId)
	}
	if b.FieldLevelEncryptionId != "" {
		args.FieldLevelEncryptionId = pulumi.String(b.FieldLevelEncryptionId)
	}
	if b.RealtimeLogConfigArn != "" {
		args.RealtimeLogConfigArn = pulumi.String(b.RealtimeLogConfigArn)
	}
	if len(b.TrustedKeyGroupIds) > 0 {
		args.TrustedKeyGroups = pulumi.ToStringArray(b.TrustedKeyGroupIds)
	}
	if len(b.TrustedSigners) > 0 {
		args.TrustedSigners = pulumi.ToStringArray(b.TrustedSigners)
	}

	if b.ForwardedValues != nil {
		fv := b.ForwardedValues
		args.ForwardedValues = &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
			QueryString:          pulumi.Bool(fv.QueryString),
			QueryStringCacheKeys: pulumi.ToStringArray(fv.QueryStringCacheKeys),
			Headers:              pulumi.ToStringArray(fv.Headers),
			Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
				Forward:          pulumi.String(fv.CookiesForward),
				WhitelistedNames: pulumi.ToStringArray(fv.WhitelistedCookieNames),
			},
		}
	}

	if len(b.FunctionAssociations) > 0 {
		var fns cloudfront.DistributionDefaultCacheBehaviorFunctionAssociationArray
		for _, fa := range b.FunctionAssociations {
			fns = append(fns, &cloudfront.DistributionDefaultCacheBehaviorFunctionAssociationArgs{
				EventType:   pulumi.String(fa.EventType),
				FunctionArn: pulumi.String(fa.FunctionArn),
			})
		}
		args.FunctionAssociations = fns
	}

	if len(b.LambdaFunctionAssociations) > 0 {
		var lambdas cloudfront.DistributionDefaultCacheBehaviorLambdaFunctionAssociationArray
		for _, la := range b.LambdaFunctionAssociations {
			lambdas = append(lambdas, &cloudfront.DistributionDefaultCacheBehaviorLambdaFunctionAssociationArgs{
				EventType:   pulumi.String(la.EventType),
				LambdaArn:   pulumi.String(la.LambdaArn),
				IncludeBody: pulumi.Bool(la.IncludeBody),
			})
		}
		args.LambdaFunctionAssociations = lambdas
	}

	if b.GrpcEnabled {
		args.GrpcConfig = &cloudfront.DistributionDefaultCacheBehaviorGrpcConfigArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	return args
}

// buildOrderedCacheBehavior is the ordered twin of buildDefaultCacheBehavior
// -- same field mapping plus the path pattern.
func buildOrderedCacheBehavior(ob *awscloudfrontv1alpha1.AwsCloudFrontOrderedCacheBehavior) *cloudfront.DistributionOrderedCacheBehaviorArgs {
	b := ob.Behavior
	parts := resolveBehaviorParts(b)

	args := &cloudfront.DistributionOrderedCacheBehaviorArgs{
		PathPattern:          pulumi.String(ob.PathPattern),
		TargetOriginId:       pulumi.String(b.TargetOriginId),
		ViewerProtocolPolicy: pulumi.String(b.ViewerProtocolPolicy),
		AllowedMethods:       parts.allowedMethods,
		CachedMethods:        parts.cachedMethods,
		Compress:             pulumi.Bool(b.Compress),
		SmoothStreaming:      pulumi.Bool(b.SmoothStreaming),
		MinTtl:               parts.minTtl,
		DefaultTtl:           parts.defaultTtl,
		MaxTtl:               parts.maxTtl,
	}

	if b.CachePolicyId != "" {
		args.CachePolicyId = pulumi.String(b.CachePolicyId)
	}
	if b.OriginRequestPolicyId != "" {
		args.OriginRequestPolicyId = pulumi.String(b.OriginRequestPolicyId)
	}
	if b.ResponseHeadersPolicyId != "" {
		args.ResponseHeadersPolicyId = pulumi.String(b.ResponseHeadersPolicyId)
	}
	if b.FieldLevelEncryptionId != "" {
		args.FieldLevelEncryptionId = pulumi.String(b.FieldLevelEncryptionId)
	}
	if b.RealtimeLogConfigArn != "" {
		args.RealtimeLogConfigArn = pulumi.String(b.RealtimeLogConfigArn)
	}
	if len(b.TrustedKeyGroupIds) > 0 {
		args.TrustedKeyGroups = pulumi.ToStringArray(b.TrustedKeyGroupIds)
	}
	if len(b.TrustedSigners) > 0 {
		args.TrustedSigners = pulumi.ToStringArray(b.TrustedSigners)
	}

	if b.ForwardedValues != nil {
		fv := b.ForwardedValues
		args.ForwardedValues = &cloudfront.DistributionOrderedCacheBehaviorForwardedValuesArgs{
			QueryString:          pulumi.Bool(fv.QueryString),
			QueryStringCacheKeys: pulumi.ToStringArray(fv.QueryStringCacheKeys),
			Headers:              pulumi.ToStringArray(fv.Headers),
			Cookies: &cloudfront.DistributionOrderedCacheBehaviorForwardedValuesCookiesArgs{
				Forward:          pulumi.String(fv.CookiesForward),
				WhitelistedNames: pulumi.ToStringArray(fv.WhitelistedCookieNames),
			},
		}
	}

	if len(b.FunctionAssociations) > 0 {
		var fns cloudfront.DistributionOrderedCacheBehaviorFunctionAssociationArray
		for _, fa := range b.FunctionAssociations {
			fns = append(fns, &cloudfront.DistributionOrderedCacheBehaviorFunctionAssociationArgs{
				EventType:   pulumi.String(fa.EventType),
				FunctionArn: pulumi.String(fa.FunctionArn),
			})
		}
		args.FunctionAssociations = fns
	}

	if len(b.LambdaFunctionAssociations) > 0 {
		var lambdas cloudfront.DistributionOrderedCacheBehaviorLambdaFunctionAssociationArray
		for _, la := range b.LambdaFunctionAssociations {
			lambdas = append(lambdas, &cloudfront.DistributionOrderedCacheBehaviorLambdaFunctionAssociationArgs{
				EventType:   pulumi.String(la.EventType),
				LambdaArn:   pulumi.String(la.LambdaArn),
				IncludeBody: pulumi.Bool(la.IncludeBody),
			})
		}
		args.LambdaFunctionAssociations = lambdas
	}

	if b.GrpcEnabled {
		args.GrpcConfig = &cloudfront.DistributionOrderedCacheBehaviorGrpcConfigArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	return args
}
