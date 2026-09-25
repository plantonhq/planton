# IaC Module Parity (Tofu <-> Pulumi)

Every cloud-resource kind ships two IaC implementations under `catalog/.../<kind>/v1/iac/`:
a Pulumi module (`pulumi/module/*.go`) and an OpenTofu module (`tf/*.tf`). For a given
`stack-input` they MUST produce the same cloud objects, names, labels, selectors,
environment, and stack outputs. A divergence here is not cosmetic: it silently changes
what gets deployed depending on which provisioner a resource happens to use.

**Neither engine is the reference.** Both must match the proto contract (`spec.proto` +
`*_outputs.proto`) and the intended behavior. When the two disagree, determine which
is correct against that contract/intent and fix the incorrect one — it can be either engine,
not always Terraform.

This note is the standing "keep an eye out for drift" practice. Read it whenever you
touch a module on either side (or add a new kind).

## What is enforced automatically (don't re-litigate by hand)

- **Stack-outputs conformance** -- `pkg/outputs/conformance_test.go`
  (`TestStackOutputsConformance`). Both engines feed the same generic transformer
  (`pkg/outputs.TransformRaw` -> `Flatten` -> `populateMessage`), so a single bar per
  kind -- "this representative output set fully populates the `StackOutputs` proto with
  nothing left unmapped" -- enforces cross-engine output parity. Add a case for each
  kind whose outputs you care about. You can also dry-run a module ad hoc:
  `planton validate-outputs --kind <Kind> --module-dir <dir> --sample-outputs <json>`,
  or run the full contract check (input surface + outputs + engine validation):
  `planton module verify --kind <Kind> --module-dir <dir>`.
- **Output transform convention** -- emit outputs that flatten to the proto field
  paths. Scalars are plain outputs; nested proto messages (e.g. `KubernetesSecretKey`)
  are emitted as nested objects (`output "password_secret" { value = { name = ..., key = ... } }`),
  which `Flatten` turns into `password_secret.name` / `password_secret.key`. Do NOT emit
  flat names like `password_secret_name` -- they never reach the nested proto field. Of
  the 364 tofu modules, zero use a `transform-outputs` executable or
  `output_transform.yaml`; the generic path is the convention. Reach for an override
  only when an upstream provider's output genuinely cannot be shaped to the proto.

## Manual parity checklist (the hand-written logic no tool can diff)

When changing a provider-resource module (`locals.tf`/`database.tf`/`main.tf` and the
matching `pulumi/module/*.go`), confirm both sides agree on:

- **Namespace source.** Use `spec.namespace` (NOT a resource id or a derived name).
- **Resource naming basis.** Both engines name the created objects (operator CRs, pod
  annotations, secret names) off the SAME field -- `metadata.name` is the established
  basis. Don't introduce a parallel `metadata.id`-based name on one side.
- **Reference resolution source.** Both engines consume the same resolved spec
  references (`StringValueOrRef` fields arrive as plain strings). Never substitute a
  provider data-source lookup on one side for a spec field the other side reads -- the
  two resolution paths can disagree (a wildcard-location cluster lookup vs the spec's
  location reference), and the ignored spec field becomes dead config on one engine
  only, which is a spec-feature-coverage defect.
- **Bridged-provider-only client-side flags.** The Pulumi provider can bridge a
  NEWER upstream line than the released GA provider the Terraform module pins. A
  client-side safety flag that exists only on that newer line (e.g. a
  deletion-protection flag that defaults TRUE) silently changes one engine's
  lifecycle behavior -- typically blocking destroy on Pulumi while Terraform
  destroys cleanly. When the released GA schema has no such attribute, explicitly
  neutralize it in the Pulumi module with a comment explaining why; never let a
  bridged-only default decide destroy semantics on one engine.
- **Bridged-provider attribute value format.** The same computed attribute can come
  back in DIFFERENT string forms on the two engines: the bridged Pulumi provider may
  return a fully qualified resource path (e.g. `projects/{p}/instanceConfigs/{name}`)
  where the released Terraform provider stores the plain name the spec passed in.
  Exporting the raw attribute as a stack output then breaks output parity — and any
  API caller or verifier consuming the output — on one engine only, invisibly to every
  offline gate. Decide the output contract from the spec's vocabulary (usually the
  plain name), normalize the divergent engine with a comment, and prove with a live
  run on BOTH engines, not just `validate-outputs` (which checks shape, not value).
- **Cross-resource addressing.** Both engines consume the spec's resolved reference
  fields (`StringValueOrRef` arrives as a plain string) to address a parent or sibling
  resource. Never re-discover a parent on one engine via a provider data-source lookup
  (e.g. a wildcard-location search by name) while the other reads the spec field -- the
  two resolution paths diverge the moment names collide across locations or projects,
  and the data-source path silently ignores the spec contract.
- **Optional-output exports in the Pulumi module.** Two failure modes surface only at
  live deploy (never at `go build`, never in `pulumi preview`'s early phase):
  (1) an `ApplyT` callback whose input type mismatches the bridged field's
  optionality — many bridged computed attributes are `*string`/`*int` even when they
  feel always-present, and a `func(v string)` callback panics at runtime
  ("applier's first input parameter must be assignable from *string"); and
  (2) chaining lazy accessors through a possibly-empty nested list
  (`.NetworkInterfaces.Index(0).AccessConfigs().Index(0)...`) — the index panics
  the whole program when the list is empty (e.g. a private VM with no access
  config). Export optional nested outputs with ONE struct-slice `ApplyT` carrying
  explicit `len(...)`/nil guards that degrade to `""`, and mirror the same
  empty-value contract in the Terraform module with `try(..., "")` so both engines
  export identical values for the absent case.
- **Labels.** Same keys and values. The resource-identity labels are the
  `kuberneteslabelkeys` set (`planton.ai/resource`, `planton.ai/name`, `planton.ai/kind`,
  `planton.ai/id`, `planton.ai/organization`, `planton.ai/environment`); the kind value
  is the `CloudResourceKind` enum string (e.g. `KubernetesPostgres`), and the id label is
  present only when `metadata.id` is set.
- **Pod / service selectors.** Selectors must match the labels the operator/helm chart
  actually puts on the workload pods (e.g. Zalando/Spilo pods are `application: spilo`),
  NOT our resource-identity labels. A wrong selector matches zero pods and silently
  breaks connectivity while still "succeeding".
- **Spec feature coverage.** Every behavior on one side exists on the other: backup,
  restore/standby, ingress, env injection, resource sizing, etc. The proto `spec` is the
  contract -- if it has a field, both modules must honor it.
- **Outputs shape.** Both engines export the same `StackOutputs` field set (see the
  automated conformance guard above).

## variables.tf (generator-owned where migrated, curated elsewhere)

`planton tofu generate-variables <Kind>` (`pkg/iac/tofu/generators`) renders
`variables.tf` from the spec proto in the committed `optional()` convention: every
attribute the tfvars renderer may omit (it emits populated fields only) is `optional()`
with a zero-value default, and an attribute stays required only when the proto marks the
field required (see the generators package doc for the full contract).

- For kinds in the drift gate's migrated set (`TestVariablesTFDrift` in
  `pkg/iac/tofu/generators`), the committed `variables.tf` IS the generator output —
  regenerate with `PLANTON_REGEN_VARIABLES=1`, never hand-edit.
- For the rest, the file is still hand-curated: when a spec field is added, add the
  matching attribute (in the `optional()` style) so partial tfvars still apply. Diffing
  against `generate-variables` output is a quick way to spot a missing field.
- `planton module verify --kind <Kind> --module-dir <dir>` checks any module's declared
  input surface against the kind's schema, and that every secret home the kind declares
  (`secret_home` on a field every viewer reads) is read by the module in both engines,
  with severities tied to what actually breaks a deployment. Its fleet test proves every
  official module reads its kind's secret homes.

## Worked example

The Postgres modules render the CloudNativePG family (Cluster + Barman Cloud
ObjectStores + ScheduledBackups + declared-credential Secrets) from one naming
root (`metadata.name`), with nested secret-handle outputs
(`username_secret{name,key}` / `password_secret{name,key}`) pointing at the
EFFECTIVE application Secret — the operator-generated `<name>-app` normally,
the module-provided `<name>-app-provided` when initdb declares an owner
password. See the conformance guard's `KubernetesPostgres` case and its
negative counterpart `TestStackOutputsConformance_DetectsFlatSecretDrift`
(which proves flat `password_secret_name`-style outputs are caught).

The `Auth0Client` `jwt_configuration.alg` default is another spec-feature-coverage parity
case: the proto documents `Default: RS256`, but both engines previously passed an omitted
`alg` through (Auth0 then defaulted HS256, which JWKS-verifying clients like NextAuth reject).
Both engines now encode the default -- tofu via `alg = optional(string, "RS256")` in
`variables.tf` (beside `secret_encoded = optional(bool, false)`), Pulumi via an else-RS256 in
`client.go`. The default is module-level rather than a proto `(options.default)` because the
proto-default applier (`internal/manifest/protodefaults.ApplyDefaults`) runs only in the CLI
manifest loader, not on the tfvars-render path used by orchestrated deploys (`pkg/iac/tofu/generators/tfvars.go`
prunes unset fields). `alg` is not a stack output, so the conformance guard is unaffected.

The `KubernetesPostgres` **backup object-store credentials** (the declared-key
arms of `spec.backup.object_store`) are a spec-feature-coverage +
secret-handling parity case. Both engines, when keys are declared, materialize
one deterministic Secret (`<name>-backup-creds`, plus `<name>-region` and
`<name>-backup-endpoint-ca` satellites where the arm needs them) and wire the
Barman Cloud ObjectStore's `SecretKeySelector` references at it — never
plaintext in a rendered custom resource. The keyless arms render the backend's
ambient-identity flag (`inheritFromIAMRole` / `gkeEnvironment` /
`inheritFromAzureAD`) and create no credential Secret at all. Secret names,
key names, and the selector wiring must match byte-for-byte across engines
because the ObjectStore CR the plugin reads is the same object either engine
applies. `secret_access_key`/`storage_key`/`connection_string`/
`service_account_key_json` carry `(options.sensitive) = true`;
`access_key_id` is an identifier (the secret-coverage heuristic does not flag
the `_id` suffix), so it needs no annotation. None of these are stack outputs,
so the conformance guard is unaffected.

The `CloudflareR2Bucket` module pins the Cloudflare provider to v5 on both engines (tofu
`~> 5.0`, Pulumi `sdk/v6`) and provisions the bucket plus its bucket-scoped sub-resources in one
module. The `location` hint is the enum value used verbatim as the provider string
(`wnam`/`enam`/`weur`/`eeur`/`apac`/`oc`); `auto` (the enum zero value) means "no hint" and is
omitted on both sides (tofu sets `null`, Pulumi leaves the `Location` arg unset). `jurisdiction`
(a validated string) and `storage_class` (an enum used verbatim) are likewise omitted when empty so
the provider applies its defaults, and `jurisdiction` is passed to every sub-resource so the whole
bucket shares one jurisdiction. `public_access` provisions `cloudflare_r2_managed_domain`
(`enabled = true`) and surfaces the r2.dev domain. `custom_domains` is a list: each enabled entry
becomes one `cloudflare_r2_custom_domain` (tofu `for_each` keyed by domain, Pulumi a loop), with the
v5 attrs `domain`/`zone_id`/`enabled = true` plus optional `min_tls`/`ciphers`; `zone_id` is a
`StringValueOrRef` resolved to a plain string before tfvars. CORS, lifecycle, and lock are each a
single sub-resource created only when their `rules` list is non-empty; the abort-multipart transition
is always an `Age` condition and storage-class transitions always target `InfrequentAccess` (the sole
supported class), hard-set identically on both engines. Stack outputs are the proto fields
`bucket_name`, `bucket_url` (the path-style `https://<account_id>.r2.cloudflarestorage.com/<bucket>`
S3 URL), `custom_domain_urls` (one per enabled custom domain), and `public_url` (the r2.dev domain
when public access is enabled) — see the conformance guard's `CloudflareR2Bucket` case.

The Workers family (`CloudflareWorker`, `CloudflareKvNamespace`, `CloudflareWorkersKvPair`,
`CloudflareD1Database`, `CloudflareHyperdriveConfig`) pins the Cloudflare provider to v5 on both
engines. `CloudflareWorker` models bindings as grouped, type-specific lists (the wrangler.toml grain);
both engines flatten them into the provider's single discriminated `bindings` array (tofu builds
uniform objects via `merge(null_attrs, ...)`, Pulumi appends `WorkersScriptBindingArgs`), each cross-
resource binding resolving a `StringValueOrRef` to a plain id. The script source is a oneof — inline
`content` or an R2 `r2_bundle` fetched through the S3-compatible provider (the AWS provider is only
configured on the bundle path). Routing folds onto the worker as `cloudflare_workers_script_subdomain`
(workers.dev), `cloudflare_workers_custom_domain` (one per hostname; `environment` is deprecated and omitted),
and `cloudflare_workers_route` (one per pattern); cron schedules fold onto
`cloudflare_workers_cron_trigger`. Stack outputs are `script_id`, `script_name`,
`custom_domain_hostnames`, `route_patterns`, plus keyed maps `custom_domain_ids` (by hostname),
`route_ids` and `route_zone_ids` (by list index) so import can reassemble `{account_id}/{domain_id}`
and `{zone_id}/{route_id}`. Bindings cover the provider's full type list (wrangler.toml grain);
`r2_bundle.bucket`, Durable Object `script_name`, `tail_consumers.service`, dispatch outbound
`service`, transferred-class `from_script`, and VPC `tunnel_id` are Worker/R2/tunnel FKs.
`migrations` (Durable Object class create/rename/transfer/delete) and nested `observability.logs` /
`observability.traces` are honored by both engines. **PARITY-EXCEPTION**: Pulumi SDK v6.17.0 has no
inputs for `cache_options`, `exports`, `package_dependencies`, rate-limit `mitigation_timeout`, or
traces `propagation_policy` — tofu honors them; Pulumi logs and skips. **Workers Static Assets** (`spec.assets`)
fold onto the same `cloudflare_workers_script` as the `assets` block: both engines set
`directory` and the `config` sub-fields (`html_handling`, `not_found_handling`, `headers`,
`redirects`) identically, and model `run_worker_first` as the provider's dynamic field — a
`[]string` of path rules when `run_worker_first_rules` is set, else a bool from
`run_worker_first` (full parity on v6.17.0, where `WorkersScriptAssetsConfig.RunWorkerFirst`
is `interface{}`; mutually-exclusive by CEL). When `assets` is set without a script source the
Worker is assets-only: both engines omit `content`/`main_module`. `assets.binding_name` appends
an `assets`-type entry to the shared bindings list so a full-stack worker can read assets via
`env.<NAME>`. No new stack outputs (the workers.dev URL is not derivable — the provider exposes
no account-subdomain lookup). The provider pins the Pulumi Cloudflare SDK at
**v6.17.0**, and tofu↔Pulumi are at **full parity** across the family: D1 `jurisdiction`, the worker
service-binding `entrypoint`, worker `limits.subrequests`, the worker custom-domain `zone_id`, and the
DNS-record `private_routing` are all modeled in the proto and honored by both engines (these were
briefly deferred against the older v6.10.1 SDK, then restored on the upgrade — see
`coding-guidelines/0004` in the project for the standing principle: the proto stays future-proof, the
lagging engine is upgraded or degraded-and-documented, never held back with proto `reserved`).
Hyperdrive's `origin.service_id` (egress through a Workers VPC Service for a private origin) is
modeled and honored by both engines (tofu `main.tf` origin block, Pulumi `originArgs.ServiceId`),
omitted when empty; it is mutually exclusive with the spec-level `mtls` block by a message CEL
(TLS is managed on the VPC Service) and is not a stack output, so the conformance guard is
unaffected. Hyperdrive's `origin.password`/`origin.access_client_secret` and the worker `secrets[].value` are
`StringValueOrRef + (sensitive)`. See the conformance guard's `CloudflareWorker`,
`CloudflareKvNamespace`, `CloudflareWorkersKvPair`, `CloudflareD1Database`, and
`CloudflareHyperdriveConfig` cases.

The Load Balancing family (`CloudflareLoadBalancer`, `CloudflareLoadBalancerPool`,
`CloudflareLoadBalancerMonitor`) pins the Cloudflare provider to v5 on both engines and mirrors
Cloudflare's own resource topology: the monitor and pool are account-scoped, reusable resources, and
the load balancer is zone-scoped and references pools by id/`StringValueOrRef`. `CloudflareLoadBalancer`
carries the full v5 steering surface — `default_pools`/`fallback_pool`, `steering_policy`,
`session_affinity` (+ `session_affinity_attributes`), `region/country/pop_pools` (modeled as
`[{code, pool_ids[]}]` and rebuilt into the provider's `{code => pool_ids}` map by both engines),
`adaptive_routing`, `location_strategy`, and `random_steering`; the `rules[]` beta surface is a recorded
skip. Both engines omit the `none`/`off` enum defaults so the provider applies its own, and
`load_balancer_cname_target` resolves to the hostname (not the opaque LB id). `CloudflareLoadBalancerPool`
carries origins (each `address` a `StringValueOrRef` with no fixed kind, plus weight/port/host-header/
virtual-network/flatten-cname), a `monitor` reference, `check_regions`, `load_shedding`,
`origin_steering`, and `notification_filter`; `monitor_group` is reserved. `CloudflareLoadBalancerMonitor`
carries the full probe surface (type, path/codes/body/method/headers, port, interval/timeout/retries,
consecutive up/down, follow-redirects, allow-insecure, probe-zone) with a CEL rule requiring a port for
tcp/udp_icmp/smtp. The Pulumi SDK is pinned at **v6.17.0** and tofu↔Pulumi are at **full parity** across
the family (no deferrals). The family has no secret-bearing fields. See the conformance guard's
`CloudflareLoadBalancer`, `CloudflareLoadBalancerPool`, and `CloudflareLoadBalancerMonitor` cases.

The Zero Trust Access family (`CloudflareZeroTrustAccessApplication`, `CloudflareZeroTrustAccessPolicy`,
`CloudflareZeroTrustAccessGroup`) pins the Cloudflare provider to v5 on both engines and mirrors
Cloudflare's own resource topology: a reusable account/zone-scoped **group** (a named bundle of access
rules) is referenced by a reusable account-scoped **policy** (decision + rules), which is referenced by
the **application** (the protected resource) via `policies[]` (`StringValueOrRef` → policy id). Policy and
group share an identical `CloudflareAccessRule` oneof (26 variants: identity, network/device, service
token, user-risk, and external evaluation) modeled independently in each component (the codebase has no
cross-component proto imports); the Terraform modules pass the rule lists straight through (proto field
names match the provider 1:1, including the nested `user_risk_score.user_risk_score`), while the Pulumi
modules map each variant explicitly. The application carries the full v5 surface — typed `type` enum,
`destinations`, app-launcher visuals, self-hosted cookie/CORS/interstitial controls, `mfa_config`,
`oauth_configuration`, `target_criteria` (with `target_attributes` rebuilt into the provider's
`{name => values}` map), and the deep `saas_app` (SAML + OIDC) and `scim_config` subtrees — and exports
`application_id`, `aud`, `domain`, and the SaaS signing/SSO material. Secret-bearing SCIM authentication
fields (`password`, `token`, `client_secret`) are `(sensitive)`.

**One tofu↔Pulumi parity gap (documented):** the `cloudflare_account_member` access-rule variant exists in
the Terraform provider (v5.21.1) but **not in the Pulumi Cloudflare SDK (v6.17.0)** for group/policy rules.
The proto models it (full source-of-truth) and the Terraform modules provision it; the Pulumi modules log a
warning and skip that one variant. Every other field is at full parity. When a newer Pulumi SDK exposes
`ZeroTrustAccess{Group,Policy}Include/Exclude/Require.CloudflareAccountMember`, wire it in and remove the
note (see each component's Pulumi `README.md`). See the conformance guard's
`CloudflareZeroTrustAccessApplication`, `CloudflareZeroTrustAccessPolicy`, and
`CloudflareZeroTrustAccessGroup` cases.

The `CloudflareRuleset` component carries the full v5 `cloudflare_ruleset` surface: the 20-value action
set, rule-level `ratelimit` / `logging` / `exposed_credential_check`, and the deep `action_parameters`
tree — `set_config` (SSL/security-level/Polish/Rocket Loader/autominify/…), the full cache surface
(`cache_key.custom_key` cookie/header/host/query_string/user, `cache_reserve`, `edge_ttl`/`browser_ttl`,
`additional_cacheable_ports`, strip/respect toggles), the `set_cache_control` directives (modeled with
three reusable shapes), `set_cache_tags`, `log_custom_field` lists, `from_list`, `algorithms`,
`matched_data`, `increment`, `serve_error`, and all five skip surfaces (`phases`, `products`, `ruleset`,
`rulesets`, and the per-rule `rules` map of ruleset ID → rule IDs — both engines unwrap the proto's
list-wrapper map values onto the provider's map-of-lists, CEL-validated as 32-hex IDs and incompatible
with the single `ruleset` option, mirroring the provider's own validators). Value-set fields (modes,
operations, sensitivity levels, content types, SSL/security-level/Polish/body-buffering) are
CEL-validated. `exposed_credential_check.password_expression` is a wirefilter expression locating the
password, not a secret, so it carries `sensitive_exempt_reason`.

**One tofu↔Pulumi parity gap (documented):** `action_parameters.vary` (variant caching keyed on response
headers) exists in the Terraform provider (v5.21.1) but **not in the pulumi-cloudflare SDK (v6.17.0)** —
there is no `vary` field on `RulesetRuleActionParametersArgs`. The proto models it (future-proof source of
truth) and the Terraform module provisions it; the Pulumi module omits it with an inline note. When a newer
Pulumi SDK exposes `RulesetRuleActionParameters.Vary`, wire it in and remove the note (see the ruleset
Pulumi `README.md`). Every other ruleset field is at full parity.

The Email Routing family (`CloudflareEmailRoutingZone`, `CloudflareEmailRoutingRule`,
`CloudflareEmailRoutingAddress`) pins the Cloudflare provider to v5 on both engines. The zone kind folds
three zone singletons: `cloudflare_email_routing_settings` (create IS enable, destroy IS disable),
`cloudflare_email_routing_catch_all` (whose provider Delete is a genuine no-op — both modules carry the
note), and `cloudflare_email_routing_dns` (instantiated by `lock_dns_records`; `dns_name` targets a
subdomain). Rules and the catch-all model the provider's generic `{type, value[]}` action LIST as typed
actions with FKs (`forward_to` → address, `worker` → Worker); both engines map each typed action back,
so one rule can forward AND invoke a Worker. The catch-all's `matchers` argument is module-supplied
(`[{type = "all"}]` — the only value the provider's validator accepts). **PARITY-EXCEPTION**: the
address's `status` (explicit verification-state override, `unverified`/`verified`) exists in the
Terraform provider (v5.23.0) but **not in the pulumi-cloudflare SDK (v6.17.0)** — `EmailRoutingAddressArgs`
carries only `AccountId` and `Email`; tofu honors it, Pulumi omits it with an inline note. Every other
field is at full parity. Stack outputs: zone `{zone_id, enabled, status, name}` (the zone id is the
import identity of all three singletons), rule `{rule_id, zone_id}`, address
`{address_id, email, verified, created}`.

The `CloudflareQueue` component models the queue plus its single (folded) consumer; the Pulumi SDK
(v6.17.0) and Terraform provider (v5.21.1) are at **full parity** (`cloudflare_queue` +
`cloudflare_queue_consumer`, both engines). The consumer is folded onto the queue because at the resource
level a queue has exactly one consumer with no independent lifecycle (the module still provisions the
separate consumer resource). The queue has no secret-bearing fields. The v5 API caps
`message_retention_period` at 86400s (1 day) despite docs implying 14 days — the CEL matches the API.
The Worker `queues` producer binding and the R2 `event_notifications` reference a `CloudflareQueue` by name
and id respectively, at full parity on both engines. See the conformance guard's `CloudflareQueue` case.

The `CloudflarePagesProject` component manages the Pages **project** (build config, optional git source,
per-environment deployment configs, and folded custom domains via `cloudflare_pages_domain`); it never
manages deployments, because the Cloudflare provider (v5.21.1) and Pulumi SDK (v6.17.0) expose no Pages
deployment resource — versions are produced out-of-band (git push for git-connected projects, or
`wrangler pages deploy` for direct-upload). Both engines are at **full parity** on the project surface, with
three behaviors that BOTH engines implement identically (learned from the live API):
1. **Env-var grain.** The proto splits env vars into plain `vars` + secret `secrets` (the secret value is
   `(sensitive)`); both engines recombine them into the provider's single `env_vars` map keyed by name, with
   `type = plain_text`/`secret_text`. This keeps the secret annotation static (vs the provider's conditionally
   sensitive `{type,value}`).
2. **Paired environments.** Cloudflare rejects a project whose `preview` and `production` configs are
   inconsistent (e.g. `fail_open` must match). When only one environment is supplied, both engines mirror it
   to the other (tofu via `dc_*_src` coalescing in `locals.tf`, Pulumi via the mirror in
   `deployment_config.go`).
3. **Empty binding maps are omitted, not `{}`.** The provider normalizes an empty map to null and flags an
   inconsistent apply otherwise; both engines send null for any empty binding group (tofu via
   `length(...) > 0 ? {...} : null`, Pulumi by only assigning a non-empty `...Map`). Bindings resolve a
   `StringValueOrRef` to a plain id (KV/D1/R2/Queue/Hyperdrive/Worker). Stack outputs are `project_name`,
   `subdomain`, `domains`, and `created_on` (no deployment-level outputs — none exist at provision time). The
   web-analytics token and secret env values are `(sensitive)`. `CloudflarePagesProject` is **not** marked
   `is_service_kind`: with the git-connected model Cloudflare is the deployer, so Service Hub drives no version
   deploys. See the conformance guard's `CloudflarePagesProject` case.
