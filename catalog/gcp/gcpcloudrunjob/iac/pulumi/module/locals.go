package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudrunjobv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrunjob/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpCloudRunJob    *gcpcloudrunjobv1alpha1.GcpCloudRunJob
	GcpLabels         map[string]string
	JobName           string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudrunjobv1alpha1.GcpCloudRunJobIacInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudRunJob = iacInput.Target

	locals.JobName = locals.GcpCloudRunJob.Spec.JobName
	if locals.JobName == "" {
		locals.JobName = locals.GcpCloudRunJob.Metadata.Name
	}

	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpCloudRunJob.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.JobName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpCloudRunJob.String())

	if locals.GcpCloudRunJob.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpCloudRunJob.Metadata.Org
	}
	if locals.GcpCloudRunJob.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpCloudRunJob.Metadata.Env
	}
	if locals.GcpCloudRunJob.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpCloudRunJob.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
