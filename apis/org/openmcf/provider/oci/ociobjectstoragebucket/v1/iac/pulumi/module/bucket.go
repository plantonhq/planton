package module

import (
	"fmt"

	"github.com/pkg/errors"
	ociobjectstoragebucketv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ociobjectstoragebucket/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/objectstorage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var accessTypeMap = map[ociobjectstoragebucketv1.OciObjectStorageBucketSpec_AccessType]string{
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_no_public_access:        "NoPublicAccess",
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_object_read:             "ObjectRead",
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_object_read_without_list: "ObjectReadWithoutList",
}

var storageTierMap = map[ociobjectstoragebucketv1.OciObjectStorageBucketSpec_StorageTier]string{
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_standard: "Standard",
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_archive:  "Archive",
}

var versioningMap = map[ociobjectstoragebucketv1.OciObjectStorageBucketSpec_Versioning]string{
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_enabled:   "Enabled",
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_disabled:  "Disabled",
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_suspended: "Suspended",
}

var autoTieringMap = map[ociobjectstoragebucketv1.OciObjectStorageBucketSpec_AutoTiering]string{
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_auto_tiering_disabled: "Disabled",
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_infrequent_access:     "InfrequentAccess",
}

func bucket(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*objectstorage.Bucket, error) {
	spec := locals.OciObjectStorageBucket.Spec

	bucketArgs := &objectstorage.BucketArgs{
		CompartmentId:       pulumi.String(spec.CompartmentId.GetValue()),
		Namespace:           pulumi.String(spec.Namespace),
		Name:                pulumi.String(spec.Name),
		ObjectEventsEnabled: pulumi.Bool(spec.ObjectEventsEnabled),
		FreeformTags:        pulumi.ToStringMap(locals.FreeformTags),
	}

	if v, ok := accessTypeMap[spec.AccessType]; ok {
		bucketArgs.AccessType = pulumi.StringPtr(v)
	}

	if v, ok := storageTierMap[spec.StorageTier]; ok {
		bucketArgs.StorageTier = pulumi.StringPtr(v)
	}

	if v, ok := versioningMap[spec.Versioning]; ok {
		bucketArgs.Versioning = pulumi.StringPtr(v)
	}

	if v, ok := autoTieringMap[spec.AutoTiering]; ok {
		bucketArgs.AutoTiering = pulumi.StringPtr(v)
	}

	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		bucketArgs.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	if len(spec.Metadata) > 0 {
		bucketArgs.Metadata = pulumi.ToStringMap(spec.Metadata)
	}

	if len(spec.RetentionRules) > 0 {
		bucketArgs.RetentionRules = buildRetentionRules(spec.RetentionRules)
	}

	createdBucket, err := objectstorage.NewBucket(ctx, locals.BucketName, bucketArgs, pulumiOciOpt(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create object storage bucket")
	}

	ctx.Export(OpBucketId, createdBucket.BucketId)

	return createdBucket, nil
}

func buildRetentionRules(rules []*ociobjectstoragebucketv1.OciObjectStorageBucketSpec_RetentionRule) objectstorage.BucketRetentionRuleArray {
	result := make(objectstorage.BucketRetentionRuleArray, len(rules))
	for i, rule := range rules {
		args := objectstorage.BucketRetentionRuleArgs{
			DisplayName: pulumi.String(rule.DisplayName),
		}

		if rule.Duration != nil {
			args.Duration = &objectstorage.BucketRetentionRuleDurationArgs{
				TimeAmount: pulumi.String(fmt.Sprintf("%d", rule.Duration.TimeAmount)),
				TimeUnit:   pulumi.String(timeUnitToString(rule.Duration.TimeUnit)),
			}
		}

		if rule.TimeRuleLocked != "" {
			args.TimeRuleLocked = pulumi.StringPtr(rule.TimeRuleLocked)
		}

		result[i] = args
	}
	return result
}

func timeUnitToString(unit ociobjectstoragebucketv1.OciObjectStorageBucketSpec_TimeUnit) string {
	switch unit {
	case ociobjectstoragebucketv1.OciObjectStorageBucketSpec_days:
		return "DAYS"
	case ociobjectstoragebucketv1.OciObjectStorageBucketSpec_years:
		return "YEARS"
	default:
		return "DAYS"
	}
}
