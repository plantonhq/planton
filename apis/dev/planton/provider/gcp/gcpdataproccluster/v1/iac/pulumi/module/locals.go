package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpdataprocclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpdataproccluster/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpDataprocCluster *gcpdataprocclusterv1.GcpDataprocCluster
	GcpLabels          map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpdataprocclusterv1.GcpDataprocClusterStackInput) *Locals {
	locals := &Locals{}
	locals.GcpDataprocCluster = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpDataprocCluster.Spec.ClusterName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpDataprocCluster.String()),
	}

	if locals.GcpDataprocCluster.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpDataprocCluster.Metadata.Org
	}
	if locals.GcpDataprocCluster.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpDataprocCluster.Metadata.Env
	}
	if locals.GcpDataprocCluster.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpDataprocCluster.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
