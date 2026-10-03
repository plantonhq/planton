// Package infrachart loads, renders, and validates InfraCharts offline — the
// parameterized bundles of catalog object manifests that live under charts/.
//
// The package mirrors the platform's server-side chart pipeline closely enough
// to be a trustworthy pre-publish gate: templates are rendered with the
// chart's default values through a Jinja engine constrained to the same
// sandboxed language subset the platform renders with, every rendered
// document is validated strictly against the protobuf schema of its kind, and
// every valueFrom reference is checked for resolvability against the
// referenced kind's actual proto surface. What it deliberately does NOT do is
// talk to a control plane — the platform's `chart build` RPC remains the
// authoritative server-side gate; this package is the fast, offline
// equivalent for authoring loops and CI.
package infrachart

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// Chart is an InfraChart loaded from a chart directory.
type Chart struct {
	// Dir is the chart directory the chart was loaded from.
	Dir string

	// Name is the human-readable chart name from Chart.yaml metadata.name.
	Name string

	// Description is spec.description from Chart.yaml.
	Description string

	// Params are the parameters declared in values.yaml, in declaration order.
	Params []Param

	// Templates are the template files under templates/, sorted by name for
	// deterministic rendering and reporting.
	Templates []TemplateFile
}

// TemplateFile is one file under the chart's templates/ directory.
type TemplateFile struct {
	// Name is the path relative to templates/ (e.g. "network.yaml").
	Name string

	// Content is the raw template source.
	Content string
}

// chartYaml models the subset of Chart.yaml the loader needs. Unknown fields
// are tolerated (Chart.yaml is a platform API whose schema may grow).
type chartYaml struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Description string `json:"description"`
	} `json:"spec"`
}

// IsChartDir reports whether dir holds an infra-chart. A Chart.yaml is the
// marker, matching how chart bundles are discovered everywhere else.
func IsChartDir(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "Chart.yaml"))
	return err == nil && !info.IsDir()
}

// LoadDir loads a chart from a directory containing Chart.yaml, values.yaml,
// and templates/.
func LoadDir(dir string) (*Chart, error) {
	chartYamlBytes, err := os.ReadFile(filepath.Join(dir, "Chart.yaml"))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read Chart.yaml in %s", dir)
	}
	var cy chartYaml
	if err := yaml.Unmarshal(chartYamlBytes, &cy); err != nil {
		return nil, errors.Wrap(err, "failed to parse Chart.yaml")
	}
	if cy.Kind != "InfraChart" {
		return nil, errors.Errorf("Chart.yaml kind must be InfraChart (got %q)", cy.Kind)
	}
	if cy.Metadata.Name == "" {
		return nil, errors.New("Chart.yaml metadata.name is required")
	}

	params, err := loadValuesFile(filepath.Join(dir, "values.yaml"))
	if err != nil {
		return nil, err
	}

	templates, err := loadTemplates(filepath.Join(dir, "templates"))
	if err != nil {
		return nil, err
	}
	if len(templates) == 0 {
		return nil, errors.Errorf("chart has no templates under %s", filepath.Join(dir, "templates"))
	}

	return &Chart{
		Dir:         dir,
		Name:        cy.Metadata.Name,
		Description: cy.Spec.Description,
		Params:      params,
		Templates:   templates,
	}, nil
}

// loadTemplates reads every .yaml/.yml file under templatesDir (recursively —
// charts may group templates in subdirectories, e.g. kubernetes/addons/).
func loadTemplates(templatesDir string) ([]TemplateFile, error) {
	var out []TemplateFile
	err := filepath.WalkDir(templatesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ext := strings.ToLower(filepath.Ext(path)); ext != ".yaml" && ext != ".yml" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(templatesDir, path)
		if err != nil {
			return err
		}
		out = append(out, TemplateFile{Name: rel, Content: string(content)})
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read templates in %s", templatesDir)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
