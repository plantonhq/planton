package module

import (
	"strings"

	alicloudrdsinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudrdsinstance/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudRdsInstance *alicloudrdsinstancev1.AlicloudRdsInstance
	Tags                map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudrdsinstancev1.AlicloudRdsInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudRdsInstance = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudRdsInstance.String()),
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
	if locals.AlicloudRdsInstance.Spec.InstanceName != "" {
		return locals.AlicloudRdsInstance.Spec.InstanceName
	}
	return locals.AlicloudRdsInstance.Metadata.Name
}

func instanceChargeType(spec *alicloudrdsinstancev1.AlicloudRdsInstanceSpec) string {
	if spec.InstanceChargeType != nil {
		return *spec.InstanceChargeType
	}
	return "Postpaid"
}

func category(spec *alicloudrdsinstancev1.AlicloudRdsInstanceSpec) string {
	if spec.Category != nil {
		return *spec.Category
	}
	return "HighAvailability"
}

func accountType(acct *alicloudrdsinstancev1.AlicloudRdsAccount) string {
	if acct.AccountType != nil {
		return *acct.AccountType
	}
	return "Normal"
}

func privilege(priv *alicloudrdsinstancev1.AlicloudRdsAccountPrivilege) string {
	if priv.Privilege != nil {
		return *priv.Privilege
	}
	return "ReadOnly"
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
