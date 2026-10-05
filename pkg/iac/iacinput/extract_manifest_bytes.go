package iacinput

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/ulidgen"
	"gopkg.in/yaml.v3"
)

// ExtractManifestFromBytes extracts the manifest from IaC input YAML bytes.
// This enables reading IaC input from sources other than files (e.g., clipboard).
// The IaC input must contain a "target" field with the manifest content.
// Returns the path to the temporary manifest file.
func ExtractManifestFromBytes(iacInputBytes []byte) (manifestPath string, err error) {
	var iacInputMap map[string]interface{}
	if err := yaml.Unmarshal(iacInputBytes, &iacInputMap); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal IaC input YAML")
	}

	targetField, ok := iacInputMap["target"]
	if !ok {
		return "", errors.New("IaC input does not contain 'target' field")
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

// IsIacInput checks if the given YAML bytes represent an IaC input (has "target" key at root).
func IsIacInput(content []byte) bool {
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		return false
	}
	_, hasTarget := parsed["target"]
	return hasTarget
}
