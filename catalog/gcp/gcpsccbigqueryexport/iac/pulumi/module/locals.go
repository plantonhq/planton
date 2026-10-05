package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpsccbigqueryexportv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccbigqueryexport/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpSccBigQueryExport *gcpsccbigqueryexportv1alpha1.GcpSccBigQueryExport
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpsccbigqueryexportv1alpha1.GcpSccBigQueryExportIacInput) *Locals {
	locals := &Locals{}
	locals.GcpSccBigQueryExport = iacInput.Target

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
