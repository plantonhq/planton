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

func initializeLocals(_ *pulumi.Context, stackInput *gcpsccbigqueryexportv1alpha1.GcpSccBigQueryExportStackInput) *Locals {
	locals := &Locals{}
	locals.GcpSccBigQueryExport = stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
