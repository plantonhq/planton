package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scaleway "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/redis"
)

// redisCluster provisions the Scaleway Redis cluster and exports the
// stack outputs (cluster ID, endpoints, certificate).
//
// The cluster is created with:
//   - The initial user (user_name + password from spec).
//   - Optional ACL rules (when NOT using Private Network).
//   - Optional Private Network attachment (when NOT using ACL).
//   - Optional TLS encryption.
//   - Optional Redis settings.
//   - Standard OpenMCF tags for resource identification.
func redisCluster(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) error {
	spec := locals.ScalewayRedisCluster.Spec

	// Build the cluster arguments.
	clusterArgs := &redis.ClusterArgs{
		Name:     pulumi.StringPtr(locals.ScalewayRedisCluster.Metadata.Name),
		Version:  pulumi.String(spec.Version),
		NodeType: pulumi.String(spec.NodeType),
		Zone:     pulumi.StringPtr(spec.Zone),
		Tags:     toPulumiStringArray(locals.ScalewayTags),
		UserName: pulumi.String(spec.UserName),
		Password: pulumi.StringPtr(spec.Password),
	}

	// Cluster size (determines standalone/HA/cluster mode).
	if spec.ClusterSize > 0 {
		clusterArgs.ClusterSize = pulumi.IntPtr(int(spec.ClusterSize))
	}

	// TLS encryption.
	if spec.TlsEnabled {
		clusterArgs.TlsEnabled = pulumi.BoolPtr(true)
	}

	// ACL rules (mutually exclusive with Private Network).
	if len(spec.AclRules) > 0 {
		aclRules := make(redis.ClusterAclArray, 0, len(spec.AclRules))
		for _, rule := range spec.AclRules {
			aclRules = append(aclRules, &redis.ClusterAclArgs{
				Ip:          pulumi.String(rule.Ip),
				Description: pulumi.StringPtr(rule.Description),
			})
		}
		clusterArgs.Acls = aclRules
	}

	// Private Network (mutually exclusive with ACL rules).
	if locals.PrivateNetworkId != "" {
		clusterArgs.PrivateNetworks = redis.ClusterPrivateNetworkArray{
			&redis.ClusterPrivateNetworkArgs{
				Id: pulumi.String(locals.PrivateNetworkId),
			},
		}
	}

	// Redis settings.
	if len(spec.Settings) > 0 {
		settingsMap := pulumi.StringMap{}
		for k, v := range spec.Settings {
			settingsMap[k] = pulumi.String(v)
		}
		clusterArgs.Settings = settingsMap
	}

	// Create the cluster.
	createdCluster, err := redis.NewCluster(
		ctx,
		"cluster",
		clusterArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create redis cluster")
	}

	// Export core output: cluster ID.
	ctx.Export(OpClusterId, createdCluster.ID())

	// Export public network endpoint (populated when NOT using PN).
	ctx.Export(OpPublicNetworkPort, createdCluster.PublicNetwork.ApplyT(func(pn redis.ClusterPublicNetwork) int {
		if pn.Port != nil {
			return *pn.Port
		}
		return 0
	}).(pulumi.IntOutput))

	ctx.Export(OpPublicNetworkIps, createdCluster.PublicNetwork.ApplyT(func(pn redis.ClusterPublicNetwork) []string {
		return pn.Ips
	}).(pulumi.StringArrayOutput))

	// Export private network endpoint (populated when using PN).
	ctx.Export(OpPrivateNetworkPort, createdCluster.PrivateNetworks.ApplyT(func(pns []redis.ClusterPrivateNetwork) int {
		if len(pns) > 0 && pns[0].Port != nil {
			return *pns[0].Port
		}
		return 0
	}).(pulumi.IntOutput))

	ctx.Export(OpPrivateNetworkIps, createdCluster.PrivateNetworks.ApplyT(func(pns []redis.ClusterPrivateNetwork) []string {
		if len(pns) > 0 {
			return pns[0].Ips
		}
		return nil
	}).(pulumi.StringArrayOutput))

	// Export TLS certificate (empty when TLS disabled).
	ctx.Export(OpCertificate, createdCluster.Certificate)

	return nil
}

// toPulumiStringArray converts a Go string slice to a Pulumi StringArray.
func toPulumiStringArray(tags []string) pulumi.StringArray {
	result := make(pulumi.StringArray, len(tags))
	for i, tag := range tags {
		result[i] = pulumi.String(tag)
	}
	return result
}
