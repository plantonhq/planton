package module

import (
	"fmt"

	"github.com/pkg/errors"
	ociobjectstoragebucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociobjectstoragebucket/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/objectstorage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var lifecycleActionMap = map[ociobjectstoragebucketv1.OciObjectStorageBucketSpec_LifecycleAction]string{
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_lifecycle_archive:           "ARCHIVE",
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_lifecycle_infrequent_access: "INFREQUENT_ACCESS",
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_lifecycle_delete:            "DELETE",
	ociobjectstoragebucketv1.OciObjectStorageBucketSpec_lifecycle_abort:             "ABORT",
}

func lifecyclePolicy(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, bucket *objectstorage.Bucket) error {
	spec := locals.OciObjectStorageBucket.Spec

	if len(spec.LifecycleRules) == 0 {
		return nil
	}

	rules := buildLifecycleRules(spec.LifecycleRules)

	policyName := fmt.Sprintf("%s-lifecycle", locals.BucketName)
	_, err := objectstorage.NewObjectLifecyclePolicy(ctx, policyName, &objectstorage.ObjectLifecyclePolicyArgs{
		Bucket:    pulumi.String(spec.Name),
		Namespace: pulumi.String(spec.Namespace),
		Rules:     rules,
	}, pulumiOciOpt(provider), pulumi.DependsOn([]pulumi.Resource{bucket}))
	if err != nil {
		return errors.Wrap(err, "failed to create object lifecycle policy")
	}

	return nil
}

func buildLifecycleRules(rules []*ociobjectstoragebucketv1.OciObjectStorageBucketSpec_LifecycleRule) objectstorage.ObjectLifecyclePolicyRuleArray {
	result := make(objectstorage.ObjectLifecyclePolicyRuleArray, len(rules))
	for i, rule := range rules {
		action := lifecycleActionMap[rule.Action]

		target := rule.Target
		if target == "" {
			target = "objects"
		}

		args := objectstorage.ObjectLifecyclePolicyRuleArgs{
			Name:       pulumi.String(rule.Name),
			Action:     pulumi.String(action),
			IsEnabled:  pulumi.Bool(rule.IsEnabled),
			TimeAmount: pulumi.String(fmt.Sprintf("%d", rule.TimeAmount)),
			TimeUnit:   pulumi.String(timeUnitToString(rule.TimeUnit)),
			Target:     pulumi.StringPtr(target),
		}

		if rule.ObjectNameFilter != nil {
			args.ObjectNameFilter = buildObjectNameFilter(rule.ObjectNameFilter)
		}

		result[i] = args
	}
	return result
}

func buildObjectNameFilter(filter *ociobjectstoragebucketv1.OciObjectStorageBucketSpec_ObjectNameFilter) *objectstorage.ObjectLifecyclePolicyRuleObjectNameFilterArgs {
	args := &objectstorage.ObjectLifecyclePolicyRuleObjectNameFilterArgs{}

	if len(filter.InclusionPatterns) > 0 {
		args.InclusionPatterns = pulumi.ToStringArray(filter.InclusionPatterns)
	}

	if len(filter.InclusionPrefixes) > 0 {
		args.InclusionPrefixes = pulumi.ToStringArray(filter.InclusionPrefixes)
	}

	if len(filter.ExclusionPatterns) > 0 {
		args.ExclusionPatterns = pulumi.ToStringArray(filter.ExclusionPatterns)
	}

	return args
}
