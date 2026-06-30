package module

import (
	"strings"

	alicloudmongodbinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudmongodbinstance/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudMongodbInstance *alicloudmongodbinstancev1.AliCloudMongodbInstance
	Tags                    map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudmongodbinstancev1.AliCloudMongodbInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudMongodbInstance = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudMongodbInstance.String()),
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
	if locals.AliCloudMongodbInstance.Spec.DbInstanceName != "" {
		return locals.AliCloudMongodbInstance.Spec.DbInstanceName
	}
	return locals.AliCloudMongodbInstance.Metadata.Name
}

func replicationFactor(spec *alicloudmongodbinstancev1.AliCloudMongodbInstanceSpec) int {
	if spec.ReplicationFactor != nil {
		return int(*spec.ReplicationFactor)
	}
	return 3
}

func storageEngine(spec *alicloudmongodbinstancev1.AliCloudMongodbInstanceSpec) string {
	if spec.StorageEngine != nil && *spec.StorageEngine != "" {
		return *spec.StorageEngine
	}
	return "WiredTiger"
}

func instanceChargeType(spec *alicloudmongodbinstancev1.AliCloudMongodbInstanceSpec) string {
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
