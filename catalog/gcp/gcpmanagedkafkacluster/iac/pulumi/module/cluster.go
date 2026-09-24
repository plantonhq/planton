package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/managedkafka"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// cluster creates the Managed Service for Apache Kafka cluster. Internet
// (public) access is not offered: pulumi-gcp v9.37.0 has no
// public_cluster_config, and a one-engine field would break parity.
func cluster(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpManagedKafkaCluster.Spec
	resourceName := locals.GcpManagedKafkaCluster.Metadata.Name

	// Enable the Managed Kafka API first so a fresh project works on the
	// first deploy. DisableOnDestroy stays false: tearing down one cluster
	// must never disable the API for every other cluster in the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("managedkafka.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpkafka-managedkafka.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable managedkafka.googleapis.com api")
	}

	networkConfigs := managedkafka.ClusterGcpConfigAccessConfigNetworkConfigArray{}
	for _, subnet := range locals.Subnets {
		networkConfigs = append(networkConfigs, &managedkafka.ClusterGcpConfigAccessConfigNetworkConfigArgs{
			Subnet: pulumi.String(subnet),
		})
	}
	gcpConfig := &managedkafka.ClusterGcpConfigArgs{
		AccessConfig: &managedkafka.ClusterGcpConfigAccessConfigArgs{
			NetworkConfigs: networkConfigs,
		},
	}
	if spec.KmsKey.GetValue() != "" {
		gcpConfig.KmsKey = pulumi.String(spec.KmsKey.GetValue())
	}

	// Counts are int64 in the spec; Google's API takes them as decimal
	// strings.
	args := &managedkafka.ClusterArgs{
		Location:  pulumi.String(spec.Location),
		ClusterId: pulumi.String(locals.ClusterId),
		CapacityConfig: &managedkafka.ClusterCapacityConfigArgs{
			VcpuCount:   pulumi.String(strconv.FormatInt(spec.CapacityConfig.VcpuCount, 10)),
			MemoryBytes: pulumi.String(strconv.FormatInt(spec.CapacityConfig.MemoryBytes, 10)),
		},
		GcpConfig: gcpConfig,
		Labels:    pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// The spec lifts the block's one leaf; the block is sent only when a
	// disk size is declared, leaving Google's default otherwise.
	if spec.BrokerDiskSizeGib > 0 {
		args.BrokerCapacityConfig = &managedkafka.ClusterBrokerCapacityConfigArgs{
			DiskSizeGib: pulumi.String(strconv.FormatInt(spec.BrokerDiskSizeGib, 10)),
		}
	}
	if spec.RebalanceMode != "" {
		args.RebalanceConfig = &managedkafka.ClusterRebalanceConfigArgs{
			Mode: pulumi.String(spec.RebalanceMode),
		}
	}

	// Declared (even empty) means sent: an empty tls_config block is how an
	// earlier mTLS configuration is cleared, and an undeclared one leaves
	// Google's current configuration alone -- the Terraform module's rule.
	if tls := spec.TlsConfig; tls != nil {
		tlsArgs := &managedkafka.ClusterTlsConfigArgs{}
		if tls.SslPrincipalMappingRules != "" {
			tlsArgs.SslPrincipalMappingRules = pulumi.String(tls.SslPrincipalMappingRules)
		}
		if len(tls.CaPools) > 0 {
			casConfigs := managedkafka.ClusterTlsConfigTrustConfigCasConfigArray{}
			for _, caPool := range tls.CaPools {
				casConfigs = append(casConfigs, &managedkafka.ClusterTlsConfigTrustConfigCasConfigArgs{
					CaPool: pulumi.String(caPool.GetValue()),
				})
			}
			tlsArgs.TrustConfig = &managedkafka.ClusterTlsConfigTrustConfigArgs{CasConfigs: casConfigs}
		}
		args.TlsConfig = tlsArgs
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := managedkafka.NewCluster(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create managed kafka cluster")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpClusterId, created.ClusterId)
	ctx.Export(OpLocation, created.Location)
	return nil
}
