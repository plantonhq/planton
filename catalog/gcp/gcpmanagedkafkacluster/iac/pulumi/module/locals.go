package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpmanagedkafkaclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmanagedkafkacluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// computeApiPrefix is what a GcpSubnetwork self link starts with; Google's
// Kafka API wants the projects/{p}/regions/{r}/subnetworks/{s} form.
const computeApiPrefix = "https://www.googleapis.com/compute/v1/"

type Locals struct {
	GcpProviderConfig      *gcpprovider.GcpProviderConfig
	GcpManagedKafkaCluster *gcpmanagedkafkaclusterv1alpha1.GcpManagedKafkaCluster
	GcpLabels              map[string]string

	// ClusterId is spec.cluster_id when set, otherwise metadata.name --
	// identical to the Terraform module's locals.cluster_id.
	ClusterId string

	// Subnets are the network configs' subnets with the compute API prefix
	// trimmed -- identical to the Terraform module's locals.subnets.
	Subnets []string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpmanagedkafkaclusterv1alpha1.GcpManagedKafkaClusterStackInput) *Locals {
	locals := &Locals{}
	locals.GcpManagedKafkaCluster = stackInput.Target
	metadata := locals.GcpManagedKafkaCluster.Metadata
	spec := locals.GcpManagedKafkaCluster.Spec

	locals.ClusterId = spec.ClusterId
	if locals.ClusterId == "" {
		locals.ClusterId = metadata.Name
	}

	for _, networkConfig := range spec.NetworkConfigs {
		locals.Subnets = append(locals.Subnets, strings.TrimPrefix(networkConfig.Subnet.GetValue(), computeApiPrefix))
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpManagedKafkaCluster.String())

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
