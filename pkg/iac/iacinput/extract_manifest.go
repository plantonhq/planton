package iacinput

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/ulidgen"
	"gopkg.in/yaml.v3"
)

// ExtractManifestFromIacInput reads an IaC input YAML file, extracts
// the "target" field, and writes it to a temporary file.
// Returns the path to the temporary manifest file.
func ExtractManifestFromIacInput(iacInputPath string) (manifestPath string, err error) {
	iacInputBytes, err := os.ReadFile(iacInputPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read IaC input file %s", iacInputPath)
	}

	var iacInputMap map[string]interface{}
	if err := yaml.Unmarshal(iacInputBytes, &iacInputMap); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal IaC input YAML")
	}

	targetField, ok := iacInputMap["target"]
	if !ok {
		return "", errors.New("IaC input file does not contain 'target' field")
	}

	targetBytes, err := yaml.Marshal(targetField)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal target field to YAML")
	}

	downloadDir, err := workspace.GetManifestDownloadDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get manifest download directory")
	}

	fileName := ulidgen.NewGenerator().Generate().String() + "-manifest.yaml"
	manifestPath = filepath.Join(downloadDir, fileName)

	if err := os.WriteFile(manifestPath, targetBytes, 0600); err != nil {
		return "", errors.Wrapf(err, "failed to write manifest to %s", manifestPath)
	}

	return manifestPath, nil
}
