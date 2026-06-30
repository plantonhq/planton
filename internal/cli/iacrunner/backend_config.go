package iacrunner

import (
	"github.com/plantonhq/planton/pkg/iac/tofu/backendconfig"
	"google.golang.org/protobuf/proto"
)

// BackendConfigInfo holds backend configuration for display purposes.
type BackendConfigInfo struct {
	// BackendType is the backend type (e.g., "s3", "gcs", "azurerm")
	BackendType string
	// Bucket is the bucket or container name
	Bucket string
	// Key is the path to the state file within the bucket
	Key string
}

// ExtractBackendConfigForDisplay extracts backend config for CLI display.
// Returns nil if no backend config is present (local backend).
func ExtractBackendConfigForDisplay(manifest proto.Message, provisioner string) *BackendConfigInfo {
	config, err := backendconfig.ExtractFromManifest(manifest, provisioner)
	if err != nil || config == nil {
		return nil
	}

	return &BackendConfigInfo{
		BackendType: config.BackendType,
		Bucket:      config.BackendBucket,
		Key:         config.BackendKey,
	}
}
