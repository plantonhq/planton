package module

import (
	"fmt"

	"github.com/pkg/errors"
	confluentkafkav1 "github.com/plantonhq/planton/apis/dev/planton/provider/confluent/confluentkafka/v1"
	"github.com/pulumi/pulumi-confluentcloud/sdk/v2/go/confluentcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates a Confluent Cloud Kafka cluster with all configured parameters
func Resources(ctx *pulumi.Context, stackInput *confluentkafkav1.ConfluentKafkaStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Setup Confluent Cloud provider with credentials from provider config
	provider, err := createProvider(ctx, stackInput)
	if err != nil {
		return errors.Wrap(err, "failed to create Confluent Cloud provider")
	}

	// Create the Kafka cluster
	createdCluster, err := createKafkaCluster(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Kafka cluster")
	}

	// Export stack outputs
	return exportOutputs(ctx, createdCluster, locals)
}

// createProvider initializes the Confluent Cloud provider with credentials
func createProvider(ctx *pulumi.Context, stackInput *confluentkafkav1.ConfluentKafkaStackInput) (*confluentcloud.Provider, error) {
	providerConfig := stackInput.ProviderConfig

	if providerConfig == nil {
		// Use default provider (assumes credentials from environment variables)
		return confluentcloud.NewProvider(ctx, "confluentcloud-provider", &confluentcloud.ProviderArgs{})
	}

	// Create provider with explicit credentials
	return confluentcloud.NewProvider(ctx, "confluentcloud-provider", &confluentcloud.ProviderArgs{
		CloudApiKey:    pulumi.String(providerConfig.ApiKey),
		CloudApiSecret: pulumi.String(providerConfig.ApiSecret),
	})
}

// createKafkaCluster creates the Confluent Cloud Kafka cluster resource
func createKafkaCluster(ctx *pulumi.Context, locals *Locals, provider *confluentcloud.Provider) (*confluentcloud.KafkaCluster, error) {
	spec := locals.ConfluentKafka.Spec

	// Determine display name (use custom display_name or fallback to metadata.name)
	displayName := spec.DisplayName
	if displayName == "" {
		displayName = locals.ConfluentKafka.Metadata.Name
	}

	// Build cluster configuration based on cluster type
	clusterArgs := &confluentcloud.KafkaClusterArgs{
		DisplayName:  pulumi.String(displayName),
		Availability: pulumi.String(spec.Availability),
		Cloud:        pulumi.String(spec.Cloud),
		Region:       pulumi.String(spec.Region),
		Environment: &confluentcloud.KafkaClusterEnvironmentArgs{
			Id: pulumi.String(spec.EnvironmentId),
		},
	}

	// Configure cluster type-specific settings
	clusterType := spec.ClusterType
	if clusterType == "" {
		clusterType = "STANDARD" // Default to STANDARD if not specified
	}

	// If network configuration is provided, associate the cluster with the network
	// This is supported for ENTERPRISE and DEDICATED cluster types
	if spec.NetworkConfig != nil {
		clusterArgs.Network = &confluentcloud.KafkaClusterNetworkArgs{
			Id: pulumi.String(spec.NetworkConfig.NetworkId),
		}
	}

	switch clusterType {
	case "BASIC":
		clusterArgs.Basic = &confluentcloud.KafkaClusterBasicArgs{}

	case "STANDARD":
		clusterArgs.Standard = &confluentcloud.KafkaClusterStandardArgs{}

	case "ENTERPRISE":
		// ENTERPRISE clusters use the STANDARD configuration with private networking
		// The network_config enables private connectivity via PrivateLink/Private Link/Private Service Connect
		clusterArgs.Standard = &confluentcloud.KafkaClusterStandardArgs{}

	case "DEDICATED":
		// Dedicated clusters require CKU configuration
		if spec.DedicatedConfig == nil {
			return nil, fmt.Errorf("dedicated_config is required when cluster_type is DEDICATED")
		}

		clusterArgs.Dedicated = &confluentcloud.KafkaClusterDedicatedArgs{
			Cku: pulumi.Int(int(spec.DedicatedConfig.Cku)),
		}

	default:
		return nil, fmt.Errorf("invalid cluster_type: %s (must be BASIC, STANDARD, ENTERPRISE, or DEDICATED)", clusterType)
	}

	// Create the Kafka cluster
	return confluentcloud.NewKafkaCluster(ctx,
		locals.ConfluentKafka.Metadata.Name,
		clusterArgs,
		pulumi.Provider(provider),
	)
}
