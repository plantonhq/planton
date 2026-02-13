package backendconfig

import (
	"fmt"
)

// ValidationResult contains validation status and details about missing fields.
type ValidationResult struct {
	// Valid is true if all required fields are present
	Valid bool
	// MissingFields contains details about each missing required field
	MissingFields []MissingField
	// Warnings contains non-fatal issues that should be displayed to the user
	Warnings []string
}

// MissingField describes a missing required field with guidance for the user.
type MissingField struct {
	// Name is the field name (e.g., "endpoint")
	Name string
	// FlagName is the CLI flag to provide this value (e.g., "--backend-endpoint")
	FlagName string
	// EnvVarName is the environment variable to provide this value (e.g., "OPENMCF_BACKEND_ENDPOINT")
	EnvVarName string
	// LabelName is the manifest label (e.g., "terraform.openmcf.org/backend.endpoint")
	LabelName string
	// Description is a human-readable description of the field
	Description string
	// Example is an example value to show the user
	Example string
	// Required indicates if this field is required
	Required bool
}

// Validate checks if the backend configuration is complete for the specified backend type.
// Returns a ValidationResult with details about any missing required fields.
func Validate(config *TofuBackendConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	switch config.BackendType {
	case "s3":
		result = validateS3Backend(config)
	case "gcs":
		result = validateGCSBackend(config)
	case "azurerm":
		result = validateAzureBackend(config)
	case "local", "":
		// Local backend has no required fields
	default:
		result.Valid = false
		result.MissingFields = append(result.MissingFields, MissingField{
			Name:        "type",
			Description: fmt.Sprintf("Unknown backend type: %s. Supported types: s3, gcs, azurerm, local", config.BackendType),
			Required:    true,
		})
	}

	return result
}

// validateS3Backend validates S3 backend configuration.
func validateS3Backend(config *TofuBackendConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Bucket is always required for S3
	if config.BackendBucket == "" {
		result.Valid = false
		result.MissingFields = append(result.MissingFields, MissingField{
			Name:        "bucket",
			FlagName:    "--backend-bucket",
			EnvVarName:  EnvBackendBucket,
			LabelName:   "terraform.openmcf.org/backend.bucket",
			Description: "S3 bucket name for state storage",
			Example:     "my-terraform-state-bucket",
			Required:    true,
		})
	}

	// Key is always required
	if config.BackendKey == "" {
		result.Valid = false
		result.MissingFields = append(result.MissingFields, MissingField{
			Name:        "key",
			FlagName:    "--backend-key",
			EnvVarName:  "", // Key is intentionally not read from env vars
			LabelName:   "terraform.openmcf.org/backend.key",
			Description: "Path to state file within bucket",
			Example:     "env/prod/terraform.tfstate",
			Required:    true,
		})
	}

	// Region is required for S3
	if config.BackendRegion == "" {
		result.Valid = false
		result.MissingFields = append(result.MissingFields, MissingField{
			Name:        "region",
			FlagName:    "--backend-region",
			EnvVarName:  EnvBackendRegion,
			LabelName:   "terraform.openmcf.org/backend.region",
			Description: "AWS region (use 'auto' for S3-compatible backends like R2)",
			Example:     "us-west-2 (or 'auto' for R2/MinIO)",
			Required:    true,
		})
	}

	// S3-compatible detection: region=auto requires endpoint
	if config.BackendRegion == "auto" && config.BackendEndpoint == "" {
		result.Valid = false
		result.MissingFields = append(result.MissingFields, MissingField{
			Name:        "endpoint",
			FlagName:    "--backend-endpoint",
			EnvVarName:  EnvBackendEndpoint,
			LabelName:   "terraform.openmcf.org/backend.endpoint",
			Description: "Custom S3-compatible endpoint (required when region is 'auto')",
			Example:     "https://<account-id>.r2.cloudflarestorage.com",
			Required:    true,
		})
		result.Warnings = append(result.Warnings,
			"Detected S3-compatible backend (region=auto). Endpoint is required for R2, MinIO, etc.")
	}

	return result
}

// validateGCSBackend validates GCS backend configuration.
func validateGCSBackend(config *TofuBackendConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Bucket is required for GCS
	if config.BackendBucket == "" {
		result.Valid = false
		result.MissingFields = append(result.MissingFields, MissingField{
			Name:        "bucket",
			FlagName:    "--backend-bucket",
			EnvVarName:  EnvBackendBucket,
			LabelName:   "terraform.openmcf.org/backend.bucket",
			Description: "GCS bucket name for state storage",
			Example:     "my-terraform-state",
			Required:    true,
		})
	}

	// Prefix (key) is required for GCS
	if config.BackendKey == "" {
		result.Valid = false
		result.MissingFields = append(result.MissingFields, MissingField{
			Name:        "key",
			FlagName:    "--backend-key",
			EnvVarName:  "", // Key is intentionally not read from env vars
			LabelName:   "terraform.openmcf.org/backend.key",
			Description: "Prefix path for state file within bucket",
			Example:     "terraform/state",
			Required:    true,
		})
	}

	return result
}

// validateAzureBackend validates Azure backend configuration.
func validateAzureBackend(config *TofuBackendConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Container name (bucket) is required for Azure
	if config.BackendBucket == "" {
		result.Valid = false
		result.MissingFields = append(result.MissingFields, MissingField{
			Name:        "bucket",
			FlagName:    "--backend-bucket",
			EnvVarName:  EnvBackendBucket,
			LabelName:   "terraform.openmcf.org/backend.bucket",
			Description: "Azure Storage container name for state storage",
			Example:     "tfstate",
			Required:    true,
		})
	}

	// Key is required for Azure
	if config.BackendKey == "" {
		result.Valid = false
		result.MissingFields = append(result.MissingFields, MissingField{
			Name:        "key",
			FlagName:    "--backend-key",
			EnvVarName:  "", // Key is intentionally not read from env vars
			LabelName:   "terraform.openmcf.org/backend.key",
			Description: "State file blob name",
			Example:     "prod.terraform.tfstate",
			Required:    true,
		})
	}

	return result
}
