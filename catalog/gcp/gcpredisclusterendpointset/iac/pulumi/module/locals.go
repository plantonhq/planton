package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpredisclusterendpointsetv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpredisclusterendpointset/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig          *gcpprovider.GcpProviderConfig
	GcpRedisClusterEndpointSet *gcpredisclusterendpointsetv1alpha1.GcpRedisClusterEndpointSet

	// ClusterName is the bare cluster name Google's resource is keyed by.
	// The spec's cluster arrives either as the cluster's full resource path
	// (a GcpRedisCluster reference resolves to its name output,
	// projects/{p}/locations/{r}/clusters/{name}) or as the bare name; the
	// last path segment is the name in both cases -- identical to the
	// Terraform module's locals.cluster_name.
	ClusterName string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpredisclusterendpointsetv1alpha1.GcpRedisClusterEndpointSetStackInput) *Locals {
	locals := &Locals{}
	locals.GcpRedisClusterEndpointSet = stackInput.Target

	cluster := locals.GcpRedisClusterEndpointSet.Spec.Cluster.GetValue()
	locals.ClusterName = cluster[strings.LastIndex(cluster, "/")+1:]

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
