package module

import (
	"github.com/pkg/errors"
	atlasmongodbv1 "github.com/plantonhq/planton/apis/dev/planton/provider/atlas/atlasmongodb/v1"
	"github.com/pulumi/pulumi-mongodbatlas/sdk/v3/go/mongodbatlas"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates a Atlas MongoDB cluster with all configured parameters
func Resources(ctx *pulumi.Context, stackInput *atlasmongodbv1.AtlasMongodbStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Setup Atlas MongoDB provider with credentials from provider config
	var provider *mongodbatlas.Provider
	var err error
	providerConfig := stackInput.ProviderConfig

	if providerConfig == nil {
		// Use default provider (assumes credentials from environment)
		provider, err = mongodbatlas.NewProvider(ctx, "atlasmongodb-provider", &mongodbatlas.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default Atlas MongoDB provider")
		}
	} else {
		// Create provider with explicit credentials
		provider, err = mongodbatlas.NewProvider(ctx, "atlasmongodb-provider", &mongodbatlas.ProviderArgs{
			PublicKey:  pulumi.String(providerConfig.PublicKey),
			PrivateKey: pulumi.String(providerConfig.PrivateKey),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create Atlas MongoDB provider with credentials")
		}
	}

	// Create the Atlas MongoDB cluster
	createdCluster, err := createCluster(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Atlas MongoDB cluster")
	}

	// Export stack outputs
	return exportOutputs(ctx, createdCluster, locals)
}
