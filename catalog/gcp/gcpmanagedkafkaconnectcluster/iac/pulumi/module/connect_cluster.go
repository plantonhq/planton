package module

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/managedkafka"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// connectCluster creates the Kafka Connect cluster attached to its Kafka
// cluster.
func connectCluster(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpManagedKafkaConnectCluster.Spec
	resourceName := locals.GcpManagedKafkaConnectCluster.Metadata.Name

	// Enable the Managed Kafka API first so a fresh project works on the
	// first deploy. DisableOnDestroy stays false: tearing down one Connect
	// cluster must never disable the API for the Kafka clusters in the
	// project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("managedkafka.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpkcon-managedkafka.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable managedkafka.googleapis.com api")
	}

	// A GcpSubnetwork reference resolves to the subnet's self link; Google
	// wants the projects/{p}/regions/{r}/subnetworks/{s} form -- the
	// Terraform module's trimprefix.
	networkConfigs := managedkafka.ConnectClusterGcpConfigAccessConfigNetworkConfigArray{}
	for _, networkConfig := range spec.NetworkConfigs {
		networkArgs := &managedkafka.ConnectClusterGcpConfigAccessConfigNetworkConfigArgs{
			PrimarySubnet: pulumi.String(strings.TrimPrefix(networkConfig.PrimarySubnet.GetValue(), computeApiPrefix)),
		}
		if len(networkConfig.DnsDomainNames) > 0 {
			networkArgs.DnsDomainNames = pulumi.ToStringArray(networkConfig.DnsDomainNames)
		}
		networkConfigs = append(networkConfigs, networkArgs)
	}

	// Counts are int64 in the spec; Google's API takes them as decimal
	// strings.
	args := &managedkafka.ConnectClusterArgs{
		Location:         pulumi.String(spec.Location),
		ConnectClusterId: pulumi.String(locals.ConnectClusterId),
		KafkaCluster:     pulumi.String(spec.KafkaCluster.GetValue()),
		CapacityConfig: &managedkafka.ConnectClusterCapacityConfigArgs{
			VcpuCount:   pulumi.String(strconv.FormatInt(spec.CapacityConfig.VcpuCount, 10)),
			MemoryBytes: pulumi.String(strconv.FormatInt(spec.CapacityConfig.MemoryBytes, 10)),
		},
		GcpConfig: &managedkafka.ConnectClusterGcpConfigArgs{
			AccessConfig: &managedkafka.ConnectClusterGcpConfigAccessConfigArgs{
				NetworkConfigs: networkConfigs,
			},
		},
		Labels: pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := managedkafka.NewConnectCluster(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create managed kafka connect cluster")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpConnectClusterId, created.ConnectClusterId)
	ctx.Export(OpLocation, created.Location)
	return nil
}
