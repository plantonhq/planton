package backendconfig

import (
	"google.golang.org/protobuf/proto"
)

// CLIBackendFlags holds backend configuration from CLI flags.
// These flags take precedence over all other sources when merging configuration.
type CLIBackendFlags struct {
	BackendType     string
	BackendBucket   string
	BackendKey      string
	BackendRegion   string
	BackendEndpoint string
}

// BuildBackendConfig merges configuration from multiple sources.
// Priority order: CLI flags > Manifest labels > Environment variables
//
// This allows users to:
// - Set defaults via environment variables (OPENMCF_BACKEND_*)
// - Override with manifest labels for resource-specific settings
// - Override with CLI flags for CI/CD or local testing
func BuildBackendConfig(
	manifest proto.Message,
	provisionerType string,
	cliFlags CLIBackendFlags,
) (*TofuBackendConfig, error) {
	// 1. Start with empty config
	config := &TofuBackendConfig{}

	// 2. Apply environment variables (lowest priority layer)
	envConfig := ReadFromEnv()
	applyEnvOverrides(config, envConfig)

	// 3. Apply manifest labels (middle layer - overrides env vars)
	manifestConfig, _ := ExtractFromManifest(manifest, provisionerType)
	if manifestConfig != nil {
		applyManifestOverrides(config, manifestConfig)
	}

	// 4. Apply CLI flags (highest priority - overrides everything)
	applyCLIOverrides(config, cliFlags)

	// Recompute S3-compatible flag after merging all layers
	config.S3Compatible = config.IsS3Compatible()

	return config, nil
}

// applyEnvOverrides applies environment variable values to the config.
// Only non-empty values are applied.
func applyEnvOverrides(config *TofuBackendConfig, env EnvBackendConfig) {
	if env.BackendType != "" {
		config.BackendType = env.BackendType
	}
	if env.BackendBucket != "" {
		config.BackendBucket = env.BackendBucket
	}
	if env.BackendRegion != "" {
		config.BackendRegion = env.BackendRegion
	}
	if env.BackendEndpoint != "" {
		config.BackendEndpoint = env.BackendEndpoint
	}
	// Note: BackendKey is intentionally NOT read from environment variables
}

// applyManifestOverrides applies manifest label values to the config.
// Only non-empty values are applied.
func applyManifestOverrides(config *TofuBackendConfig, manifest *TofuBackendConfig) {
	if manifest.BackendType != "" {
		config.BackendType = manifest.BackendType
	}
	if manifest.BackendBucket != "" {
		config.BackendBucket = manifest.BackendBucket
	}
	if manifest.BackendKey != "" {
		config.BackendKey = manifest.BackendKey
	}
	if manifest.BackendRegion != "" {
		config.BackendRegion = manifest.BackendRegion
	}
	if manifest.BackendEndpoint != "" {
		config.BackendEndpoint = manifest.BackendEndpoint
	}
}

// applyCLIOverrides applies CLI flag values to the config.
// Only non-empty values are applied. CLI flags have highest priority.
func applyCLIOverrides(config *TofuBackendConfig, cli CLIBackendFlags) {
	if cli.BackendType != "" && cli.BackendType != "local" {
		config.BackendType = cli.BackendType
	}
	if cli.BackendBucket != "" {
		config.BackendBucket = cli.BackendBucket
	}
	if cli.BackendKey != "" {
		config.BackendKey = cli.BackendKey
	}
	if cli.BackendRegion != "" {
		config.BackendRegion = cli.BackendRegion
	}
	if cli.BackendEndpoint != "" {
		config.BackendEndpoint = cli.BackendEndpoint
	}
}
