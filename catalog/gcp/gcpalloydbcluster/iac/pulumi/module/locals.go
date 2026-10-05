package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpalloydbclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpalloydbcluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpAlloydbCluster *gcpalloydbclusterv1alpha1.GcpAlloydbCluster
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpalloydbclusterv1alpha1.GcpAlloydbClusterIacInput) *Locals {
	locals := &Locals{}
	locals.GcpAlloydbCluster = iacInput.Target
	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module. Applied
	// to the cluster AND its bundled primary instance.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpAlloydbCluster.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpAlloydbCluster.Spec.ClusterName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpAlloydbCluster.String())

	if locals.GcpAlloydbCluster.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpAlloydbCluster.Metadata.Org
	}
	if locals.GcpAlloydbCluster.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpAlloydbCluster.Metadata.Env
	}
	if locals.GcpAlloydbCluster.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpAlloydbCluster.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
