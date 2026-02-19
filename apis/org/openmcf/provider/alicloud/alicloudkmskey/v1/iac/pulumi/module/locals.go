package module

import (
	"strings"

	alicloudkmskeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudkmskey/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudKmsKey *alicloudkmskeyv1.AlicloudKmsKey
	Tags           map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudkmskeyv1.AlicloudKmsKeyStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudKmsKey = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudKmsKey.String()),
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

func keySpec(spec *alicloudkmskeyv1.AlicloudKmsKeySpec) string {
	if spec.KeySpec != nil {
		return *spec.KeySpec
	}
	return "Aliyun_AES_256"
}

func keyUsage(spec *alicloudkmskeyv1.AlicloudKmsKeySpec) string {
	if spec.KeyUsage != nil {
		return *spec.KeyUsage
	}
	return "ENCRYPT/DECRYPT"
}

func protectionLevel(spec *alicloudkmskeyv1.AlicloudKmsKeySpec) string {
	if spec.ProtectionLevel != nil {
		return *spec.ProtectionLevel
	}
	return "SOFTWARE"
}

func automaticRotation(spec *alicloudkmskeyv1.AlicloudKmsKeySpec) string {
	if spec.AutomaticRotation != nil && *spec.AutomaticRotation {
		return "Enabled"
	}
	return "Disabled"
}

func pendingWindowInDays(spec *alicloudkmskeyv1.AlicloudKmsKeySpec) int {
	if spec.PendingWindowInDays != nil {
		return int(*spec.PendingWindowInDays)
	}
	return 30
}

func deletionProtection(spec *alicloudkmskeyv1.AlicloudKmsKeySpec) string {
	if spec.DeletionProtection != nil && *spec.DeletionProtection {
		return "Enabled"
	}
	return "Disabled"
}
