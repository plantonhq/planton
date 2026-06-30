package module

import (
	"fmt"

	"github.com/pkg/errors"
	awsneptuneclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsneptunecluster/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS Neptune cluster related resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsneptuneclusterv1.AwsNeptuneClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsNeptuneCluster.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdSg, err := securityGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "security group")
	}

	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "subnet group")
	}

	createdParamGroup, err := clusterParameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "cluster parameter group")
	}

	cluster, err := neptuneCluster(ctx, locals, provider, createdSg, createdSubnetGroup, createdParamGroup)
	if err != nil {
		return errors.Wrap(err, "neptune cluster")
	}

	err = neptuneClusterInstances(ctx, locals, provider, cluster)
	if err != nil {
		return errors.Wrap(err, "neptune cluster instances")
	}

	ctx.Export(OpClusterEndpoint, cluster.Endpoint)
	ctx.Export(OpClusterReaderEndpoint, cluster.ReaderEndpoint)
	ctx.Export(OpClusterId, cluster.ID())
	ctx.Export(OpClusterArn, cluster.Arn)
	ctx.Export(OpClusterResourceId, cluster.ClusterResourceId)
	ctx.Export(OpClusterPort, cluster.Port)
	ctx.Export(OpHostedZoneId, cluster.HostedZoneId)

	if createdSubnetGroup != nil {
		ctx.Export(OpDbSubnetGroupName, createdSubnetGroup.Name)
	}
	if createdSg != nil {
		ctx.Export(OpSecurityGroupId, createdSg.ID())
	}
	if createdParamGroup != nil {
		ctx.Export(OpClusterParameterGroupName, createdParamGroup.Name)
	}

	return nil
}

// getEffectivePort returns the port to use, either from spec or default (8182).
func getEffectivePort(spec *awsneptuneclusterv1.AwsNeptuneClusterSpec) int {
	port := spec.GetPort()
	if port == 0 {
		return 8182
	}
	return int(port)
}

// getParameterGroupFamily derives the Neptune parameter group family from the engine version.
// Neptune families: "neptune1", "neptune1.2", "neptune1.3"
func getParameterGroupFamily(engineVersion string) string {
	if engineVersion == "" {
		return "neptune1.3"
	}
	if len(engineVersion) >= 3 {
		major := engineVersion[:3]
		return fmt.Sprintf("neptune%s", major)
	}
	return fmt.Sprintf("neptune%s", engineVersion)
}
