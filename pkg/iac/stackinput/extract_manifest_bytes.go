package stackinput

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/ulidgen"
	"gopkg.in/yaml.v3"
)

// ExtractManifestFromBytes extracts the manifest from stack input YAML bytes.
// This enables reading stack input from sources other than files (e.g., clipboard).
// The stack input must contain a "target" field with the manifest content.
// Returns the path to the temporary manifest file.
func ExtractManifestFromBytes(stackInputBytes []byte) (manifestPath string, err error) {
	var stackInputMap map[string]interface{}
	if err := yaml.Unmarshal(stackInputBytes, &stackInputMap); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal stack input YAML")
	}

	targetField, ok := stackInputMap["target"]
	if !ok {
		return "", errors.New("stack input does not contain 'target' field")
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

// IsStackInput checks if the given YAML bytes represent a stack input (has "target" key at root).
func IsStackInput(content []byte) bool {
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		return false
	}
	_, hasTarget := parsed["target"]
	return hasTarget
}
