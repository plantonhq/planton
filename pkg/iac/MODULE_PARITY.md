# IaC Module Parity (Tofu <-> Pulumi)

Every cloud-resource kind ships two IaC implementations under `apis/.../<kind>/v1/iac/`:
a Pulumi module (`pulumi/module/*.go`) and an OpenTofu module (`tf/*.tf`). For a given
`stack-input` they MUST produce the same cloud objects, names, labels, selectors,
environment, and stack outputs. A divergence here is not cosmetic: it silently changes
what gets deployed depending on which provisioner a resource happens to use.

**Neither engine is the reference.** Both must match the proto contract (`spec.proto` +
`*_stack_outputs.proto`) and the intended behavior. When the two disagree, determine which
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
  `openmcf validate-outputs --kind <Kind> --module-dir <dir> --sample-outputs <json>`.
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

## variables.tf (a generated *scaffold*, curated in practice)

`openmcf tofu generate-variables <Kind>` (`pkg/iac/tofu/generators`) renders a starting
`variables.tf` from the spec proto, but the committed convention is the curated
`optional()` form (used by the large majority of modules, e.g. `kubernetesnamespace`,
`kubernetescronjob`). The generator's raw output makes every field required, which is not
runtime-compatible with the generated `terraform.tfvars` (it omits unset fields). So:

- Treat the generator as a reference for *coverage*, not a file to commit verbatim.
- When a spec field is added, add the matching `variable` (in the curated `optional()`
  style) so partial tfvars still apply. Diffing against `generate-variables` output is a
  quick way to spot a missing field.

## Worked example

The Postgres tofu module was brought to parity with its Pulumi counterpart (correct
namespace source, `metadata.name` naming basis, `application: spilo` LB selector,
`planton.ai/*` labels, backup + disaster-recovery standby/env, and nested secret
outputs). See the conformance guard's `KubernetesPostgres` case and its negative
counterpart `TestStackOutputsConformance_DetectsFlatSecretDrift`.

The `Auth0Client` `jwt_configuration.alg` default is another spec-feature-coverage parity
case: the proto documents `Default: RS256`, but both engines previously passed an omitted
`alg` through (Auth0 then defaulted HS256, which JWKS-verifying clients like NextAuth reject).
Both engines now encode the default -- tofu via `alg = optional(string, "RS256")` in
`variables.tf` (beside `secret_encoded = optional(bool, false)`), Pulumi via an else-RS256 in
`client.go`. The default is module-level rather than a proto `(options.default)` because the
proto-default applier (`internal/manifest/protodefaults.ApplyDefaults`) runs only in the CLI
manifest loader, not on the tfvars-render path used by orchestrated deploys (`pkg/iac/tofu/generators/tfvars.go`
prunes unset fields). `alg` is not a stack output, so the conformance guard is unaffected.

The `KubernetesPostgres` per-database **backup R2 credentials** (`spec.backup_config.credentials`,
plus the symmetric `spec.backup_config.restore.credentials`) are a spec-feature-coverage +
secret-handling parity case. Both engines, when `credentials` is set, create a Kubernetes Secret
holding `access_key_id`/`secret_access_key` and inject the credentials into the Spilo `spec.env`
via `secretKeyRef` (never plaintext in the postgresql CR/pod) -- Pulumi in `backup_config.go`
(`r2CredentialEnvVars`, shared with `restore_config.go`), tofu via `kubernetes_secret_v1` in
`credentials.tf` referenced from the `valueFrom.secretKeyRef` env entries in `locals.tf`. The
backup target is composed identically on both engines from `backup_config.bucket` +
`backup_config.object_prefix` as `WALG_S3_PREFIX = s3://<bucket>/<object_prefix>/$(SCOPE)/$(PGVERSION)`
(the `object_prefix` segment is dropped when empty), and restore composes
`s3://<bucket>/<object_prefix>` for the standby `s3_wal_path`. The non-secret backup env
(`AWS_ENDPOINT`/`AWS_REGION=auto`/`AWS_FORCE_PATH_STYLE=true`/`WALG_S3_PREFIX`/`USE_WALG_BACKUP`/
`USE_WALG_RESTORE`/`BACKUP_SCHEDULE`/`BACKUP_NUM_TO_RETAIN`) and the standby `STANDBY_AWS_*` set
match across engines, as does the standby-env-first / backup-env-second merge order.
`backup_config.bucket` is a `StringValueOrRef` resolved to a plain bucket name before tfvars (like
`spec.namespace`), so both engines receive an identical literal. `secret_access_key` carries
`(options.sensitive) = true`; `access_key_id` is an identifier (the secret-coverage heuristic does
not flag the `_id` suffix), so it needs no annotation. The cluster-wide
`KubernetesZalandoPostgresOperator` backup config mirrors the same `bucket` + `object_prefix` +
`credentials` shape, composing the identical `WALG_S3_PREFIX` into its operator configmap on both
engines. None of these are stack outputs, so the conformance guard is unaffected.

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
(workers.dev), `cloudflare_workers_custom_domain` (one per hostname, `environment = "production"`),
and `cloudflare_workers_route` (one per pattern); cron schedules fold onto
`cloudflare_workers_cron_trigger`. Stack outputs are `script_id`, `script_name`,
`custom_domain_hostnames`, and `route_patterns`. The provider pins the Pulumi Cloudflare SDK at
**v6.17.0**, and tofu↔Pulumi are at **full parity** across the family: D1 `jurisdiction`, the worker
service-binding `entrypoint`, worker `limits.subrequests`, the worker custom-domain `zone_id`, and the
DNS-record `private_routing` are all modeled in the proto and honored by both engines (these were
briefly deferred against the older v6.10.1 SDK, then restored on the upgrade — see
`coding-guidelines/0004` in the project for the standing principle: the proto stays future-proof, the
lagging engine is upgraded or degraded-and-documented, never held back with proto `reserved`).
Hyperdrive's `origin.password`/`origin.access_client_secret` and the worker `secrets[].value` are
`StringValueOrRef + (sensitive)`. See the conformance guard's `CloudflareWorker`,
`CloudflareKvNamespace`, `CloudflareWorkersKvPair`, `CloudflareD1Database`, and
`CloudflareHyperdriveConfig` cases.
