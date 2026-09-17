//go:build !codegen
// +build !codegen

// The total-accounting check: every configurable, non-deprecated argument of
// every consumed provider resource is exact-matched to a spec field, mapped
// by recorded judgment, or excluded with a reason -- and, in reverse, every
// spec leaf field lands on a provider argument or carries an exclusion. The
// reverse direction is not symmetry for its own sake: specs are authored
// against a moving provider, and a field the provider no longer serves is
// drift the forward walk alone can never see.
//
// Anything unaccounted in either direction is a Finding, and Findings gate
// through the burn-down baseline (baseline.go) in the pkg/anatomy /
// pkg/secretcoverage grain: new findings fail, fixed-but-still-listed
// baseline entries fail, and the ledger only ever shrinks truthfully.

package providerparity

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// machineryArgPaths are Terraform's own operational surface, present on
// essentially every resource and deliberately outside parity accounting: a
// spec models the RESOURCE, not the tool driving it. One standing exclusion
// here beats the same boilerplate copied into every kind's manifest -- and
// keeps the parity denominator principled (resource configuration only).
//
//   - "id": the Terraform meta-argument legacy SDK resources expose as an
//     optional attribute.
//   - "timeouts.*": per-operation apply/destroy deadlines, an engine
//     concern the platform owns globally.
func isMachineryArg(path string) bool {
	return path == "id" || path == "timeouts" || strings.HasPrefix(path, "timeouts.")
}

// iamResourceRe classifies the provider's per-resource IAM triplets
// (*_iam_member / *_iam_binding / *_iam_policy). The catalog's standing
// design covers this class with additive iam_members fields on the owning
// kinds (authoritative binding/policy forms are deliberately not modeled),
// so the class is dispositioned by pattern, never by 300+ ledger entries.
var iamResourceRe = regexp.MustCompile(`_iam_(member|binding|policy)$`)

// Disposition names for the breadth accounting. Every GA resource carries
// exactly one.
const (
	// DispositionModeled: consumed by at least one kind's Terraform module.
	DispositionModeled = "modeled"
	// DispositionIamCovered: an *_iam_* triplet resource, covered by the
	// owning kind's additive iam_members field.
	DispositionIamCovered = "iam-covered"
	// DispositionExcludedDeprecated: deprecated surface, excluded from
	// parity (schema-flagged automatically; doc-level deprecations enter
	// via the ledger).
	DispositionExcludedDeprecated = "excluded-deprecated"
	// DispositionComposed: covered by fields of an existing kind rather
	// than a kind of its own (ledger-recorded, with the covering kind
	// named in the reason).
	DispositionComposed = "composed"
	// DispositionModelPlanned: judged to be covered by a planned kind that
	// is not built yet -- a kind of its own, or a planned kind it composes
	// into (ledger-recorded, with the planned kind named in the reason).
	// When the kind ships, the entry flips to modeled (computed) or
	// composed, and the staleness findings force the cleanup.
	DispositionModelPlanned = "model-planned"
	// DispositionDeferred: deliberately not offered, with the reason
	// recorded (ledger-recorded).
	DispositionDeferred = "deferred"
)

// ledgerDispositions are the judgments the ledger may record. The other two
// classes (modeled, iam-covered) are computed, never hand-written.
var ledgerDispositions = map[string]bool{
	DispositionComposed:           true,
	DispositionModelPlanned:       true,
	DispositionDeferred:           true,
	DispositionExcludedDeprecated: true,
}

// Finding is one unaccounted or stale item. Findings group into the
// burn-down baseline by BaselineKey, so the baseline reads as a work list
// (one line per kind or resource), not a field dump.
type Finding struct {
	// BaselineKey is "kind:<Kind>" (depth accounting), "resource:<name>"
	// (breadth disposition), "provider:<cloud>" (provider-block
	// accounting), or "admission:<resource>" (a recorded secondary-channel
	// admission that has gone stale).
	BaselineKey string `json:"baselineKey"`
	// Detail names the exact gap and, where possible, the fix.
	Detail string `json:"detail"`
}

// KindAccounting is one kind's depth accounting against the pinned provider.
type KindAccounting struct {
	Kind        string `json:"kind"`
	HasManifest bool   `json:"hasManifest"`
	// MissingModule marks a kind with no Terraform module directory: its
	// depth cannot be measured at all, which is itself the finding -- the
	// per-argument walk is skipped so the one real gap is not buried under
	// a spec-field noise storm.
	MissingModule bool `json:"missingModule,omitempty"`
	// TotalArgs is the surface under accounting: non-deprecated,
	// non-machinery configurable arguments across the kind's consumed
	// resources (internal-dispositioned resources excluded).
	TotalArgs int `json:"totalArgs"`
	// MatchedArgs matched their derived spec path exactly.
	MatchedArgs int `json:"matchedArgs"`
	// MappedArgs matched through a recorded mapping.
	MappedArgs int `json:"mappedArgs"`
	// ExcludedArgs carry a recorded exclusion.
	ExcludedArgs int `json:"excludedArgs"`
	// InternalResources are consumed resources dispositioned as module
	// plumbing by the manifest.
	InternalResources []string `json:"internalResources,omitempty"`
	// ExternalResources are consumed resources dispositioned as externally
	// specified by the manifest: the kind's primary surface whose contract
	// lives outside every loaded provider schema by recorded admission
	// (e.g. a raw-ARM azapi resource at a pinned type@api-version). They
	// carry no argument walk, and a kind consuming ONLY external/internal
	// resources runs no reverse spec walk either.
	ExternalResources []string `json:"externalResources,omitempty"`
	// AdmittedResources are consumed resources served only by a secondary
	// channel's schema (never by the GA baseline) and admitted for this
	// kind by the admission list. They ARE argument-walked -- against the
	// channel's schema -- so the kind's depth is total across both
	// channels; the admission is what makes that walk legitimate rather
	// than an accident of which artifacts happen to be loaded.
	AdmittedResources []string `json:"admittedResources,omitempty"`
	// AdmissionGaps are consumed resources whose channel and admission
	// disagree: a secondary-channel resource with no admission for this
	// kind, an admitted resource the module does not attach through the
	// channel's provider, or a channel attachment on a resource the GA
	// baseline serves. Each is a Finding.
	AdmissionGaps []string `json:"admissionGaps,omitempty"`
	// UnaccountedArgs ("resource: arg") have no match, mapping, or
	// exclusion -- each is a Finding.
	UnaccountedArgs []string `json:"unaccountedArgs,omitempty"`
	// UncoveredSpecFields are spec leaves no argument matched and no
	// exclusion covers -- reverse drift candidates. Each is a Finding.
	UncoveredSpecFields []string `json:"uncoveredSpecFields,omitempty"`
	// ManifestStale are manifest entries referencing surface that no
	// longer exists (after a pin bump or spec change). Each is a Finding.
	ManifestStale []string `json:"manifestStale,omitempty"`
}

// Accounted reports whether the kind is at total accounting.
func (k KindAccounting) Accounted() bool {
	return !k.MissingModule &&
		len(k.UnaccountedArgs) == 0 && len(k.UncoveredSpecFields) == 0 &&
		len(k.ManifestStale) == 0 && len(k.AdmissionGaps) == 0
}

// ResourceDisposition is one GA resource's recorded breadth judgment.
type ResourceDisposition struct {
	Resource    string `json:"resource"`
	Disposition string `json:"disposition,omitempty"`
	// Detail carries the consuming kinds (modeled), the recorded reason
	// (ledger), or the classifying rule (pattern/schema classes).
	Detail string `json:"detail,omitempty"`
}

// Accounting is the total-accounting result for one cloud provider's
// catalog: per-kind depth accounting plus the GA breadth disposition, with
// every gap surfaced as a Finding.
type Accounting struct {
	CloudProvider string `json:"cloudProvider"`
	// GASchema/GASchemaVersion name the parity baseline the accounting ran
	// against. Secondary-channel capability (google-beta) enters per kind
	// through the admission list below; an admitted resource is
	// argument-walked against its channel's schema, everything else against
	// the baseline.
	GASchema        string `json:"gaSchema"`
	GASchemaVersion string `json:"gaSchemaVersion"`
	// Admissions is the recorded secondary-channel admission list this
	// accounting ran with (admissions/<channel>.yaml for this baseline),
	// sorted by resource then kind. Rendered on the public page so every
	// admission and its reason is a reader-visible fact.
	Admissions []AdmissionEntry `json:"admissions,omitempty"`
	// Kinds is the per-kind depth accounting, sorted by kind.
	Kinds []KindAccounting `json:"kinds"`
	// Dispositions covers every GA resource, sorted by name. Resources
	// with no recorded disposition have an empty Disposition and a
	// Finding.
	Dispositions []ResourceDisposition `json:"dispositions"`
	// DispositionTotals counts resources per disposition ("" counts the
	// undispositioned).
	DispositionTotals map[string]int `json:"dispositionTotals"`
	// ProviderConfig is the provider-block accounting -- present exactly when
	// the provider ships a provider-config manifest (enrollment by file
	// presence, the per-kind manifests' convention). Its findings ride the
	// same Findings slice under the "provider:<cloud>" baseline-key class.
	ProviderConfig *ProviderConfigAccounting `json:"providerConfig,omitempty"`
	// Findings is every gap, sorted by baseline key then detail.
	Findings []Finding `json:"findings,omitempty"`
}

// BuildAccounting runs the total-accounting check for one cloud provider's
// catalog: censuses and manifests from the tree, schemas from the committed
// artifacts, the dispositions ledger from dispositionsPath and the admission
// list from admissionsDir (empty strings for the defaults). gaSchema names
// the parity-baseline schema (e.g. "google").
func BuildAccounting(repoRoot string, provider cloudresourcekind.CloudResourceProvider, schemas map[string]*Schema, gaSchema, dispositionsPath, admissionsDir string) (Accounting, error) {
	if _, ok := schemas[gaSchema]; !ok {
		return Accounting{}, errors.Errorf("GA schema %q is not among the loaded schemas", gaSchema)
	}
	spec := SpecCensus(provider)
	modules, err := ModuleCensusForProvider(repoRoot, provider)
	if err != nil {
		return Accounting{}, err
	}
	manifests := map[string]*Manifest{}
	for _, m := range modules {
		kind := crkreflect.KindFromString(m.Kind)
		manifest, err := LoadKindManifest(repoRoot, provider, kind)
		if err != nil {
			return Accounting{}, err
		}
		if manifest != nil {
			manifests[m.Kind] = manifest
		}
	}
	if dispositionsPath == "" {
		dispositionsPath = filepath.Join(repoRoot, DefaultDispositionsDir, gaSchema+".yaml")
	}
	ledger, err := LoadLedger(dispositionsPath, gaSchema)
	if err != nil {
		return Accounting{}, err
	}
	if admissionsDir == "" {
		admissionsDir = filepath.Join(repoRoot, DefaultAdmissionsDir)
	}
	admissions, err := LoadAdmissions(admissionsDir, gaSchema)
	if err != nil {
		return Accounting{}, err
	}
	acc := buildAccounting(crkreflect.ProviderDirName(provider), spec, modules, schemas, gaSchema, manifests, ledger, admissions)

	// Provider-block accounting, enrolled by manifest presence. Composed here
	// (the I/O boundary) rather than inside the pure join so the existing
	// hermetic accounting tests keep their surface; the provider-config join
	// itself is pure and hermetic-testable on its own.
	pcManifest, err := LoadProviderConfigManifest(repoRoot, provider)
	if err != nil {
		return Accounting{}, err
	}
	if pcManifest != nil {
		if schemas[gaSchema].ProviderConfig == nil {
			return Accounting{}, errors.Errorf(
				"provider %s is enrolled for provider-config accounting but schema artifact %s@%s carries no provider block -- regenerate the artifacts",
				acc.CloudProvider, gaSchema, schemas[gaSchema].Version)
		}
		configPaths, err := ProviderConfigCensus(provider)
		if err != nil {
			return Accounting{}, err
		}
		// Every provider block a module configures is judged: the baseline's
		// block always, and an admitted secondary channel's block too (a
		// module that attaches `provider = google-beta` configures that
		// alias -- quota-project attribution, credentials -- and an argument
		// set there is exactly as much provider behavior as one set on the
		// baseline block). Only channels the admission list names count;
		// an unadmitted channel's block is already a per-kind finding.
		judgedAliases := map[string]bool{gaSchema: true}
		for _, a := range admissions {
			judgedAliases[a.Provider] = true
		}
		moduleSetArgs := map[string][]string{}
		for _, m := range modules {
			for alias, args := range m.ProviderBlockArgs {
				if !judgedAliases[alias] {
					continue
				}
				for _, arg := range args {
					moduleSetArgs[arg] = append(moduleSetArgs[arg], m.Kind)
				}
			}
		}
		for arg, kinds := range moduleSetArgs {
			sort.Strings(kinds)
			moduleSetArgs[arg] = slices.Compact(kinds)
		}
		pc, findings := buildProviderConfigAccounting(acc.CloudProvider, configPaths,
			schemas[gaSchema].ProviderConfig, pcManifest, moduleSetArgs)
		acc.ProviderConfig = &pc
		acc.Findings = append(acc.Findings, findings...)
		sort.Slice(acc.Findings, func(i, j int) bool {
			if acc.Findings[i].BaselineKey != acc.Findings[j].BaselineKey {
				return acc.Findings[i].BaselineKey < acc.Findings[j].BaselineKey
			}
			return acc.Findings[i].Detail < acc.Findings[j].Detail
		})
	}
	return acc, nil
}

// buildAccounting is the pure join -- everything I/O-free so the hermetic
// tests can drive every accounting shape without a repo tree.
func buildAccounting(cloudProvider string, spec []KindCensus, modules []ModuleCensus, schemas map[string]*Schema, gaSchema string, manifests map[string]*Manifest, ledger []LedgerEntry, admissions []AdmissionEntry) Accounting {
	schemaNames := make([]string, 0, len(schemas))
	for name := range schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	specPaths := map[string][]string{}
	manifestOnly := map[string][]string{}
	for _, k := range spec {
		specPaths[k.Kind] = k.SpecFieldPaths
		manifestOnly[k.Kind] = k.ManifestOnlyPaths
	}

	acc := Accounting{
		CloudProvider:     cloudProvider,
		GASchema:          gaSchema,
		GASchemaVersion:   schemas[gaSchema].Version,
		Admissions:        admissions,
		DispositionTotals: map[string]int{},
	}

	consumedBy := map[string][]string{}
	for _, m := range modules {
		for _, res := range m.Resources {
			consumedBy[res] = append(consumedBy[res], m.Kind)
		}
	}

	// Admissions indexed per kind (the per-kind walk reads its own), and
	// checked for staleness here: an admission is judgment about a
	// resource's channel at the pin, and the two ways it goes stale are
	// both mechanical -- the baseline now serves the resource (the
	// promotion the entry was waiting for), or the kind no longer consumes
	// it. Either way the entry must go; leaving it would be an exception
	// nobody re-evaluates.
	admittedByKind := map[string]map[string]string{}
	ga := schemas[gaSchema]
	for _, a := range admissions {
		if admittedByKind[a.Kind] == nil {
			admittedByKind[a.Kind] = map[string]string{}
		}
		admittedByKind[a.Kind][a.Resource] = a.Provider
		key := "admission:" + a.Resource
		if _, servedByBaseline := ga.Resources[a.Resource]; servedByBaseline {
			acc.Findings = append(acc.Findings, Finding{key,
				fmt.Sprintf("stale admission: %s is served by the baseline schema %s@%s -- move %s's module to the baseline provider and remove the entry", a.Resource, gaSchema, ga.Version, a.Kind)})
		}
		if channel, ok := schemas[a.Provider]; !ok {
			acc.Findings = append(acc.Findings, Finding{key,
				fmt.Sprintf("admission names channel %q, which has no loaded schema artifact -- distill it or fix the admissions file", a.Provider)})
		} else if _, servedByChannel := channel.Resources[a.Resource]; !servedByChannel {
			acc.Findings = append(acc.Findings, Finding{key,
				fmt.Sprintf("stale admission: %s@%s does not serve %s -- the resource was removed or renamed at the pin", a.Provider, channel.Version, a.Resource)})
		}
		if !slices.Contains(consumedBy[a.Resource], a.Kind) {
			acc.Findings = append(acc.Findings, Finding{key,
				fmt.Sprintf("stale admission: %s's module does not consume %s -- remove the entry (an admission exists only for a module that attaches the resource)", a.Kind, a.Resource)})
		}
	}

	for _, m := range modules {
		if m.MissingModule {
			acc.Kinds = append(acc.Kinds, KindAccounting{Kind: m.Kind, MissingModule: true})
			acc.Findings = append(acc.Findings, Finding{"kind:" + m.Kind,
				fmt.Sprintf("no Terraform module directory at %s -- forge the module, or record the debt in the anatomy baseline and here", m.ModuleDir)})
			continue
		}
		manifest := manifests[m.Kind]
		if manifest == nil {
			manifest = &Manifest{}
		}
		ka := accountKind(m, specPaths[m.Kind], manifestOnly[m.Kind], schemas, schemaNames, gaSchema, manifests[m.Kind] != nil, manifest, admittedByKind[m.Kind])
		acc.Kinds = append(acc.Kinds, ka)
		key := "kind:" + m.Kind
		for _, gap := range ka.AdmissionGaps {
			acc.Findings = append(acc.Findings, Finding{key, gap})
		}
		for _, arg := range ka.UnaccountedArgs {
			acc.Findings = append(acc.Findings, Finding{key,
				fmt.Sprintf("unaccounted provider argument %s -- match it, map it, or exclude it with a reason in the kind's %s", arg, ManifestFileName)})
		}
		for _, field := range ka.UncoveredSpecFields {
			acc.Findings = append(acc.Findings, Finding{key,
				fmt.Sprintf("spec field %s reaches no provider argument -- reverse drift, a missing mapping, or a platform field needing a specExclusions entry", field)})
		}
		for _, stale := range ka.ManifestStale {
			acc.Findings = append(acc.Findings, Finding{key, stale})
		}
	}

	// Breadth: every GA resource carries exactly one disposition.
	ledgerByResource := map[string]LedgerEntry{}
	for _, e := range ledger {
		ledgerByResource[e.Resource] = e
	}
	gaNames := make([]string, 0, len(ga.Resources))
	for name := range ga.Resources {
		gaNames = append(gaNames, name)
	}
	sort.Strings(gaNames)
	// Computed classes always win over the ledger, and a ledger entry
	// shadowed by a computed class is a staleness finding: recorded judgment
	// that duplicates what the instrument derives on its own is judgment
	// nobody re-evaluates -- exactly the rot class this package exists to
	// eliminate.
	for _, name := range gaNames {
		d := ResourceDisposition{Resource: name}
		entry, inLedger := ledgerByResource[name]
		switch {
		case len(consumedBy[name]) > 0:
			d.Disposition = DispositionModeled
			kinds := append([]string(nil), consumedBy[name]...)
			sort.Strings(kinds)
			d.Detail = "consumed by " + strings.Join(kinds, ", ")
			if inLedger {
				acc.Findings = append(acc.Findings, Finding{"resource:" + name,
					"stale ledger entry: the resource is consumed by a module (modeled) -- remove it from the ledger"})
			}
		case iamResourceRe.MatchString(name):
			d.Disposition = DispositionIamCovered
			d.Detail = "per-resource IAM triplet, covered by the owning kind's additive iam_members field"
			if inLedger {
				acc.Findings = append(acc.Findings, Finding{"resource:" + name,
					"stale ledger entry: the IAM triplet is covered by the computed iam-covered class -- remove it from the ledger"})
			}
		case ga.Resources[name].Deprecated:
			d.Disposition = DispositionExcludedDeprecated
			d.Detail = "deprecated in the provider schema"
			if inLedger {
				acc.Findings = append(acc.Findings, Finding{"resource:" + name,
					"stale ledger entry: the schema already flags the resource deprecated (computed) -- the ledger's excluded-deprecated is only for doc-level deprecations; remove it"})
			}
		case inLedger:
			d.Disposition = entry.Disposition
			d.Detail = entry.Reason
		default:
			acc.Findings = append(acc.Findings, Finding{"resource:" + name,
				"GA resource has no recorded disposition -- consume it, or record composed/model-planned/deferred/excluded-deprecated in the dispositions ledger"})
		}
		acc.DispositionTotals[d.Disposition]++
		acc.Dispositions = append(acc.Dispositions, d)
	}
	for _, e := range ledger {
		if _, ok := ga.Resources[e.Resource]; !ok {
			acc.Findings = append(acc.Findings, Finding{"resource:" + e.Resource,
				fmt.Sprintf("stale ledger entry: no resource %s in %s@%s -- it was removed or renamed at the pin", e.Resource, gaSchema, ga.Version)})
		}
	}

	sort.Slice(acc.Findings, func(i, j int) bool {
		if acc.Findings[i].BaselineKey != acc.Findings[j].BaselineKey {
			return acc.Findings[i].BaselineKey < acc.Findings[j].BaselineKey
		}
		return acc.Findings[i].Detail < acc.Findings[j].Detail
	})
	return acc
}

// argMatcher applies one resource's recorded judgment: exclusions first
// (exact path or subtree prefix — one judgment covers an excluded block and
// everything under it, mirroring mappings and specExclusions), then the
// longest recorded mapping prefix, then the default spec root. No name
// heuristics, by design.
type argMatcher struct {
	specRoot   string
	mappings   []Mapping // sorted by arg length, longest first
	exclusions []string  // arg paths; each excludes itself and its subtree
}

func newArgMatcher(rm *ResourceManifest) argMatcher {
	m := argMatcher{specRoot: specPathRoot}
	if rm == nil {
		return m
	}
	if rm.SpecRoot != "" {
		m.specRoot = rm.SpecRoot
	}
	m.mappings = append(m.mappings, rm.Mappings...)
	sort.Slice(m.mappings, func(i, j int) bool { return len(m.mappings[i].Arg) > len(m.mappings[j].Arg) })
	for _, ex := range rm.Exclusions {
		m.exclusions = append(m.exclusions, ex.Arg)
	}
	return m
}

// excluded reports whether an argument is covered by a recorded exclusion,
// either exactly or as a descendant of an excluded subtree. Blocks never
// appear as arguments themselves (the census walks leaves), so a
// block-naming exclusion is meaningful ONLY through this prefix form.
func (m argMatcher) excluded(argPath string) bool {
	for _, ex := range m.exclusions {
		if argPath == ex || strings.HasPrefix(argPath, ex+".") {
			return true
		}
	}
	return false
}

// derive returns the spec path(s) an argument must match, and whether a
// recorded mapping produced them. An argument normally derives to exactly
// one path; a FAN-IN argument (several mappings recording the same arg —
// e.g. a map-typed limits argument realizing several honest spec fields)
// derives to every mapped path. Only the longest matching arg prefix
// contributes, preserving the longest-mapping-wins contract.
func (m argMatcher) derive(argPath string) ([]string, bool) {
	var specs []string
	matchLen := -1
	for _, mp := range m.mappings { // sorted by arg length, longest first
		if matchLen >= 0 && len(mp.Arg) < matchLen {
			break
		}
		if argPath == mp.Arg {
			specs = append(specs, mp.Spec)
			matchLen = len(mp.Arg)
		} else if strings.HasPrefix(argPath, mp.Arg+".") {
			if mp.Collapse {
				// Subtree-to-leaf judgment: everything under the arg
				// subtree is the one spec leaf (recursive grammars).
				specs = append(specs, mp.Spec)
			} else {
				specs = append(specs, mp.Spec+argPath[len(mp.Arg):])
			}
			matchLen = len(mp.Arg)
		}
	}
	if len(specs) > 0 {
		return specs, true
	}
	return []string{m.specRoot + "." + argPath}, false
}

// resolveResourceSchema names the loaded schema that serves a resource type:
// the baseline first when one is named (GA is the parity yardstick, so a
// resource both channels serve is always accounted against GA), then every
// other loaded schema in sorted name order. Empty means no schema knows the
// resource. Shared by the accounting and the report so the two never
// disagree about which artifact a resource was measured against.
func resolveResourceSchema(schemas map[string]*Schema, schemaNames []string, baseline, resource string) (*Block, string) {
	if baseline != "" {
		if b, ok := schemas[baseline].Resources[resource]; ok {
			return b, baseline
		}
	}
	for _, name := range schemaNames {
		if name == baseline {
			continue
		}
		if b, ok := schemas[name].Resources[resource]; ok {
			return b, name
		}
	}
	return nil, ""
}

// accountKind runs both accounting directions for one kind. manifestOnlyPaths
// are the spec leaves the proto marks (dev.planton.shared.options.manifest_only);
// the reverse direction reads each as an exclusion the schema itself declares.
// admitted maps the secondary-channel resources the admission list admits
// for THIS kind to the channel (schema name) each is admitted from.
func accountKind(m ModuleCensus, kindSpecPaths, manifestOnlyPaths []string, schemas map[string]*Schema, schemaNames []string, gaSchema string, hasManifest bool, manifest *Manifest, admitted map[string]string) KindAccounting {
	ka := KindAccounting{Kind: m.Kind, HasManifest: hasManifest}
	specSet := map[string]bool{}
	for _, p := range kindSpecPaths {
		specSet[p] = true
	}
	underPath := func(paths []string, prefix string) bool {
		for _, p := range paths {
			if p == prefix || strings.HasPrefix(p, prefix+".") {
				return true
			}
		}
		return false
	}

	coveredSpec := map[string]bool{}
	consumed := map[string]bool{}
	// schemaWalked flips when at least one consumed resource is argument-
	// walked against a loaded schema; it decides whether the reverse spec
	// walk below has any surface to check against.
	schemaWalked := false
	for _, res := range m.Resources {
		consumed[res] = true

		// An internal judgment is the whole judgment -- module plumbing
		// with no arguments to account. Checked BEFORE the schema lookup
		// so utility-provider resources (e.g. hashicorp/time's time_sleep,
		// which no cloud-provider schema knows) can be judged internal
		// instead of tripping the stale-artifact guard.
		rm := manifest.Resources[res]
		if rm != nil && rm.Internal != "" {
			ka.InternalResources = append(ka.InternalResources, res)
			continue
		}

		block, servedBy := resolveResourceSchema(schemas, schemaNames, gaSchema, res)

		// Channel discipline. A resource the baseline serves is accounted
		// against the baseline, full stop -- an explicit attachment to a
		// secondary channel there is a module reaching for beta it does not
		// need. A resource only a secondary channel serves is accounted
		// against that channel exactly when the admission list admits it
		// for this kind AND the module attaches it through that channel's
		// provider (the attachment is what makes the module actually
		// deploy through the channel; the admission is what makes that a
		// recorded decision). Any other combination is a gap, and a gap
		// still walks the resource's arguments so one finding never hides
		// the depth accounting behind it.
		attachment := m.ProviderAttachments[res]
		switch {
		case block == nil:
			// Unknown to every schema: judged below (internal/external) or a
			// stale-artifact finding.
		case servedBy == gaSchema:
			if attachment != "" && attachment != gaSchema {
				ka.AdmissionGaps = append(ka.AdmissionGaps,
					fmt.Sprintf("%s attaches through provider %q but the baseline schema %s serves it at the pin -- drop the attachment and use the baseline provider", res, attachment, gaSchema))
			}
		default:
			channel, isAdmitted := admitted[res]
			switch {
			case !isAdmitted:
				ka.AdmissionGaps = append(ka.AdmissionGaps,
					fmt.Sprintf("%s is served only by the %s schema and carries no admission for this kind -- record it in %s/%s.yaml (resource, kind, reason, promotionTracking) or model it without the secondary channel", res, servedBy, DefaultAdmissionsDir, servedBy))
			case channel != servedBy:
				ka.AdmissionGaps = append(ka.AdmissionGaps,
					fmt.Sprintf("%s is admitted from channel %q but the %s schema is what serves it -- fix the admission's channel", res, channel, servedBy))
			case attachment != servedBy:
				ka.AdmissionGaps = append(ka.AdmissionGaps,
					fmt.Sprintf("%s is admitted from %s but the module attaches it through %q -- set `provider = %s` on the resource block so the module deploys through the admitted channel", res, servedBy, attachment, servedBy))
			default:
				ka.AdmittedResources = append(ka.AdmittedResources, res)
			}
		}

		if rm != nil && rm.External != "" {
			if block != nil {
				// The judgment claims the resource lives outside the loaded
				// schemas, but a schema serves it at the pin -- the external
				// disposition is stale (the provider shipped native support);
				// account the resource against the schema instead.
				ka.ManifestStale = append(ka.ManifestStale,
					fmt.Sprintf("%s: external judgment is stale -- the resource is served by a loaded schema at the pin; account it there", res))
				continue
			}
			ka.ExternalResources = append(ka.ExternalResources, res)
			continue
		}
		if block == nil {
			ka.ManifestStale = append(ka.ManifestStale,
				fmt.Sprintf("consumed resource %s is unknown to every loaded schema -- the module outruns its pin or the artifacts are stale", res))
			continue
		}
		schemaWalked = true
		matcher := newArgMatcher(rm)

		argPaths := map[string]bool{}
		for _, arg := range block.ConfigurableArgs("") {
			if arg.Deprecated || isMachineryArg(arg.Path) {
				continue
			}
			argPaths[arg.Path] = true
			ka.TotalArgs++
			if matcher.excluded(arg.Path) {
				ka.ExcludedArgs++
				continue
			}
			derived, mapped := matcher.derive(arg.Path)
			// A fan-in argument covers every mapped spec field that
			// exists; per-entry manifest hygiene below still flags any
			// mapping whose spec side has gone stale.
			matchedAny := false
			for _, d := range derived {
				if specSet[d] {
					coveredSpec[d] = true
					matchedAny = true
				}
			}
			if matchedAny {
				if mapped {
					ka.MappedArgs++
				} else {
					ka.MatchedArgs++
				}
				continue
			}
			ka.UnaccountedArgs = append(ka.UnaccountedArgs, res+": "+arg.Path)
		}

		if rm == nil {
			continue
		}
		// Manifest hygiene: judgment referencing surface that no longer
		// exists is a finding, not a warning -- after a pin bump these ARE
		// the migration work list.
		sortedArgs := make([]string, 0, len(argPaths))
		for p := range argPaths {
			sortedArgs = append(sortedArgs, p)
		}
		sort.Strings(sortedArgs)
		if rm.SpecRoot != "" && !underPath(kindSpecPaths, rm.SpecRoot) {
			ka.ManifestStale = append(ka.ManifestStale,
				fmt.Sprintf("%s: specRoot %s matches no spec field", res, rm.SpecRoot))
		}
		for _, mp := range rm.Mappings {
			if !underPath(sortedArgs, mp.Arg) {
				ka.ManifestStale = append(ka.ManifestStale,
					fmt.Sprintf("%s: mapping arg %s matches no configurable argument at the pin", res, mp.Arg))
			}
			if !underPath(kindSpecPaths, mp.Spec) {
				ka.ManifestStale = append(ka.ManifestStale,
					fmt.Sprintf("%s: mapping spec %s matches no spec field", res, mp.Spec))
			}
			// A collapse mapping folds a whole arg subtree onto ONE leaf, so
			// its spec side must BE a census leaf — a subtree would make the
			// fold ambiguous and overclaim silently.
			if mp.Collapse && !slices.Contains(kindSpecPaths, mp.Spec) {
				ka.ManifestStale = append(ka.ManifestStale,
					fmt.Sprintf("%s: collapse mapping spec %s must name a spec census leaf, not a subtree", res, mp.Spec))
			}
		}
		for _, ex := range rm.Exclusions {
			if !underPath(sortedArgs, ex.Arg) {
				ka.ManifestStale = append(ka.ManifestStale,
					fmt.Sprintf("%s: exclusion %s matches no configurable argument at the pin -- remove it", res, ex.Arg))
			}
		}
	}

	for res := range manifest.Resources {
		if !consumed[res] {
			ka.ManifestStale = append(ka.ManifestStale,
				fmt.Sprintf("manifest judges resource %s, which the module no longer consumes", res))
		}
	}

	// Reverse direction: every spec leaf reaches provider surface or
	// carries a recorded exclusion. When the kind's ENTIRE consumed surface
	// is externally specified (plus any internal plumbing), no loaded
	// schema serves its spec fields and the walk is suspended -- the spec's
	// depth is accounted against the external contract the manifest's
	// external judgment names. A kind that mixes schema-served and external
	// resources keeps the walk: its schema-side fields still must reach
	// provider surface, and external-fed fields need specExclusions naming
	// the external resource.
	//
	// A leaf the proto marks manifest_only is excluded by the schema itself:
	// the field exists so the manifest can say what the wire spells by the
	// absence of its siblings, and no engine forwards it. The manifest never
	// repeats that fact -- a specExclusions entry for such a leaf is stale.
	manifestOnly := map[string]bool{}
	for _, p := range manifestOnlyPaths {
		manifestOnly[p] = true
	}
	specExcluded := func(path string) bool {
		if manifestOnly[path] {
			return true
		}
		for _, ex := range manifest.SpecExclusions {
			if path == ex.Field || strings.HasPrefix(path, ex.Field+".") {
				return true
			}
		}
		return false
	}
	if schemaWalked || len(ka.ExternalResources) == 0 {
		for _, p := range kindSpecPaths {
			if !coveredSpec[p] && !specExcluded(p) {
				ka.UncoveredSpecFields = append(ka.UncoveredSpecFields, p)
			}
		}
	}
	for _, ex := range manifest.SpecExclusions {
		switch {
		case !underPath(kindSpecPaths, ex.Field):
			ka.ManifestStale = append(ka.ManifestStale,
				fmt.Sprintf("specExclusions: %s matches no spec field", ex.Field))
		case manifestOnly[ex.Field]:
			ka.ManifestStale = append(ka.ManifestStale,
				fmt.Sprintf("specExclusions: %s is manifest_only in the proto, which already excludes it -- remove the entry", ex.Field))
		}
	}

	sort.Strings(ka.InternalResources)
	sort.Strings(ka.ExternalResources)
	sort.Strings(ka.AdmittedResources)
	sort.Strings(ka.AdmissionGaps)
	sort.Strings(ka.UnaccountedArgs)
	sort.Strings(ka.UncoveredSpecFields)
	sort.Strings(ka.ManifestStale)
	return ka
}
