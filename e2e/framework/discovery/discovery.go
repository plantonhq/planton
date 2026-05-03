// Package discovery scans the openmcf repository to find testable components
// and their associated IaC modules and hack manifests.
package discovery

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Component represents a discovered OpenMCF component that can be E2E tested.
type Component struct {
	// Name is the component name in lowercase (e.g., "kubernetesnamespace").
	Name string

	// Provider is the cloud provider (e.g., "kubernetes", "aws", "gcp").
	Provider string

	// ManifestPath is the absolute path to iac/hack/manifest.yaml.
	ManifestPath string

	// PulumiDir is the absolute path to iac/pulumi/ (empty if not present).
	PulumiDir string

	// TerraformDir is the absolute path to iac/tf/ (empty if not present).
	TerraformDir string
}

// DiscoverComponents scans the apis directory tree to find all components
// that have an iac/hack/manifest.yaml file (meaning they're testable).
func DiscoverComponents(repoRoot string) ([]Component, error) {
	apisDir := filepath.Join(repoRoot, "apis", "org", "openmcf", "provider")

	var components []Component

	providerDirs, err := os.ReadDir(apisDir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read provider directory %s", apisDir)
	}

	for _, providerEntry := range providerDirs {
		if !providerEntry.IsDir() {
			continue
		}
		providerName := providerEntry.Name()
		providerPath := filepath.Join(apisDir, providerName)

		componentDirs, err := os.ReadDir(providerPath)
		if err != nil {
			continue
		}

		for _, componentEntry := range componentDirs {
			if !componentEntry.IsDir() {
				continue
			}
			componentName := componentEntry.Name()

			// Components live at provider/{component}/v1/iac/
			iacBase := filepath.Join(providerPath, componentName, "v1", "iac")
			manifestPath := filepath.Join(iacBase, "hack", "manifest.yaml")

			if _, err := os.Stat(manifestPath); err != nil {
				continue
			}

			comp := Component{
				Name:         componentName,
				Provider:     providerName,
				ManifestPath: manifestPath,
			}

			pulumiDir := filepath.Join(iacBase, "pulumi")
			if info, err := os.Stat(pulumiDir); err == nil && info.IsDir() {
				comp.PulumiDir = pulumiDir
			}

			tfDir := filepath.Join(iacBase, "tf")
			if info, err := os.Stat(tfDir); err == nil && info.IsDir() {
				comp.TerraformDir = tfDir
			}

			components = append(components, comp)
		}
	}

	return components, nil
}

// DiscoverByProvider filters discovered components to a single provider.
func DiscoverByProvider(repoRoot, providerName string) ([]Component, error) {
	all, err := DiscoverComponents(repoRoot)
	if err != nil {
		return nil, err
	}

	var filtered []Component
	for _, c := range all {
		if strings.EqualFold(c.Provider, providerName) {
			filtered = append(filtered, c)
		}
	}
	return filtered, nil
}

// DiscoverByName finds a single component by name (case-insensitive).
func DiscoverByName(repoRoot, componentName string) (*Component, error) {
	all, err := DiscoverComponents(repoRoot)
	if err != nil {
		return nil, err
	}

	for _, c := range all {
		if strings.EqualFold(c.Name, componentName) {
			return &c, nil
		}
	}
	return nil, errors.Errorf("component %q not found", componentName)
}
