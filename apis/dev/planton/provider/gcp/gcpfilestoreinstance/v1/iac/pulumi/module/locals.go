package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpfilestoreinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpfilestoreinstance/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpFilestoreInstance *gcpfilestoreinstancev1.GcpFilestoreInstance
	GcpLabels            map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpfilestoreinstancev1.GcpFilestoreInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.GcpFilestoreInstance = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpFilestoreInstance.Spec.InstanceName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpFilestoreInstance.String()),
	}

	if locals.GcpFilestoreInstance.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpFilestoreInstance.Metadata.Org
	}
	if locals.GcpFilestoreInstance.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpFilestoreInstance.Metadata.Env
	}
	if locals.GcpFilestoreInstance.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpFilestoreInstance.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
