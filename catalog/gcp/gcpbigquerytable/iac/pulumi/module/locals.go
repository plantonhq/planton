package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpbigquerytablev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbigquerytable/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpBigQueryTable  *gcpbigquerytablev1alpha1.GcpBigQueryTable
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpbigquerytablev1alpha1.GcpBigQueryTableIacInput) *Locals {
	locals := &Locals{}
	locals.GcpBigQueryTable = iacInput.Target

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpBigQueryTable.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpBigQueryTable.Spec.TableId
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpBigQueryTable.String())

	if locals.GcpBigQueryTable.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpBigQueryTable.Metadata.Org
	}
	if locals.GcpBigQueryTable.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpBigQueryTable.Metadata.Env
	}
	if locals.GcpBigQueryTable.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpBigQueryTable.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
