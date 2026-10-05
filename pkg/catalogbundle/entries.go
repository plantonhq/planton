package catalogbundle

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"sigs.k8s.io/yaml"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// Catalog entries are the bundle's display-and-deploy-coordinates cargo: one
// document per user-facing kind (entries/<provider>/<kind>.yaml) carrying the
// kind's title, description, URL slug, logo, contract links, and the
// official IaC module directories the release ships for it. Every field is a
// projection of the release's own tree -- the catalog page's H1 and intro,
// the iac/ directories that actually exist -- authored once beside the
// kind and never restated by hand, so consumers render and deploy the
// catalog from data without re-deriving any of it.
//
// Entries carry only facts the release owns. Anything a consuming surface
// layers on top (which provider badge to show, how to compose a detail page)
// stays with that surface.

const (
	ossRepoURL  = "https://github.com/plantonhq/planton"
	ossRawURL   = "https://raw.githubusercontent.com/plantonhq/planton/refs/heads/main"
	bsrDocsURL  = "https://buf.build/planton/planton/docs/main"
	logoBaseURL = "https://downloads.planton.dev/catalog/logos"

	// entriesPrefix keys entry documents inside the zip, beside conversions/
	// and presets/.
	entriesPrefix = "entries/"

	// testProviderName is the provider whose kinds exist only to exercise the
	// versioning machinery. They never ship to users, so they never get an
	// entry.
	testProviderName = "_test"
)

// CatalogEntry is one kind's catalog metadata as the bundle carries it.
type CatalogEntry struct {
	// Kind is the PascalCase kind name -- the entry's identity, unique across
	// the catalog and constant across the kind's versions.
	Kind string `json:"kind"`
	// Title is the kind's display name, authored as its catalog page's H1.
	Title string `json:"title"`
	// Description is the first sentence of the catalog page's intro paragraph.
	Description string `json:"description"`
	// Slug is the kebab-case URL segment consoles address the kind by
	// (AwsS3Bucket -> aws-s3-bucket). Derived uniformly from the kind name
	// with the provider kept atomic; unique across the catalog.
	Slug string `json:"slug"`
	// ServiceGroup is the provider-console service group the kind is
	// browsed under, as the registry's enum value name (e.g. "aws_compute");
	// display labels resolve from the group's own metadata. Absent when the
	// kind's provider has no service taxonomy (managed-service providers) --
	// an honest absence mirroring the registry, never a guess.
	ServiceGroup string `json:"serviceGroup,omitempty"`
	// LogoUrl is the kind's own logo at the versionless key the release
	// lane publishes it to.
	LogoUrl string `json:"logoUrl"`
	// WebLinks point at the kind's versioned API contract.
	WebLinks CatalogEntryWebLinks `json:"webLinks"`
	// IacModules names the official module directories the release ships.
	IacModules CatalogEntryIacModules `json:"iacModules"`
	// CostSummary is the kind's cost anatomy at a glance, projected
	// from its cost profile and generated preset estimates (the costs/ and
	// estimates/ cargo). Absent when the kind ships no fact-sheets --
	// absence means "not yet covered", never "free".
	CostSummary *CatalogEntryCostSummary `json:"costSummary,omitempty"`
	// ControlSummary counts the kind's posture across the central
	// control catalog, projected from its control profile (the controls/
	// cargo). Absent when the kind ships no fact-sheets.
	ControlSummary *CatalogEntryControlSummary `json:"controlSummary,omitempty"`
	// PermissionsProvenance says how the kind's permission manifest
	// (the permissions/ cargo) was established: "derived" (static analysis
	// of the official modules), "proven" (observed from live provisioning),
	// or "mixed". Presence signals a downloadable least-privilege manifest
	// exists; the value carries the trust distinction. Absent when the
	// kind ships no fact-sheets.
	PermissionsProvenance string `json:"permissionsProvenance,omitempty"`
}

// CatalogEntryCostSummary is the price-tag projection: enough for a card
// chip ("~$18-140/mo"), with the full story (per-preset line items, sources,
// exclusions) in the estimates/ cargo. The range spans the kind's
// priced preset estimates at published list prices; the bounds echo the
// estimate documents' own decimal strings.
type CatalogEntryCostSummary struct {
	// BillingModel classifies how cost accrues, in the cost profile's own
	// vocabulary: always_on, usage_based, hybrid, free, or cluster_capacity.
	BillingModel string `json:"billingModel"`
	// Currency is the ISO 4217 currency of the range below. Empty when no
	// priced preset estimate exists (rate-delegated and cluster-capacity
	// kinds state no dollar figure -- an honest absence, never 0).
	Currency string `json:"currency,omitempty"`
	// MonthlyMin and MonthlyMax bound the monthly totals across the
	// kind's priced preset estimates, as decimal strings. A genuine
	// zero-committed preset yields "0.00" -- that is a verified number, not
	// a missing one.
	MonthlyMin string `json:"monthlyMin,omitempty"`
	MonthlyMax string `json:"monthlyMax,omitempty"`
}

// CatalogEntryControlSummary counts the kind's stance per control
// status across the ENTIRE central control catalog -- the control-profile
// gate guarantees every catalog control is examined, so these counts always
// sum to the catalog's size and "examined" is never in question.
type CatalogEntryControlSummary struct {
	// EnforcedByDefault counts controls the official modules (or the
	// provider) enforce on every deployment.
	EnforcedByDefault int `json:"enforcedByDefault"`
	// Configurable counts controls exposed as spec choices.
	Configurable int `json:"configurable"`
	// NotApplicable counts controls with no meaning for this kind
	// class.
	NotApplicable int `json:"notApplicable"`
}

// CatalogEntryWebLinks point at the kind's API contract, as source and
// as rendered documentation.
type CatalogEntryWebLinks struct {
	SourceCode    CatalogEntryContractLinks `json:"sourceCode"`
	Documentation CatalogEntryContractLinks `json:"documentation"`
}

// CatalogEntryContractLinks address the contract's parts: the version
// directory as a whole plus the spec, iac-input, and outputs protos.
type CatalogEntryContractLinks struct {
	Root     string `json:"root"`
	Spec     string `json:"spec"`
	IacInput string `json:"iacInput"`
	Outputs  string `json:"outputs"`
}

// CatalogEntryIacModules names the official module directories the release
// ships for a kind. An engine whose directory is absent from the tree
// stays empty here, and deploy paths refuse that engine for the kind --
// presence is truth from the release's own tree, never an assumption. A
// kind ships one engine's module only when its kind declares that
// (kind_meta.provisioners), and the anatomy gate holds the tree to the
// declaration, so this projection and the declaration always agree.
type CatalogEntryIacModules struct {
	// TerraformModuleDir is the repository-relative directory of the
	// terraform-authored module (also executed by OpenTofu). Example:
	// catalog/aws/awsalb/iac/tf
	TerraformModuleDir string `json:"terraformModuleDir,omitempty"`
	// PulumiModuleDir is the repository-relative directory of the pulumi
	// module (Go). Example: catalog/aws/awsalb/iac/pulumi
	PulumiModuleDir string `json:"pulumiModuleDir,omitempty"`
}

// projectEntries derives every user-facing kind's entry from the catalog
// tree plus the compiled kind registry. The projection is total and proven:
// a registry kind without a kind directory, an iac/ directory without
// its engine's entry point, or a tree yielding zero entries fails the build
// with the exact list -- stale or guessed deploy coordinates are unshippable
// by construction. Kinds with fact-sheet cargo additionally carry
// cost/controls/permissions summaries, computed from the same parsed
// documents the cargo packs.
func projectEntries(catalogDir string, cargo map[string]*kindCargo) (map[string][]byte, error) {
	entries := map[string][]byte{}
	var missing, liveness []string

	for _, kind := range catalogkindreflect.KindsList() {
		provider := catalogkindreflect.GetProvider(kind)
		if provider.String() == testProviderName {
			continue
		}
		providerDir := strings.ReplaceAll(provider.String(), "_", "")
		kindName := catalogkindreflect.ExtractKindNameByKind(kind)
		kindDir := strings.ToLower(kindName)
		kindPath := filepath.Join(catalogDir, providerDir, kindDir)
		if _, err := os.Stat(kindPath); err != nil {
			missing = append(missing, fmt.Sprintf("%s/%s (kind %s)", providerDir, kindDir, kindName))
			continue
		}
		versionDir, err := catalogkindreflect.KindVersion(kind)
		if err != nil {
			missing = append(missing, fmt.Sprintf("%s/%s: no version in registry", providerDir, kindDir))
			continue
		}

		title, description := readCatalogPage(filepath.Join(kindPath, "catalog.md"), kindName)
		entry := buildCatalogEntry(kindName, providerDir, kindDir, versionDir, title, description)

		serviceGroup, err := catalogkindreflect.ServiceGroup(kind)
		if err != nil {
			return nil, fmt.Errorf("resolving the %s service group: %w", kindName, err)
		}
		if serviceGroup != catalogkind.CatalogProviderServiceGroup_catalog_provider_service_group_unspecified {
			entry.ServiceGroup = serviceGroup.String()
		}

		if c := cargo[providerDir+"/"+kindDir]; c != nil {
			if err := applyCargoSummaries(&entry, c); err != nil {
				return nil, fmt.Errorf("projecting the %s fact-sheet summaries: %w", kindName, err)
			}
		}

		for _, engine := range []struct {
			dir, entryGlob, label string
			target                *string
		}{
			// Every pulumi module builds from main.go (the runner executes a
			// prebuilt binary and writes its own Pulumi.yaml); a terraform
			// module is any directory of .tf files.
			{"pulumi", "main.go", "Pulumi", &entry.IacModules.PulumiModuleDir},
			{"tf", "*.tf", "Terraform", &entry.IacModules.TerraformModuleDir},
		} {
			engineDir := filepath.Join(kindPath, "iac", engine.dir)
			if _, err := os.Stat(engineDir); err != nil {
				continue
			}
			if matches, _ := filepath.Glob(filepath.Join(engineDir, engine.entryGlob)); len(matches) == 0 {
				liveness = append(liveness, fmt.Sprintf(
					"%s/%s/iac/%s exists without its %s entry point (%s)",
					providerDir, kindDir, engine.dir, engine.label, engine.entryGlob))
				continue
			}
			*engine.target = fmt.Sprintf("catalog/%s/%s/iac/%s", providerDir, kindDir, engine.dir)
		}

		raw, err := yaml.Marshal(entry)
		if err != nil {
			return nil, fmt.Errorf("marshaling the %s catalog entry: %w", kindName, err)
		}
		entries[entriesPrefix+providerDir+"/"+kindDir+".yaml"] = raw
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf(
			"%d registry kind(s) have no kind directory in the catalog tree -- the registry and the tree must agree:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
	if len(liveness) > 0 {
		sort.Strings(liveness)
		return nil, fmt.Errorf(
			"%d module path(s) fail the liveness check:\n  %s",
			len(liveness), strings.Join(liveness, "\n  "))
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("the catalog tree under %s yielded zero entries -- refusing to build a bundle without its catalog-entry cargo", catalogDir)
	}
	return entries, nil
}

func buildCatalogEntry(kindName, providerDir, kindDir, versionDir, title, description string) CatalogEntry {
	entry := CatalogEntry{
		Kind:        kindName,
		Title:       title,
		Description: description,
		Slug:        entrySlug(kindName, providerDir),
		LogoUrl:     fmt.Sprintf("%s/%s/%s/logo.svg", logoBaseURL, providerDir, kindDir),
	}

	contractBase := fmt.Sprintf("catalog/%s/%s/%s", providerDir, kindDir, versionDir)
	// The proto package carries the version segment
	// (dev.planton.aws.awss3bucket.v1alpha1); BSR symbol anchors are full
	// names, so the package appears in both the page path and the fragment.
	pkg := fmt.Sprintf("dev.planton.%s.%s.%s", providerDir, kindDir, versionDir)
	entry.WebLinks.SourceCode = CatalogEntryContractLinks{
		Root:     fmt.Sprintf("%s/tree/main/%s", ossRepoURL, contractBase),
		Spec:     fmt.Sprintf("%s/%s/spec.proto", ossRawURL, contractBase),
		IacInput: fmt.Sprintf("%s/%s/input.proto", ossRawURL, contractBase),
		Outputs:  fmt.Sprintf("%s/%s/outputs.proto", ossRawURL, contractBase),
	}
	entry.WebLinks.Documentation = CatalogEntryContractLinks{
		Root:     fmt.Sprintf("%s:%s", bsrDocsURL, pkg),
		Spec:     fmt.Sprintf("%s:%s#%s.%sSpec", bsrDocsURL, pkg, pkg, kindName),
		IacInput: fmt.Sprintf("%s:%s#%s.%sIacInput", bsrDocsURL, pkg, pkg, kindName),
		Outputs:  fmt.Sprintf("%s:%s#%s.%sOutputs", bsrDocsURL, pkg, pkg, kindName),
	}
	return entry
}

// readCatalogPage extracts the kind's canonical display name (the H1)
// and its one-line description (the first sentence of the intro paragraph)
// from the kind-root catalog page. A kind without a page falls back to
// its kind name -- the anatomy lint owns page presence, not the bundle.
func readCatalogPage(path, kindName string) (title, description string) {
	title = kindName
	description = "Deploy " + kindName
	data, err := os.ReadFile(path)
	if err != nil {
		return title, description
	}
	lines := strings.Split(string(data), "\n")
	i := 0
	for ; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "# ") {
			title = strings.TrimSpace(strings.TrimPrefix(lines[i], "# "))
			description = "Deploy " + title
			i++
			break
		}
	}
	for ; i < len(lines); i++ {
		text := strings.TrimSpace(lines[i])
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		if idx := strings.Index(text, ". "); idx > 0 {
			text = text[:idx+1]
		}
		description = text
		break
	}
	return title, description
}

// entrySlug derives a kind's URL slug: the provider directory name kept
// as one atomic word, then the kind name's remaining words kebab-cased
// (DigitalOceanDroplet -> digitalocean-droplet, AwsS3Bucket -> aws-s3-bucket).
// A kind whose name does not start with its provider kebabs whole. One rule
// for every kind -- uniqueness is gated at conformance.
func entrySlug(kindName, providerDir string) string {
	if rest, ok := strings.CutPrefix(strings.ToLower(kindName), providerDir); ok && rest != "" {
		return providerDir + "-" + kebabCase(kindName[len(kindName)-len(rest):])
	}
	return kebabCase(kindName)
}

// kebabCase converts a CamelCase name (S3Bucket) to kebab (s3-bucket): a
// hyphen lands before every uppercase rune that starts a new word -- after a
// lowercase or digit, or before an Upper->lower transition inside an acronym
// run.
func kebabCase(kindName string) string {
	runes := []rune(kindName)
	var b strings.Builder
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextLower) {
				b.WriteRune('-')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// parseCatalogEntries decodes and shape-checks every entry document in the
// bundle. Strict decoding plus identity checks here mirror the checksum
// posture: a bundle carrying a malformed entry is refused whole. Agreement
// with the kind registry (and slug uniqueness across entries) is the
// conformance gate's job -- loading stays registry-free by design.
func parseCatalogEntries(entries map[string][]byte) ([]CatalogEntry, error) {
	names := make([]string, 0, len(entries))
	for name := range entries {
		if strings.HasPrefix(name, entriesPrefix) {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	parsed := make([]CatalogEntry, 0, len(names))
	for _, name := range names {
		var entry CatalogEntry
		if err := yaml.UnmarshalStrict(entries[name], &entry); err != nil {
			return nil, fmt.Errorf("bundle entry %s is not a valid catalog entry: %w", name, err)
		}
		if entry.Kind == "" || entry.Title == "" || entry.Slug == "" {
			return nil, fmt.Errorf("bundle entry %s is missing its identity (kind, title, and slug are required)", name)
		}
		wantSuffix := "/" + strings.ToLower(entry.Kind) + ".yaml"
		if !strings.HasSuffix(name, wantSuffix) {
			return nil, fmt.Errorf("bundle entry %s declares kind %s but is not keyed by it -- the bundle was mis-built", name, entry.Kind)
		}
		parsed = append(parsed, entry)
	}

	sort.Slice(parsed, func(i, j int) bool { return parsed[i].Kind < parsed[j].Kind })
	return parsed, nil
}
