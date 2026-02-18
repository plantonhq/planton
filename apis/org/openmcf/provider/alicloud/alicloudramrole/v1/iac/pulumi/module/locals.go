package module

import (
	"strings"

	alicloudramrolev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudramrole/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudRamRole *alicloudramrolev1.AlicloudRamRole
	Tags            map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudramrolev1.AlicloudRamRoleStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudRamRole = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudRamRole.String()),
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

func maxSessionDuration(spec *alicloudramrolev1.AlicloudRamRoleSpec) int {
	if spec.MaxSessionDuration != nil {
		return int(*spec.MaxSessionDuration)
	}
	return 3600
}

func forceDelete(spec *alicloudramrolev1.AlicloudRamRoleSpec) bool {
	if spec.Force != nil {
		return *spec.Force
	}
	return false
}

func policyType(pa *alicloudramrolev1.AlicloudRamRolePolicyAttachment) string {
	if pa.PolicyType != nil {
		return *pa.PolicyType
	}
	return "System"
}
