package module

import (
	kubernetesclickhousev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesclickhouse/v1"
	altinityv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/altinityoperator/kubernetes/clickhouse/v1"
)

// buildCoordinationConfig generates the coordination (ZooKeeper/Keeper) configuration
// Handles both new 'coordination' field and deprecated 'zookeeper' field for backward compatibility
//
// Priority Order:
//  1. coordination field (new API)
//  2. zookeeper field (deprecated, backward compat)
//  3. default (auto-managed ClickHouse Keeper)
func buildCoordinationConfig(spec *kubernetesclickhousev1.KubernetesClickHouseSpec, keeperServiceName string) *altinityv1.ClickHouseInstallationSpecConfigurationZookeeperArgs {
	// Priority: coordination field takes precedence over deprecated zookeeper field
	if spec.Coordination != nil {
		return buildCoordinationFromNewField(spec.Coordination, keeperServiceName)
	}

	// Backward compatibility: fall back to deprecated zookeeper field
	if spec.Zookeeper != nil {
		return buildCoordinationFromDeprecatedZookeeperField(spec.Zookeeper)
	}

	// Default: auto-managed ClickHouse Keeper with single replica
	// Note: ClickHouse Keeper must be deployed separately via ClickHouseKeeperInstallation
	// This references the Keeper service that should exist
	return buildDefaultKeeperReference(keeperServiceName)
}

// buildCoordinationFromNewField handles the new coordination configuration
// Routes to appropriate builder based on coordination type
func buildCoordinationFromNewField(coordination *kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig, keeperServiceName string) *altinityv1.ClickHouseInstallationSpecConfigurationZookeeperArgs {
	coordinationType := coordination.Type

	// Default unspecified to keeper
	if coordinationType == kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig_unspecified {
		coordinationType = kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig_keeper
	}

	switch coordinationType {
	case kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig_keeper:
		return buildAutoManagedKeeperReference(coordination, keeperServiceName)

	case kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig_external_keeper:
		return buildExternalKeeperReference(coordination, keeperServiceName)

	case kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig_external_zookeeper:
		return buildExternalZookeeperReference(coordination, keeperServiceName)
	}

	// Fallback to default
	return buildDefaultKeeperReference(keeperServiceName)
}
