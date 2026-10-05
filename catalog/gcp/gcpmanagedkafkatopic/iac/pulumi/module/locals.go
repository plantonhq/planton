package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpmanagedkafkatopicv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmanagedkafkatopic/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpManagedKafkaTopic *gcpmanagedkafkatopicv1alpha1.GcpManagedKafkaTopic

	// ClusterId is the bare cluster id Google's resource is keyed by. The
	// spec's cluster arrives either as the cluster's full resource path (a
	// GcpManagedKafkaCluster reference resolves to its name output) or as
	// the bare id; the last path segment is the id in both cases --
	// identical to the Terraform module's locals.cluster.
	ClusterId string

	// TopicId is spec.topic_id when set, otherwise metadata.name --
	// identical to the Terraform module's locals.topic_id.
	TopicId string
}

// initializeLocals derives the bare cluster id and the defaulted topic
// name. Topics carry no labels, so there is no attribution label set.
func initializeLocals(_ *pulumi.Context, iacInput *gcpmanagedkafkatopicv1alpha1.GcpManagedKafkaTopicIacInput) *Locals {
	locals := &Locals{}
	locals.GcpManagedKafkaTopic = iacInput.Target
	spec := locals.GcpManagedKafkaTopic.Spec

	cluster := spec.Cluster.GetValue()
	locals.ClusterId = cluster[strings.LastIndex(cluster, "/")+1:]

	locals.TopicId = spec.TopicId
	if locals.TopicId == "" {
		locals.TopicId = locals.GcpManagedKafkaTopic.Metadata.Name
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
