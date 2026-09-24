package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpmanagedkafkaconnectorv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmanagedkafkaconnector/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig        *gcpprovider.GcpProviderConfig
	GcpManagedKafkaConnector *gcpmanagedkafkaconnectorv1alpha1.GcpManagedKafkaConnector

	// ConnectClusterId is the bare Connect cluster id Google's resource is
	// keyed by. The spec's connect_cluster arrives either as the Connect
	// cluster's full resource path (a GcpManagedKafkaConnectCluster
	// reference resolves to its name output) or as the bare id; the last
	// path segment is the id in both cases -- identical to the Terraform
	// module's locals.connect_cluster.
	ConnectClusterId string

	// ConnectorId is spec.connector_id when set, otherwise metadata.name --
	// identical to the Terraform module's locals.connector_id.
	ConnectorId string
}

// initializeLocals derives the bare Connect cluster id and the defaulted
// connector id. Connectors carry no labels, so there is no attribution
// label set.
func initializeLocals(_ *pulumi.Context, stackInput *gcpmanagedkafkaconnectorv1alpha1.GcpManagedKafkaConnectorStackInput) *Locals {
	locals := &Locals{}
	locals.GcpManagedKafkaConnector = stackInput.Target
	spec := locals.GcpManagedKafkaConnector.Spec

	connectCluster := spec.ConnectCluster.GetValue()
	locals.ConnectClusterId = connectCluster[strings.LastIndex(connectCluster, "/")+1:]

	locals.ConnectorId = spec.ConnectorId
	if locals.ConnectorId == "" {
		locals.ConnectorId = locals.GcpManagedKafkaConnector.Metadata.Name
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
