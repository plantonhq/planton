package module

import (
	"strings"

	alicloudceninstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudceninstance/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudCenInstance *alicloudceninstancev1.AlicloudCenInstance
	Tags                map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudceninstancev1.AlicloudCenInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudCenInstance = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudCenInstance.String()),
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
