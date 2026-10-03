package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpgkenodepoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgkenodepool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpGkeNodePool    *gcpgkenodepoolv1alpha1.GcpGkeNodePool
	// GcpLabels carries the platform attribution labels applied as GCE
	// resource labels on the node VMs, merged over any user resource
	// labels so attribution can never be clobbered.
	GcpLabels map[string]string
	// NodePoolName is the cloud-side pool name: spec.node_pool_name when
	// set, otherwise metadata.name (the spec-level contract). Empty when
	// name_prefix drives naming — GKE then generates the full name.
	NodePoolName string
	// AttributionName keys the name label: the explicit pool name, or the
	// prefix for prefixed pools whose final name is only known after
	// create.
	AttributionName string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpgkenodepoolv1alpha1.GcpGkeNodePoolIacInput) *Locals {
	locals := &Locals{}
	locals.GcpGkeNodePool = iacInput.Target

	if locals.GcpGkeNodePool.Spec.NamePrefix != "" {
		locals.NodePoolName = ""
		locals.AttributionName = locals.GcpGkeNodePool.Spec.NamePrefix
	} else {
		locals.NodePoolName = locals.GcpGkeNodePool.Spec.NodePoolName
		if locals.NodePoolName == "" {
			locals.NodePoolName = locals.GcpGkeNodePool.Metadata.Name
		}
		locals.AttributionName = locals.NodePoolName
	}

	// User resource labels merge in first so the platform attribution labels
	// can never be clobbered by a spec label with the same key. The pool
	// name (not metadata.name) keys the name label so the label matches
	// what is visible in the GCP console.
	locals.GcpLabels = map[string]string{}
	if nodeConfig := locals.GcpGkeNodePool.Spec.NodeConfig; nodeConfig != nil {
		for key, value := range nodeConfig.ResourceLabels {
			locals.GcpLabels[key] = value
		}
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.AttributionName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpGkeNodePool.String())

	if locals.GcpGkeNodePool.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpGkeNodePool.Metadata.Org
	}
	if locals.GcpGkeNodePool.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpGkeNodePool.Metadata.Env
	}
	if locals.GcpGkeNodePool.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpGkeNodePool.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
