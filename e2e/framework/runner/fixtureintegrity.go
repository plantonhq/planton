package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/e2e/profile"
	"github.com/plantonhq/planton/pkg/manifestgraph"
	componentv1 "github.com/plantonhq/planton/qa/componente2eprofile/v1"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// This file is the OFFLINE half of the reference-resolution contract that
// refresolve.go and dependencies.go execute at deploy time. A value_from
// reference that cannot resolve against the prerequisite chain fails only
// DEEP into a live run -- after minutes of fixture deploys (an ambiguous
// name), or worse, silently: an unresolvable reference is left untouched by
// design, the module receives no value, and the failure surfaces as a
// provider validation error at DEPLOY that reads like a module defect (see
// e2e/README.md, "Bare polymorphic references"). None of the other offline
// gates see this class: the manifest validates and the plan renders without
// the reference ever being followed.
//
// CheckScenarioFixtureIntegrity therefore replays, statically, exactly what
// DeployDependencies + ResolveManifestRefs will do: resolve the chain, walk
// each install manifest's references against the instances deployed BEFORE
// it, then walk the scenario's references against the full chain. The rules
// mirror lookupRefValue one for one -- including the sole-instance fallback
// (a name that matches no deployed instance still resolves when exactly one
// instance of the kind exists; scenario references name real-world topology,
// not install-profile files, so that is legal by design and NOT a finding).

// FixtureIntegrityFinding describes one reference (or chain-resolution
// failure) that would break a live run. ManifestPath is the manifest carrying
// the problem -- the scenario itself or one of its chain's install manifests.
type FixtureIntegrityFinding struct {
	ScenarioPath string
	ManifestPath string
	Field        string
	RefKind      string
	RefName      string
	Reason       string
}

func (f FixtureIntegrityFinding) String() string {
	ref := ""
	if f.RefKind != "" || f.RefName != "" {
		ref = fmt.Sprintf(" ref %s/%q on field %q:", f.RefKind, f.RefName, f.Field)
	}
	manifest := f.ManifestPath
	if manifest == f.ScenarioPath {
		manifest = "scenario"
	}
	return fmt.Sprintf("%s [%s]%s %s", f.ScenarioPath, manifest, ref, f.Reason)
}

// CheckScenarioFixtureIntegrity statically verifies that every value_from
// reference in the scenario -- and in every install manifest of its resolved
// prerequisite chain -- will resolve when the runner deploys that chain.
// Findings are defects to fix in the manifests (or the kind's registry
// prerequisites); the returned error is reserved for I/O-level failures of
// the checker itself.
func CheckScenarioFixtureIntegrity(repoRoot, componentProvider, component, scenarioPath string) ([]FixtureIntegrityFinding, error) {
	deps, err := ResolveDependencies(repoRoot, componentProvider, component, scenarioPath)
	if err != nil {
		// A chain that cannot even resolve (missing install manifest, unknown
		// annotation kind, cycle) is the first thing a live run would die on.
		return []FixtureIntegrityFinding{{
			ScenarioPath: scenarioPath,
			ManifestPath: scenarioPath,
			Reason:       "prerequisite chain resolution failed: " + err.Error(),
		}}, nil
	}

	// deployed accumulates (kind -> metadata.name set) in deploy order,
	// mirroring the `accumulated` outputs map in DeployDependencies: an
	// install manifest's references resolve only against instances deployed
	// before it; the scenario's resolve against the whole chain.
	deployed := make(map[cloudresourcekind.CloudResourceKind]map[string]bool)
	var findings []FixtureIntegrityFinding

	for _, dep := range deps {
		docPaths, err := splitManifestDocuments(dep.ManifestPath)
		if err != nil {
			findings = append(findings, FixtureIntegrityFinding{
				ScenarioPath: scenarioPath,
				ManifestPath: dep.ManifestPath,
				Reason:       "install profile cannot be split into documents: " + err.Error(),
			})
			continue
		}
		kind := crkreflect.KindFromString(dep.KindSlug)
		for _, docPath := range docPaths {
			findings = append(findings, manifestRefFindings(scenarioPath, dep.ManifestPath, docPath, deployed)...)

			name, err := manifestMetadataName(docPath)
			if err != nil {
				findings = append(findings, FixtureIntegrityFinding{
					ScenarioPath: scenarioPath,
					ManifestPath: dep.ManifestPath,
					Reason:       "install manifest has no readable metadata.name: " + err.Error(),
				})
				continue
			}
			if deployed[kind] == nil {
				deployed[kind] = make(map[string]bool)
			}
			deployed[kind][name] = true
		}
	}

	findings = append(findings, manifestRefFindings(scenarioPath, scenarioPath, scenarioPath, deployed)...)
	return findings, nil
}

// manifestRefFindings loads one manifest document and checks every value_from
// reference in its spec against the instances deployed so far. reportPath is
// the on-disk file to name in findings (docPath may be a temp per-document
// split of a multi-document profile).
func manifestRefFindings(scenarioPath, reportPath, docPath string, deployed map[cloudresourcekind.CloudResourceKind]map[string]bool) []FixtureIntegrityFinding {
	loadPath, err := withRunClockExpanded(docPath)
	if err != nil {
		return []FixtureIntegrityFinding{{
			ScenarioPath: scenarioPath,
			ManifestPath: reportPath,
			Reason:       "run-clock token cannot expand: " + err.Error(),
		}}
	}
	manifestObject, err := manifest.LoadManifest(loadPath)
	if err != nil {
		return []FixtureIntegrityFinding{{
			ScenarioPath: scenarioPath,
			ManifestPath: reportPath,
			Reason:       "manifest failed to load: " + err.Error(),
		}}
	}

	var findings []FixtureIntegrityFinding
	// A read-only replay of the deploy-time resolution through the SAME
	// traversal it uses (pkg/manifestgraph's walker -- map-typed reference
	// fields included, which the old spec-tree walk never saw).
	for _, use := range manifestgraph.CollectRefUses(manifestObject) {
		if finding := checkRefResolvable(use, deployed); finding != nil {
			finding.ScenarioPath = scenarioPath
			finding.ManifestPath = reportPath
			findings = append(findings, *finding)
		}
	}
	return findings
}

// checkRefResolvable mirrors the deploy-time lookup's resolution rules
// against the statically known instance names instead of live outputs.
// Returns nil when the reference will resolve. The finding's Field keeps the
// BARE field name -- it is a segment of the baseline's stable key shape, and
// renaming it would stale every committed entry at once.
func checkRefResolvable(use manifestgraph.RefUse, deployed map[cloudresourcekind.CloudResourceKind]map[string]bool) *FixtureIntegrityFinding {
	kind := manifestgraph.EffectiveKind(use)
	if kind == cloudresourcekind.CloudResourceKind_unspecified {
		return &FixtureIntegrityFinding{
			Field:   string(use.Field.Name()),
			RefName: use.Ref.GetName(),
			Reason: "reference carries no kind and the field declares no default_kind -- the runner leaves it " +
				"unresolved and the module deploys without the value; add an explicit `kind:` to the valueFrom",
		}
	}

	instances := deployed[kind]
	if len(instances) == 0 {
		return &FixtureIntegrityFinding{
			Field:   string(use.Field.Name()),
			RefKind: kind.String(),
			RefName: use.Ref.GetName(),
			Reason: "the prerequisite chain deploys no instance of this kind before this manifest -- the reference " +
				"can never resolve; add the kind to the component's registry prerequisites or the scenario's " +
				"planton.dev/e2e-prerequisites annotation",
		}
	}
	if instances[use.Ref.GetName()] || len(instances) == 1 {
		// Exact name match, or the runner's sole-instance fallback.
		return nil
	}
	names := make([]string, 0, len(instances))
	for n := range instances {
		names = append(names, n)
	}
	sort.Strings(names)
	return &FixtureIntegrityFinding{
		Field:   string(use.Field.Name()),
		RefKind: kind.String(),
		RefName: use.Ref.GetName(),
		Reason: fmt.Sprintf("name matches none of the %d deployed instances (%s) -- the runner fails the run on "+
			"this ambiguity after the fixtures are already deployed", len(instances), strings.Join(names, ", ")),
	}
}

// CheckCatalogFixtureIntegrity runs the scenario check across every component
// scenario in the repository's catalog and returns all findings, keeping one
// component's defect from hiding another's.
//
// Components whose E2E profile records `status: deferred` are skipped: a
// deferral is the kind's own record that its lanes cannot run (a wall-class
// prerequisite may be structurally unshippable — e.g. a fixture that would
// mutate the shared account irreversibly), so demanding a deployable fixture
// chain from it asserts a contract nobody holds. The gate re-arms the moment
// the profile leaves `deferred` — which is exactly when the chain must work.
// A missing or unreadable profile never skips: absence of a record is not a
// deferral.
func CheckCatalogFixtureIntegrity(repoRoot string) ([]FixtureIntegrityFinding, error) {
	catalogDir := filepath.Join(repoRoot, "catalog")
	providers, err := os.ReadDir(catalogDir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read catalog dir %s", catalogDir)
	}

	var findings []FixtureIntegrityFinding
	for _, provider := range providers {
		if !provider.IsDir() {
			continue
		}
		providerDir := filepath.Join(catalogDir, provider.Name())
		components, err := os.ReadDir(providerDir)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read provider dir %s", providerDir)
		}
		for _, component := range components {
			if !component.IsDir() {
				continue
			}
			if componentProfileDeferred(repoRoot, provider.Name(), component.Name()) {
				continue
			}
			scenariosDir := filepath.Join(providerDir, component.Name(), "e2e", "scenarios")
			scenarios, err := os.ReadDir(scenariosDir)
			if err != nil {
				continue // no scenarios -- nothing to check
			}
			for _, scenario := range scenarios {
				if scenario.IsDir() || !strings.HasSuffix(scenario.Name(), ".yaml") {
					continue
				}
				scenarioPath := filepath.Join(scenariosDir, scenario.Name())
				scenarioFindings, err := CheckScenarioFixtureIntegrity(repoRoot, provider.Name(), component.Name(), scenarioPath)
				if err != nil {
					return nil, errors.Wrapf(err, "fixture integrity check failed for %s", scenarioPath)
				}
				findings = append(findings, scenarioFindings...)
			}
		}
	}
	return findings, nil
}

// componentProfileDeferred reports whether a component's E2E profile records
// `status: deferred`. Any load failure (no profile, unreadable, unregistered
// directory name) returns false so the gate still checks the component --
// only an explicit deferral record earns the skip.
func componentProfileDeferred(repoRoot, provider, component string) bool {
	p, err := profile.LoadComponentProfile(repoRoot, provider, component)
	if err != nil {
		return false
	}
	return p.GetSpec().GetStatus() == componentv1.ComponentE2EProfileSpec_deferred
}

// withRunClockExpanded returns a manifest whose run-clock tokens hold timestamps, as a lane expands
// them before parsing: a token standing in a number field (an expiry, a start date) cannot load
// into the kind's message as text. A manifest without one passes through untouched; the copy
// keeps the base name and lives beside no scenario.
func withRunClockExpanded(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(raw), TimeTokenPrefix) {
		return path, nil
	}
	expanded, err := expandRunClock(string(raw), LaneClock())
	if err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp("", "planton-e2e-fixture-*")
	if err != nil {
		return "", err
	}
	copyPath := filepath.Join(dir, filepath.Base(path))
	return copyPath, os.WriteFile(copyPath, []byte(expanded), 0o600)
}
