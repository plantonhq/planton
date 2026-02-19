package module

import (
	"strings"

	alicloudmongodbinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudmongodbinstance/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudMongodbInstance *alicloudmongodbinstancev1.AlicloudMongodbInstance
	Tags                    map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudmongodbinstancev1.AlicloudMongodbInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudMongodbInstance = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudMongodbInstance.String()),
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
	if locals.AlicloudMongodbInstance.Spec.DbInstanceName != "" {
		return locals.AlicloudMongodbInstance.Spec.DbInstanceName
	}
	return locals.AlicloudMongodbInstance.Metadata.Name
}

func replicationFactor(spec *alicloudmongodbinstancev1.AlicloudMongodbInstanceSpec) int {
	if spec.ReplicationFactor != nil {
		return int(*spec.ReplicationFactor)
	}
	return 3
}

func storageEngine(spec *alicloudmongodbinstancev1.AlicloudMongodbInstanceSpec) string {
	if spec.StorageEngine != nil && *spec.StorageEngine != "" {
		return *spec.StorageEngine
	}
	return "WiredTiger"
}

func instanceChargeType(spec *alicloudmongodbinstancev1.AlicloudMongodbInstanceSpec) string {
	if spec.InstanceChargeType != nil && *spec.InstanceChargeType != "" {
		return *spec.InstanceChargeType
	}
	return "PostPaid"
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
