package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpredisclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcprediscluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpRedisCluster   *gcpredisclusterv1alpha1.GcpRedisCluster
	GcpLabels         map[string]string

	// ClusterName is the cluster's GCP name: spec.cluster_name when set,
	// otherwise metadata.name -- the same fallback the Terraform module
	// applies in locals.tf.
	ClusterName string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpredisclusterv1alpha1.GcpRedisClusterStackInput) *Locals {
	locals := &Locals{}
	locals.GcpRedisCluster = stackInput.Target

	locals.ClusterName = locals.GcpRedisCluster.Spec.ClusterName
	if locals.ClusterName == "" {
		locals.ClusterName = locals.GcpRedisCluster.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpRedisCluster.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.ClusterName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpRedisCluster.String())

	if locals.GcpRedisCluster.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpRedisCluster.Metadata.Org
	}
	if locals.GcpRedisCluster.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpRedisCluster.Metadata.Env
	}
	if locals.GcpRedisCluster.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpRedisCluster.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
