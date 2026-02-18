package module

import (
	"strings"

	alicloudrampolicyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudrampolicy/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudRamPolicy *alicloudrampolicyv1.AlicloudRamPolicy
	Tags              map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudrampolicyv1.AlicloudRamPolicyStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudRamPolicy = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudRamPolicy.String()),
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

func forceDelete(spec *alicloudrampolicyv1.AlicloudRamPolicySpec) bool {
	if spec.Force != nil {
		return *spec.Force
	}
	return false
}
