package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpspannerinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpspannerinstance/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpSpannerInstance *gcpspannerinstancev1.GcpSpannerInstance
	GcpLabels          map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpspannerinstancev1.GcpSpannerInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.GcpSpannerInstance = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpSpannerInstance.Spec.InstanceName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpSpannerInstance.String()),
	}

	if locals.GcpSpannerInstance.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpSpannerInstance.Metadata.Org
	}
	if locals.GcpSpannerInstance.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpSpannerInstance.Metadata.Env
	}
	if locals.GcpSpannerInstance.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpSpannerInstance.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
