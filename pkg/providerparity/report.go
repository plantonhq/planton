//go:build !codegen
// +build !codegen

package providerparity

import (
	"sort"

	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// Report is the machine-readable provider-parity measurement for one cloud
// provider's catalog: the three censuses joined into per-kind numbers plus
// catalog-level aggregates. It carries measurement only -- matching spec
// fields to provider arguments and gating unaccounted surface is the
// total-accounting layer's job -- and is the single aggregation every
// consumer (CLI output, CI gate, the public parity report generator) renders
// from, so the numbers can never disagree between surfaces.
type Report struct {
	// CloudProvider is the catalog provider measured, e.g. "gcp".
	CloudProvider string `json:"cloudProvider"`
	// SchemaVersions names the exact pinned release of each loaded Terraform
	// provider schema, e.g. {"google": "6.50.0"} -- the versions parity is
	// declared against.
	SchemaVersions map[string]string `json:"schemaVersions"`
	Kinds          int               `json:"kinds"`
	// DistinctResources is the union of resource types consumed across all
	// kinds' Terraform modules.
	DistinctResources int `json:"distinctResources"`
	// TotalSpecFields is the catalog's authored surface: the sum of every
	// kind's spec leaf-field count.
	TotalSpecFields int `json:"totalSpecFields"`
	// TotalConfigurableArgs is the provider-side surface of the consumed
	// resources: the sum of non-deprecated configurable arguments across
	// DistinctResources, at the pinned versions.
	TotalConfigurableArgs int `json:"totalConfigurableArgs"`
	// UnknownResources are consumed resource types absent from every loaded
	// schema -- a module using a resource its declared pin does not serve
	// (or a schema artifact behind the modules). Always investigate.
	UnknownResources []string `json:"unknownResources,omitempty"`
	// PinDistribution counts modules per declared provider constraint,
	// e.g. {"google": {"~> 6.0": 78, "~> 5.0": 1}} -- a single-pin catalog
	// shows exactly one constraint per provider.
	PinDistribution map[string]map[string]int `json:"pinDistribution"`
	// KindReports is the per-kind breakdown, sorted by kind.
	KindReports []KindReport `json:"kindReports"`
}

// KindReport is one kind's slice of the measurement.
type KindReport struct {
	Kind       string            `json:"kind"`
	SpecFields int               `json:"specFields"`
	Resources  []ResourceUse     `json:"resources"`
	Pins       map[string]string `json:"pins"`
}

// ResourceUse is one consumed resource resolved against the loaded schemas.
type ResourceUse struct {
	Name string `json:"name"`
	// Schema is the loaded schema that defines the resource, resolved in
	// sorted provider-name order (so "google" wins over "google-beta" when
	// both define it -- GA is the parity baseline). Empty means unknown.
	Schema string `json:"schema,omitempty"`
	// ConfigurableArgs is the resource's non-deprecated configurable
	// argument count at the pinned version.
	ConfigurableArgs int  `json:"configurableArgs"`
	Deprecated       bool `json:"deprecated,omitempty"`
}

// BuildReport measures one cloud provider's catalog: the spec and module
// censuses are run here; schemas are the committed artifacts (LoadSchemas).
func BuildReport(repoRoot string, provider cloudresourcekind.CloudResourceProvider, schemas map[string]*Schema) (Report, error) {
	spec := SpecCensus(provider)
	modules, err := ModuleCensusForProvider(repoRoot, provider)
	if err != nil {
		return Report{}, err
	}
	return buildReport(crkreflect.ProviderDirName(provider), spec, modules, schemas), nil
}

func buildReport(cloudProvider string, spec []KindCensus, modules []ModuleCensus, schemas map[string]*Schema) Report {
	schemaNames := make([]string, 0, len(schemas))
	versions := map[string]string{}
	for name, s := range schemas {
		schemaNames = append(schemaNames, name)
		versions[name] = s.Version
	}
	sort.Strings(schemaNames)

	specFields := map[string]int{}
	for _, k := range spec {
		specFields[k.Kind] = len(k.SpecFieldPaths)
	}

	report := Report{
		CloudProvider:   cloudProvider,
		SchemaVersions:  versions,
		Kinds:           len(modules),
		PinDistribution: map[string]map[string]int{},
	}

	distinct := map[string]ResourceUse{}
	for _, m := range modules {
		kr := KindReport{
			Kind:       m.Kind,
			SpecFields: specFields[m.Kind],
			Pins:       m.Pins,
		}
		for _, res := range m.Resources {
			use := ResourceUse{Name: res}
			// The report measures without knowing the provider's baseline
			// (enrollment names it); sorted order places each canonical
			// schema before its secondary channel ("google" < "google-beta"),
			// and the accounting -- which does know the baseline -- is the
			// arbiter of admission.
			if block, name := resolveResourceSchema(schemas, schemaNames, "", res); block != nil {
				use.Schema = name
				use.ConfigurableArgs = block.ConfigurableArgCount()
				use.Deprecated = block.Deprecated
			}
			kr.Resources = append(kr.Resources, use)
			distinct[res] = use
		}
		for pinName, constraint := range m.Pins {
			byConstraint := report.PinDistribution[pinName]
			if byConstraint == nil {
				byConstraint = map[string]int{}
				report.PinDistribution[pinName] = byConstraint
			}
			byConstraint[constraint]++
		}
		report.TotalSpecFields += kr.SpecFields
		report.KindReports = append(report.KindReports, kr)
	}

	report.DistinctResources = len(distinct)
	names := make([]string, 0, len(distinct))
	for name := range distinct {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		use := distinct[name]
		if use.Schema == "" {
			report.UnknownResources = append(report.UnknownResources, name)
			continue
		}
		report.TotalConfigurableArgs += use.ConfigurableArgs
	}
	return report
}
