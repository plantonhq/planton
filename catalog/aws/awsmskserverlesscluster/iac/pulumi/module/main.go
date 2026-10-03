package module

import (
	"github.com/pkg/errors"
	awsmskserverlessclusterv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmskserverlesscluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the AWS MSK Serverless Cluster and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsmskserverlessclusterv1alpha1.AwsMskServerlessClusterIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsMskServerlessCluster.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdCluster, err := cluster(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "msk serverless cluster")
	}

	ctx.Export(OpClusterArn, createdCluster.Arn)
	ctx.Export(OpClusterName, createdCluster.ClusterName)
	ctx.Export(OpClusterUuid, createdCluster.ClusterUuid)
	ctx.Export(OpBootstrapBrokersSaslIam, createdCluster.BootstrapBrokersSaslIam)

	return nil
}
