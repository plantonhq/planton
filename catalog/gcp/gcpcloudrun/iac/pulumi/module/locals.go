package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudrunv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrun/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpCloudRun       *gcpcloudrunv1alpha1.GcpCloudRun
	// GcpLabels carries the platform attribution labels applied on the
	// service object, merged over any user labels so attribution can never
	// be clobbered.
	GcpLabels map[string]string
	// ServiceName is the cloud-side service name: spec.service_name when
	// set, otherwise metadata.name (the spec-level contract).
	ServiceName string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudrunv1alpha1.GcpCloudRunIacInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudRun = iacInput.Target

	locals.ServiceName = locals.GcpCloudRun.Spec.ServiceName
	if locals.ServiceName == "" {
		locals.ServiceName = locals.GcpCloudRun.Metadata.Name
	}

	// User labels merge in first so the platform attribution labels can
	// never be clobbered by a spec label with the same key. The service
	// name (not metadata.name) keys the name label so the label matches
	// what is visible in the GCP console.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpCloudRun.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.ServiceName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpCloudRun.String())

	if locals.GcpCloudRun.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpCloudRun.Metadata.Org
	}
	if locals.GcpCloudRun.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpCloudRun.Metadata.Env
	}
	if locals.GcpCloudRun.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpCloudRun.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
