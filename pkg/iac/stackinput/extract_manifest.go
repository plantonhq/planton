package stackinput

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/ulidgen"
	"gopkg.in/yaml.v3"
)

// ExtractManifestFromStackInput reads a stack input YAML file, extracts
// the "target" field, and writes it to a temporary file.
// Returns the path to the temporary manifest file.
func ExtractManifestFromStackInput(stackInputPath string) (manifestPath string, err error) {
	stackInputBytes, err := os.ReadFile(stackInputPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read stack input file %s", stackInputPath)
	}

	var stackInputMap map[string]interface{}
	if err := yaml.Unmarshal(stackInputBytes, &stackInputMap); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal stack input YAML")
	}

	targetField, ok := stackInputMap["target"]
	if !ok {
		return "", errors.New("stack input file does not contain 'target' field")
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
