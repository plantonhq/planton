package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpbigquerydatasetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpbigquerydataset/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
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
