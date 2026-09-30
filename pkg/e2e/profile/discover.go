package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/plantonhq/planton/pkg/crkreflect"
	componentv1 "github.com/plantonhq/planton/qa/componente2eprofile/v1"
	providerv1 "github.com/plantonhq/planton/qa/providere2eprofile/v1"
	sharedpb "github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// ComponentEntry pairs a component name with its loaded profile.
type ComponentEntry struct {
	Name    string
	Profile *componentv1.ComponentE2EProfile
}

// FilterOpts controls which components are included in discovery.
type FilterOpts struct {
	// Only include components with this status. Empty means all.
	Status componentv1.ComponentE2EProfileSpec_Status
	// Only include components in this tier. 0 means all.
	Tier int32
	// Only include components that have been validated with this provisioner. 0 means all.
	Provisioner sharedpb.IacProvisioner
}

// DiscoverResult holds the full discovery output for a provider.
type DiscoverResult struct {
	Provider   *providerv1.ProviderE2EProfile
	Components []ComponentEntry
}

// Discover scans all component E2E profiles under a provider and applies filters.
func Discover(repoRoot, provider string, opts FilterOpts) (*DiscoverResult, error) {
	pp, err := LoadProviderProfile(repoRoot, provider)
	if err != nil {
		return nil, err
	}

	provDir := ProviderDir(repoRoot, provider)
	entries, err := os.ReadDir(provDir)
	if err != nil {
		return nil, errors.Wrapf(err, "reading provider directory %s", provDir)
	}

	var components []ComponentEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		componentName := entry.Name()

		// A directory that does not resolve to a registered kind is not a
		// component (e.g. the provider's aa_e2e/ or aa_import/ folders) —
		// skip it, exactly as a component without a profile is skipped.
		profilePath, err := ComponentProfilePath(repoRoot, provider, componentName)
		if err != nil {
			continue
		}
		if _, err := os.Stat(profilePath); err != nil {
			continue
		}

		cp, err := LoadComponentProfile(repoRoot, provider, componentName)
		if err != nil {
			return nil, err
		}
		if err := checkValidatedProvisioners(componentName, cp); err != nil {
			return nil, errors.Wrapf(err, "%s", profilePath)
		}

		if !matchesFilter(cp, opts) {
			continue
		}

		components = append(components, ComponentEntry{
			Name:    componentName,
			Profile: cp,
		})
	}

	sort.Slice(components, func(i, j int) bool {
		ti, tj := components[i].Profile.Spec.Tier, components[j].Profile.Spec.Tier
		if ti != tj {
			return ti < tj
		}
		return components[i].Name < components[j].Name
	})

	return &DiscoverResult{Provider: pp, Components: components}, nil
}

// checkValidatedProvisioners refuses a profile that claims an engine its kind
// does not run on (kind_meta.provisioners): the matrix would schedule a lane
// the CLI refuses before it starts.
func checkValidatedProvisioners(componentName string, cp *componentv1.ComponentE2EProfile) error {
	kind := crkreflect.KindFromString(componentName)
	if kind == cloudresourcekind.CloudResourceKind_unspecified || cp.GetSpec() == nil {
		return nil
	}
	for _, vp := range cp.GetSpec().GetValidatedProvisioners() {
		ok, err := crkreflect.RunsOn(kind, vp)
		if err != nil {
			return err
		}
		if !ok {
			return errors.Errorf("validated_provisioners lists %s, but %s does not run on it (kind_meta.provisioners); remove it", vp, kind)
		}
	}
	return nil
}

func matchesFilter(cp *componentv1.ComponentE2EProfile, opts FilterOpts) bool {
	spec := cp.Spec
	if spec == nil {
		return false
	}

	if opts.Status != componentv1.ComponentE2EProfileSpec_status_unspecified {
		if spec.Status != opts.Status {
			return false
		}
	}

	if opts.Tier > 0 && spec.Tier != opts.Tier {
		return false
	}

	if opts.Provisioner != sharedpb.IacProvisioner_iac_provisioner_unspecified {
		found := false
		for _, vp := range spec.ValidatedProvisioners {
			if vp == opts.Provisioner {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// MatrixCell represents one GitHub Actions matrix entry.
type MatrixCell struct {
	Name           string `json:"name"`
	Tier           int32  `json:"tier"`
	Engine         string `json:"engine"`
	Timeout        int32  `json:"timeout"`
	RunRegex       string `json:"run_regex"`
	ComponentCount int    `json:"component_count"`
}

// Matrix is the top-level GitHub Actions matrix JSON structure.
type Matrix struct {
	Include []MatrixCell `json:"include"`
}

// BuildGitHubMatrix generates the GitHub Actions matrix JSON from discovery results.
// Groups components by tier and provisioner, constructs -run regexes for go test.
func BuildGitHubMatrix(result *DiscoverResult) *Matrix {
	type groupKey struct {
		tier        int32
		provisioner sharedpb.IacProvisioner
	}

	groups := make(map[groupKey][]string)
	timeouts := make(map[groupKey]int32)

	for _, ce := range result.Components {
		spec := ce.Profile.Spec
		if spec == nil || spec.Status != componentv1.ComponentE2EProfileSpec_green {
			continue
		}

		for _, vp := range spec.ValidatedProvisioners {
			key := groupKey{tier: spec.Tier, provisioner: vp}
			groups[key] = append(groups[key], ce.Name)
			if spec.TimeoutMinutes > timeouts[key] {
				timeouts[key] = spec.TimeoutMinutes
			}
		}
	}

	var cells []MatrixCell
	for key, names := range groups {
		engineName := strings.ToLower(sharedpb.IacProvisioner_name[int32(key.provisioner)])
		tierTimeout := timeouts[key]
		if tierTimeout < 15 {
			tierTimeout = 15
		}
		// Buffer: at least 15 min overhead for kind setup/teardown + per-component overhead
		groupTimeout := tierTimeout*int32(len(names)) + 15

		runRegex := buildRunRegex(names, engineName)

		cells = append(cells, MatrixCell{
			Name:           fmt.Sprintf("Tier %d %s", key.tier, capitalize(engineName)),
			Tier:           key.tier,
			Engine:         engineName,
			Timeout:        groupTimeout,
			RunRegex:       runRegex,
			ComponentCount: len(names),
		})
	}

	sort.Slice(cells, func(i, j int) bool {
		if cells[i].Tier != cells[j].Tier {
			return cells[i].Tier < cells[j].Tier
		}
		return cells[i].Engine < cells[j].Engine
	})

	return &Matrix{Include: cells}
}

// MatrixJSON returns the GitHub Actions matrix as a JSON string.
func MatrixJSON(m *Matrix) (string, error) {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", errors.Wrap(err, "marshaling matrix JSON")
	}
	return string(b), nil
}

// buildRunRegex constructs a go test -run regex that matches all test functions
// for the given components and engine. Component names are converted to PascalCase
// Go test function names (e.g., "kubernetesvalkey" -> "KubernetesValkey").
func buildRunRegex(components []string, engine string) string {
	var parts []string
	for _, name := range components {
		parts = append(parts, toPascalCase(name))
	}
	engineSuffix := capitalize(engine)
	return fmt.Sprintf("Test(%s)_%s", strings.Join(parts, "|"), engineSuffix)
}

// testNameOverrides maps the component directory names whose Go test
// entrypoint capitalization deviates from the registry's enum-derived kind
// name. The registry is the source of truth for every other kind (component
// directories are the lowercased enum names, and test entrypoints follow the
// enum name), so keep this list to genuine, verified deviations only.
var testNameOverrides = map[string]string{
	// The enum is KubernetesArgocd but the test entrypoints (and the product
	// spelling) use TestKubernetesArgoCD_*.
	"kubernetesargocd": "KubernetesArgoCD",
}

// toPascalCase converts a lowercase component directory name to the PascalCase
// used in Go test entrypoint names (e.g. "awslambda" -> "AwsLambda" for
// TestAwsLambda_Pulumi), resolving through the kind registry so no
// hand-maintained per-kind table is needed.
func toPascalCase(name string) string {
	if name == "" {
		return ""
	}

	if override, ok := testNameOverrides[name]; ok {
		return override
	}

	kind := crkreflect.KindFromString(name)
	if kind != cloudresourcekind.CloudResourceKind_unspecified {
		if pascal := crkreflect.ExtractKindNameByKind(kind); pascal != "" {
			return pascal
		}
	}

	// Unregistered directory (should not happen for real components):
	// capitalize the first letter so the regex at least stays syntactically
	// valid; it will match zero tests, which the profile gates surface.
	return strings.ToUpper(name[:1]) + name[1:]
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// StatusCounts tallies components by status.
type StatusCounts struct {
	Green        int
	Deferred     int
	Skip         int
	Stub         int
	RealCluster  int
	PendingProof int
	Total        int
}

// CountByStatus counts components in the discovery result by their E2E status.
func CountByStatus(result *DiscoverResult) StatusCounts {
	var sc StatusCounts
	for _, ce := range result.Components {
		sc.Total++
		if ce.Profile.Spec == nil {
			continue
		}
		switch ce.Profile.Spec.Status {
		case componentv1.ComponentE2EProfileSpec_green:
			sc.Green++
		case componentv1.ComponentE2EProfileSpec_deferred:
			sc.Deferred++
		case componentv1.ComponentE2EProfileSpec_skip:
			sc.Skip++
		case componentv1.ComponentE2EProfileSpec_stub:
			sc.Stub++
		case componentv1.ComponentE2EProfileSpec_real_cluster:
			sc.RealCluster++
		case componentv1.ComponentE2EProfileSpec_pending_proof:
			sc.PendingProof++
		}
	}
	return sc
}
