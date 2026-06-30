package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpbigtableinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpbigtableinstance/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig   *gcpprovider.GcpProviderConfig
	GcpBigtableInstance *gcpbigtableinstancev1.GcpBigtableInstance
	GcpLabels           map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpbigtableinstancev1.GcpBigtableInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.GcpBigtableInstance = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpBigtableInstance.Spec.InstanceName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpBigtableInstance.String()),
	}

	if locals.GcpBigtableInstance.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpBigtableInstance.Metadata.Org
	}
	if locals.GcpBigtableInstance.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpBigtableInstance.Metadata.Env
	}
	if locals.GcpBigtableInstance.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpBigtableInstance.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
