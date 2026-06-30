package module

import (
	kubernetesclickhousev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesclickhouse/v1"
)

// shouldCreateClickHouseKeeper determines if we should auto-create ClickHouseKeeperInstallation
// Returns true only when:
// - Clustering is enabled (requires coordination)
// - coordination.type = keeper (auto-managed)
func shouldCreateClickHouseKeeper(spec *kubernetesclickhousev1.KubernetesClickHouseSpec) bool {
	// Only create Keeper for clustered deployments
	if spec.Cluster == nil || !spec.Cluster.IsEnabled {
		return false
	}

	// Check if using new coordination field
	if spec.Coordination != nil {
		coordinationType := spec.Coordination.Type
		// Default unspecified to keeper
		if coordinationType == kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig_unspecified {
			coordinationType = kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig_keeper
		}
		// Only create for keeper type (not external_keeper or external_zookeeper)
		return coordinationType == kubernetesclickhousev1.KubernetesClickHouseCoordinationConfig_keeper
	}

	// If using deprecated zookeeper field, don't auto-create Keeper
	if spec.Zookeeper != nil {
		return false
	}

	// Default behavior: clustered without explicit coordination = create Keeper
	return true
}

// getKeeperConfig extracts the keeper configuration from spec
// Returns nil if not specified (will use defaults)
func getKeeperConfig(spec *kubernetesclickhousev1.KubernetesClickHouseSpec) *kubernetesclickhousev1.KubernetesClickHouseKeeperConfig {
	if spec.Coordination != nil && spec.Coordination.KeeperConfig != nil {
		return spec.Coordination.KeeperConfig
	}
	return nil
}
