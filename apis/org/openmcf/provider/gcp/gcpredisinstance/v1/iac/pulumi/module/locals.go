package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	gcpredisinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpredisinstance/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpRedisInstance  *gcpredisinstancev1.GcpRedisInstance
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpredisinstancev1.GcpRedisInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.GcpRedisInstance = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpRedisInstance.Spec.InstanceName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpRedisInstance.String()),
	}

	if locals.GcpRedisInstance.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpRedisInstance.Metadata.Org
	}
	if locals.GcpRedisInstance.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpRedisInstance.Metadata.Env
	}
	if locals.GcpRedisInstance.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpRedisInstance.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
