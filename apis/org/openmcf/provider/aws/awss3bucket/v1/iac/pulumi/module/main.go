package module

import (
	"fmt"

	"github.com/pkg/errors"
	awss3bucketv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awss3bucket/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awss3bucketv1.AwsS3BucketStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AwsS3Bucket.Spec

	// Create explicit AWS provider from stack input credentials
	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(awsProviderConfig.GetRegion()),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// Prepare tags by merging labels and user-provided tags
	tags := pulumi.StringMap{}
	for k, v := range locals.Labels {
		tags[k] = pulumi.String(v)
	}
	for k, v := range spec.Tags {
		tags[k] = pulumi.String(v)
	}

	// Create S3 bucket
	bucket, err := s3.NewBucketV2(ctx, "bucket", &s3.BucketV2Args{
		Bucket:       pulumi.String(locals.AwsS3Bucket.Metadata.Name),
		ForceDestroy: pulumi.Bool(spec.ForceDestroy),
		Tags:         tags,
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create S3 bucket")
	}

	// Configure versioning
	if spec.VersioningEnabled {
		_, err = s3.NewBucketVersioningV2(ctx, "versioning", &s3.BucketVersioningV2Args{
			Bucket: bucket.ID(),
			VersioningConfiguration: &s3.BucketVersioningV2VersioningConfigurationArgs{
				Status: pulumi.String("Enabled"),
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to enable bucket versioning")
		}
	}

	// Configure encryption
	encryptionType := spec.EncryptionType
	if encryptionType == awss3bucketv1.AwsS3BucketSpec_ENCRYPTION_TYPE_UNSPECIFIED {
		// Default to SSE-S3 for security
		encryptionType = awss3bucketv1.AwsS3BucketSpec_ENCRYPTION_TYPE_SSE_S3
	}

	var serverSideEncryptionRule *s3.BucketServerSideEncryptionConfigurationV2RuleArgs
	switch encryptionType {
	case awss3bucketv1.AwsS3BucketSpec_ENCRYPTION_TYPE_SSE_S3:
		serverSideEncryptionRule = &s3.BucketServerSideEncryptionConfigurationV2RuleArgs{
			ApplyServerSideEncryptionByDefault: &s3.BucketServerSideEncryptionConfigurationV2RuleApplyServerSideEncryptionByDefaultArgs{
				SseAlgorithm: pulumi.String("AES256"),
			},
			BucketKeyEnabled: pulumi.Bool(true),
		}
	case awss3bucketv1.AwsS3BucketSpec_ENCRYPTION_TYPE_SSE_KMS:
		if spec.KmsKeyId == "" {
			return errors.New("kms_key_id is required when encryption_type is SSE_KMS")
		}
		serverSideEncryptionRule = &s3.BucketServerSideEncryptionConfigurationV2RuleArgs{
			ApplyServerSideEncryptionByDefault: &s3.BucketServerSideEncryptionConfigurationV2RuleApplyServerSideEncryptionByDefaultArgs{
				SseAlgorithm:   pulumi.String("aws:kms"),
				KmsMasterKeyId: pulumi.String(spec.KmsKeyId),
			},
			BucketKeyEnabled: pulumi.Bool(true),
		}
	}

	_, err = s3.NewBucketServerSideEncryptionConfigurationV2(ctx, "encryption", &s3.BucketServerSideEncryptionConfigurationV2Args{
		Bucket: bucket.ID(),
		Rules: s3.BucketServerSideEncryptionConfigurationV2RuleArray{
			serverSideEncryptionRule,
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to configure bucket encryption")
	}

	// Configure public access block
	_, err = s3.NewBucketPublicAccessBlock(ctx, "public-access-block", &s3.BucketPublicAccessBlockArgs{
		Bucket:                bucket.ID(),
		BlockPublicAcls:       pulumi.Bool(!spec.IsPublic),
		BlockPublicPolicy:     pulumi.Bool(!spec.IsPublic),
		IgnorePublicAcls:      pulumi.Bool(!spec.IsPublic),
		RestrictPublicBuckets: pulumi.Bool(!spec.IsPublic),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to configure public access block")
	}

	// Configure ownership controls (disable ACLs - bucket owner enforced)
	_, err = s3.NewBucketOwnershipControls(ctx, "ownership-controls", &s3.BucketOwnershipControlsArgs{
		Bucket: bucket.ID(),
		Rule: &s3.BucketOwnershipControlsRuleArgs{
			ObjectOwnership: pulumi.String("BucketOwnerEnforced"),
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to configure ownership controls")
	}

	// Configure lifecycle rules if specified
	if len(spec.LifecycleRules) > 0 {
		lifecycleRules := s3.BucketLifecycleConfigurationV2RuleArray{}
		for _, rule := range spec.LifecycleRules {
			if rule.Id == "" {
				return errors.New("lifecycle rule id is required")
			}

			status := "Disabled"
			if rule.Enabled {
				status = "Enabled"
			}

			lifecycleRule := &s3.BucketLifecycleConfigurationV2RuleArgs{
				Id:     pulumi.String(rule.Id),
				Status: pulumi.String(status),
			}

			// Add filter if prefix is specified
			if rule.Prefix != "" {
				lifecycleRule.Filter = &s3.BucketLifecycleConfigurationV2RuleFilterArgs{
					Prefix: pulumi.String(rule.Prefix),
				}
			}

			// Add transition if specified
			if rule.TransitionDays > 0 && rule.TransitionStorageClass != awss3bucketv1.AwsS3BucketSpec_STORAGE_CLASS_UNSPECIFIED {
				storageClass, err := mapStorageClass(rule.TransitionStorageClass)
				if err != nil {
					return err
				}
				lifecycleRule.Transitions = s3.BucketLifecycleConfigurationV2RuleTransitionArray{
					&s3.BucketLifecycleConfigurationV2RuleTransitionArgs{
						Days:         pulumi.Int(rule.TransitionDays),
						StorageClass: pulumi.String(storageClass),
					},
				}
			}

			// Add expiration if specified
			if rule.ExpirationDays > 0 {
				lifecycleRule.Expiration = &s3.BucketLifecycleConfigurationV2RuleExpirationArgs{
					Days: pulumi.Int(rule.ExpirationDays),
				}
			}

			// Add noncurrent version expiration if specified
			if rule.NoncurrentVersionExpirationDays > 0 {
				lifecycleRule.NoncurrentVersionExpiration = &s3.BucketLifecycleConfigurationV2RuleNoncurrentVersionExpirationArgs{
					NoncurrentDays: pulumi.Int(rule.NoncurrentVersionExpirationDays),
				}
			}

			// Add abort incomplete multipart upload if specified
			if rule.AbortIncompleteMultipartUploadDays > 0 {
				lifecycleRule.AbortIncompleteMultipartUpload = &s3.BucketLifecycleConfigurationV2RuleAbortIncompleteMultipartUploadArgs{
					DaysAfterInitiation: pulumi.Int(rule.AbortIncompleteMultipartUploadDays),
				}
			}

			lifecycleRules = append(lifecycleRules, lifecycleRule)
		}

		_, err = s3.NewBucketLifecycleConfigurationV2(ctx, "lifecycle", &s3.BucketLifecycleConfigurationV2Args{
			Bucket: bucket.ID(),
			Rules:  lifecycleRules,
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to configure lifecycle rules")
		}
	}

	// NOTE: Replication configuration is not implemented in Pulumi due to SDK compatibility issues.
	// Use Terraform or AWS Console/CLI to configure replication if needed.
	// For production use, prefer Terraform module which has full replication support.
	if spec.Replication != nil && spec.Replication.Enabled {
		ctx.Log.Warn("Replication configuration is specified but not supported in Pulumi implementation. "+
			"Please use Terraform module for replication support.", nil)
	}

	// Configure logging if specified
	if spec.Logging != nil && spec.Logging.Enabled {
		if spec.Logging.TargetBucket == "" {
			return errors.New("logging target_bucket is required")
		}

		_, err = s3.NewBucketLoggingV2(ctx, "logging", &s3.BucketLoggingV2Args{
			Bucket:       bucket.ID(),
			TargetBucket: pulumi.String(spec.Logging.TargetBucket),
			TargetPrefix: pulumi.String(spec.Logging.TargetPrefix),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to configure logging")
		}
	}

	// Configure CORS if specified
	if spec.Cors != nil && len(spec.Cors.CorsRules) > 0 {
		corsRules := s3.BucketCorsConfigurationV2CorsRuleArray{}
		for _, rule := range spec.Cors.CorsRules {
			if len(rule.AllowedMethods) == 0 || len(rule.AllowedOrigins) == 0 {
				return errors.New("CORS rule must have at least one allowed method and origin")
			}

			allowedMethods := pulumi.StringArray{}
			for _, method := range rule.AllowedMethods {
				allowedMethods = append(allowedMethods, pulumi.String(method))
			}

			allowedOrigins := pulumi.StringArray{}
			for _, origin := range rule.AllowedOrigins {
				allowedOrigins = append(allowedOrigins, pulumi.String(origin))
			}

			corsRule := &s3.BucketCorsConfigurationV2CorsRuleArgs{
				AllowedMethods: allowedMethods,
				AllowedOrigins: allowedOrigins,
			}

			if len(rule.AllowedHeaders) > 0 {
				allowedHeaders := pulumi.StringArray{}
				for _, header := range rule.AllowedHeaders {
					allowedHeaders = append(allowedHeaders, pulumi.String(header))
				}
				corsRule.AllowedHeaders = allowedHeaders
			}

			if len(rule.ExposeHeaders) > 0 {
				exposeHeaders := pulumi.StringArray{}
				for _, header := range rule.ExposeHeaders {
					exposeHeaders = append(exposeHeaders, pulumi.String(header))
				}
				corsRule.ExposeHeaders = exposeHeaders
			}

			if rule.MaxAgeSeconds > 0 {
				corsRule.MaxAgeSeconds = pulumi.Int(rule.MaxAgeSeconds)
			}

			corsRules = append(corsRules, corsRule)
		}

		_, err = s3.NewBucketCorsConfigurationV2(ctx, "cors", &s3.BucketCorsConfigurationV2Args{
			Bucket:    bucket.ID(),
			CorsRules: corsRules,
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to configure CORS")
		}
	}

	// Export outputs
	ctx.Export(OpBucketId, bucket.Bucket)
	ctx.Export(OpBucketArn, bucket.Arn)
	ctx.Export(OpRegion, pulumi.String(spec.AwsRegion))
	ctx.Export(OpBucketRegionalDomainName, bucket.BucketRegionalDomainName)
	ctx.Export(OpHostedZoneId, bucket.HostedZoneId)

	return nil
}

// mapStorageClass maps proto enum to AWS storage class string
func mapStorageClass(storageClass awss3bucketv1.AwsS3BucketSpec_StorageClass) (string, error) {
	switch storageClass {
	case awss3bucketv1.AwsS3BucketSpec_STORAGE_CLASS_STANDARD:
		return "STANDARD", nil
	case awss3bucketv1.AwsS3BucketSpec_STORAGE_CLASS_STANDARD_IA:
		return "STANDARD_IA", nil
	case awss3bucketv1.AwsS3BucketSpec_STORAGE_CLASS_ONE_ZONE_IA:
		return "ONEZONE_IA", nil
	case awss3bucketv1.AwsS3BucketSpec_STORAGE_CLASS_INTELLIGENT_TIERING:
		return "INTELLIGENT_TIERING", nil
	case awss3bucketv1.AwsS3BucketSpec_STORAGE_CLASS_GLACIER_INSTANT_RETRIEVAL:
		return "GLACIER_IR", nil
	case awss3bucketv1.AwsS3BucketSpec_STORAGE_CLASS_GLACIER_FLEXIBLE_RETRIEVAL:
		return "GLACIER", nil
	case awss3bucketv1.AwsS3BucketSpec_STORAGE_CLASS_GLACIER_DEEP_ARCHIVE:
		return "DEEP_ARCHIVE", nil
	default:
		return "", fmt.Errorf("unsupported storage class: %v", storageClass)
	}
}
