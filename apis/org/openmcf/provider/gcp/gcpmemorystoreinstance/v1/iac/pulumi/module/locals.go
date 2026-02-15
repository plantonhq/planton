package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	gcpmemorystoreinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpmemorystoreinstance/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig      *gcpprovider.GcpProviderConfig
	GcpMemorystoreInstance *gcpmemorystoreinstancev1.GcpMemorystoreInstance
	GcpLabels              map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpmemorystoreinstancev1.GcpMemorystoreInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.GcpMemorystoreInstance = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpMemorystoreInstance.Spec.InstanceName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpMemorystoreInstance.String()),
	}

	if locals.GcpMemorystoreInstance.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpMemorystoreInstance.Metadata.Org
	}
	if locals.GcpMemorystoreInstance.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpMemorystoreInstance.Metadata.Env
	}
	if locals.GcpMemorystoreInstance.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpMemorystoreInstance.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
