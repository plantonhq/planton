package outputs

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// OutputMapping is the schema for output_transform.yaml, the declarative
// key-remapping mechanism for custom IaC modules whose output names don't
// match the proto StackOutputs field names.
type OutputMapping struct {
	// Version must be "v1". Reserved for future schema evolution.
	Version string `yaml:"version"`

	// Mappings maps module output keys (left) to proto field names (right).
	// Source keys may use dot-path notation (e.g., "connection.host").
	Mappings map[string]string `yaml:"mappings"`

	// Skip lists output keys that should be excluded from transformation.
	// Matching is performed after Flatten() and after key renaming.
	Skip []string `yaml:"skip"`
}

// loadMapping reads and parses an output_transform.yaml from moduleDir.
// Returns an error if the file cannot be read or parsed, or if the
// version field is unsupported.
func loadMapping(moduleDir string) (*OutputMapping, error) {
	path := filepath.Join(moduleDir, mappingFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", path)
	}

	var m OutputMapping
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", path)
	}

	if m.Version != "v1" {
		return nil, errors.Errorf("%s: unsupported version %q (expected \"v1\")", path, m.Version)
	}

	return &m, nil
}

// applyMapping renames and filters keys in a flat output map according to
// an OutputMapping. The original map is not modified.
//
// Processing order:
//  1. Rename: for each entry in Mappings, if the source key (left) exists
//     in outputs, the value is moved to the target key (right).
//  2. Skip: any key listed in Skip is removed from the result.
func applyMapping(outputs map[string]string, m *OutputMapping) map[string]string {
	if m == nil {
		return outputs
	}

	result := make(map[string]string, len(outputs))
	for k, v := range outputs {
		result[k] = v
	}

	for src, dst := range m.Mappings {
		if val, ok := result[src]; ok {
			delete(result, src)
			result[dst] = val
		}
	}

	skipSet := make(map[string]struct{}, len(m.Skip))
	for _, key := range m.Skip {
		skipSet[key] = struct{}{}
	}
	for key := range result {
		if _, skip := skipSet[key]; skip {
			delete(result, key)
		}
	}

	return result
}
