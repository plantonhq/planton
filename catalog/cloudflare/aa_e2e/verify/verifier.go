// Package verify checks that Cloudflare resources created by an E2E scenario
// exist after DEPLOY and are gone after DESTROY, by GETting each resource on
// Cloudflare's own v4 REST API (the SaaS-provider pattern: one HTTP client,
// per-kind GET paths -- no cloud CLI).
//
// Standing convention: verification reads everything it needs from the
// component's outputs. A zone-scoped kind's outputs must therefore
// carry the zone_id alongside the resource's own identifier (a Cloudflare
// resource's API identity is compound -- zones/{zone_id}/<collection>/{id}).
// When enrolling a kind whose outputs lack its scope, add the output to the
// contract and both engine modules first; never fish the scope out of the
// manifest here.
package verify

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// EnvelopePresence selects which 200-body signals count as ABSENT. The zero
// value treats any 200 as present (same answer as ResourceExists, but
// ResourcePresent still parses the v4 envelope -- never use it on raw-body
// endpoints).
type EnvelopePresence struct {
	// SoftDeleted: a 200 whose result.deleted_at is set counts as absent
	// (cloudflared tunnels, teamnet routes).
	SoftDeleted bool
	// AbsentStatuses: a 200 whose result.status is one of these counts as
	// absent (certificate packs, custom hostnames, fallback origin --
	// pending_deletion / deleted).
	AbsentStatuses []string
	// EmptyResultArray: a 200 whose result is an empty JSON array counts as
	// absent. Used by whole-collection resources whose destroy empties the
	// list while the endpoint keeps answering (snippet rules -- a zone
	// singleton whose Delete wipes every rule and GET then returns []).
	EmptyResultArray bool
	// IsDeletedFlag: a 200 whose result.is_deleted is a non-zero number
	// counts as absent. The Workflows API's variant of soft-delete: its GET
	// response carries is_deleted as a required numeric field instead of the
	// deleted_at timestamp the SoftDeleted probe reads.
	IsDeletedFlag bool
	// RevokedAt: a 200 whose result.revoked_at is set counts as absent.
	// The Origin CA variant of soft-delete: DELETE /certificates/{id} is a
	// REVOKE, and the revoked certificate keeps answering GET 200 with the
	// full body plus revoked_at (measured live 2026-08-28) -- neither the
	// plain 404 probe nor SoftDeleted (deleted_at) sees the revocation.
	RevokedAt bool
}

// API is the surface verifiers need from the harness client: an
// existence probe plus the account scope resolved at Setup.
type API interface {
	// ResourceExists GETs a path relative to /client/v4/ and reports
	// presence (200) vs absence (404, or one of Cloudflare's 400
	// unknown-object answers: code 7003 "could not route", code 1001
	// "Invalid zone identifier" -- the zones/{id} answer for a deleted
	// zone, code 1103 "Location ID is invalid" -- the
	// gateway/locations/{id} answer for a deleted DNS location, code 1472
	// "Certificate not found." -- the mtls_certificates/{id} answer for a
	// deleted certificate, and the CODE-LESS "requested snippet not found"
	// -- the snippets endpoint's deleted answer carries a message but no
	// numeric code, all measured live); any other status is an error.
	ResourceExists(ctx context.Context, path string) (bool, error)
	// ResourceActive is the soft-delete-aware sibling: a 200 whose envelope
	// carries a non-null result.deleted_at counts as ABSENT. Used only by
	// verifiers whose family soft-deletes (cloudflared tunnels, teamnet
	// routes) -- it parses the v4 envelope, so it must never replace
	// ResourceExists on raw-body endpoints (e.g. the KV value endpoint).
	ResourceActive(ctx context.Context, path string) (bool, error)
	// ResourcePresent is the envelope-aware probe: 404 / the 400
	// unknown-object codes (7003, 1001, 1103, 1472) are
	// always absent; a 200 is absent when SoftDeleted sees deleted_at or
	// result.status matches AbsentStatuses. Parses the v4 envelope --
	// never use on raw-body endpoints.
	ResourcePresent(ctx context.Context, path string, opts EnvelopePresence) (bool, error)
	// AccountID returns the Cloudflare account the harness is scoped to.
	AccountID() string
}

// Verifier checks a single component's Cloudflare resource for
// existence/absence from its outputs.
type Verifier interface {
	// IDOutputKey is the output key carrying the resource's primary
	// identifier -- used to confirm the deploy produced a verifiable handle.
	IDOutputKey() string
	// VerifyExists returns an error unless the resource exists.
	VerifyExists(ctx context.Context, api API, outputs map[string]string) error
	// VerifyAbsent returns an error unless the resource is gone.
	VerifyAbsent(ctx context.Context, api API, outputs map[string]string) error
}

// apiPathVerifier is the common implementation: one GET path template whose
// placeholders are filled, in order, from the named output keys.
// Account-scoped resources set accountScoped, which prepends the harness
// account to the path ("accounts/%s/..." with the first placeholder filled
// from API.AccountID(), never from outputs -- account IDs are harness scope,
// not deploy results).
type apiPathVerifier struct {
	kind          string
	pathFormat    string
	outputKeys    []string
	accountScoped bool
	// softDeleted opts into the deleted_at-aware probe for families whose
	// destroy never yields a 404: the GET keeps answering 200 with a
	// non-null deleted_at (measured on cloudflared tunnels and teamnet
	// routes at provider v5.23.0 -- the provider's own destroy check asserts
	// deleted_at != nil, and its test sweepers filter on it). Without this,
	// verify-destroyed would report a false "still exists" on every lane.
	softDeleted bool
	// absentStatuses opts into the status-enum absence probe: a 200 whose
	// result.status is one of these counts as gone (certificate packs,
	// custom hostnames, fallback origin -- pending_deletion / deleted).
	absentStatuses []string
	// absentRetries is the number of VerifyAbsent probes, 1s apart. List
	// items delete through an async bulk POST the provider never polls; a
	// single GET immediately after destroy races the 404. Zero/one means
	// a single probe (everyone else's default).
	absentRetries int
	// emptyResultArray opts into the empty-collection absence probe: a
	// 200 whose result is [] counts as gone (snippet rules -- destroy
	// empties the zone's table and the GET keeps answering).
	emptyResultArray bool
	// isDeletedFlag opts into the is_deleted-flag absence probe: a 200
	// whose result.is_deleted is a non-zero number counts as gone. The
	// Workflows API's variant of soft-delete -- its GET response carries
	// is_deleted as a required numeric field (not deleted_at), so neither
	// the plain 404 probe nor softDeleted would see the deletion.
	isDeletedFlag bool
	// revokedAt opts into the revoked_at-aware absence probe: a 200 whose
	// result.revoked_at is set counts as gone. Origin CA certificates
	// REVOKE rather than delete, and a revoked certificate keeps answering
	// GET 200 with revoked_at set (measured live 2026-08-28).
	revokedAt bool
	// zonePathFormat opts a DUAL-SCOPE resource into scope-aware probing:
	// when the deploy's outputs carry a non-empty zone_id, the probe uses
	// this zone-rooted template (zone_id fills its first placeholder, then
	// the declared outputKeys) instead of the account-rooted pathFormat.
	// Cloudflare's dual-scope objects (IP access rules) live in separate
	// per-scope collections -- an account GET cannot see a zone-scoped rule
	// -- and the outputs contract publishes exactly one non-empty scope id
	// per deployment precisely so consumers can derive the scope.
	zonePathFormat string
}

// IDOutputKey returns the last output key -- the resource's own identifier
// (earlier keys are parent scope, e.g. zone_id).
func (v *apiPathVerifier) IDOutputKey() string {
	return v.outputKeys[len(v.outputKeys)-1]
}

func (v *apiPathVerifier) VerifyExists(ctx context.Context, api API, outputs map[string]string) error {
	path, err := v.buildPath(api, outputs)
	if err != nil {
		return err
	}
	exists, err := v.probe(ctx, api, path)
	if err != nil {
		return errors.Wrapf(err, "%s verify-exists failed", v.kind)
	}
	if !exists {
		return errors.Errorf("%s %s not found after deploy (GET %s)",
			v.kind, outputs[v.IDOutputKey()], path)
	}
	return nil
}

func (v *apiPathVerifier) VerifyAbsent(ctx context.Context, api API, outputs map[string]string) error {
	path, err := v.buildPath(api, outputs)
	if err != nil {
		return err
	}
	attempts := v.absentRetries
	if attempts < 1 {
		attempts = 1
	}
	var exists bool
	for i := 0; i < attempts; i++ {
		if i > 0 {
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		var probeErr error
		exists, probeErr = v.probe(ctx, api, path)
		if probeErr != nil {
			return errors.Wrapf(probeErr, "%s verify-absent failed", v.kind)
		}
		if !exists {
			return nil
		}
	}
	return errors.Errorf("%s %s still exists after destroy (GET %s)",
		v.kind, outputs[v.IDOutputKey()], path)
}

// probe selects the existence check: the plain status-code probe, or the
// envelope-aware one for soft-deleting / status-enum families.
func (v *apiPathVerifier) probe(ctx context.Context, api API, path string) (bool, error) {
	if v.softDeleted || len(v.absentStatuses) > 0 || v.emptyResultArray || v.isDeletedFlag || v.revokedAt {
		return api.ResourcePresent(ctx, path, EnvelopePresence{
			SoftDeleted:      v.softDeleted,
			AbsentStatuses:   v.absentStatuses,
			EmptyResultArray: v.emptyResultArray,
			IsDeletedFlag:    v.isDeletedFlag,
			RevokedAt:        v.revokedAt,
		})
	}
	return api.ResourceExists(ctx, path)
}

// buildPath fills the path template from the account scope (when account-
// scoped) and the declared output keys, failing loudly on any missing output
// -- a blank segment would probe the wrong endpoint and report a false
// absence.
func (v *apiPathVerifier) buildPath(api API, outputs map[string]string) (string, error) {
	pathFormat := v.pathFormat
	accountScoped := v.accountScoped
	values := make([]interface{}, 0, len(v.outputKeys)+1)
	if v.zonePathFormat != "" && outputs["zone_id"] != "" {
		// Dual-scope resource deployed zone-scoped: probe the zone
		// collection (the account collection cannot see this object).
		pathFormat = v.zonePathFormat
		accountScoped = false
		values = append(values, outputs["zone_id"])
	}
	if accountScoped {
		if api.AccountID() == "" {
			return "", errors.Errorf("%s is account-scoped but the harness has no account ID", v.kind)
		}
		values = append(values, api.AccountID())
	}
	for _, key := range v.outputKeys {
		value := outputs[key]
		if value == "" {
			return "", errors.Errorf("%s outputs carry no %q -- cannot verify", v.kind, key)
		}
		values = append(values, value)
	}
	return fmt.Sprintf(pathFormat, values...), nil
}

// verifiers maps a component directory name to its verifier. A kind
// registers here in the wave that enrolls it for E2E (its profile +
// scenarios), and every kind that appears in another kind's registry
// prerequisites needs its entry before that consumer's lane can run --
// fixtures are verified right after install.
var verifiers = map[string]Verifier{
	// The zone is the central fixture: most zone-scoped lanes install it as
	// their prerequisite before the component under test deploys.
	"cloudflarednszone": &apiPathVerifier{
		kind:       "cloudflarednszone",
		pathFormat: "zones/%s",
		outputKeys: []string{"zone_id"},
	},
	"cloudflarednsrecord": &apiPathVerifier{
		kind:       "cloudflarednsrecord",
		pathFormat: "zones/%s/dns_records/%s",
		outputKeys: []string{"zone_id", "record_id"},
	},
	// Existence GET is the provider Read path (accounts/.../workers/services/{name}),
	// not workers/scripts/{name}. Account comes from the harness, never from outputs.
	"cloudflareworker": &apiPathVerifier{
		kind:          "cloudflareworker",
		pathFormat:    "accounts/%s/workers/services/%s",
		outputKeys:    []string{"script_name"},
		accountScoped: true,
	},
	"cloudflareloadbalancer": &apiPathVerifier{
		kind:       "cloudflareloadbalancer",
		pathFormat: "zones/%s/load_balancers/%s",
		outputKeys: []string{"zone_id", "load_balancer_id"},
	},
	// The pool is account-scoped (it is the load balancer's registry
	// prerequisite, so its fixture is verified right after install).
	"cloudflareloadbalancerpool": &apiPathVerifier{
		kind:          "cloudflareloadbalancerpool",
		pathFormat:    "accounts/%s/load_balancers/pools/%s",
		outputKeys:    []string{"pool_id"},
		accountScoped: true,
	},
	// Email Routing is a zone singleton: the GET returns the zone's routing
	// settings object rather than 404ing for a zone that never enabled it.
	// MEASURED 2026-08-26 via the REST API: after a disable the GET still
	// answers 200 with enabled=false, so when this kind's lanes unblock (its
	// profile records the upstream provider defect deferring them), the
	// absent-check MUST become a value-aware probe reading `enabled` -- the
	// plain existence probe below would fail verify-absent against a
	// perfectly clean teardown.
	"cloudflareemailroutingzone": &apiPathVerifier{
		kind:       "cloudflareemailroutingzone",
		pathFormat: "zones/%s/email/routing",
		outputKeys: []string{"zone_id"},
	},
	"cloudflareemailroutingrule": &apiPathVerifier{
		kind:       "cloudflareemailroutingrule",
		pathFormat: "zones/%s/email/routing/rules/%s",
		outputKeys: []string{"zone_id", "rule_id"},
	},
	// Destination addresses are account-scoped (shared across zones).
	"cloudflareemailroutingaddress": &apiPathVerifier{
		kind:          "cloudflareemailroutingaddress",
		pathFormat:    "accounts/%s/email/routing/addresses/%s",
		outputKeys:    []string{"address_id"},
		accountScoped: true,
	},
	// Rulesets are dual-scope; the live scenarios are zone-scoped, and the
	// outputs carry zone_id only for zone-scoped rulesets (an account-scoped
	// arm would need an accounts/%s/rulesets/%s variant reading the harness
	// account).
	"cloudflareruleset": &apiPathVerifier{
		kind:       "cloudflareruleset",
		pathFormat: "zones/%s/rulesets/%s",
		outputKeys: []string{"zone_id", "ruleset_id"},
	},
	// The namespace is also the KV pair's install fixture, so this entry
	// serves both its own lanes and the pair's fixture verification.
	"cloudflarekvnamespace": &apiPathVerifier{
		kind:          "cloudflarekvnamespace",
		pathFormat:    "accounts/%s/storage/kv/namespaces/%s",
		outputKeys:    []string{"namespace_id"},
		accountScoped: true,
	},
	// A KV entry's existence GET is the value endpoint (200 with the raw
	// value when present). Keys are path-escaped by the API client; scenario
	// keys stay slash-free so the import ID stays parseable too.
	"cloudflareworkerskvpair": &apiPathVerifier{
		kind:          "cloudflareworkerskvpair",
		pathFormat:    "accounts/%s/storage/kv/namespaces/%s/values/%s",
		outputKeys:    []string{"namespace_id", "key_name"},
		accountScoped: true,
	},
	// Note the singular "database" in the D1 path -- the provider's own Read
	// path, not the plural most Cloudflare collections use.
	"cloudflared1database": &apiPathVerifier{
		kind:          "cloudflared1database",
		pathFormat:    "accounts/%s/d1/database/%s",
		outputKeys:    []string{"database_id"},
		accountScoped: true,
	},
	// The queue is also the R2 event-notification scenario's install fixture
	// (a scenario-declared prerequisite), so this entry serves both.
	"cloudflarequeue": &apiPathVerifier{
		kind:          "cloudflarequeue",
		pathFormat:    "accounts/%s/queues/%s",
		outputKeys:    []string{"queue_id"},
		accountScoped: true,
	},
	// R2's identity is the bucket NAME (no separate id). Jurisdiction rides
	// a cf-r2-jurisdiction HEADER, not the path -- this probe is honest only
	// for default-jurisdiction buckets, which is all the live scenarios
	// create; a non-default arm would need a header-aware client extension.
	"cloudflarer2bucket": &apiPathVerifier{
		kind:          "cloudflarer2bucket",
		pathFormat:    "accounts/%s/r2/buckets/%s",
		outputKeys:    []string{"bucket_name"},
		accountScoped: true,
	},
	"cloudflarehyperdriveconfig": &apiPathVerifier{
		kind:          "cloudflarehyperdriveconfig",
		pathFormat:    "accounts/%s/hyperdrive/configs/%s",
		outputKeys:    []string{"hyperdrive_id"},
		accountScoped: true,
	},
	// The Access kinds are dual-scope (account XOR zone) but their live
	// scenarios are account-scoped -- the common, reusable case. A
	// zone-scoped arm would need zones/%s/access/... variants reading a
	// zone_id the outputs do not carry today (the ruleset precedent:
	// recorded as not-wired, a follow-up if a zone-scoped arm is ever
	// queued).
	"cloudflarezerotrustaccessapplication": &apiPathVerifier{
		kind:          "cloudflarezerotrustaccessapplication",
		pathFormat:    "accounts/%s/access/apps/%s",
		outputKeys:    []string{"application_id"},
		accountScoped: true,
	},
	"cloudflarezerotrustaccessgroup": &apiPathVerifier{
		kind:          "cloudflarezerotrustaccessgroup",
		pathFormat:    "accounts/%s/access/groups/%s",
		outputKeys:    []string{"group_id"},
		accountScoped: true,
	},
	// The policy is also the Access application's self-hosted fixture (a
	// scenario-declared prerequisite), so this entry serves both.
	"cloudflarezerotrustaccesspolicy": &apiPathVerifier{
		kind:          "cloudflarezerotrustaccesspolicy",
		pathFormat:    "accounts/%s/access/policies/%s",
		outputKeys:    []string{"policy_id"},
		accountScoped: true,
	},
	// The tunnel is also the tunnel route's registry-prerequisite fixture.
	// Cloudflared tunnels SOFT-DELETE: the GET answers 200 with deleted_at
	// set after destroy, so the probe is the deleted_at-aware one.
	"cloudflarezerotrusttunnel": &apiPathVerifier{
		kind:          "cloudflarezerotrusttunnel",
		pathFormat:    "accounts/%s/cfd_tunnel/%s",
		outputKeys:    []string{"tunnel_id"},
		accountScoped: true,
		softDeleted:   true,
	},
	// Routes and virtual networks live under teamnet/, not cfd_tunnel/.
	// Routes soft-delete like tunnels; virtual networks 404 honestly.
	"cloudflarezerotrusttunnelroute": &apiPathVerifier{
		kind:          "cloudflarezerotrusttunnelroute",
		pathFormat:    "accounts/%s/teamnet/routes/%s",
		outputKeys:    []string{"route_id"},
		accountScoped: true,
		softDeleted:   true,
	},
	// Also the route's isolated-vnet fixture (a scenario-declared
	// prerequisite), so this entry serves both.
	"cloudflarezerotrusttunnelvirtualnetwork": &apiPathVerifier{
		kind:          "cloudflarezerotrusttunnelvirtualnetwork",
		pathFormat:    "accounts/%s/teamnet/virtual_networks/%s",
		outputKeys:    []string{"virtual_network_id"},
		accountScoped: true,
	},
	// Certificate packs, custom hostnames, and the fallback origin answer
	// GET 200 with status pending_deletion/deleted after destroy rather
	// than 404ing (measured in the provider's own destroy checks at
	// v5.23.0). The status-enum probe treats those as gone.
	"cloudflarecertificatepack": &apiPathVerifier{
		kind:           "cloudflarecertificatepack",
		pathFormat:     "zones/%s/ssl/certificate_packs/%s",
		outputKeys:     []string{"zone_id", "certificate_pack_id"},
		absentStatuses: []string{"pending_deletion", "deleted"},
	},
	"cloudflarecustomhostname": &apiPathVerifier{
		kind:           "cloudflarecustomhostname",
		pathFormat:     "zones/%s/custom_hostnames/%s",
		outputKeys:     []string{"zone_id", "custom_hostname_id"},
		absentStatuses: []string{"pending_deletion", "deleted"},
	},
	// Zone singleton: no resource id -- zone_id is the identity and the
	// only output the verifier keys on (email-routing-settings precedent).
	"cloudflarecustomhostnamefallbackorigin": &apiPathVerifier{
		kind:           "cloudflarecustomhostnamefallbackorigin",
		pathFormat:     "zones/%s/custom_hostnames/fallback_origin",
		outputKeys:     []string{"zone_id"},
		absentStatuses: []string{"pending_deletion", "deleted"},
	},
	// The zone-settings singleton class (settings surfaces that always exist
	// while the zone does; destroy is a NO-OP -- per-class contract in
	// zone_settings.go). Each probes its own family's settings endpoint.
	"cloudflarezonesettings": &settingsSingletonVerifier{
		kind:       "cloudflarezonesettings",
		pathFormat: "zones/%s/settings",
	},
	"cloudflarecachesettings": &settingsSingletonVerifier{
		kind: "cloudflarecachesettings",
		// Smart tiered cache is the family's one surface measured answering
		// on EVERY plan (editable:true on a Free zone even before any
		// write). The natural-looking cache_reserve surface is plan-gated
		// even for reads -- 1135 "not available for your plan type" on Free
		// AND Pro zones (measured 2026-08-27) -- so probing it would fail
		// every honest verify on the free fixture zone.
		pathFormat: "zones/%s/cache/tiered_cache_smart_topology_enable",
	},
	"cloudflarezonetlssettings": &settingsSingletonVerifier{
		kind:       "cloudflarezonetlssettings",
		pathFormat: "zones/%s/ssl/universal/settings",
	},
	"cloudflarelist": &apiPathVerifier{
		kind:          "cloudflarelist",
		pathFormat:    "accounts/%s/rules/lists/%s",
		outputKeys:    []string{"list_id"},
		accountScoped: true,
	},
	// List-item delete is an async bulk POST the provider never polls;
	// a single GET immediately after destroy races the 404.
	"cloudflarelistitem": &apiPathVerifier{
		kind:          "cloudflarelistitem",
		pathFormat:    "accounts/%s/rules/lists/%s/items/%s",
		outputKeys:    []string{"list_id", "item_id"},
		accountScoped: true,
		absentRetries: 5,
	},
	"cloudflareloadbalancermonitor": &apiPathVerifier{
		kind:          "cloudflareloadbalancermonitor",
		pathFormat:    "accounts/%s/load_balancers/monitors/%s",
		outputKeys:    []string{"monitor_id"},
		accountScoped: true,
	},
	// Identity is the project NAME, not a UUID -- the provider's Read is
	// GET accounts/{a}/pages/projects/{name}.
	"cloudflarepagesproject": &apiPathVerifier{
		kind:          "cloudflarepagesproject",
		pathFormat:    "accounts/%s/pages/projects/%s",
		outputKeys:    []string{"project_name"},
		accountScoped: true,
	},
	// Identity is the sitekey. Path segment is challenges/widgets, not
	// turnstile/.
	"cloudflareturnstilewidget": &apiPathVerifier{
		kind:          "cloudflareturnstilewidget",
		pathFormat:    "accounts/%s/challenges/widgets/%s",
		outputKeys:    []string{"sitekey"},
		accountScoped: true,
	},
	// User-scoped (no accounts/ or zones/ prefix). Delete is a REVOKE, and
	// a revoked certificate keeps answering GET 200 with revoked_at set
	// (measured live 2026-08-28: revoke succeeded, the immediate GET
	// returned the full body plus revoked_at) -- the probe reads revoked_at
	// as absence. Revocation reaches reads within seconds (measured 4-15s
	// live 2026-08-28); the retry budget rides out that beat.
	"cloudflareorigincacertificate": &apiPathVerifier{
		kind:          "cloudflareorigincacertificate",
		pathFormat:    "certificates/%s",
		outputKeys:    []string{"certificate_id"},
		revokedAt:     true,
		absentRetries: 15,
	},
	// Access identity providers and service tokens delete for real and 404
	// honestly on the read after (the account-scoped arm is the one the live
	// scenarios run; both resources are dual-scope at the provider).
	"cloudflarezerotrustaccessidentityprovider": &apiPathVerifier{
		kind:          "cloudflarezerotrustaccessidentityprovider",
		pathFormat:    "accounts/%s/access/identity_providers/%s",
		outputKeys:    []string{"identity_provider_id"},
		accountScoped: true,
	},
	"cloudflarezerotrustaccessservicetoken": &apiPathVerifier{
		kind:          "cloudflarezerotrustaccessservicetoken",
		pathFormat:    "accounts/%s/access/service_tokens/%s",
		outputKeys:    []string{"service_token_id"},
		accountScoped: true,
	},
	// Gateway policies delete for real, but the schema carries a computed
	// deleted_at the provider itself never reads -- the soft-delete-aware
	// probe is honest either way: a clean 404 AND a tombstoned 200 with
	// deleted_at set both count as absent (the tunnel/route precedent).
	"cloudflarezerotrustgatewaypolicy": &apiPathVerifier{
		kind:          "cloudflarezerotrustgatewaypolicy",
		pathFormat:    "accounts/%s/gateway/rules/%s",
		outputKeys:    []string{"policy_id"},
		accountScoped: true,
		softDeleted:   true,
	},
	// Zero Trust lists (gateway/lists/, distinct from the rules/lists/ family
	// above) delete for real and 404 honestly.
	"cloudflarezerotrustlist": &apiPathVerifier{
		kind:          "cloudflarezerotrustlist",
		pathFormat:    "accounts/%s/gateway/lists/%s",
		outputKeys:    []string{"list_id"},
		accountScoped: true,
	},
	// IP Access rules delete for real and 404 honestly. The kind is
	// DUAL-SCOPE (account is canonical, the zone arm is scenario-edged) and
	// the two scopes are separate API collections, so the probe follows the
	// deploy's own scope: a non-empty zone_id output selects the zone path.
	"cloudflareipaccessrule": &apiPathVerifier{
		kind:           "cloudflareipaccessrule",
		pathFormat:     "accounts/%s/firewall/access_rules/rules/%s",
		zonePathFormat: "zones/%s/firewall/access_rules/rules/%s",
		outputKeys:     []string{"rule_id"},
		accountScoped:  true,
	},
	// Bot Management is a zone settings singleton: destroy is a NO-OP
	// (empty Delete body). Verify-absent asserts the surface still
	// answers -- the settings-singleton class.
	"cloudflarebotmanagement": &settingsSingletonVerifier{
		kind:       "cloudflarebotmanagement",
		pathFormat: "zones/%s/bot_management",
	},
	// Snippets delete for real (by name) and 404 honestly.
	"cloudflaresnippet": &apiPathVerifier{
		kind:       "cloudflaresnippet",
		pathFormat: "zones/%s/snippets/%s",
		outputKeys: []string{"zone_id", "snippet_name"},
	},
	// Snippet rules are a zone singleton whose destroy empties the
	// table. GET keeps answering 200 with an empty result array -- a
	// plain existence probe would false-fail verify-absent.
	"cloudflaresnippetrules": &apiPathVerifier{
		kind:             "cloudflaresnippetrules",
		pathFormat:       "zones/%s/snippets/snippet_rules",
		outputKeys:       []string{"zone_id"},
		emptyResultArray: true,
	},
	// Standalone health checks delete for real and 404 honestly.
	"cloudflarehealthcheck": &apiPathVerifier{
		kind:       "cloudflarehealthcheck",
		pathFormat: "zones/%s/healthchecks/%s",
		outputKeys: []string{"zone_id", "healthcheck_id"},
	},
	// Waiting rooms delete for real and 404 honestly. The folded
	// bypass-rules list dies with the room.
	"cloudflarewaitingroom": &apiPathVerifier{
		kind:       "cloudflarewaitingroom",
		pathFormat: "zones/%s/waiting_rooms/%s",
		outputKeys: []string{"zone_id", "waiting_room_id"},
	},
	// Waiting-room events delete for real and 404 honestly.
	"cloudflarewaitingroomevent": &apiPathVerifier{
		kind:       "cloudflarewaitingroomevent",
		pathFormat: "zones/%s/waiting_rooms/%s/events/%s",
		outputKeys: []string{"zone_id", "waiting_room_id", "event_id"},
	},
	// Custom SSL deletes for real, but the status enum carries deletion
	// states (deployment and deletion are asynchronous) -- a 200 whose
	// status reads deleted counts as absent.
	"cloudflarecustomsslcertificate": &apiPathVerifier{
		kind:           "cloudflarecustomsslcertificate",
		pathFormat:     "zones/%s/custom_certificates/%s",
		outputKeys:     []string{"zone_id", "certificate_id"},
		absentStatuses: []string{"deleted"},
	},
	// mTLS certificates are account-scoped uploads with a real delete;
	// a deleted certificate answers 400 code 1472 "Certificate not found."
	// (measured 2026-08-28), which the client classifies as absent.
	"cloudflaremtlscertificate": &apiPathVerifier{
		kind:          "cloudflaremtlscertificate",
		pathFormat:    "accounts/%s/mtls_certificates/%s",
		outputKeys:    []string{"certificate_id"},
		accountScoped: true,
	},
	// Authenticated Origin Pulls is a zone-singleton surface: the zone-wide
	// toggle has NO delete at Cloudflare (destroy abandons the value) and
	// associations revert rather than delete -- verify-absent asserts the
	// settings surface still answers, per the settings-singleton contract.
	"cloudflareauthenticatedoriginpulls": &settingsSingletonVerifier{
		kind:       "cloudflareauthenticatedoriginpulls",
		pathFormat: "zones/%s/origin_tls_client_auth/settings",
	},
	// AOP client certificates delete for real but asynchronously: the API
	// answers 200 with pending_deletion/deleted before the record goes.
	// The live arm is the hostname-scoped upload (the zone-scoped surface
	// is plan-proven offline), so the probe speaks the hostname path.
	"cloudflareauthenticatedoriginpullscertificate": &apiPathVerifier{
		kind:           "cloudflareauthenticatedoriginpullscertificate",
		pathFormat:     "zones/%s/origin_tls_client_auth/hostnames/certificates/%s",
		outputKeys:     []string{"zone_id", "certificate_id"},
		absentStatuses: []string{"pending_deletion", "deleted"},
	},
	// Workflows delete for real. Upstream's model documents a GET-200-with-
	// is_deleted answer for deleted workflows, but live (2026-08-27) a
	// deleted workflow answered an honest 404 with code 10200
	// "workflows.api.error.workflow.not_found" -- the flag probe never
	// fired. It stays enabled because it is a strict superset (a 404 is
	// always absent) and upstream's modeling says the flag form exists.
	// Identity is the workflow NAME -- the provider's Read is
	// GET accounts/{a}/workflows/{name}.
	"cloudflareworkflow": &apiPathVerifier{
		kind:          "cloudflareworkflow",
		pathFormat:    "accounts/%s/workflows/%s",
		outputKeys:    []string{"workflow_name"},
		accountScoped: true,
		isDeletedFlag: true,
	},
	// The Secrets Store deletes for real and has a GET-by-id
	// (accounts/{a}/secrets_store/stores/{id} -- confirmed in
	// cloudflare-go v7). One store per account: a create-conflict on an
	// account with an existing store is a lane fact, not a verifier one.
	"cloudflaresecretsstore": &apiPathVerifier{
		kind:          "cloudflaresecretsstore",
		pathFormat:    "accounts/%s/secrets_store/stores/%s",
		outputKeys:    []string{"store_id"},
		accountScoped: true,
	},
	// Store secrets delete for real; the value never round-trips (write-
	// only) but the record itself 404s honestly once gone.
	"cloudflaresecretsstoresecret": &apiPathVerifier{
		kind:          "cloudflaresecretsstoresecret",
		pathFormat:    "accounts/%s/secrets_store/stores/%s/secrets/%s",
		outputKeys:    []string{"store_id", "secret_id"},
		accountScoped: true,
	},
	// AI gateways delete for real. Identity is the user-chosen gateway id
	// (the URL slug). Dynamic routes are folded resources that die with
	// the gateway; the gateway probe is the honest single handle.
	"cloudflareaigateway": &apiPathVerifier{
		kind:          "cloudflareaigateway",
		pathFormat:    "accounts/%s/ai-gateway/gateways/%s",
		outputKeys:    []string{"gateway_id"},
		accountScoped: true,
	},
	// The Zero Trust organization is an ACCOUNT-scoped settings singleton:
	// create==update (PUT upsert) and destroy is a literal no-op at the
	// provider -- verify-absent asserts the organization surface still
	// answers. The folded key-rotation cadence rides the same account.
	"cloudflarezerotrustorganization": &settingsSingletonVerifier{
		kind:       "cloudflarezerotrustorganization",
		pathFormat: "accounts/%s/access/organizations",
		idKey:      "account_id",
	},
	// Infrastructure targets delete for real and 404 honestly (the
	// generated Read removes on 404; no tombstone fields exist).
	"cloudflarezerotrustaccessinfrastructuretarget": &apiPathVerifier{
		kind:          "cloudflarezerotrustaccessinfrastructuretarget",
		pathFormat:    "accounts/%s/infrastructure/targets/%s",
		outputKeys:    []string{"target_id"},
		accountScoped: true,
	},
	// MCP portals delete for real and 404 honestly. Identity is the
	// user-chosen portal id (slug).
	"cloudflarezerotrustmcpportal": &apiPathVerifier{
		kind:          "cloudflarezerotrustmcpportal",
		pathFormat:    "accounts/%s/access/ai-controls/mcp/portals/%s",
		outputKeys:    []string{"portal_id"},
		accountScoped: true,
	},
	// MCP servers delete for real and 404 honestly. Identity is the
	// user-chosen server id; the status enum on the object is sync state,
	// never a deletion tombstone.
	"cloudflarezerotrustmcpserver": &apiPathVerifier{
		kind:          "cloudflarezerotrustmcpserver",
		pathFormat:    "accounts/%s/access/ai-controls/mcp/servers/%s",
		outputKeys:    []string{"server_id"},
		accountScoped: true,
	},
	// The Gateway configuration is an ACCOUNT-scoped settings singleton
	// (both the settings and logging folds have literal no-op deletes);
	// the PAC-file rows delete for real and ride the same lane. The
	// configuration surface is the honest single handle.
	"cloudflarezerotrustgatewaysettings": &settingsSingletonVerifier{
		kind:       "cloudflarezerotrustgatewaysettings",
		pathFormat: "accounts/%s/gateway/configuration",
		idKey:      "account_id",
	},
	// Gateway DNS locations delete for real and 404 honestly (no
	// tombstone fields in the SDK's Location struct).
	"cloudflarezerotrustdnslocation": &apiPathVerifier{
		kind:          "cloudflarezerotrustdnslocation",
		pathFormat:    "accounts/%s/gateway/locations/%s",
		outputKeys:    []string{"location_id"},
		accountScoped: true,
	},
	// The default device profile is an ACCOUNT-scoped settings singleton:
	// create==update (PATCH upsert) and destroy is a literal no-op at the
	// provider -- verify-absent asserts the profile surface still answers.
	// The folded fallback-domain list and zone-certificate toggle are also
	// no-op-destroy surfaces riding the same lane.
	"cloudflarezerotrustdevicedefaultprofile": &settingsSingletonVerifier{
		kind:       "cloudflarezerotrustdevicedefaultprofile",
		pathFormat: "accounts/%s/devices/policy",
		idKey:      "account_id",
	},
	// Custom device profiles delete for real and 404 honestly (no
	// tombstone fields; the folded per-profile fallback list rides the
	// profile and retires with it).
	"cloudflarezerotrustdevicecustomprofile": &apiPathVerifier{
		kind:          "cloudflarezerotrustdevicecustomprofile",
		pathFormat:    "accounts/%s/devices/policy/%s",
		outputKeys:    []string{"policy_id"},
		accountScoped: true,
	},
	// Posture rules delete for real and 404 honestly (the rule's enabled
	// flag is server-computed sync state, never a deletion tombstone).
	"cloudflarezerotrustdeviceposturerule": &apiPathVerifier{
		kind:          "cloudflarezerotrustdeviceposturerule",
		pathFormat:    "accounts/%s/devices/posture/%s",
		outputKeys:    []string{"rule_id"},
		accountScoped: true,
	},
	// Logpush jobs delete for real and 404 honestly. The kind is
	// dual-scope; the verifier matches the live arm's scope (zone-scoped,
	// per the IpAccessRule precedent) with the account arm plan-proven
	// offline. The folded ownership challenge is a one-shot POST with no
	// read surface at all -- nothing to verify, nothing to orphan.
	"cloudflarelogpushjob": &apiPathVerifier{
		kind:       "cloudflarelogpushjob",
		pathFormat: "zones/%s/logpush/jobs/%s",
		outputKeys: []string{"zone_id", "job_id"},
	},
	// Notification policies delete for real and 404 honestly.
	"cloudflarenotificationpolicy": &apiPathVerifier{
		kind:          "cloudflarenotificationpolicy",
		pathFormat:    "accounts/%s/alerting/v3/policies/%s",
		outputKeys:    []string{"policy_id"},
		accountScoped: true,
	},
	// Webhook destinations delete for real and 404 honestly (the type
	// field is a server-side echo inferred from the URL, never a
	// tombstone).
	"cloudflarenotificationwebhook": &apiPathVerifier{
		kind:          "cloudflarenotificationwebhook",
		pathFormat:    "accounts/%s/alerting/v3/destinations/webhooks/%s",
		outputKeys:    []string{"webhook_id"},
		accountScoped: true,
	},
	// Web Analytics sites delete for real and 404 honestly, keyed by the
	// site tag. The folded rules ride the site (deleting the site retires
	// its ruleset and rules).
	"cloudflarewebanalyticssite": &apiPathVerifier{
		kind:          "cloudflarewebanalyticssite",
		pathFormat:    "accounts/%s/rum/site_info/%s",
		outputKeys:    []string{"site_tag"},
		accountScoped: true,
	},
	// Account API tokens delete for real and 404 honestly (expired and
	// revoked are status values on a live object, never deletion
	// tombstones).
	"cloudflareaccountapitoken": &apiPathVerifier{
		kind:          "cloudflareaccountapitoken",
		pathFormat:    "accounts/%s/tokens/%s",
		outputKeys:    []string{"token_id"},
		accountScoped: true,
	},
}

// GetVerifier returns the verifier for a component, or an error if none is
// registered.
func GetVerifier(kind string) (Verifier, error) {
	v, ok := verifiers[kind]
	if !ok {
		return nil, errors.Errorf("no Cloudflare verifier registered for kind %q", kind)
	}
	return v, nil
}
