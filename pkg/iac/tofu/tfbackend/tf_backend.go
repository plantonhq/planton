package tfbackend

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/iac/terraform"
)

// WriteBackendFile creates a `backend.tf` file in projectDir using the backend type specified by tofuBackendType.
func WriteBackendFile(projectDir string, tofuBackendType terraform.TerraformBackendType) error {
	backendName := tofuBackendType.String()

	// Construct a minimal backend configuration.
	// The user can further configure the backend by editing this file or using -backend-config flags.
	backendContent := fmt.Sprintf(`terraform {
  backend "%s" {}
}
`, backendName)

	backendFilePath := filepath.Join(projectDir, "backend.tf")
	if err := os.WriteFile(backendFilePath, []byte(backendContent), 0644); err != nil {
		return errors.Wrap(err, "failed to write backend file")
	}

	return nil
}

func BackendTypeFromString(backendTypeStr string) terraform.TerraformBackendType {
	switch backendTypeStr {
	case "local":
		return terraform.TerraformBackendType_local
	case "s3":
		return terraform.TerraformBackendType_s3
	case "gcs":
		return terraform.TerraformBackendType_gcs
	case "azurerm":
		return terraform.TerraformBackendType_azurerm
	default:
		return terraform.TerraformBackendType_terraform_backend_type_unspecified
	}
}
