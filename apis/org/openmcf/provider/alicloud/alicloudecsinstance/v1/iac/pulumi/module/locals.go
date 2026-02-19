package module

import (
	"strings"

	alicloudecsinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudecsinstance/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudEcsInstance *alicloudecsinstancev1.AlicloudEcsInstance
	Tags                map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudecsinstancev1.AlicloudEcsInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudEcsInstance = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudEcsInstance.String()),
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

func instanceName(locals *Locals) string {
	if locals.AlicloudEcsInstance.Spec.InstanceName != "" {
		return locals.AlicloudEcsInstance.Spec.InstanceName
	}
	return locals.AlicloudEcsInstance.Metadata.Name
}

func instanceChargeType(spec *alicloudecsinstancev1.AlicloudEcsInstanceSpec) string {
	if spec.InstanceChargeType != nil && *spec.InstanceChargeType != "" {
		return *spec.InstanceChargeType
	}
	return "PostPaid"
}

func systemDiskCategory(spec *alicloudecsinstancev1.AlicloudEcsInstanceSpec) string {
	if spec.SystemDisk != nil && spec.SystemDisk.Category != nil && *spec.SystemDisk.Category != "" {
		return *spec.SystemDisk.Category
	}
	return "cloud_essd"
}

func systemDiskSize(spec *alicloudecsinstancev1.AlicloudEcsInstanceSpec) int {
	if spec.SystemDisk != nil && spec.SystemDisk.Size != nil {
		return int(*spec.SystemDisk.Size)
	}
	return 40
}

func dataDiskCategory(disk *alicloudecsinstancev1.AlicloudEcsDataDisk) string {
	if disk.Category != nil && *disk.Category != "" {
		return *disk.Category
	}
	return "cloud_essd"
}

func dataDiskDeleteWithInstance(disk *alicloudecsinstancev1.AlicloudEcsDataDisk) bool {
	if disk.DeleteWithInstance != nil {
		return *disk.DeleteWithInstance
	}
	return true
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}

func optionalBool(b *bool) pulumi.BoolPtrInput {
	if b == nil {
		return nil
	}
	return pulumi.Bool(*b)
}

func optionalInt(i *int32) pulumi.IntPtrInput {
	if i == nil {
		return nil
	}
	return pulumi.Int(int(*i))
}

func optionalStringPtr(s *string) pulumi.StringPtrInput {
	if s == nil || *s == "" {
		return nil
	}
	return pulumi.String(*s)
}

func optionalFloat64Ptr(f *float64) pulumi.Float64PtrInput {
	if f == nil {
		return nil
	}
	return pulumi.Float64(*f)
}
