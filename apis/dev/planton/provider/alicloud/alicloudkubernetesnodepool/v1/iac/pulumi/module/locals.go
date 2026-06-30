package module

import (
	"strings"

	alicloudkubernetesnodepoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudkubernetesnodepool/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudKubernetesNodePool *alicloudkubernetesnodepoolv1.AliCloudKubernetesNodePool
	ClusterId                  string
	VswitchIds                 []string
	SecurityGroupIds           []string
	Tags                       map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudkubernetesnodepoolv1.AliCloudKubernetesNodePoolStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudKubernetesNodePool = stackInput.Target
	target := stackInput.Target

	locals.ClusterId = target.Spec.ClusterId.GetValue()

	for _, ref := range target.Spec.VswitchIds {
		locals.VswitchIds = append(locals.VswitchIds, ref.GetValue())
	}

	for _, ref := range target.Spec.SecurityGroupIds {
		locals.SecurityGroupIds = append(locals.SecurityGroupIds, ref.GetValue())
	}

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudKubernetesNodePool.String()),
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

func imageType(spec *alicloudkubernetesnodepoolv1.AliCloudKubernetesNodePoolSpec) string {
	if spec.ImageType != nil {
		return *spec.ImageType
	}
	return "AliyunLinux3"
}

func systemDiskCategory(disk *alicloudkubernetesnodepoolv1.AliCloudKubernetesNodePoolSystemDisk) string {
	if disk != nil && disk.Category != nil {
		return *disk.Category
	}
	return "cloud_essd"
}

func systemDiskSize(disk *alicloudkubernetesnodepoolv1.AliCloudKubernetesNodePoolSystemDisk) int {
	if disk != nil && disk.Size != nil {
		return int(*disk.Size)
	}
	return 120
}

func instanceChargeType(spec *alicloudkubernetesnodepoolv1.AliCloudKubernetesNodePoolSpec) string {
	if spec.InstanceChargeType != nil {
		return *spec.InstanceChargeType
	}
	return "PostPaid"
}

func installCloudMonitor(spec *alicloudkubernetesnodepoolv1.AliCloudKubernetesNodePoolSpec) bool {
	if spec.InstallCloudMonitor != nil {
		return *spec.InstallCloudMonitor
	}
	return true
}
