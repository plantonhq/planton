package module

import (
	"strings"

	alicloudlogprojectv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudlogproject/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudLogProject *alicloudlogprojectv1.AliCloudLogProject
	Tags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudlogprojectv1.AliCloudLogProjectStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudLogProject = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudLogProject.String()),
	}

	if target.Metadata.Id != "" {
		locals.Tags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Tags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Tags["environment"] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.Tags[k] = v
	}

	return locals
}

// logStoreRetentionDays returns the retention period, defaulting to 30 when unset.
func logStoreRetentionDays(ls *alicloudlogprojectv1.AliCloudLogStore) int {
	if ls.RetentionDays != nil {
		return int(*ls.RetentionDays)
	}
	return 30
}

// logStoreShardCount returns the shard count, defaulting to 2 when unset.
func logStoreShardCount(ls *alicloudlogprojectv1.AliCloudLogStore) int {
	if ls.ShardCount != nil {
		return int(*ls.ShardCount)
	}
	return 2
}

// logStoreAutoSplit returns the auto_split flag, defaulting to true when unset.
func logStoreAutoSplit(ls *alicloudlogprojectv1.AliCloudLogStore) bool {
	if ls.AutoSplit != nil {
		return *ls.AutoSplit
	}
	return true
}

// logStoreMaxSplitShardCount returns the max split shard count, defaulting to 64 when unset.
func logStoreMaxSplitShardCount(ls *alicloudlogprojectv1.AliCloudLogStore) int {
	if ls.MaxSplitShardCount != nil {
		return int(*ls.MaxSplitShardCount)
	}
	return 64
}

// logStoreEnableIndex returns the enable_index flag, defaulting to true when unset.
func logStoreEnableIndex(ls *alicloudlogprojectv1.AliCloudLogStore) bool {
	if ls.EnableIndex != nil {
		return *ls.EnableIndex
	}
	return true
}

// logStoreAppendMeta returns the append_meta flag, defaulting to true when unset.
func logStoreAppendMeta(ls *alicloudlogprojectv1.AliCloudLogStore) bool {
	if ls.AppendMeta != nil {
		return *ls.AppendMeta
	}
	return true
}
