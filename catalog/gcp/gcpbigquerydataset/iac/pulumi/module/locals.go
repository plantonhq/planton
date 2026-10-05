package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpbigquerydatasetv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbigquerydataset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpBigQueryDataset *gcpbigquerydatasetv1alpha1.GcpBigQueryDataset
	GcpLabels          map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpbigquerydatasetv1alpha1.GcpBigQueryDatasetIacInput) *Locals {
	locals := &Locals{}
	locals.GcpBigQueryDataset = iacInput.Target

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpBigQueryDataset.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpBigQueryDataset.Spec.DatasetId
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpBigQueryDataset.String())

	if locals.GcpBigQueryDataset.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpBigQueryDataset.Metadata.Org
	}
	if locals.GcpBigQueryDataset.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpBigQueryDataset.Metadata.Env
	}
	if locals.GcpBigQueryDataset.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpBigQueryDataset.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
