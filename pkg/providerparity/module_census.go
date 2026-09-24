//go:build !codegen
// +build !codegen

// Module census: the catalog's consumption side of the parity measurement.
// For every registered kind of one cloud provider, which Terraform resources
// its module actually declares and which provider releases it pins.
//
// The scan reads EVERY `*.tf` file in the module directory, never main.tf
// alone: modules may split resources across sibling files, and a main.tf-only
// scan undercounts silently -- the worst failure mode a parity instrument can
// have. Resource declarations are matched by the same top-level pattern the
// import-map conformance guard scans.

package providerparity

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// catalogRoot is where components live, relative to the repo root. A
// package-local constant by convention (pkg/anatomy, pkg/e2e/profile,
// pkg/iac/importmap do the same); the version segment never appears in
// module paths -- iac/ sits at the component root.
const catalogRoot = "catalog"

// tfResourceRe matches top-level terraform resource declarations.
var tfResourceRe = regexp.MustCompile(`(?m)^resource\s+"([a-z0-9_]+)"\s+"`)

// tfProviderPinRe pulls each required_providers entry's local name, source,
// and version constraint. The catalog's modules declare pins in one uniform
// block shape, so a structural regex is sufficient here, as it is for the
// tf-provider-pins CI guard.
var tfProviderPinRe = regexp.MustCompile(`(?s)([A-Za-z0-9_-]+)\s*=\s*\{[^{}]*?source\s*=\s*"([^"]+)"[^{}]*?version\s*=\s*"([^"]+)"[^{}]*?\}`)

// ModuleCensus is one kind's Terraform-module consumption.
type ModuleCensus struct {
	// Kind is the registry name, e.g. "GcpGcsBucket".
	Kind string
	// ModuleDir is the repo-root-relative Terraform module directory.
	ModuleDir string
	// MissingModule marks a registered, implemented kind with no Terraform
	// module directory at all -- an anatomy-baselined gap in some catalogs.
	// Recorded here (and surfaced as an accounting Finding) rather than
	// aborting the census, so one kind's debt never hides every other
	// kind's numbers.
	MissingModule bool
	// Resources is the sorted distinct resource types the module declares.
	Resources []string
	// Pins maps each declared provider's local name to its version
	// constraint, e.g. {"google": "~> 6.0"}.
	Pins map[string]string
	// ProviderBlockArgs maps each `provider "<name>" {}` block's local name
	// to the sorted dotted argument paths the module sets inside it (e.g.
	// azurerm's "features.machine_learning.purge_soft_deleted_workspace_on_destroy").
	// Parsed with real HCL, not regex: provider-block bodies nest, and a
	// missed argument here is an invisible provider-behavior leak -- the
	// silent-omission class the provider-config accounting exists to
	// eliminate. Empty blocks (the catalog's canonical shape) yield no entry.
	ProviderBlockArgs map[string][]string
	// ProviderAttachments maps a resource type to the provider local name
	// its blocks attach through the `provider = <name>` meta-argument
	// (e.g. "google_firebase_project" -> "google-beta"). Only explicit
	// attachments are recorded: a resource with no meta-argument uses the
	// provider its type prefix implies, and the map has no entry for it.
	// This is the census fact the beta-admission accounting reads -- a
	// secondary-channel resource enters a module only through an explicit
	// attachment, so an attachment with no admission (or an admission with
	// no attachment) is visible here rather than inferred. Parsed with real
	// HCL alongside the provider blocks. A resource type attached through
	// two different providers in one module is an error: the module's
	// intent is ambiguous and the accounting cannot pick.
	ProviderAttachments map[string]string
}

// ModuleCensusForProvider scans the Terraform module of every registered kind
// of one cloud provider, sorted by kind. A registered kind whose module
// directory is missing is a per-kind FINDING, never a silent skip and never
// a run-abort: the finding gates through the burn-down baseline, so an
// anatomy-clean catalog (zero baseline entries) still fails CI on the first
// missing module, while a catalog carrying recorded anatomy debt can be
// measured without that debt hiding every other kind's numbers.
func ModuleCensusForProvider(repoRoot string, provider cloudresourcekind.CloudResourceProvider) ([]ModuleCensus, error) {
	var out []ModuleCensus
	for _, kind := range crkreflect.KindsList() {
		if crkreflect.GetProvider(kind) != provider {
			continue
		}
		if _, err := crkreflect.NewInstance(kind); err != nil {
			continue // enum value exists but the kind is not implemented yet
		}
		moduleRel := filepath.Join(catalogRoot, crkreflect.ProviderDirName(provider), strings.ToLower(kind.String()), "iac", "tf")
		moduleAbs := filepath.Join(repoRoot, moduleRel)
		if _, err := os.Stat(moduleAbs); os.IsNotExist(err) {
			out = append(out, ModuleCensus{Kind: kind.String(), ModuleDir: moduleRel, MissingModule: true})
			continue
		}
		census, err := ScanModule(moduleAbs)
		if err != nil {
			return nil, errors.Wrapf(err, "kind %s", kind.String())
		}
		census.Kind = kind.String()
		census.ModuleDir = moduleRel
		out = append(out, census)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out, nil
}

// ScanModule reads every top-level `*.tf` file in one module directory and
// returns its declared resources, provider pins, and provider-block
// arguments. Exposed so tests can drive it against fixture modules in
// isolation.
func ScanModule(moduleDir string) (ModuleCensus, error) {
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		return ModuleCensus{}, errors.Wrapf(err, "reading module dir %s", moduleDir)
	}
	seen := map[string]bool{}
	census := ModuleCensus{Pins: map[string]string{}}
	providerArgs := map[string]map[string]bool{}
	attachments := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}
		path := filepath.Join(moduleDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return ModuleCensus{}, errors.Wrapf(err, "reading %s", entry.Name())
		}
		for _, m := range tfResourceRe.FindAllStringSubmatch(string(content), -1) {
			if !seen[m[1]] {
				seen[m[1]] = true
				census.Resources = append(census.Resources, m[1])
			}
		}
		for _, m := range tfProviderPinRe.FindAllStringSubmatch(string(content), -1) {
			census.Pins[m[1]] = m[3]
		}
		if err := collectProviderBlockArgs(path, content, providerArgs); err != nil {
			return ModuleCensus{}, err
		}
		if err := collectProviderAttachments(path, content, attachments); err != nil {
			return ModuleCensus{}, err
		}
	}
	if len(attachments) > 0 {
		census.ProviderAttachments = attachments
	}
	for name, args := range providerArgs {
		if len(args) == 0 {
			continue
		}
		sorted := make([]string, 0, len(args))
		for a := range args {
			sorted = append(sorted, a)
		}
		sort.Strings(sorted)
		if census.ProviderBlockArgs == nil {
			census.ProviderBlockArgs = map[string][]string{}
		}
		census.ProviderBlockArgs[name] = sorted
	}
	sort.Strings(census.Resources)
	return census, nil
}

// collectProviderBlockArgs HCL-parses one .tf file and records the dotted
// argument paths set inside each top-level `provider "<name>" {}` block. A
// parse failure is a hard error: the catalog's modules are valid HCL by
// construction (they deploy), so an unparseable file means the census would
// otherwise under-report -- fail loudly instead.
func collectProviderBlockArgs(path string, content []byte, out map[string]map[string]bool) error {
	file, diags := hclparse.NewParser().ParseHCL(content, path)
	if diags.HasErrors() {
		return errors.Errorf("parsing %s: %s", path, diags.Error())
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return errors.Errorf("parsing %s: unexpected body type %T", path, file.Body)
	}
	for _, block := range body.Blocks {
		if block.Type != "provider" || len(block.Labels) != 1 {
			continue
		}
		name := block.Labels[0]
		if out[name] == nil {
			out[name] = map[string]bool{}
		}
		collectBodyArgPaths(block.Body, "", out[name])
	}
	return nil
}

// collectProviderAttachments HCL-parses one .tf file and records, for every
// top-level `resource "<type>" "<name>" {}` block that sets the `provider`
// meta-argument, the provider local name it attaches through. The
// meta-argument's value is a bare traversal (`google-beta`, or
// `google-beta.alias` for an aliased configuration); the root name is the
// provider local name the required_providers block pins, which is the
// identity the admission accounting and the pin guard both key on.
//
// Two blocks of one resource type attaching through different providers is
// a hard error rather than a last-wins pick: the census reports per
// resource TYPE, and a type split across providers has no single truth to
// account against.
func collectProviderAttachments(path string, content []byte, out map[string]string) error {
	file, diags := hclparse.NewParser().ParseHCL(content, path)
	if diags.HasErrors() {
		return errors.Errorf("parsing %s: %s", path, diags.Error())
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return errors.Errorf("parsing %s: unexpected body type %T", path, file.Body)
	}
	for _, block := range body.Blocks {
		if block.Type != "resource" || len(block.Labels) != 2 {
			continue
		}
		attr, set := block.Body.Attributes["provider"]
		if !set {
			continue
		}
		traversal, ok := attr.Expr.(*hclsyntax.ScopeTraversalExpr)
		if !ok || len(traversal.Traversal) == 0 {
			return errors.Errorf("parsing %s: resource %q %q sets provider to a non-traversal expression -- the meta-argument must be a bare provider reference", path, block.Labels[0], block.Labels[1])
		}
		providerName := traversal.Traversal.RootName()
		resourceType := block.Labels[0]
		if prev, exists := out[resourceType]; exists && prev != providerName {
			return errors.Errorf("parsing %s: resource type %s attaches through both %q and %q -- one provider per resource type per module", path, resourceType, prev, providerName)
		}
		out[resourceType] = providerName
	}
	return nil
}

// collectBodyArgPaths walks one HCL body, recording attribute paths and
// descending nested blocks ("features" with an empty body is still recorded:
// an empty required block is an argument the module sets).
func collectBodyArgPaths(body *hclsyntax.Body, prefix string, out map[string]bool) {
	join := func(name string) string {
		if prefix == "" {
			return name
		}
		return prefix + "." + name
	}
	for name := range body.Attributes {
		out[join(name)] = true
	}
	for _, block := range body.Blocks {
		blockPath := join(block.Type)
		if len(block.Body.Attributes) == 0 && len(block.Body.Blocks) == 0 {
			out[blockPath] = true
			continue
		}
		collectBodyArgPaths(block.Body, blockPath, out)
	}
}
