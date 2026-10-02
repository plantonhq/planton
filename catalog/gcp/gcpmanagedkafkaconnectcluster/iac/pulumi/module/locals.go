package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpmanagedkafkaconnectclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmanagedkafkaconnectcluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// computeApiPrefix is what a GcpSubnetwork self link starts with; Google's
// Kafka API wants the projects/{p}/regions/{r}/subnetworks/{s} form.
const computeApiPrefix = "https://www.googleapis.com/compute/v1/"

type Locals struct {
	GcpProviderConfig             *gcpprovider.GcpProviderConfig
	GcpManagedKafkaConnectCluster *gcpmanagedkafkaconnectclusterv1alpha1.GcpManagedKafkaConnectCluster
	GcpLabels                     map[string]string

	// ConnectClusterId is spec.connect_cluster_id when set, otherwise
	// metadata.name -- identical to the Terraform module's
	// locals.connect_cluster_id.
	ConnectClusterId string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpmanagedkafkaconnectclusterv1alpha1.GcpManagedKafkaConnectClusterStackInput) *Locals {
	locals := &Locals{}
	locals.GcpManagedKafkaConnectCluster = stackInput.Target
	metadata := locals.GcpManagedKafkaConnectCluster.Metadata
	spec := locals.GcpManagedKafkaConnectCluster.Spec

	locals.ConnectClusterId = spec.ConnectClusterId
	if locals.ConnectClusterId == "" {
		locals.ConnectClusterId = metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpManagedKafkaConnectCluster.String())

	if metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = metadata.Env
	}
	if metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
