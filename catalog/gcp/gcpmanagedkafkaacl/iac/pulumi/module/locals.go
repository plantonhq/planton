package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpmanagedkafkaaclv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmanagedkafkaacl/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpManagedKafkaAcl *gcpmanagedkafkaaclv1alpha1.GcpManagedKafkaAcl

	// ClusterId is the bare cluster id Google's resource is keyed by. The
	// spec's cluster arrives either as the cluster's full resource path (a
	// GcpManagedKafkaCluster reference resolves to its name output) or as
	// the bare id; the last path segment is the id in both cases --
	// identical to the Terraform module's locals.cluster.
	ClusterId string
}

// initializeLocals derives the bare cluster id. ACLs carry no labels, so
// there is no attribution label set.
func initializeLocals(_ *pulumi.Context, stackInput *gcpmanagedkafkaaclv1alpha1.GcpManagedKafkaAclStackInput) *Locals {
	locals := &Locals{}
	locals.GcpManagedKafkaAcl = stackInput.Target

	cluster := locals.GcpManagedKafkaAcl.Spec.Cluster.GetValue()
	locals.ClusterId = cluster[strings.LastIndex(cluster, "/")+1:]

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
