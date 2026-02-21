package module

import (
	"github.com/pkg/errors"
	alicloudstoragebucketv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudstoragebucket/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/oss"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudstoragebucketv1.AliCloudStorageBucketStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudStorageBucket.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	bucketArgs := &oss.BucketArgs{
		Bucket:          pulumi.String(spec.BucketName),
		Acl:             optionalStringFromPtr(spec.Acl),
		StorageClass:    optionalStringFromPtr(spec.StorageClass),
		RedundancyType:  optionalStringFromPtr(spec.RedundancyType),
		ForceDestroy:    pulumi.Bool(spec.ForceDestroy),
		ResourceGroupId: optionalString(spec.ResourceGroupId),
		Tags:            pulumi.ToStringMap(locals.Tags),
	}

	if spec.VersioningEnabled {
		bucketArgs.Versioning = &oss.BucketVersioningTypeArgs{
			Status: pulumi.String("Enabled"),
		}
	}

	if spec.ServerSideEncryption != nil {
		bucketArgs.ServerSideEncryptionRule = &oss.BucketServerSideEncryptionRuleArgs{
			SseAlgorithm:   pulumi.String(spec.ServerSideEncryption.SseAlgorithm),
			KmsMasterKeyId: optionalString(spec.ServerSideEncryption.KmsMasterKeyId),
		}
	}

	if spec.Logging != nil {
		bucketArgs.Logging = &oss.BucketLoggingTypeArgs{
			TargetBucket: pulumi.String(spec.Logging.TargetBucket),
			TargetPrefix: optionalString(spec.Logging.TargetPrefix),
		}
	}

	if len(spec.CorsRules) > 0 {
		bucketArgs.CorsRules = buildCorsRules(spec.CorsRules)
	}

	if len(spec.LifecycleRules) > 0 {
		bucketArgs.LifecycleRules = buildLifecycleRules(spec.LifecycleRules)
	}

	bucket, err := oss.NewBucket(ctx, spec.BucketName, bucketArgs, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create OSS bucket %s", spec.BucketName)
	}

	ctx.Export(OpBucketName, bucket.Bucket)
	ctx.Export(OpExtranetEndpoint, bucket.ExtranetEndpoint)
	ctx.Export(OpIntranetEndpoint, bucket.IntranetEndpoint)

	return nil
}

func buildCorsRules(rules []*alicloudstoragebucketv1.AliCloudStorageBucketCorsRule) oss.BucketCorsRuleArray {
	var result oss.BucketCorsRuleArray
	for _, r := range rules {
		result = append(result, &oss.BucketCorsRuleArgs{
			AllowedOrigins: pulumi.ToStringArray(r.AllowedOrigins),
			AllowedMethods: pulumi.ToStringArray(r.AllowedMethods),
			AllowedHeaders: pulumi.ToStringArray(r.AllowedHeaders),
			ExposeHeaders:  pulumi.ToStringArray(r.ExposeHeaders),
			MaxAgeSeconds:  pulumi.Int(int(r.MaxAgeSeconds)),
		})
	}
	return result
}

func buildLifecycleRules(rules []*alicloudstoragebucketv1.AliCloudStorageBucketLifecycleRule) oss.BucketLifecycleRuleArray {
	var result oss.BucketLifecycleRuleArray
	for i, r := range rules {
		ruleArgs := &oss.BucketLifecycleRuleArgs{
			Prefix:  pulumi.String(r.Prefix),
			Enabled: pulumi.Bool(r.Enabled),
			Id:      pulumi.Sprintf("rule-%d", i),
		}

		if r.ExpirationDays > 0 {
			ruleArgs.Expirations = oss.BucketLifecycleRuleExpirationArray{
				&oss.BucketLifecycleRuleExpirationArgs{
					Days: pulumi.Int(int(r.ExpirationDays)),
				},
			}
		}

		if len(r.Transitions) > 0 {
			var transitions oss.BucketLifecycleRuleTransitionArray
			for _, t := range r.Transitions {
				transitions = append(transitions, &oss.BucketLifecycleRuleTransitionArgs{
					Days:         pulumi.Int(int(t.Days)),
					StorageClass: pulumi.String(t.StorageClass),
				})
			}
			ruleArgs.Transitions = transitions
		}

		if r.AbortMultipartUploadDays > 0 {
			ruleArgs.AbortMultipartUploads = oss.BucketLifecycleRuleAbortMultipartUploadArray{
				&oss.BucketLifecycleRuleAbortMultipartUploadArgs{
					Days: pulumi.Int(int(r.AbortMultipartUploadDays)),
				},
			}
		}

		if r.NoncurrentVersionExpirationDays > 0 {
			ruleArgs.NoncurrentVersionExpirations = oss.BucketLifecycleRuleNoncurrentVersionExpirationArray{
				&oss.BucketLifecycleRuleNoncurrentVersionExpirationArgs{
					Days: pulumi.Int(int(r.NoncurrentVersionExpirationDays)),
				},
			}
		}

		result = append(result, ruleArgs)
	}
	return result
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}

func optionalStringFromPtr(s *string) pulumi.StringPtrInput {
	if s == nil || *s == "" {
		return nil
	}
	return pulumi.String(*s)
}
