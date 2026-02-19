package module

import (
	"strings"

	alicloudsecuritygroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudsecuritygroup/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudSecurityGroup *alicloudsecuritygroupv1.AlicloudSecurityGroup
	Tags                  map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudsecuritygroupv1.AlicloudSecurityGroupStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudSecurityGroup = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudSecurityGroup.String()),
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

func rulePortRange(rule *alicloudsecuritygroupv1.AlicloudSecurityGroupRule) string {
	if rule.PortRange != nil {
		return *rule.PortRange
	}
	return "-1/-1"
}

func rulePriority(rule *alicloudsecuritygroupv1.AlicloudSecurityGroupRule) int {
	if rule.Priority != nil {
		return int(*rule.Priority)
	}
	return 1
}

func rulePolicy(rule *alicloudsecuritygroupv1.AlicloudSecurityGroupRule) string {
	if rule.Policy != nil {
		return *rule.Policy
	}
	return "accept"
}

func innerAccessPolicy(spec *alicloudsecuritygroupv1.AlicloudSecurityGroupSpec) string {
	if spec.InnerAccessPolicy != nil {
		return *spec.InnerAccessPolicy
	}
	return "Accept"
}
