package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpdataprocclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdataproccluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpDataprocCluster *gcpdataprocclusterv1alpha1.GcpDataprocCluster
	GcpLabels          map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpdataprocclusterv1alpha1.GcpDataprocClusterIacInput) *Locals {
	locals := &Locals{}
	locals.GcpDataprocCluster = iacInput.Target

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpDataprocCluster.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpDataprocCluster.Spec.ClusterName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpDataprocCluster.String())

	if locals.GcpDataprocCluster.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpDataprocCluster.Metadata.Org
	}
	if locals.GcpDataprocCluster.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpDataprocCluster.Metadata.Env
	}
	if locals.GcpDataprocCluster.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpDataprocCluster.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
