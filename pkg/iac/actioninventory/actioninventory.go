// Package actioninventory holds the committed snapshots of the cloud
// providers' own permission-action inventories and the matching rules the
// permissions conformance gate uses to prove that every action a runner
// permissions manifest (catalog/<provider>/<kind>/iac/permissions.yaml)
// names actually exists. A least-privilege policy naming an action the
// provider never defined is a fabrication that no syntax check can catch --
// the inventory makes that class of wrong structurally impossible.
//
// The snapshot is generated data used only by this gate: like
// pkg/anatomy's baseline.yaml, it lives beside the code that reads it and
// is refreshed by a make target (`make generate-action-inventory`) that
// talks to the provider's published inventory -- network in make, never in
// CI. CI validates manifests against the committed snapshot. The snapshot
// deliberately covers ONLY the services the committed manifests reference:
// a manifest naming a service the snapshot lacks fails the gate with the
// refresh command, and a refresh that cannot find the service in the
// provider's inventory is a hard error -- a genuinely wrong prefix.
//
// AWS is the first inventory arm, read from AWS's machine-readable service
// reference (https://servicereference.us-east-1.amazonaws.com). Unlike the
// price books' immutable versioned offer documents, the reference serves
// latest-only URLs, so each service records the retrieval date and the
// index's own modification stamp instead of claiming an immutability the
// source does not offer. The reference also publishes, per action, the
// resource types the action can be scoped to; actions publishing NONE are
// evaluated by IAM against Resource "*" only, and the snapshot records
// them (non_scopable_actions) so the scopability gate can refuse the
// quietest least-privilege mistake: a statement scoped to an ARN pattern
// its own actions can never match, which reads tighter than required and
// denies at runtime instead.
//
// Azure is the second arm, read from ARM's provider-operations metadata
// (the inventory behind `az provider operation list`). Azure separates
// management-plane operations from data-plane operations (a role
// definition's `actions` vs `dataActions`), and the permissions schema
// mirrors that split -- so the snapshot records each plane separately and
// the gate holds each manifest field to ITS plane: an operation on the
// wrong plane is a modeling error, not just a typo. ARM publishes no
// modification stamp for the inventory, so Azure services carry only the
// retrieval date -- provenance never claims a fact the source does not
// offer.
//
// GCP is the third arm, read from IAM's testable-permissions inventory
// (the list behind `gcloud iam list-testable-permissions`). GCP
// permissions are flat dotted names ("container.clusters.create"); the
// snapshot keys them by their service segment (the part before the first
// dot) with the remainder as the action, and like ARM the API publishes
// no modification stamp. The inventory is SCOPE-TYPED -- a permission is
// listed only under the resource types it can be tested on, so
// permissions that authorize exclusively on organizations/folders
// (resourcemanager.projects.create) or billing accounts
// (billing.resourceAssociations.create) never appear in a
// project-anchored query -- so the fetcher queries the project, an
// organization, and a billing-account scope and unions the results;
// permissions a provider enforces but publishes under NO scope (Cloud
// Run's domainmappings family) stay teachable only in manifest notes,
// exactly as before.
//
// DigitalOcean is the fourth arm, read from the provider's published
// token-scope reference (docs.digitalocean.com serves it as
// machine-readable markdown; DigitalOcean exposes no scope-inventory
// API). A token scope is "resource:action" -- the same prefix-colon-name
// grammar as an AWS IAM action -- so the snapshot rides the Service
// shape verbatim: prefix = the scope's resource segment, actions = its
// verbs (which go beyond CRUD: view_credentials, access_cluster, admin).
// Matching is EXACT, never MatchAction's IAM glob semantics --
// DigitalOcean does not evaluate wildcards, and borrowing IAM's matcher
// would claim semantics the provider does not have. The docs page
// publishes no modification stamp, so provenance is the URL plus the
// retrieval date.
//
// Cloudflare is the fifth arm and the first on the GroupInventory shape:
// its inventory is a FLAT catalog of named permission groups (the units
// an API token is assembled from), each with a stable id and the scope
// levels it applies at, served whole by one authenticated endpoint
// (GET /accounts/{account_id}/tokens/permission_groups; the catalog is
// global -- every account sees the same list). Group names are NOT
// unique across scopes, so the gate proves each manifest's (name, scope)
// pair; a group renamed by Cloudflare surfaces as a gate failure at the
// next refresh, which is the staleness detection working.
//
// Auth0 is the sixth arm, back on the Service shape. A Management API
// scope is "verb:resource" -- DigitalOcean's grammar reversed -- so the
// prefix is the RESOURCE (the segment after the colon, e.g.
// "connections_options") and the actions are its verbs ("read",
// "update"). The inventory is the tenant's own definition of its
// Management API (the system resource server, which a tenant cannot
// edit), because that is the set Auth0 checks a grant against. Auth0's
// published API reference is deliberately not the source: it omits scopes
// the tenant enforces, so a snapshot drawn from it would refuse a correct
// manifest. Matching is exact; the definition publishes no modification
// stamp, so provenance is the route plus the retrieval date.
//
// Providers without a machine-readable inventory arm (kubernetes) are
// exempt from existence checking (their structural validation lives in
// pkg/iac/permissions) -- exemption is stated here, never silent.
package actioninventory

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/plantonhq/planton/pkg/yamlemit"
)

// AwsFileName is the AWS inventory snapshot's name inside this package's
// directory.
const AwsFileName = "aws.yaml"

// AzureFileName is the Azure inventory snapshot's name inside this
// package's directory.
const AzureFileName = "azure.yaml"

// GcpFileName is the GCP inventory snapshot's name inside this package's
// directory.
const GcpFileName = "gcp.yaml"

// DigitalOceanFileName is the DigitalOcean inventory snapshot's name
// inside this package's directory.
const DigitalOceanFileName = "digitalocean.yaml"

// CloudflareFileName is the Cloudflare permission-group inventory
// snapshot's name inside this package's directory.
const CloudflareFileName = "cloudflare.yaml"

// Auth0FileName is the Auth0 Management API scope inventory snapshot's
// name inside this package's directory.
const Auth0FileName = "auth0.yaml"

// Inventory is one provider's committed action-inventory snapshot.
type Inventory struct {
	// Provider is the catalog provider directory name (e.g. "aws").
	Provider string `yaml:"provider"`
	// Services are the per-service action lists, sorted by prefix.
	Services []Service `yaml:"services"`
}

// Service is one cloud service's action list as the provider publishes it.
type Service struct {
	// Prefix is the service's action-name prefix in the provider's own
	// vocabulary -- for AWS, the IAM prefix before the colon in
	// "lambda:CreateFunction"; for Azure, the ARM namespace before the
	// first slash in "Microsoft.Network/dnsZones/read".
	Prefix string `yaml:"prefix"`
	// SourceURL is the provider inventory document the actions were read
	// from. The URL serves latest-only content; provenance is the URL plus
	// the dates below, never a claim of immutability.
	SourceURL string `yaml:"source_url"`
	// SourceModified is the provider's own last-modified date for this
	// service's inventory document (UTC, date-only), when the source
	// publishes one (AWS's reference index does; ARM does not). Empty
	// means the source offers no stamp -- provenance never invents one.
	SourceModified string `yaml:"source_modified,omitempty"`
	// RetrievedOn is the date this service's actions were fetched (UTC).
	RetrievedOn string `yaml:"retrieved_on"`
	// Actions are the action names the provider defines for this service,
	// sorted, without the service prefix. For Azure these are the
	// management-plane operations (a role definition's `actions`).
	Actions []string `yaml:"actions"`
	// NonScopableActions is the subset of Actions whose provider
	// reference declares NO resource types. IAM evaluates such actions
	// against Resource "*" only: an ARN-scoped grant NEVER matches them,
	// so a policy statement that looks tighter than the provider supports
	// silently denies at runtime -- the scopability gate holds every
	// statement carrying one of these to exactly the "*" resource. Sorted;
	// every entry must also appear in Actions. Only providers whose
	// reference publishes per-action resource types (AWS's service
	// reference) populate it.
	NonScopableActions []string `yaml:"non_scopable_actions,omitempty"`
	// DataActions are the provider's data-plane operation names for this
	// service, sorted, without the service prefix. Only providers that
	// split planes (Azure) populate it; a manifest's data_actions are held
	// to THIS list, never to Actions.
	DataActions []string `yaml:"data_actions,omitempty"`
}

// LoadAws reads and strictly parses the committed AWS inventory snapshot
// and enforces its structural invariants (sorted unique services, sorted
// unique non-empty action lists) so gate results can never depend on
// snapshot ordering accidents.
func LoadAws(dir string) (*Inventory, error) {
	// AWS's reference index publishes a modification stamp, so AWS
	// services must carry one.
	return load(dir, AwsFileName, "aws", true)
}

// LoadAzure reads and strictly parses the committed Azure inventory
// snapshot under the same structural invariants. ARM publishes no
// modification stamp, so Azure services carry only the retrieval date.
func LoadAzure(dir string) (*Inventory, error) {
	return load(dir, AzureFileName, "azure", false)
}

// LoadGcp reads and strictly parses the committed GCP inventory snapshot
// under the same structural invariants. The IAM inventory publishes no
// modification stamp, so GCP services carry only the retrieval date.
func LoadGcp(dir string) (*Inventory, error) {
	return load(dir, GcpFileName, "gcp", false)
}

// LoadDigitalOcean reads and strictly parses the committed DigitalOcean
// inventory snapshot under the same structural invariants. A DigitalOcean
// token scope is "resource:action" -- the same prefix-colon-name grammar
// as an AWS IAM action -- so the scope inventory rides the Service shape
// verbatim: prefix = the scope's resource segment, actions = its verbs.
// The scopes reference publishes no modification stamp, so DigitalOcean
// services carry only the retrieval date.
func LoadDigitalOcean(dir string) (*Inventory, error) {
	return load(dir, DigitalOceanFileName, "digitalocean", false)
}

// LoadAuth0 reads and strictly parses the committed Auth0 inventory
// snapshot under the same structural invariants. An Auth0 scope is
// "verb:resource", so a Service's prefix is the scope's resource segment
// and its actions are the verbs. The tenant's Management API definition
// publishes no modification stamp, so Auth0 services carry only the
// retrieval date.
func LoadAuth0(dir string) (*Inventory, error) {
	return load(dir, Auth0FileName, "auth0", false)
}

func load(dir, fileName, provider string, requireSourceModified bool) (*Inventory, error) {
	raw, err := os.ReadFile(path.Join(dir, fileName))
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(strings.NewReader(string(raw)))
	decoder.KnownFields(true)
	inv := &Inventory{}
	if err := decoder.Decode(inv); err != nil {
		return nil, fmt.Errorf("parse %s: %w", fileName, err)
	}
	if inv.Provider != provider {
		return nil, fmt.Errorf("%s: provider is %q, want %q", fileName, inv.Provider, provider)
	}
	if len(inv.Services) == 0 {
		return nil, fmt.Errorf("%s: no services -- run `make generate-action-inventory`", fileName)
	}
	seen := map[string]bool{}
	for i, svc := range inv.Services {
		if svc.Prefix == "" || svc.SourceURL == "" || svc.RetrievedOn == "" {
			return nil, fmt.Errorf("%s: service %q: prefix, source_url, and retrieved_on are all required", fileName, svc.Prefix)
		}
		if requireSourceModified && svc.SourceModified == "" {
			return nil, fmt.Errorf("%s: service %q: prefix, source_url, source_modified, and retrieved_on are all required", fileName, svc.Prefix)
		}
		if seen[svc.Prefix] {
			return nil, fmt.Errorf("%s: duplicate service prefix %q", fileName, svc.Prefix)
		}
		seen[svc.Prefix] = true
		if i > 0 && inv.Services[i-1].Prefix > svc.Prefix {
			return nil, fmt.Errorf("%s: services not sorted at %q", fileName, svc.Prefix)
		}
		if len(svc.Actions) == 0 && len(svc.DataActions) == 0 {
			return nil, fmt.Errorf("%s: service %q has no actions", fileName, svc.Prefix)
		}
		if err := checkActionList(fileName, svc.Prefix, "actions", svc.Actions); err != nil {
			return nil, err
		}
		if err := checkActionList(fileName, svc.Prefix, "data_actions", svc.DataActions); err != nil {
			return nil, err
		}
		if err := checkActionList(fileName, svc.Prefix, "non_scopable_actions", svc.NonScopableActions); err != nil {
			return nil, err
		}
		// The subset invariant: a non-scopable entry that is not a
		// published action would let the scopability gate reason about a
		// name the existence gate never admitted.
		if len(svc.NonScopableActions) > 0 {
			published := make(map[string]bool, len(svc.Actions))
			for _, action := range svc.Actions {
				published[action] = true
			}
			for _, action := range svc.NonScopableActions {
				if !published[action] {
					return nil, fmt.Errorf("%s: service %q non_scopable_actions entry %q is not in actions", fileName, svc.Prefix, action)
				}
			}
		}
	}
	return inv, nil
}

// checkActionList enforces one plane's list invariants: no empties, no
// duplicates, sorted.
func checkActionList(fileName, prefix, plane string, actions []string) error {
	for j, action := range actions {
		if action == "" {
			return fmt.Errorf("%s: service %q has an empty action in %s", fileName, prefix, plane)
		}
		if j > 0 {
			if prev := actions[j-1]; prev == action {
				return fmt.Errorf("%s: service %q duplicates action %q in %s", fileName, prefix, action, plane)
			} else if prev > action {
				return fmt.Errorf("%s: service %q %s not sorted at %q", fileName, prefix, plane, action)
			}
		}
	}
	return nil
}

// ServiceActions returns the snapshot's action list for a service prefix,
// or nil when the snapshot does not cover the service.
func (inv *Inventory) ServiceActions(prefix string) []string {
	if svc := inv.Lookup(prefix); svc != nil {
		return svc.Actions
	}
	return nil
}

// Lookup returns the snapshot's service for a prefix, or nil when the
// snapshot does not cover it. Callers that must tell "service not
// covered" apart from "covered but empty on one plane" (the Azure gate)
// use this instead of ServiceActions.
func (inv *Inventory) Lookup(prefix string) *Service {
	for i := range inv.Services {
		if inv.Services[i].Prefix == prefix {
			return &inv.Services[i]
		}
	}
	return nil
}

// MatchAction reports how many of a service's published actions an IAM
// action name matches. IAM evaluates action names case-insensitively and
// supports "*" (any run) and "?" (any one character) wildcards, so the
// gate does too: an exact name matches its published spelling in any case,
// and a wildcard pattern must match at least one published action -- a
// pattern matching NOTHING is the fabrication class the inventory exists
// to catch, and wildcards are where it hides best.
func MatchAction(published []string, name string) int {
	pattern := strings.ToLower(name)
	matches := 0
	for _, action := range published {
		if globMatch(pattern, strings.ToLower(action)) {
			matches++
		}
	}
	return matches
}

// globMatch implements IAM's action wildcard semantics ("*" any run, "?"
// any single character) over lowercased inputs. Iterative with
// backtracking on the last "*", the classic linear-scan shape.
func globMatch(pattern, s string) bool {
	pIdx, sIdx := 0, 0
	starIdx, starMatch := -1, 0
	for sIdx < len(s) {
		switch {
		case pIdx < len(pattern) && (pattern[pIdx] == '?' || pattern[pIdx] == s[sIdx]):
			pIdx++
			sIdx++
		case pIdx < len(pattern) && pattern[pIdx] == '*':
			starIdx = pIdx
			starMatch = sIdx
			pIdx++
		case starIdx != -1:
			pIdx = starIdx + 1
			starMatch++
			sIdx = starMatch
		default:
			return false
		}
	}
	for pIdx < len(pattern) && pattern[pIdx] == '*' {
		pIdx++
	}
	return pIdx == len(pattern)
}

// header is the snapshot's leading comment, written by Render so the
// fetcher and any future regeneration path can never disagree on it.
const header = `# GENERATED -- DO NOT EDIT. Refresh with ` + "`make generate-action-inventory`" + `.
# Committed snapshot of the provider's own permission-action inventory,
# scoped to exactly the services the committed runner permissions manifests
# (catalog/*/*/iac/permissions.yaml) reference. The conformance gate in
# this package proves every manifest action exists here -- an action name
# the provider never defined cannot ship. Source URLs serve latest-only
# content; provenance is the URL plus the retrieval date and the source's
# own modification stamp.
`

// Render writes an inventory in its canonical byte form: header, sorted
// services, sorted actions, dates quoted. One renderer home keeps refresh
// diffs meaningful.
func Render(inv *Inventory) string {
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("provider: " + inv.Provider + "\n")
	b.WriteString("services:\n")
	services := append([]Service(nil), inv.Services...)
	sort.Slice(services, func(i, j int) bool { return services[i].Prefix < services[j].Prefix })
	for _, svc := range services {
		b.WriteString("  - prefix: " + svc.Prefix + "\n")
		b.WriteString("    source_url: " + svc.SourceURL + "\n")
		if svc.SourceModified != "" {
			b.WriteString("    source_modified: \"" + svc.SourceModified + "\"\n")
		}
		b.WriteString("    retrieved_on: \"" + svc.RetrievedOn + "\"\n")
		if len(svc.Actions) > 0 {
			b.WriteString("    actions:\n")
			actions := append([]string(nil), svc.Actions...)
			sort.Strings(actions)
			for _, action := range actions {
				b.WriteString("      - " + action + "\n")
			}
		}
		if len(svc.NonScopableActions) > 0 {
			b.WriteString("    non_scopable_actions:\n")
			nonScopable := append([]string(nil), svc.NonScopableActions...)
			sort.Strings(nonScopable)
			for _, action := range nonScopable {
				b.WriteString("      - " + action + "\n")
			}
		}
		if len(svc.DataActions) > 0 {
			b.WriteString("    data_actions:\n")
			dataActions := append([]string(nil), svc.DataActions...)
			sort.Strings(dataActions)
			for _, action := range dataActions {
				b.WriteString("      - " + action + "\n")
			}
		}
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// The permission-group inventory shape.
//
// Cloudflare's inventory does not fit the Service shape: there is no
// service prefix and no action list -- the provider publishes a FLAT
// catalog of named permission groups, each with a stable id and the scope
// levels it applies at, from ONE endpoint. Forcing that into prefix ->
// actions would mismodel the provider, so group-shaped inventories get
// their own snapshot type with the same full treatment (strict loader,
// canonical renderer, structural invariants). Cloudflare is the first
// and only arm on this shape today.
// ---------------------------------------------------------------------------

// GroupInventory is one provider's committed permission-group inventory
// snapshot. Unlike the Service shape's per-service provenance, the whole
// inventory comes from a single provider endpoint, so provenance lives at
// the top level.
type GroupInventory struct {
	// Provider is the catalog provider directory name (e.g. "cloudflare").
	Provider string `yaml:"provider"`
	// SourceURL is the provider inventory endpoint the groups were read
	// from, in its documented route form (account ids are placeholders --
	// the catalog is global, and the snapshot must not vary by which
	// operator account fetched it).
	SourceURL string `yaml:"source_url"`
	// RetrievedOn is the date the groups were fetched (UTC). The endpoint
	// publishes no modification stamp -- provenance never invents one.
	RetrievedOn string `yaml:"retrieved_on"`
	// Groups are the provider's permission groups, sorted by name then id.
	// Names are NOT unique -- the provider defines same-named groups at
	// different scopes -- so consumers key on (name, scope), never name
	// alone.
	Groups []PermissionGroup `yaml:"groups"`
}

// PermissionGroup is one named permission group as the provider defines
// it.
type PermissionGroup struct {
	// ID is the provider's stable identifier for the group. The provider
	// documents names as cosmetic and ids as the durable key, so the
	// snapshot records both: manifests speak names (the human vocabulary),
	// and a future token-creation renderer joins name -> id here without
	// a second fetch.
	ID string `yaml:"id"`
	// Name is the group's display name verbatim (e.g. "DNS Write").
	Name string `yaml:"name"`
	// Scopes are the resource levels the group applies at, in the
	// provider's own identifier spelling (e.g.
	// "com.cloudflare.api.account.zone"), sorted.
	Scopes []string `yaml:"scopes"`
}

// LoadCloudflare reads and strictly parses the committed Cloudflare
// permission-group inventory snapshot and enforces its structural
// invariants (sorted unique groups, sorted unique non-empty scope lists)
// so gate results can never depend on snapshot ordering accidents.
func LoadCloudflare(dir string) (*GroupInventory, error) {
	return loadGroups(dir, CloudflareFileName, "cloudflare")
}

func loadGroups(dir, fileName, provider string) (*GroupInventory, error) {
	raw, err := os.ReadFile(path.Join(dir, fileName))
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(strings.NewReader(string(raw)))
	decoder.KnownFields(true)
	inv := &GroupInventory{}
	if err := decoder.Decode(inv); err != nil {
		return nil, fmt.Errorf("parse %s: %w", fileName, err)
	}
	if inv.Provider != provider {
		return nil, fmt.Errorf("%s: provider is %q, want %q", fileName, inv.Provider, provider)
	}
	if inv.SourceURL == "" || inv.RetrievedOn == "" {
		return nil, fmt.Errorf("%s: source_url and retrieved_on are both required", fileName)
	}
	if len(inv.Groups) == 0 {
		return nil, fmt.Errorf("%s: no groups -- run `make generate-action-inventory`", fileName)
	}
	seenIDs := map[string]bool{}
	for i, group := range inv.Groups {
		if group.ID == "" || group.Name == "" {
			return nil, fmt.Errorf("%s: group %d: id and name are both required", fileName, i)
		}
		if seenIDs[group.ID] {
			return nil, fmt.Errorf("%s: duplicate group id %q", fileName, group.ID)
		}
		seenIDs[group.ID] = true
		if i > 0 {
			prev := inv.Groups[i-1]
			if prev.Name > group.Name || (prev.Name == group.Name && prev.ID > group.ID) {
				return nil, fmt.Errorf("%s: groups not sorted at %q (%s)", fileName, group.Name, group.ID)
			}
		}
		if len(group.Scopes) == 0 {
			return nil, fmt.Errorf("%s: group %q (%s) has no scopes", fileName, group.Name, group.ID)
		}
		for j, scope := range group.Scopes {
			if scope == "" {
				return nil, fmt.Errorf("%s: group %q (%s) has an empty scope", fileName, group.Name, group.ID)
			}
			if j > 0 {
				if prev := group.Scopes[j-1]; prev == scope {
					return nil, fmt.Errorf("%s: group %q (%s) duplicates scope %q", fileName, group.Name, group.ID, scope)
				} else if prev > scope {
					return nil, fmt.Errorf("%s: group %q (%s) scopes not sorted at %q", fileName, group.Name, group.ID, scope)
				}
			}
		}
	}
	return inv, nil
}

// HasGroup reports whether the inventory defines a group with this exact
// name at this exact scope. Matching is exact and case-sensitive: group
// names are not IAM-evaluated patterns, so no wildcard or case semantics
// apply.
func (inv *GroupInventory) HasGroup(name, scope string) bool {
	for _, group := range inv.Groups {
		if group.Name != name {
			continue
		}
		for _, s := range group.Scopes {
			if s == scope {
				return true
			}
		}
	}
	return false
}

// GroupScopes returns the union of scopes the inventory defines for a
// group name, or nil when no group carries the name. Callers that must
// tell "name invented" apart from "name real, scope wrong" (the
// Cloudflare gate's two distinct diagnostics) use this beside HasGroup.
func (inv *GroupInventory) GroupScopes(name string) []string {
	var scopes []string
	seen := map[string]bool{}
	for _, group := range inv.Groups {
		if group.Name != name {
			continue
		}
		for _, s := range group.Scopes {
			if !seen[s] {
				seen[s] = true
				scopes = append(scopes, s)
			}
		}
	}
	sort.Strings(scopes)
	return scopes
}

// groupsHeader is the group-shaped snapshot's leading comment, written by
// RenderGroups so the fetcher and any future regeneration path can never
// disagree on it.
const groupsHeader = `# GENERATED -- DO NOT EDIT. Refresh with ` + "`make generate-action-inventory`" + `.
# Committed snapshot of the provider's own permission-group inventory,
# scoped to exactly the group names the committed runner permissions
# manifests (catalog/*/*/iac/permissions.yaml) reference. The conformance
# gate in this package proves every manifest (name, scope) pair exists
# here -- a permission group the provider never defined cannot ship. The
# endpoint serves latest-only content; provenance is the URL plus the
# retrieval date.
`

// RenderGroups writes a group inventory in its canonical byte form:
// header, top-level provenance, groups sorted by name then id, scopes
// sorted, dates quoted. Scalar quoting rides pkg/yamlemit (the one home
// of the generators' quoting decision -- group names carry ": " and would
// change meaning emitted plain). One renderer home keeps refresh diffs
// meaningful.
func RenderGroups(inv *GroupInventory) string {
	var b strings.Builder
	b.WriteString(groupsHeader)
	yamlemit.WriteKV(&b, 0, "provider", inv.Provider, false)
	yamlemit.WriteKV(&b, 0, "source_url", inv.SourceURL, false)
	yamlemit.WriteKV(&b, 0, "retrieved_on", inv.RetrievedOn, true)
	b.WriteString("groups:\n")
	groups := append([]PermissionGroup(nil), inv.Groups...)
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Name != groups[j].Name {
			return groups[i].Name < groups[j].Name
		}
		return groups[i].ID < groups[j].ID
	})
	for _, group := range groups {
		yamlemit.WriteKV(&b, 2, "- id", group.ID, false)
		yamlemit.WriteKV(&b, 4, "name", group.Name, false)
		b.WriteString("    scopes:\n")
		scopes := append([]string(nil), group.Scopes...)
		sort.Strings(scopes)
		for _, scope := range scopes {
			yamlemit.WriteListItem(&b, 6, scope)
		}
	}
	return b.String()
}
