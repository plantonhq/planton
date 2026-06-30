package module

import (
	confluentkafkav1 "github.com/plantonhq/planton/apis/dev/planton/provider/confluent/confluentkafka/v1"
	"github.com/pulumi/pulumi-confluentcloud/sdk/v2/go/confluentcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the Kafka cluster outputs to Pulumi stack outputs
func exportOutputs(ctx *pulumi.Context, cluster *confluentcloud.KafkaCluster, locals *Locals) error {
	// Export the cluster ID
	ctx.Export("id", cluster.ID())

	// Export the bootstrap endpoint
	ctx.Export("bootstrap_endpoint", cluster.BootstrapEndpoint)

	// Export the Confluent Resource Name (CRN)
	ctx.Export("crn", cluster.RbacCrn)

	// Export the REST endpoint
	ctx.Export("rest_endpoint", cluster.RestEndpoint)

	// Create and export the outputs structure
	outputs := &confluentkafkav1.ConfluentKafkaStackOutputs{}

	cluster.ID().ApplyT(func(id string) error {
		outputs.Id = id
		return nil
	})

	cluster.BootstrapEndpoint.ApplyT(func(endpoint string) error {
		outputs.BootstrapEndpoint = endpoint
		return nil
	})

	cluster.RbacCrn.ApplyT(func(crn string) error {
		outputs.Crn = crn
		return nil
	})

	cluster.RestEndpoint.ApplyT(func(endpoint string) error {
		outputs.RestEndpoint = endpoint
		return nil
	})

	return nil
}

const (
	OpOutputKey = "key"
)
