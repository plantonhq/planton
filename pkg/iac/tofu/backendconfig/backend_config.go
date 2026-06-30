package backendconfig

import (
	"fmt"

	"github.com/plantonhq/planton/pkg/iac/tofu/tofulabels"
	"github.com/plantonhq/planton/pkg/reflection/metadatareflect"
	"google.golang.org/protobuf/proto"
)

// TofuBackendConfig represents the Terraform/Tofu backend configuration
type TofuBackendConfig struct {
	// BackendType specifies the backend type (e.g., "s3", "gcs", "azurerm")
	BackendType string
	// BackendBucket specifies the bucket or container name for remote backends
	BackendBucket string
	// BackendKey specifies the state file path within the bucket
	BackendKey string
	// BackendRegion specifies the region for S3 backends
	BackendRegion string
	// BackendEndpoint specifies a custom S3-compatible endpoint (for R2, MinIO, etc.)
	BackendEndpoint string
	// S3Compatible indicates this is an S3-compatible backend requiring skip flags
	S3Compatible bool
}

// IsS3Compatible returns true if this is an S3-compatible backend (R2, MinIO, etc.)
// Detection signals: explicit endpoint is set OR region is "auto"
func (c *TofuBackendConfig) IsS3Compatible() bool {
	return c.BackendEndpoint != "" || c.BackendRegion == "auto"
}

// ExtractFromManifest extracts Terraform/Tofu backend configuration from manifest labels.
// The provisionerType should be "terraform" or "tofu" to determine which label prefix to use.
// It first checks for provisioner-specific labels (e.g., tofu.planton.dev/backend.type),
// then falls back to legacy terraform.* labels for backward compatibility.
func ExtractFromManifest(manifest proto.Message, provisionerType string) (*TofuBackendConfig, error) {
	labels := metadatareflect.ExtractLabels(manifest)
	if labels == nil {
		return nil, fmt.Errorf("no labels found in manifest")
	}

	// Try provisioner-specific labels first
	backendType, hasType := labels[tofulabels.BackendTypeLabelKey(provisionerType)]
	backendBucket, hasBucket := labels[tofulabels.BackendBucketLabelKey(provisionerType)]
	backendKey, hasKey := labels[tofulabels.BackendKeyLabelKey(provisionerType)]
	backendRegion, _ := labels[tofulabels.BackendRegionLabelKey(provisionerType)]
	backendEndpoint, _ := labels[tofulabels.BackendEndpointLabelKey(provisionerType)]

	// If provisioner-specific labels not found, fall back to legacy terraform.* labels
	// This ensures backward compatibility for existing manifests
	if !hasType && !hasBucket && !hasKey {
		backendType, hasType = labels[tofulabels.LegacyBackendTypeLabelKey]
		backendBucket, hasBucket = labels[tofulabels.LegacyBackendBucketLabelKey]
		// Try backend.key first, then fall back to deprecated backend.object
		backendKey, hasKey = labels[tofulabels.LegacyBackendKeyLabelKey]
		if !hasKey {
			backendKey, hasKey = labels[tofulabels.LegacyBackendObjectLabelKey]
		}
		backendRegion, _ = labels[tofulabels.LegacyBackendRegionLabelKey]
		backendEndpoint, _ = labels[tofulabels.LegacyBackendEndpointLabelKey]
	}

	// Return nil if no backend labels are present
	if !hasType && !hasBucket && !hasKey {
		return nil, nil
	}

	// Extract whatever labels are present - validation happens later via Validate()
	config := &TofuBackendConfig{
		BackendType:     backendType,
		BackendBucket:   backendBucket,
		BackendKey:      backendKey,
		BackendRegion:   backendRegion,
		BackendEndpoint: backendEndpoint,
	}
	// Compute S3-compatible flag based on endpoint or region=auto
	config.S3Compatible = config.IsS3Compatible()

	return config, nil
}
