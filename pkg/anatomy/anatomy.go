// Package anatomy is the machine-enforced constitution of the catalog's
// shape: every component folder follows ONE canonical anatomy, checked file
// by file, so anatomy drift is unshippable instead of auditable.
//
// The canonical anatomy (one rule: version dirs hold only the versioned
// contract; the component root holds the living component):
//
//	catalog/<provider>/<kind>/
//	├── README.md          the GitHub-facing component page       (required)
//	├── catalog.md         the catalog page                       (required)
//	├── logo.svg           the component logo                     (required)
//	├── GUIDE.md           authored operational judgment          (optional)
//	├── cost.yaml          the component's cost profile           (required)
//	├── controls.yaml      the component's control profile        (required)
//	├── iac/               ONE live module set per component      (required)
//	│   ├── pulumi/        with README.md, no Makefile            (required*)
//	│   ├── tf/            with README.md, no .gitignore          (required*)
//	│   ├── permissions.yaml   runner least-privilege manifest    (required)
//	│   ├── import-map.yaml                                       (optional)
//	│   └── provider-parity.yaml   recorded parity judgment       (optional)
//	├── presets/           .yaml manifests + .md sidecar pairs    (required)
//	├── e2e/               test manifest, profile, scenarios      (optional, tiered)
//	├── conversions/       cross-version conversion specs         (optional)
//	└── <version>/         the versioned contract ONLY:
//	    api.proto, spec.proto, input.proto, outputs.proto,
//	    their .pb.go stubs, BUILD.bazel, spec_test.go, reference.md
//
// * Both engine modules are required unless the kind's kind_meta.provisioners
// declares fewer engines; then the tree holds exactly the modules those engines
// run (pulumi/ for pulumi, tf/ for tofu or terraform) and nothing else, so a
// module for an engine the kind refuses can never ship.
//
// Two prefix conventions coexist deliberately: underscore dirs at the catalog
// root (_docs/, _patterns/, _compliance/, _pricing/) hold non-component
// content Go tooling must ignore (_compliance/ carries the authored control
// catalog and framework crosswalks the per-component controls.yaml files
// reference; _pricing/ carries the per-preset cost estimates priced from the
// components' cost.yaml profiles at published list prices),
// while aa_-prefixed dirs inside providers (aa_e2e/, aa_eval/, aa_import/)
// hold provider infrastructure that CONTAINS buildable Go -- Go tooling skips
// underscore dirs entirely, so Go-bearing infrastructure cannot use one.
// (_test is underscore-prefixed AND a registered provider; walkers key off
// the registry, never the prefix.)
//
// The walk is keyed off the kind registry (crkreflect), never directory
// prefixes: a directory that does not resolve to a registered kind is a
// violation, and a registered kind without a directory is one too.
//
// The descriptor of accepted gaps lives in baseline.yaml -- a burn-down list
// mirroring pkg/secretcoverage's baseline (a reader who knows one knows
// both). The CI lane is .github/workflows/lint.component-anatomy.yaml.
package anatomy

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// versionDirRe is the maturity-channel grammar. Never a bare v* glob: that
// would match non-version names (the aa_e2e/verify lesson).
var versionDirRe = regexp.MustCompile(`^v[0-9]+((alpha|beta)[0-9]+)?$`)

// Violation is one anatomy rule broken at one path.
type Violation struct {
	// Path is repo-root-relative (a component dir, or a file inside one).
	Path string
	// Rule is the stable rule identifier (baseline entries key on it).
	Rule string
	// Detail says what the fix is, in plain language.
	Detail string
}

// ID is the stable baseline identifier: "<path>:<rule>".
func (v Violation) ID() string { return v.Path + ":" + v.Rule }

// Rule identifiers. Stable: baseline.yaml entries reference them.
const (
	RuleUnregisteredDir       = "unregistered-component-dir"
	RuleMissingComponent      = "missing-component-dir"
	RuleUnexpectedEntry       = "unexpected-entry"
	RuleMissingReadme         = "missing-readme"
	RuleMissingCatalogPage    = "missing-catalog-md"
	RuleMissingLogo           = "missing-logo"
	RuleMissingIac            = "missing-iac"
	RuleMissingPulumi         = "missing-pulumi-module"
	RuleMissingTf             = "missing-tf-module"
	RuleUndeclaredEngine      = "module-for-undeclared-engine"
	RuleMissingIacReadme      = "missing-iac-readme"
	RuleForbiddenFile         = "forbidden-file"
	RuleMissingPresets        = "missing-presets"
	RuleMissingSidecar        = "missing-preset-sidecar"
	RuleMissingProto          = "missing-proto"
	RuleMissingStub           = "missing-stub"
	RuleMissingSpecTest       = "missing-spec-test"
	RuleMissingReference      = "missing-reference"
	RuleMissingCostProfile    = "missing-cost-profile"
	RuleMissingControlProfile = "missing-control-profile"
	RuleMissingPermissions    = "missing-permissions"
)

// componentEntries is the CLOSED set of names allowed at a component root
// (plus version dirs, matched by grammar). An entry outside this set is the
// exact drift class this gate exists to catch.
var componentEntries = map[string]bool{
	"README.md":     true,
	"catalog.md":    true,
	"GUIDE.md":      true,
	"logo.svg":      true,
	"cost.yaml":     true,
	"controls.yaml": true,
	"iac":           true,
	"presets":       true,
	"e2e":           true,
	"conversions":   true,
}

// versionEntries is the CLOSED set of names allowed inside a version dir:
// the versioned contract, nothing living.
var versionEntries = map[string]bool{
	"api.proto":     true,
	"api.pb.go":     true,
	"spec.proto":    true,
	"spec.pb.go":    true,
	"input.proto":   true,
	"input.pb.go":   true,
	"outputs.proto": true,
	"outputs.pb.go": true,
	"BUILD.bazel":   true,
	"spec_test.go":  true,
	"reference.md":  true,
}

// requiredContractProtos are the version-dir protos every kind must serve.
// input/outputs carry the KEPT {Kind}StackInput / {Kind}StackOutputs
// messages -- the filenames are layout, the message names are identity.
var requiredContractProtos = []string{
	"api.proto", "spec.proto", "input.proto", "outputs.proto",
}

// e2eDirs are the directories the e2e tier may carry (files there are
// free-form manifests: manifest.yaml, variants, loose fixtures).
var e2eDirs = map[string]bool{"scenarios": true, "fixtures": true, "prerequisites": true}

// Check walks the catalog at repoRoot and returns every anatomy violation,
// sorted by ID. It never consults the baseline -- Gate does the comparison.
func Check(repoRoot string) ([]Violation, error) {
	catalogDir := filepath.Join(repoRoot, "catalog")
	var vs []Violation
	add := func(rel, rule, detail string) {
		vs = append(vs, Violation{Path: rel, Rule: rule, Detail: detail})
	}

	// Every registered kind must be found on disk (completeness half).
	unseenKinds := map[cloudresourcekind.CloudResourceKind]bool{}
	for _, k := range crkreflect.KindsList() {
		if k != cloudresourcekind.CloudResourceKind_unspecified {
			unseenKinds[k] = true
		}
	}

	providers, err := os.ReadDir(catalogDir)
	if err != nil {
		return nil, err
	}
	for _, p := range providers {
		if p.Name() == "_docs" || p.Name() == "_patterns" || p.Name() == "_compliance" || p.Name() == "_pricing" {
			continue // non-component homes at the catalog root
		}
		if !p.IsDir() {
			add(filepath.Join("catalog", p.Name()), RuleUnexpectedEntry,
				"the catalog root holds only provider directories (and _docs/, _patterns/, _compliance/, _pricing/)")
			continue
		}
		providerRel := filepath.Join("catalog", p.Name())
		entries, err := os.ReadDir(filepath.Join(repoRoot, providerRel))
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			rel := filepath.Join(providerRel, e.Name())
			if !e.IsDir() {
				// Provider support files: protos + stubs, cli help, docs,
				// build files, and the provider-block parity manifest
				// (provider-config-parity.yaml -- its presence enrolls the
				// provider for provider-block accounting; see
				// pkg/providerparity). Anything else does not belong here.
				switch {
				case e.Name() == "BUILD.bazel",
					e.Name() == "provider-config-parity.yaml",
					strings.HasSuffix(e.Name(), ".proto"),
					strings.HasSuffix(e.Name(), ".go"),
					strings.HasSuffix(e.Name(), ".md"):
				default:
					add(rel, RuleUnexpectedEntry, "not a provider support file (proto/go/md/BUILD.bazel)")
				}
				continue
			}
			if strings.HasPrefix(e.Name(), "aa_") {
				continue // declared provider infrastructure (Go-bearing, so not underscore-prefixed)
			}
			kind := crkreflect.KindFromString(e.Name())
			if kind == cloudresourcekind.CloudResourceKind_unspecified {
				add(rel, RuleUnregisteredDir,
					"directory does not resolve to a registered kind -- typo'd, renamed, or missing its registry entry")
				continue
			}
			delete(unseenKinds, kind)
			checkComponent(repoRoot, rel, kind, add)
		}
	}

	var missing []string
	for k := range unseenKinds {
		missing = append(missing, k.String())
	}
	sort.Strings(missing)
	for _, name := range missing {
		add(name, RuleMissingComponent, "registered kind has no component directory in the catalog")
	}

	sort.Slice(vs, func(i, j int) bool { return vs[i].ID() < vs[j].ID() })
	return vs, nil
}

func checkComponent(repoRoot, componentRel string, kind cloudresourcekind.CloudResourceKind, add func(rel, rule, detail string)) {
	dir := filepath.Join(repoRoot, componentRel)
	entries, _ := os.ReadDir(dir)

	names := map[string]bool{}
	versionDirs := 0
	for _, e := range entries {
		names[e.Name()] = true
		switch {
		case e.IsDir() && versionDirRe.MatchString(e.Name()):
			versionDirs++
			checkVersionDir(repoRoot, filepath.Join(componentRel, e.Name()), add)
		case componentEntries[e.Name()]:
			// allowed living-component entry
		default:
			add(filepath.Join(componentRel, e.Name()), RuleUnexpectedEntry,
				"not part of the component anatomy -- the living component holds only its declared classes")
		}
	}

	if versionDirs == 0 {
		add(componentRel, RuleMissingProto, "component has no version directory serving its contract")
	}
	for name, rule := range map[string]string{
		"README.md":     RuleMissingReadme,
		"catalog.md":    RuleMissingCatalogPage,
		"logo.svg":      RuleMissingLogo,
		"cost.yaml":     RuleMissingCostProfile,
		"controls.yaml": RuleMissingControlProfile,
	} {
		if !names[name] {
			add(componentRel, rule, "required at the component root")
		}
	}

	// iac/: one live module set -- both engines, or exactly the modules the
	// kind's declared engines run -- each with a README, no build-system or
	// VCS residue.
	if !names["iac"] {
		add(componentRel, RuleMissingIac, "every component ships its module set at the root")
	} else {
		iacRel := filepath.Join(componentRel, "iac")
		iacEntries, _ := os.ReadDir(filepath.Join(repoRoot, iacRel))
		hasPermissions := false
		for _, e := range iacEntries {
			switch e.Name() {
			case "permissions.yaml":
				hasPermissions = true
			case "pulumi", "tf", "import-map.yaml", "provider-parity.yaml":
			default:
				// A module is its engine directory and reads nothing beside it:
				// releases zip exactly iac/tf or iac/pulumi, and the Pulumi
				// binary lane runs from a generated workspace with no files at
				// all. Anything a module applies (CRDs above all) is derived
				// from the pinned artifact at apply time, never staged here.
				add(filepath.Join(iacRel, e.Name()), RuleUnexpectedEntry,
					"iac/ holds exactly pulumi/, tf/, permissions.yaml, optionally import-map.yaml and provider-parity.yaml; a module derives what it applies and stages nothing beside itself")
			}
		}
		if !hasPermissions {
			add(iacRel, RuleMissingPermissions,
				"every module set declares the runner permissions it needs (iac/permissions.yaml)")
		}
		wanted, declared := moduleFamilies(kind)
		for _, engine := range []string{"pulumi", "tf"} {
			engineRel := filepath.Join(iacRel, engine)
			engineDir := filepath.Join(repoRoot, engineRel)
			_, statErr := os.Stat(engineDir)
			present := statErr == nil
			switch {
			case !present && wanted[engine]:
				rule := RuleMissingPulumi
				if engine == "tf" {
					rule = RuleMissingTf
				}
				detail := "one live module set per component means both engines"
				if declared != "" {
					detail = kind.String() + " runs on " + declared + ", which needs iac/" + engine
				}
				add(componentRel, rule, detail)
				continue
			case present && !wanted[engine]:
				add(engineRel, RuleUndeclaredEngine,
					kind.String()+" runs on "+declared+" only (kind_meta.provisioners), so iac/"+engine+
						" would ship a module for an engine the kind refuses -- delete it, or declare the engine")
				continue
			case !present:
				continue
			}
			if _, err := os.Stat(filepath.Join(engineDir, "README.md")); err != nil {
				add(engineRel, RuleMissingIacReadme, "module READMEs ship in the console catalog and the iac-source zip")
			}
			for _, forbidden := range []string{"Makefile", ".gitignore"} {
				if _, err := os.Stat(filepath.Join(engineDir, forbidden)); err == nil {
					add(filepath.Join(engineRel, forbidden), RuleForbiddenFile,
						"nothing consumes it -- the class was deleted with the anatomy redesign")
				}
			}
		}
	}

	// presets/: complete manifests with their load-bearing .md sidecars
	// (the site silently drops a preset without one).
	if !names["presets"] {
		add(componentRel, RuleMissingPresets, "every component ships presets at the root")
	} else {
		presetsRel := filepath.Join(componentRel, "presets")
		presetEntries, _ := os.ReadDir(filepath.Join(repoRoot, presetsRel))
		for _, e := range presetEntries {
			switch {
			case strings.HasSuffix(e.Name(), ".md"):
			case strings.HasSuffix(e.Name(), ".yaml"):
				sidecar := strings.TrimSuffix(e.Name(), ".yaml") + ".md"
				if _, err := os.Stat(filepath.Join(repoRoot, presetsRel, sidecar)); err != nil {
					add(filepath.Join(presetsRel, e.Name()), RuleMissingSidecar,
						"a preset without its .md sidecar is silently dropped from the site")
				}
			default:
				add(filepath.Join(presetsRel, e.Name()), RuleUnexpectedEntry,
					"presets/ holds .yaml manifests and their .md sidecars")
			}
		}
	}

	// e2e/: declared-optional (tiered), but its shape is still closed at the
	// directory level: free-form manifest yamls plus the three known dirs.
	if names["e2e"] {
		e2eRel := filepath.Join(componentRel, "e2e")
		e2eEntries, _ := os.ReadDir(filepath.Join(repoRoot, e2eRel))
		for _, e := range e2eEntries {
			switch {
			case e.IsDir() && e2eDirs[e.Name()]:
			case !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml"):
			default:
				add(filepath.Join(e2eRel, e.Name()), RuleUnexpectedEntry,
					"e2e/ holds manifest yamls plus scenarios/, fixtures/, prerequisites/")
			}
		}
	}
}

func checkVersionDir(repoRoot, versionRel string, add func(rel, rule, detail string)) {
	dir := filepath.Join(repoRoot, versionRel)
	entries, _ := os.ReadDir(dir)
	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name()] = true
		if !versionEntries[e.Name()] {
			add(filepath.Join(versionRel, e.Name()), RuleUnexpectedEntry,
				"version dirs hold only the versioned contract -- living-component content belongs at the component root")
		}
	}
	for _, proto := range requiredContractProtos {
		if !names[proto] {
			add(versionRel, RuleMissingProto, fmt.Sprintf("versioned contract requires %s", proto))
			continue
		}
		stub := strings.TrimSuffix(proto, ".proto") + ".pb.go"
		if !names[stub] {
			// A proto without its regenerated stub is the silent-staleness
			// class the build doc warns about.
			add(versionRel, RuleMissingStub, fmt.Sprintf("%s has no regenerated %s", proto, stub))
		}
	}
	if !names["spec_test.go"] {
		add(versionRel, RuleMissingSpecTest, "every served contract carries its spec test")
	}
	if !names["reference.md"] {
		add(versionRel, RuleMissingReference, "every served contract carries its generated reference page")
	}
}

// moduleFamilies answers which module directories a kind's iac/ must hold:
// both, for a kind that declares no provisioners; otherwise pulumi/ exactly
// when it declares pulumi and tf/ exactly when it declares tofu or terraform
// (one HCL module serves both). declared names the engines for messages, and
// is empty for an undeclared kind. A declaration naming no real engine is the
// registry tests' finding; here it reads as undeclared.
func moduleFamilies(kind cloudresourcekind.CloudResourceKind) (wanted map[string]bool, declared string) {
	provisioners, err := crkreflect.Provisioners(kind)
	if err != nil || len(provisioners) == 0 {
		return map[string]bool{"pulumi": true, "tf": true}, ""
	}
	wanted = map[string]bool{}
	names := make([]string, 0, len(provisioners))
	for _, p := range provisioners {
		names = append(names, p.String())
		switch p {
		case shared.IacProvisioner_pulumi:
			wanted["pulumi"] = true
		case shared.IacProvisioner_tofu, shared.IacProvisioner_terraform:
			wanted["tf"] = true
		}
	}
	return wanted, strings.Join(names, " and ")
}
