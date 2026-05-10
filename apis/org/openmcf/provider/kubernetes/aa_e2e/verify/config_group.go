package verify

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ConfigGroupVerifier handles KubernetesManifest components that apply
// arbitrary multi-document YAML via Pulumi's yamlv2.ConfigGroup.
// It parses the embedded spec.manifest_yaml, extracts each document's
// kind and name, then verifies all resources exist (or are absent).
type ConfigGroupVerifier struct {
	Namespace    string
	ManifestPath string
}

func (v *ConfigGroupVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	resources, err := ParseEmbeddedManifests(v.ManifestPath)
	if err != nil {
		return err
	}
	if len(resources) == 0 {
		return errors.New("no embedded resources found in spec.manifest_yaml")
	}

	for _, res := range resources {
		ns := res.Namespace
		if ns == "" {
			ns = v.Namespace
		}
		if err := KubectlResourceExists(ctx, kubeconfig, res.Kind, res.Name, ns); err != nil {
			return errors.Wrapf(err, "embedded resource %s/%s not found", res.Kind, res.Name)
		}
	}
	return nil
}

func (v *ConfigGroupVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	resources, err := ParseEmbeddedManifests(v.ManifestPath)
	if err != nil {
		return err
	}

	for _, res := range resources {
		ns := res.Namespace
		if ns == "" {
			ns = v.Namespace
		}
		if err := KubectlResourceAbsent(ctx, kubeconfig, res.Kind, res.Name, ns); err != nil {
			return errors.Wrapf(err, "embedded resource %s/%s still exists", res.Kind, res.Name)
		}
	}
	return nil
}

// EmbeddedResource represents a single Kubernetes resource parsed from
// the spec.manifest_yaml field of a KubernetesManifest component.
type EmbeddedResource struct {
	Kind      string
	Name      string
	Namespace string
}

// ParseEmbeddedManifests reads a KubernetesManifest's YAML file, extracts
// the spec.manifest_yaml field, splits on document separators, and returns
// the kind + name of each embedded Kubernetes resource.
func ParseEmbeddedManifests(manifestPath string) ([]EmbeddedResource, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read manifest %s", manifestPath)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, errors.Wrapf(err, "failed to parse manifest YAML %s", manifestPath)
	}

	spec, ok := raw["spec"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("manifest %s has no spec section", manifestPath)
	}

	manifestYAML, ok := spec["manifest_yaml"].(string)
	if !ok || manifestYAML == "" {
		return nil, errors.Errorf("manifest %s has no spec.manifest_yaml field", manifestPath)
	}

	docs := splitYAMLDocuments(manifestYAML)
	var resources []EmbeddedResource

	for _, doc := range docs {
		trimmed := strings.TrimSpace(doc)
		if trimmed == "" {
			continue
		}

		var obj map[string]interface{}
		if err := yaml.Unmarshal([]byte(trimmed), &obj); err != nil {
			continue
		}

		res := EmbeddedResource{}
		if kind, ok := obj["kind"].(string); ok {
			res.Kind = strings.ToLower(kind)
		}
		if metadata, ok := obj["metadata"].(map[string]interface{}); ok {
			if name, ok := metadata["name"].(string); ok {
				res.Name = name
			}
			if ns, ok := metadata["namespace"].(string); ok {
				res.Namespace = ns
			}
		}

		if res.Kind != "" && res.Name != "" {
			resources = append(resources, res)
		}
	}

	return resources, nil
}

// splitYAMLDocuments splits a multi-document YAML string on "---" separators.
func splitYAMLDocuments(content string) []string {
	var docs []string
	var current bytes.Buffer

	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "---" {
			if current.Len() > 0 {
				docs = append(docs, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteString(line)
		current.WriteByte('\n')
	}

	if current.Len() > 0 {
		docs = append(docs, current.String())
	}

	return docs
}
