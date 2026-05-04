// Package verify provides manifest-driven resource verification for
// Kubernetes E2E tests. Each verifier type checks that a specific class
// of Kubernetes resource (namespace, workload, Helm chart, CRD workload,
// operator) is present and healthy after deployment, or absent after destroy.
package verify

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ManifestInfo holds the parsed fields from a manifest needed for verification.
type ManifestInfo struct {
	Kind      string
	Name      string
	Namespace string
}

// ParseManifestInfo extracts kind, name, and namespace from a manifest YAML file
// to drive dynamic verification without hardcoded values.
func ParseManifestInfo(manifestPath string) (*ManifestInfo, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read manifest %s", manifestPath)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, errors.Wrapf(err, "failed to parse manifest YAML %s", manifestPath)
	}

	info := &ManifestInfo{}

	if kind, ok := raw["kind"].(string); ok {
		info.Kind = kind
	}

	if metadata, ok := raw["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"].(string); ok {
			info.Name = name
		}
	}

	if spec, ok := raw["spec"].(map[string]interface{}); ok {
		if name, ok := spec["name"].(string); ok {
			info.Name = name
		}

		switch ns := spec["namespace"].(type) {
		case string:
			info.Namespace = ns
		case map[string]interface{}:
			if val, ok := ns["value"].(string); ok {
				info.Namespace = val
			}
		}
	}

	if info.Namespace == "" {
		info.Namespace = "default"
	}

	return info, nil
}
