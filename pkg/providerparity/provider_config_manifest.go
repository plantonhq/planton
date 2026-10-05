//go:build !codegen
// +build !codegen

// Provider-level mapping manifest: the recorded-judgment side of
// PROVIDER-BLOCK parity, the per-kind manifests' sibling one level up. One
// file per cloud provider at catalog/<provider>/provider-config-parity.yaml,
// and file presence IS enrollment (the per-kind manifests' convention): a
// provider with a manifest is held to total provider-block accounting; a
// provider without one is simply not measured on this axis yet.
//
// The judgment vocabulary mirrors the per-kind ResourceManifest where the
// concepts coincide (mappings with collapse, exclusions with mandatory
// reasons) and adds ONE class the provider block uniquely needs: moduleOwned
// -- a provider-block argument set INSIDE catalog modules' own provider
// blocks by recorded judgment (e.g. a behavior flag every module of a cloud
// must carry). The HCL module census detects such arguments mechanically;
// this class is where their judgment lives, so a module-set provider
// argument is always a decision on record, never an invisible leak.

package providerparity

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
	"gopkg.in/yaml.v3"
)

// ProviderConfigManifestFileName is the provider-level manifest, at the
// provider's catalog root beside terraform-parity.md.
const ProviderConfigManifestFileName = "provider-config-parity.yaml"

// ProviderConfigManifest is one provider's recorded provider-block judgment.
type ProviderConfigManifest struct {
	// Mappings record name divergences between the provider block and the
	// config proto: arg (relative to the provider block, e.g. "assume_role")
	// is config (a config-rooted path, e.g. "config.assume_role_chain").
	// Exact matching resumes below a mapped subtree; Collapse folds an arg
	// subtree onto ONE config leaf (e.g. the per-service endpoints block
	// onto the one endpoints map field).
	Mappings []ConfigMapping `yaml:"mappings,omitempty"`
	// Exclusions are provider-block arguments deliberately not expressible
	// through the provider config, each with its mandatory reason.
	Exclusions []ArgExclusion `yaml:"exclusions,omitempty"`
	// ExclusionPatterns exclude an argument CLASS by regular expression --
	// ONE recorded judgment for a family the provider stamps out per service
	// (e.g. google's ~180 per-service *_custom_endpoint arguments). The
	// per-resource IAM-triplet disposition rides the same by-pattern
	// principle; here the pattern is recorded judgment rather than code so
	// it stays reviewable next to the rest of the manifest. A pattern
	// matching zero arguments at the pin is stale.
	ExclusionPatterns []PatternExclusion `yaml:"exclusionPatterns,omitempty"`
	// ModuleOwned are provider-block arguments set inside catalog modules'
	// own provider blocks by recorded judgment. Each names the argument (or
	// argument subtree) and the reason the modules own it.
	ModuleOwned []ArgExclusion `yaml:"moduleOwned,omitempty"`
	// ConfigExclusions are config proto leaf fields with no provider-block
	// counterpart -- platform-level concepts (e.g. the account identity the
	// control plane validates) the reverse accounting must not flag.
	ConfigExclusions []ConfigExclusion `yaml:"configExclusions,omitempty"`
}

// ConfigMapping is one recorded rename between a provider-block argument and
// a provider-config proto field (the provider-level twin of Mapping).
type ConfigMapping struct {
	Config   string `yaml:"config"`
	Arg      string `yaml:"arg"`
	Collapse bool   `yaml:"collapse,omitempty"`
}

// ConfigExclusion is one config proto leaf with no provider-block
// counterpart.
type ConfigExclusion struct {
	Field  string `yaml:"field"`
	Reason string `yaml:"reason"`
}

// PatternExclusion is one argument-class exclusion by regular expression.
type PatternExclusion struct {
	Pattern string `yaml:"pattern"`
	Reason  string `yaml:"reason"`
}

// ProviderConfigManifestPath composes the manifest location for one provider.
func ProviderConfigManifestPath(repoRoot string, provider catalogkind.CatalogProvider) string {
	return filepath.Join(repoRoot, catalogRoot, catalogkindreflect.ProviderDirName(provider), ProviderConfigManifestFileName)
}

// LoadProviderConfigManifest loads one provider's manifest, returning
// (nil, nil) when the provider ships none -- absence is the enrollment
// signal, not an error. Parsing is strict and validation failures are hard
// errors, matching the per-kind manifests' contract.
func LoadProviderConfigManifest(repoRoot string, provider catalogkind.CatalogProvider) (*ProviderConfigManifest, error) {
	path := ProviderConfigManifestPath(repoRoot, provider)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "opening provider-config manifest %s", path)
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	var m ProviderConfigManifest
	if err := dec.Decode(&m); err != nil {
		return nil, errors.Wrapf(err, "provider-config manifest %s is not valid (strict) YAML", path)
	}
	if err := m.validate(); err != nil {
		return nil, errors.Wrapf(err, "provider-config manifest %s", path)
	}
	return &m, nil
}

func isConfigPath(p string) bool {
	return p == configPathRoot || len(p) > len(configPathRoot) && p[:len(configPathRoot)+1] == configPathRoot+"."
}

func (m *ProviderConfigManifest) validate() error {
	seenArgs := map[string]bool{}
	seenPairs := map[string]bool{}
	for _, mp := range m.Mappings {
		if mp.Arg == "" {
			return errors.Errorf("mapping for config %q names no arg", mp.Config)
		}
		if !isConfigPath(mp.Config) {
			return errors.Errorf("mapping for arg %q must be config-rooted, got %q", mp.Arg, mp.Config)
		}
		pair := mp.Config + " -> " + mp.Arg
		if seenPairs[pair] {
			return errors.Errorf("mapping %q is recorded twice", pair)
		}
		seenPairs[pair] = true
		seenArgs[mp.Arg] = true
	}
	for _, class := range []struct {
		name    string
		entries []ArgExclusion
	}{{"exclusion", m.Exclusions}, {"moduleOwned", m.ModuleOwned}} {
		for _, ex := range class.entries {
			if ex.Arg == "" {
				return errors.Errorf("%s names no arg", class.name)
			}
			if ex.Reason == "" {
				return errors.Errorf("%s of %q carries no reason -- an unexplained omission is the failure class this file exists to prevent", class.name, ex.Arg)
			}
			if seenArgs[ex.Arg] {
				return errors.Errorf("arg %q is judged twice", ex.Arg)
			}
			seenArgs[ex.Arg] = true
		}
	}
	for _, pe := range m.ExclusionPatterns {
		if pe.Pattern == "" {
			return errors.New("exclusionPatterns entry names no pattern")
		}
		if _, err := regexp.Compile(pe.Pattern); err != nil {
			return errors.Wrapf(err, "exclusionPatterns: %q is not a valid regular expression", pe.Pattern)
		}
		if pe.Reason == "" {
			return errors.Errorf("exclusionPatterns: %q carries no reason", pe.Pattern)
		}
	}
	seenFields := map[string]bool{}
	for _, ex := range m.ConfigExclusions {
		if !isConfigPath(ex.Field) {
			return errors.Errorf("configExclusions: field %q must be config-rooted", ex.Field)
		}
		if ex.Reason == "" {
			return errors.Errorf("configExclusions: %q carries no reason", ex.Field)
		}
		if seenFields[ex.Field] {
			return errors.Errorf("configExclusions: %q is listed twice", ex.Field)
		}
		seenFields[ex.Field] = true
	}
	return nil
}
