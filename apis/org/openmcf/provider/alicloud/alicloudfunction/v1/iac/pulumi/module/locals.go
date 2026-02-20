package module

import (
	"strings"

	alicloudfunctionv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudfunction/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudFunction *alicloudfunctionv1.AlicloudFunction
	Tags             map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudfunctionv1.AlicloudFunctionStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudFunction = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudFunction.String()),
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

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}

func optionalInt(v *int32) pulumi.IntPtrInput {
	if v == nil {
		return nil
	}
	return pulumi.Int(int(*v))
}

func optionalFloat64(v *float64) pulumi.Float64PtrInput {
	if v == nil {
		return nil
	}
	return pulumi.Float64(*v)
}

func optionalBool(v *bool) pulumi.BoolPtrInput {
	if v == nil {
		return nil
	}
	return pulumi.Bool(*v)
}
