package module

import (
	kubernetesclickhousev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesclickhouse/v1"
	altinityv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/altinityoperator/kubernetes/clickhouse/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildAutoManagedKeeperReference creates configuration for auto-managed ClickHouse Keeper
// The module creates a ClickHouseKeeperInstallation CRD with a computed name
// which creates a service following pattern "keeper-<chk-name>"
func buildAutoManagedKeeperReference(coordination *kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig, keeperServiceName string) *altinityv1.ClickHouseInstallationSpecConfigurationZookeeperArgs {
	// Service name pattern: keeper-<chk-name>
	// The keeper service name is computed and passed in to avoid hardcoding
	return &altinityv1.ClickHouseInstallationSpecConfigurationZookeeperArgs{
		Nodes: altinityv1.ClickHouseInstallationSpecConfigurationZookeeperNodesArray{
			&altinityv1.ClickHouseInstallationSpecConfigurationZookeeperNodesArgs{
				Host: pulumi.String(keeperServiceName),
				Port: pulumi.IntPtr(vars.ZookeeperPort),
			},
		},
	}
}

// buildExternalKeeperReference creates configuration for external ClickHouse Keeper
// Connects to existing ClickHouse Keeper infrastructure
func buildExternalKeeperReference(coordination *kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig, keeperServiceName string) *altinityv1.ClickHouseInstallationSpecConfigurationZookeeperArgs {
	if coordination.ExternalConfig == nil || len(coordination.ExternalConfig.Nodes) == 0 {
		// No external nodes specified, fallback to default
		return buildDefaultKeeperReference(keeperServiceName)
	}

	// Build node references from external config
	nodes := make(altinityv1.ClickHouseInstallationSpecConfigurationZookeeperNodesArray, len(coordination.ExternalConfig.Nodes))
	for i, node := range coordination.ExternalConfig.Nodes {
		nodes[i] = &altinityv1.ClickHouseInstallationSpecConfigurationZookeeperNodesArgs{
			Host: pulumi.String(node),
		}
	}

	return &altinityv1.ClickHouseInstallationSpecConfigurationZookeeperArgs{
		Nodes: nodes,
	}
}

// buildDefaultKeeperReference creates a reference to default auto-managed ClickHouse Keeper
// Used when no coordination configuration is specified
//
// The module auto-creates ClickHouseKeeperInstallation with a computed name
// which creates service following the pattern "keeper-<chk-name>"
func buildDefaultKeeperReference(keeperServiceName string) *altinityv1.ClickHouseInstallationSpecConfigurationZookeeperArgs {
	return &altinityv1.ClickHouseInstallationSpecConfigurationZookeeperArgs{
		Nodes: altinityv1.ClickHouseInstallationSpecConfigurationZookeeperNodesArray{
			&altinityv1.ClickHouseInstallationSpecConfigurationZookeeperNodesArgs{
				Host: pulumi.String(keeperServiceName),
				Port: pulumi.IntPtr(vars.ZookeeperPort),
			},
		},
	}
}
