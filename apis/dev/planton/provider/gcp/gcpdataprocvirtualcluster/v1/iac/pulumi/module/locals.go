package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpdataprocvirtualclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpdataprocvirtualcluster/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values used across the module.
type Locals struct {
	GcpProviderConfig         *gcpprovider.GcpProviderConfig
	GcpDataprocVirtualCluster *gcpdataprocvirtualclusterv1.GcpDataprocVirtualCluster
	GcpLabels                 map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpdataprocvirtualclusterv1.GcpDataprocVirtualClusterStackInput) *Locals {
	locals := &Locals{}
	locals.GcpDataprocVirtualCluster = stackInput.Target

	// Determine resource name for labels.
	resourceName := locals.GcpDataprocVirtualCluster.Spec.ClusterName
	if resourceName == "" && locals.GcpDataprocVirtualCluster.Metadata != nil {
		resourceName = locals.GcpDataprocVirtualCluster.Metadata.Name
	}

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: resourceName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpDataprocVirtualCluster.String()),
	}

	if locals.GcpDataprocVirtualCluster.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpDataprocVirtualCluster.Metadata.Org
	}
	if locals.GcpDataprocVirtualCluster.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpDataprocVirtualCluster.Metadata.Env
	}
	if locals.GcpDataprocVirtualCluster.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpDataprocVirtualCluster.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
