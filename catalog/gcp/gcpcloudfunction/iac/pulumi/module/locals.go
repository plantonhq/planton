package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudfunctionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudfunction/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpCloudFunction  *gcpcloudfunctionv1alpha1.GcpCloudFunction
	GcpLabels         map[string]string
	FunctionName      string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudfunctionv1alpha1.GcpCloudFunctionIacInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudFunction = iacInput.Target

	// Function name defaults to metadata.name.
	locals.FunctionName = locals.GcpCloudFunction.Spec.FunctionName
	if locals.FunctionName == "" {
		locals.FunctionName = locals.GcpCloudFunction.Metadata.Name
	}

	// User labels merge BENEATH the platform attribution labels so the
	// platform's keys always win — the same order the Terraform module
	// applies.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpCloudFunction.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.FunctionName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpCloudFunction.String())

	if locals.GcpCloudFunction.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpCloudFunction.Metadata.Org
	}
	if locals.GcpCloudFunction.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpCloudFunction.Metadata.Env
	}
	if locals.GcpCloudFunction.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpCloudFunction.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
