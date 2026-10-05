package module

import (
	"strings"

	gcpcloudrundomainmappingv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrundomainmapping/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpCloudRunDomainMapping *gcpcloudrundomainmappingv1alpha1.GcpCloudRunDomainMapping

	// GcpLabels carries the platform attribution labels stored on the
	// mapping object, merged over any user labels so attribution can never
	// be clobbered. metadata.name keys the name label (the GcpDnsZone
	// basis for domain-named kinds): the mapping's cloud-side name is the
	// domain itself, whose dots and 253-char budget don't fit the
	// Knative/K8s 63-char label-value contract.
	GcpLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudrundomainmappingv1alpha1.GcpCloudRunDomainMappingIacInput) *Locals {
	target := iacInput.Target

	gcpLabels := map[string]string{}
	for key, value := range target.Spec.Labels {
		gcpLabels[key] = value
	}
	gcpLabels[gcplabelkeys.Resource] = "true"
	gcpLabels[gcplabelkeys.ResourceName] = target.Metadata.Name
	gcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpCloudRunDomainMapping.String())

	if target.Metadata.Org != "" {
		gcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		gcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}
	if target.Metadata.Id != "" {
		gcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}

	return &Locals{
		GcpCloudRunDomainMapping: target,
		GcpLabels:                gcpLabels,
	}
}
