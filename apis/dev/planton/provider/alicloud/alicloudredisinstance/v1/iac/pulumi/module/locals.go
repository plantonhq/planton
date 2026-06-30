package module

import (
	"strings"

	alicloudredisinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudredisinstance/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudRedisInstance *alicloudredisinstancev1.AliCloudRedisInstance
	Tags                  map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudredisinstancev1.AliCloudRedisInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudRedisInstance = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudRedisInstance.String()),
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
	if locals.AliCloudRedisInstance.Spec.DbInstanceName != "" {
		return locals.AliCloudRedisInstance.Spec.DbInstanceName
	}
	return locals.AliCloudRedisInstance.Metadata.Name
}

func engineVersion(spec *alicloudredisinstancev1.AliCloudRedisInstanceSpec) string {
	if spec.EngineVersion != nil && *spec.EngineVersion != "" {
		return *spec.EngineVersion
	}
	return "7.0"
}

func instanceType(spec *alicloudredisinstancev1.AliCloudRedisInstanceSpec) string {
	if spec.InstanceType != nil && *spec.InstanceType != "" {
		return *spec.InstanceType
	}
	return "Redis"
}

func paymentType(spec *alicloudredisinstancev1.AliCloudRedisInstanceSpec) string {
	if spec.PaymentType != nil && *spec.PaymentType != "" {
		return *spec.PaymentType
	}
	return "PostPaid"
}

func vpcAuthMode(spec *alicloudredisinstancev1.AliCloudRedisInstanceSpec) string {
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
