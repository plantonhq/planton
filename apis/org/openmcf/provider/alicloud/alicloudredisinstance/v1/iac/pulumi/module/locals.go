package module

import (
	"strings"

	alicloudredisinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudredisinstance/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudRedisInstance *alicloudredisinstancev1.AlicloudRedisInstance
	Tags                  map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudredisinstancev1.AlicloudRedisInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudRedisInstance = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudRedisInstance.String()),
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
	if locals.AlicloudRedisInstance.Spec.DbInstanceName != "" {
		return locals.AlicloudRedisInstance.Spec.DbInstanceName
	}
	return locals.AlicloudRedisInstance.Metadata.Name
}

func engineVersion(spec *alicloudredisinstancev1.AlicloudRedisInstanceSpec) string {
	if spec.EngineVersion != nil && *spec.EngineVersion != "" {
		return *spec.EngineVersion
	}
	return "7.0"
}

func instanceType(spec *alicloudredisinstancev1.AlicloudRedisInstanceSpec) string {
	if spec.InstanceType != nil && *spec.InstanceType != "" {
		return *spec.InstanceType
	}
	return "Redis"
}

func paymentType(spec *alicloudredisinstancev1.AlicloudRedisInstanceSpec) string {
	if spec.PaymentType != nil && *spec.PaymentType != "" {
		return *spec.PaymentType
	}
	return "PostPaid"
}

func vpcAuthMode(spec *alicloudredisinstancev1.AlicloudRedisInstanceSpec) string {
	if spec.VpcAuthMode != nil && *spec.VpcAuthMode != "" {
		return *spec.VpcAuthMode
	}
	return "Open"
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
