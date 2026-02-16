package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	gcpbigquerydatasetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpbigquerydataset/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpBigQueryDataset *gcpbigquerydatasetv1.GcpBigQueryDataset
	GcpLabels          map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpbigquerydatasetv1.GcpBigQueryDatasetStackInput) *Locals {
	locals := &Locals{}
	locals.GcpBigQueryDataset = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpBigQueryDataset.Spec.DatasetId,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpBigQueryDataset.String()),
	}

	if locals.GcpBigQueryDataset.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpBigQueryDataset.Metadata.Org
	}
	if locals.GcpBigQueryDataset.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpBigQueryDataset.Metadata.Env
	}
	if locals.GcpBigQueryDataset.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpBigQueryDataset.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
