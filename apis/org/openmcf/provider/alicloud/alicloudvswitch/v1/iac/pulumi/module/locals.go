package module

import (
	"strings"

	alicloudvswitchv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudvswitch/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudVswitch *alicloudvswitchv1.AliCloudVswitch
	VpcId           string
	Tags            map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudvswitchv1.AliCloudVswitchStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudVswitch = stackInput.Target
	target := stackInput.Target

	// Resolve vpc_id from StringValueOrRef.
	// At IaC runtime, value_from references have already been resolved by the
	// platform, so GetValue() returns the final literal VPC ID.
	locals.VpcId = target.Spec.VpcId.GetValue()

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudVswitch.String()),
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
