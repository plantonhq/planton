package module

import (
	"strings"

	alicloudrampolicyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudrampolicy/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudRamPolicy *alicloudrampolicyv1.AliCloudRamPolicy
	Tags              map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudrampolicyv1.AliCloudRamPolicyStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudRamPolicy = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudRamPolicy.String()),
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

func forceDelete(spec *alicloudrampolicyv1.AliCloudRamPolicySpec) bool {
	if spec.Force != nil {
		return *spec.Force
	}
	return false
}
