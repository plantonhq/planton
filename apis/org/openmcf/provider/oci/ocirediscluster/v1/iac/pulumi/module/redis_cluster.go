package module

import (
	"strings"

	"github.com/pkg/errors"
	ociredisclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocirediscluster/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/redis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func redisCluster(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciRedisCluster.Spec

	args := &redis.RedisClusterArgs{
		CompartmentId:   pulumi.String(spec.CompartmentId.GetValue()),
		DisplayName:     pulumi.String(locals.DisplayName),
		SubnetId:        pulumi.String(spec.SubnetId.GetValue()),
		NodeCount:       pulumi.Int(int(spec.NodeCount)),
		NodeMemoryInGbs: pulumi.Float64(float64(spec.NodeMemoryInGbs)),
		SoftwareVersion: pulumi.String(spec.SoftwareVersion),
		FreeformTags:    pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.ClusterMode != ociredisclusterv1.OciRedisClusterSpec_cluster_mode_unspecified {
		args.ClusterMode = pulumi.StringPtr(strings.ToUpper(spec.ClusterMode.String()))
	}

	if spec.ShardCount > 0 {
		args.ShardCount = pulumi.IntPtr(int(spec.ShardCount))
	}

	if len(spec.NsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(spec.NsgIds))
		for i, nsg := range spec.NsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		args.NsgIds = nsgIds
	}

	if spec.ConfigSetId != nil {
		args.OciCacheConfigSetId = pulumi.StringPtr(spec.ConfigSetId.GetValue())
	}

	cluster, err := redis.NewRedisCluster(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci redis cluster")
	}

	ctx.Export(OpClusterId, cluster.ID())
	ctx.Export(OpPrimaryFqdn, cluster.PrimaryFqdn)
	ctx.Export(OpPrimaryEndpointIpAddress, cluster.PrimaryEndpointIpAddress)
	ctx.Export(OpReplicasFqdn, cluster.ReplicasFqdn)
	ctx.Export(OpDiscoveryFqdn, cluster.DiscoveryFqdn)

	return nil
}
