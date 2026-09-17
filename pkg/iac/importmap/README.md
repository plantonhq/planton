# importmap — import-ID recipes, loaded and resolved

State import needs two pieces of knowledge per resource: **which address**
the engine tracks it under, and **which identifier** the provider imports it
by. Addresses are never authored — the engine enumerates them per spec at
import time (a read-only preview lists them), which handles
module-constructed names, repeated (`for_each`/`count`) resources, and
conditional resources by construction. Only the identifier knowledge is
authored, in two tiers:

| Tier | Kind | Location | Owns |
|------|------|----------|------|
| Provider | `ProviderImportCatalog` | `catalog/{provider}/aa_import/catalog.yaml` | Import-ID **format** per resource type (`"{bucket}"`, `"{vpc_id}"`) — or, for types whose upstream provider ships no importer, the recorded `not_importable_upstream_reason` — plus `config_only_attributes`, attributes that exist only in IaC configuration and can never round-trip through import |
| Component | `ComponentImportMap` | `{component}/v1/iac/import-map.yaml` | The **value source** per `{placeholder}`: metadata.name (optionally with a literal suffix or prefix, for convention-named satellites like `<name>-hpa` and provider-composed IDs like `table/<name>`), a spec field, a stack output, a pasted ARN's part, the enumerated address's instance key, or a module-hardcoded literal (a typed-CR module's apiVersion/kind) — with "where to find this" guidance for anything only the user can supply |

An id_format has two optionality forms, with deliberately different literal
handling: `{name?}` renders as the empty string when unresolved and KEEPS its
surrounding literal delimiters (the DynamoDB `index:{index_name?}/` shape),
while a bracketed segment group like `[//{namespace}]` disappears WHOLESALE —
delimiters included — when its placeholder does not resolve. The group form
exists for composed IDs whose provider rejects a trailing delimiter:
`kubectl_manifest` imports namespaced CRs as
`apiVersion//kind//name//namespace` and cluster-scoped CRs as the 3-part
`apiVersion//kind//name`, never with a trailing `//`.

When a module declares SEVERAL resources of one type whose ID placeholders
carry different values (a control-plane module installing multiple Helm
releases — each release's `{release_name}` is a different constant), a value
declaration can be scoped to one Terraform logical resource name via
`tofu_resource_name`. At resolve time the declaration scoped to the address's
logical name wins; addresses without a scoped declaration fall back to the
unscoped declaration of the same placeholder. The offline conformance guard
rejects a scope that names no real module resource (a typo'd scope would
otherwise silently fall back and import the wrong resource).

A module that applies a MULTI-GVK manifest bundle through one `for_each`
resource (a release-manifest operator install) keys its instances by each
document's own composed identity — `apiVersion//kind//name[//namespace]` —
and derives every composed-ID placeholder from a segment of the key itself
via `from_address_key_segment` (0-based, `//`-delimited; an index past the
key's segment count resolves empty, so the namespace group drops for
cluster-scoped documents). Per-document literals cannot serve there: the
bundle spans many apiVersion/kind pairs. Single-GVK typed-CR modules keep
their literal declarations — the segment arm exists for the bundle class.

A keyed satellite whose import ID is CLOUD-GENERATED — the instance key is
config-time identity but the ID is assigned at creation (a VPC secondary
CIDR's `vpc-cidr-assoc-...` association id keyed by the CIDR, a KMS grant's
generated grant id keyed by list position) — derives via
`from_stack_output_keyed_by_address`: the module exports a map output keyed
by the SAME key as the resource's `for_each` instances, and the enumerated
address selects the entry (the flattened `<output>.<address key>` lookup).
`from_address_key` cannot serve these (the key is not the ID), and a plain
`from_stack_output` cannot either (one static key cannot vary per
instance). The offline guard requires the named output field to exist on
the kind's StackOutputs AND be a map; keep a `where_to_find` paste fallback
for fresh adoption, where no prior deploy's outputs exist.

An import ID that IS secret material — the canonical case is a
`random_password` resource, whose provider import ID is the password value
itself — derives through `from_cluster_secret_key`: the module materialized
the value into a convention-named Kubernetes Secret
(`<metadata.name><name_suffix>`, the `from_metadata_name_suffix` naming
convention), and a cluster-connected resolving context reads the declared
key from it — exactly the recipe the row's `where_to_find` gives a human.
Contexts without cluster credentials leave the value unresolved and fall
back to asking. The resolved value feeds the import operation and is never
logged or persisted.

Both are proto-backed KRM documents (`iac.planton.dev/v1`, protos under
`iac/`), parsed through `pkg/protobufyaml` like the E2E
profiles.

## Correctness is machine-proven, never review-trusted

- **Offline** (`conformance_test.go`, runs in `make test`): every resource
  type a mapped component's OpenTofu module declares has an id_format;
  every placeholder those formats reference is declared by the component
  map; `from_spec_field` paths resolve to scalar leaves on the kind's spec
  proto; `from_stack_output` keys are real StackOutputs fields.
- **Live** (the E2E `IMPORT-RT` phase, opt-in via
  `PLANTON_E2E_IMPORT_ROUNDTRIP=1`): deploy the fixture, set its state
  aside, re-import every resource *blind* through these recipes, and
  require the follow-up plan to propose no real change — in-place updates
  are tolerated only for declared `config_only_attributes`. One
  framework-owned pruning happens before declarations apply: attributes the
  plan itself marks WHOLLY unknown post-apply (`after_unknown` literal true —
  plugin-framework providers recompute `id`/`metadata`-class computed
  attributes on every in-place update) are never treated as drift; partially
  unknown objects and all known drift stay under the oracle
  (e.g. `aws_s3_bucket.force_destroy`, engine delete behavior with no
  cloud-side existence) and `write_normalized_attributes`. A tolerance may
  be a dotted sub-path (`"spec.update_strategy"`) when a provider's importer
  fails to read back one nested block — the oracle then prunes exactly that
  sub-path from both plan sides and requires the remainder identical, so
  sibling drift still fails. The destroy that follows runs through the
  re-imported state, proving it fully owns the resources.

A resource type whose upstream provider ships NO importer at the pinned
version declares `not_importable_upstream_reason` in the provider catalog
instead of an `id_format` (mutually exclusive; the conformance guard
enforces exactly one). The canonical case is
`aws_appautoscaling_scheduled_action`: its target and policy siblings both
carry an import handler, the scheduled action does not. The round-trip
skips these addresses at import and proves the ADOPTER'S contract for them
instead — the post-import plan proposes re-creating exactly those
resources, and the reconcile-apply executes it. That is honest only when
the type's create path CONVERGES on an existing cloud resource (an
upsert-style Put, as PutScheduledAction is); a type whose create would
conflict with its own survivor needs different treatment, and the live
lane is what verifies the convergence.

A component map may additionally declare `import_normalized` entries — the
narrowest tolerance vocabulary, scoped to ONE of the module's own logical
resources and ONE dotted sub-path each, with a mandatory reason. It exists
for values that cannot round-trip BY PROVIDER CONSTRUCTION rather than by
importer read-back gaps: the canonical case is a Secret data key wiring
`random_password`'s `bcrypt_hash`, which the random provider recomputes with
a fresh salt on import — the first post-adoption apply rewrites the key to
an equivalent hash of the same password, a functional no-op. Provider-wide
classes stay in the catalog; a kind's own irreducible normalization lives in
the kind's map, beside the resource it describes.

Sub-path grammar (shared by catalog sub-paths and `import_normalized`
paths): segments are dot-separated, and a segment whose KEY itself
contains dots is written bracket-quoted — `data["password.db"]` —
because a plain dotted path would walk `data → password → db` and never
match the real key (Kubernetes Secret data keys like htpasswd file
names are the canonical case; proven live on a blind round-trip that
failed its declared tolerance until the grammar could express the key).

**REPLACES are never tolerable — and one class of config can only plan a
replace after import.** Every tolerance above covers in-place updates; a
post-import plan proposing destroy+create always fails the oracle, because
silently "passing" a plan that would destroy the adopted resource is worse
than any recipe bug. A scenario whose config carries **write-only ForceNew
SECRETS** (first live case: a container group's secure environment
variables) sits exactly there: the blind import necessarily lacks the
secret, the config supplies it, the attribute forces replacement — and
`ignore_changes` must never paper it over, because rotating such a secret
is SUPPOSED to replace the resource. Such a scenario opts out of the
round-trip with the reason-carrying manifest annotation
`planton.dev/e2e-import-roundtrip-skip: "<why>"` (an empty value does not
skip; the runner prints the reason into the lane log), while the kind's
recipes stay enrolled and prove on its secret-free scenarios. The kind's
user docs must teach the adoption consequence: the first apply after
importing a secret-bearing instance replaces it once, converging to the
manifest's secrets.

## Enrollment is the file itself

The import-map file's presence is the single enrollment signal everywhere:
the offline conformance guard auto-discovers every
`{component}/v1/iac/import-map.yaml` (`DiscoverComponentImportMaps`), the
E2E round-trip gate keys off the same file, and the platform's catalog
pipeline bundles the same file into the import wizard's data. There is
deliberately no allowlist: an authored map that ships to users while
dodging validation — or passes validation while never shipping — is a
divergence this single signal makes impossible.

**The merge discipline that keeps the ledger honest: a new or changed
import map does not merge without a green live round-trip lane.** The
offline guard proves recipes are well-formed; only the live lane proves
they are CORRECT (they import the real resource, not a plausible one).

## Live-proven ledger

The round-trip lane is the only honest enforcement of "proven" — this
table is the human-readable record of which kinds have run it, so a reader
never has to guess which maps are correctness-proven versus merely
well-formed. A recipe for a resource no scenario deploys is offline-validated
only (noted per row); the lane proves exactly what the fixtures exercise.

| Component | Live round-trip | Tolerated declared updates / notes |
|-----------|-----------------|------------------------------------|
| `awss3bucket` | 2026-08-11, both scenarios re-proven post-depth-closure (full-surface: 14 resources blind incl. the abac row and the three new `{bucket}:{name}` composites — analytics/inventory/metric — via `for_each` keys; run-scoped bucket names via `${E2E_RUN_ID}`) | `aws_s3_bucket.force_destroy` (config-only); the metadata-configuration row is offline-validated only (its arm is a recorded live deferral) |
| `awsvpc` | 2026-08-11, minimal re-proven post-depth-closure — the ipv4 association id derives BLIND via `from_stack_output_keyed_by_address` over the module's association-id map output (previously user-supplied/offline-only) | ipv6-association and encryption-control rows offline-validated only (no scenario deploys those arms; recorded deferrals) |
| `awssecuritygroup` | 2026-07-10, rules-rich | `aws_security_group.revoke_rules_on_delete` (config-only) |
| `awsecrrepo` | 2026-07-10, full-surface (repository + lifecycle policy) | `aws_ecr_repository.force_delete` (config-only) |
| `awssubnet` | 2026-07-10, minimal + routed (subnet + route table + `{subnet_id}/{route_table_id}` association composite) | — |
| `awsinternetgateway` | 2026-07-10, smoke | — |
| `awsnatgateway` | 2026-07-10, minimal | — |
| `awsiamrole` | 2026-07-10, smoke (role + inline `{role_name}:{inline_policy_name}` + attachment `{role_name}/{managed_policy_arn}`, both via `for_each` keys) | `aws_iam_role.force_detach_policies` declared (provider-documented config-only; scenario does not set it) |
| `awsdynamodb` | 2026-08-13, BOTH scenarios post-capacity-fold (on-demand-full-surface: table + resource policy + table- and GSI-level contributor insights incl. the optional `{index_name?}` segment and account id; provisioned-autoscaled: table + both `aws_appautoscaling_target`s and the read `aws_appautoscaling_policy` blind via the composed `{service_namespace}/{scaling_resource_id}/{scalable_dimension}` IDs — the `from_metadata_name_prefix` arm's and the scoped-literal declarations' first live proof — with the scheduled action's skip-and-recreate path the FIRST live exercise of `not_importable_upstream_reason`). Earlier: 2026-07-10, on-demand-full-surface | `aws_dynamodb_resource_policy.policy`/`revision_id` (write-normalized); `aws_dynamodb_kinesis_streaming_destination` not composed by the scenario, offline-validated only; `aws_appautoscaling_scheduled_action` not importable upstream (recorded reason; reconcile-apply re-puts it) |
| `awssqsqueue` | 2026-07-10, fifo-full-surface (queue URL id) | — |
| `awssnstopic` | 2026-07-10, standard-topic (topic ARN id) | `aws_sns_topic_data_protection_policy` not composed by the scenario, offline-validated only |
| `awskmskey` | 2026-08-11, minimal re-proven post-depth-closure (key UUID + `alias/...` via `for_each` key + the NEW `{kms_key_id}:{grant_id}` composite — the cloud-generated grant id derives BLIND via `from_stack_output_keyed_by_address` over the grant_ids output) | `aws_kms_key.deletion_window_in_days` declared (provider-documented config-only; scenario does not set it) |
| `kubernetesdeployment` | 2026-07-21, all 13 scenarios (workload + service + HPA + PDB + env Secret + created namespace) | `kubernetes_deployment_v1.wait_for_rollout`, `kubernetes_service_v1.wait_for_load_balancer`, `kubernetes_secret_v1.wait_for_service_account_token` (config-only) |
| `kubernetesstatefulset` | 2026-07-21, all 13 scenarios | same config-only knobs; `kubernetes_stateful_set_v1` `spec.update_strategy` (write-normalized dotted sub-path: the provider importer does not read the block back) |
| `kubernetesdaemonset` | 2026-07-21, all 5 scenarios | `kubernetes_daemon_set_v1.wait_for_rollout` (config-only) |
| `kubernetesjob` | 2026-07-21, all 5 scenarios | — |
| `kubernetescronjob` | 2026-07-21, all 4 scenarios | — |
| `kubernetesservice` | 2026-07-21, all 13 scenarios | `kubernetes_service_v1.wait_for_load_balancer` (config-only) |
| `kubernetesingress` | 2026-07-21, all 4 scenarios (incl. the composed real-backend fixture) | `kubernetes_ingress_v1.wait_for_load_balancer` (config-only) |
| `kubernetesnetworkpolicy` | 2026-07-21, all 4 scenarios | — |
| `kubernetespersistentvolumeclaim` | 2026-07-22, all 3 scenarios (incl. the composed StorageClass fixture and the pinned-empty-class static-binding shape) | `kubernetes_persistent_volume_claim_v1.wait_until_bound` (config-only; deliberately false in the module) |
| `kubernetesstorageclass` | 2026-07-22, both scenarios | — |
| `kubernetesresourcequota` | 2026-07-22, all 3 scenarios (governed-namespace proves the quota + companion LimitRange pair) | — |
| `kubernetespriorityclass` | 2026-07-22, both scenarios | — |
| `kubernetespoddisruptionbudget` | 2026-07-22, both scenarios | — |
| `kuberneteshorizontalpodautoscaler` | 2026-07-22, all 3 scenarios (incl. the composed Deployment target fixture) | — |
| `kubernetescertmanager` | 2026-07-22, both scenarios (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kuberneteshelmrelease` | 2026-07-22, all 3 scenarios (HTTPS repo, OCI registry, values-override). The module-owned CRD rows (`helm_crds`: the chart's `crds/` directory, derived from the pinned chart at plan time and keyed by each CRD's own name as cluster-scoped 3-part composed IDs; empty for a chart without one, absent when `crds.install` is false) are declared and offline-validated; a CRD-bearing round-trip lane has not yet exercised them live. | `helm_release` install-time attributes (repository, values/set/set_sensitive, lifecycle knobs — Helm does not persist how a release was installed; provider-documented, declared config-only) + `kubectl_manifest` provider-side knobs incl. `apply_only` on the CRD rows (config-only). Computed attributes the plan marks wholly after-unknown (id, metadata) are pruned by the oracle itself, not declared. |
| `kubernetesexternaldns` | 2026-07-22, all 3 scenarios (Helm release + created namespace; the credential-Secret arms are exercised offline by the hack manifest — the kind lanes run credential-less by design) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesexternalsecretsoperator` | 2026-07-22, both scenarios (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesingressnginx` | 2026-07-22, all 3 scenarios (Helm release + created namespace; multi-instance scenario re-imports alongside a live sibling controller) | `helm_release` install-time attributes (config-only, see the catalog row); the module's controller-Service data source is a read, not an imported resource |
| `kubernetesmetricsserver` | 2026-07-22, both scenarios (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesgatewayclass` | 2026-07-22, both scenarios (the composed 3-part cluster-scoped ID — the `[//{namespace}]` segment group drops) | `kubectl_manifest` provider-side knobs (config-only) + `yaml_body`/`yaml_body_parsed` (write-normalized: the importer stores the stripped live object), see the catalog row |
| `kubernetesgateway` | 2026-07-22, all 3 scenarios (incl. the composed GatewayClass-fixture FK) | same `kubectl_manifest` tolerances |
| `kuberneteslistenerset` | 2026-07-22, both scenarios (incl. the composed Gateway-fixture parent FK) | same `kubectl_manifest` tolerances |
| `kuberneteshttproute` | 2026-07-22, all 3 scenarios (incl. the composed Gateway + Service fixture chain) | same `kubectl_manifest` tolerances |
| `kubernetesgrpcroute` | 2026-07-22, both scenarios | same `kubectl_manifest` tolerances |
| `kubernetestcproute` | 2026-07-22, both scenarios | same `kubectl_manifest` tolerances |
| `kubernetesudproute` | 2026-07-22, both scenarios | same `kubectl_manifest` tolerances |
| `kubernetestlsroute` | 2026-07-22, both scenarios | same `kubectl_manifest` tolerances |
| `kubernetesreferencegrant` | 2026-07-22, both scenarios | same `kubectl_manifest` tolerances |
| `kubernetesclusterissuer` | 2026-07-22, self-signed scenario (cluster-scoped 3-part composed ID) | same `kubectl_manifest` tolerances; credential Secrets derive blind via `from_address_key` (the module keys them by Secret name) — the self-signed scenario materializes none, so those recipes are offline-validated only |
| `kubernetesissuer` | 2026-07-22, self-signed + composed CA-chain scenarios | same `kubectl_manifest` tolerances; credential-Secret recipes offline-validated only (see `kubernetesclusterissuer`) |
| `kubernetescertificate` | 2026-07-22, minimal + full-surface scenarios (real issuance fixture chain) | same `kubectl_manifest` tolerances |
| `kubernetesclustersecretstore` | 2026-07-22, fake-backend scenario (cluster-scoped 3-part composed ID) | same `kubectl_manifest` tolerances; the conditional `<name>-credentials` Secret derives via `from_address_key` / `from_metadata_name_suffix` — the fake backend declares no credentials, so that recipe is offline-validated only |
| `kubernetessecretstore` | 2026-07-22, fake-backend scenario | same as `kubernetesclustersecretstore` |
| `kubernetesexternalsecret` | 2026-07-22, both scenarios (incl. the real sync-loop fixture chain) | same `kubectl_manifest` tolerances |
| `kubernetesistio` | 2026-07-22, minimal + ambient scenarios (20 resources: FOUR `tofu_resource_name`-scoped Helm releases incl. the ambient cni/ztunnel pair, 15 module-owned CRDs blind-derived via `from_address_key` — the module keys each `kubectl_manifest` by the CRD's own name — and the created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) + `kubectl_manifest` tolerances |
| `kubernetesdestinationrule` | 2026-07-22, minimal scenario | same `kubectl_manifest` tolerances |
| `kubernetesserviceentry` | 2026-07-22, minimal scenario | same `kubectl_manifest` tolerances |
| `kubernetespeerauthentication` | 2026-07-22, minimal scenario | same `kubectl_manifest` tolerances |
| `kubernetesrequestauthentication` | 2026-07-22, minimal scenario | same `kubectl_manifest` tolerances |
| `kubernetesauthorizationpolicy` | 2026-07-22, minimal + behavioral-deny scenarios (re-imported alongside a live mesh) | same `kubectl_manifest` tolerances |
| `kubernetestelemetry` | 2026-07-22, minimal scenario | same `kubectl_manifest` tolerances |
| `kubernetesenvoyfilter` | 2026-07-22, minimal scenario | same `kubectl_manifest` tolerances |
| `kubernetescilium` | 2026-07-23, both scenarios on the cilium-cni cluster profile (Helm release; kube-system installs import no namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kuberneteskeda` | 2026-07-23, both scenarios (Helm release + created namespace; behavioral-scaling re-imports with the live scale-target fixture present) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesbackendtlspolicy` | 2026-07-23, all 3 scenarios (4-part namespaced composed ID; composed-target resolves the Service-fixture FK) | same `kubectl_manifest` tolerances |
| `kubernetesclusterautoscaler` | 2026-07-23, minimal scenario on the kwok arm (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesvelero` | 2026-07-23, both scenarios (Helm release + created namespace; behavioral re-imports alongside the live MinIO fixture) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetescloudnativepgoperator` | 2026-07-23, both scenarios (two tofu_resource_name-scoped Helm releases — operator + Barman Cloud plugin — plus the created namespace; the plugin scenario re-imports alongside the live cert-manager fixture) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetespostgres` | 2026-07-23, all three scenarios (composed 4-part IDs for the Cluster / ObjectStore / ScheduledBackup CRs, for_each Secrets via from_address_key, count-indexed singleton Secrets via scoped from_metadata_name_suffix — the with-backup lane's 10-resource blind re-import is the largest kubectl_manifest family proven to date) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetesstrimzikafkaoperator` | 2026-07-23, both scenarios (Helm release + created namespace; tuned-full re-imports alongside its watched-namespace fixtures) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kuberneteskafka` | 2026-07-23, all three scenarios (two tofu_resource_name-scoped kubectl_manifest families — the Kafka CR singleton and the pool-name-keyed KafkaNodePool for_each — plus the count-indexed metrics ConfigMap via scoped from_metadata_name_suffix and the created namespace; the durability lane re-imports a live 3-node cluster) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kuberneteskafkatopic` | 2026-07-23, both scenarios (4-part namespaced composed ID; the scenarios resolve the kafka_cluster FK against a live fixture cluster) | same `kubectl_manifest` tolerances |
| `kuberneteskafkauser` | 2026-07-23, both scenarios (4-part namespaced composed ID; behavioral-auth re-imports alongside the operator-generated credentials Secret, which is deliberately outside the module's state) | same `kubectl_manifest` tolerances |
| `kubernetesvalkey` | 2026-07-23, all three scenarios (Helm release + created namespace; the `-auth` Secret row on the auth-declaring lanes — declared ACL passwords materialize it) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesperconamysqloperator` | 2026-07-23, both scenarios (Helm release + created namespace + the module-owned PXC validating-webhook row on the widened-watch arm, fixed cluster-scoped name) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesmysql` | 2026-07-23, all three scenarios (the namespaced 4-part composed ID for the PerconaXtraDBCluster CR + declared user-password Secrets + the anchor namespace; the durability lane's round-trip caught the operators' co-owned `percona.com/<cluster>-<user>-hash` rotation annotations — both engines now ignore annotation drift on exactly those Secrets) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetesperconamongooperator` | 2026-07-23, both scenarios (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesmongodb` | 2026-07-23, all three scenarios (the namespaced 4-part composed ID for the PerconaServerMongoDB CR + declared user-password Secrets + the anchor namespace; the co-owned rotation-annotation drift ignored — see `kubernetesmysql`) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kuberneteskafkaconnect` | 2026-07-24, all three scenarios (the namespaced 4-part composed ID for the KafkaConnect CR + the metrics ConfigMap + the anchor namespace; re-imported alongside the live fixture cluster) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kuberneteskafkaconnector` | 2026-07-24, both scenarios (the namespaced 4-part composed ID; the data-flow lane re-imports alongside the live self-mirror pipe) | same `kubectl_manifest` tolerances |
| `kuberneteskafkamirrormaker2` | 2026-07-24, both scenarios (the namespaced 4-part composed ID for the KafkaMirrorMaker2 CR + the metrics ConfigMap + the anchor namespace; re-imported alongside the live two-cluster migration fixtures) | same `kubectl_manifest` tolerances |
| `kuberneteskarapace` | 2026-07-24, both scenarios (module-owned typed manifests — the registry and REST-proxy Deployment/Service pairs keyed per resource — plus the conditional SASL-password Secret and the anchor namespace) | typed-resource config-only knobs (rollout waits; see the catalog rows) |
| `kuberneteskafkaui` | 2026-07-24, both scenarios (Helm release + the console credentials Secret + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesopensearch` | 2026-07-24, scenario lanes within the search wave's nine blind round-trip lanes (the namespaced 4-part composed ID for the OpenSearchCluster CR + the anchor namespace; re-imported alongside the live operator fixture) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetessolr` | 2026-07-24, scenario lanes within the search wave's nine blind round-trip lanes (the namespaced 4-part composed ID for the SolrCloud CR + the anchor namespace; re-imported alongside the live operator fixture) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetesneo4j` | 2026-07-24, scenario lanes within the search wave's nine blind round-trip lanes (Helm release + the `-auth` Secret via scoped suffix + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesaltinityoperator` | 2026-07-25, both scenarios (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesclickhouse` | 2026-07-25, all three scenarios — the FIRST two-GVK composed-ID map: the ClickHouseInstallation plus the `-keeper`-suffixed ClickHouseKeeperInstallation via scoped `from_metadata_name_suffix`, the `-clickhouse-auth` Secret, and the anchor namespace | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetesseaweedfs` | 2026-07-25, all three scenarios (Helm release + created namespace + the `-admin-auth` console Secret via scoped suffix and its `random_password` companion — the FIRST secret-material import by VALUE via `from_cluster_secret_key`, secret IDs redacted from progress output) | `helm_release` install-time attributes (config-only, see the catalog row); `random_password` generation-shape args ignored by module design (imported credentials never regenerate) |
| `kubernetesqdrant` | 2026-07-25, all three scenarios (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesrabbitmqoperator` | 2026-07-25, both scenarios (the 16-document release-manifest bundle blind-imported via `from_address_key_segment` — the FIRST multi-GVK bundle import) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetesrabbitmq` | 2026-07-25, all three scenarios (the namespaced 4-part composed ID for the RabbitmqCluster CR + the anchor namespace; re-imported alongside the live operator fixture) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kuberneteskubeprometheusstack` | 2026-07-25, all three scenarios (Helm release + created namespace; the alerting lane re-imports with the operator-reconciled StatefulSets live) | `helm_release` install-time attributes (config-only, see the catalog row); the count-indexed `<name>-remote-write-auth` Secret is not composed by any scenario (none declares remote-write basic auth), offline-validated only |
| `kubernetesgrafana` | 2026-07-25, all three scenarios (Helm release + created namespace; behavioral-persistence re-imports alongside the live PVC-backed dashboard) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesloki` | 2026-07-25, all three scenarios (Helm release + created namespace; full-surface re-imports with the gateway auth-guarded in multi-tenant mode) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetestempo` | 2026-07-25, all three scenarios (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetessignoz` | 2026-07-26, all three scenarios (Helm release + created namespace; re-imported alongside the live composed-ClickHouse fixture chain — the module owns no credential Secret) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesargocd` | 2026-07-26, all three scenarios (Helm release + created namespace; the behavioral-gitops lane re-imports with the live Application and kept CRDs present) | `helm_release` install-time attributes (config-only, see the catalog row); the application-owned `argocd-initial-admin-secret` is deliberately outside the module's state — no row exists for it |
| `kubernetesargoworkflows` | 2026-07-26, all three scenarios (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row); workflow credentials are references the module never owns — no Secret rows |
| `kubernetestemporal` | 2026-07-26, all three scenarios (Helm release + anchor namespace; re-imported alongside the live composed-Postgres fixture chain) | `helm_release` install-time attributes (config-only, see the catalog row); every database credential is a reference the module never owns — no Secret rows |
| `kubernetesnats` | 2026-07-26, all three scenarios — the FIRST live exercise of the keyed-collection secret import: the per-user `random_password` collection re-imports via `from_cluster_secret_key` + `key_from_address_key` (each password read from the auth Secret under its own username key, secret IDs redacted from progress output) plus the count-indexed `-auth` Secret and the created namespace | `helm_release` install-time attributes (config-only, see the catalog row); `random_password` generation-shape args ignored by module design (imported credentials never regenerate) |
| `kubernetestektonoperator` | 2026-07-26, both scenarios (the 41-document multi-GVK bundle blind-imported across all three ordered resource groups via `from_address_key_segment` — the composed-ID segments serve every group identically) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row); the round-trip's reconcile-apply restores them into state before destroy |
| `kubernetestekton` | 2026-07-26, all three scenarios (the cluster-scoped 3-part literal composed ID `operator.tekton.dev/v1alpha1//TektonConfig//config`; re-imported alongside the live operator fixture) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetesgharunnerscalesetcontroller` | 2026-07-26, both scenarios (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesgharunnerscaleset` | 2026-07-26, minimal scenario (Helm release + created namespace + the count-indexed `-github-auth` Secret on the declared-PAT arm); the behavioral-github lane's re-import awaits the owner-arranged GitHub credentials (the scenario's required-env skip) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kuberneteskyverno` | 2026-08-01, all three scenarios (Helm release + the module-owned webhook-GC sentinel ConfigMap + created namespace; no Secret rows — the engine's webhook certificates are runtime-generated and travel with the release) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesgatekeeper` | 2026-08-01, all three scenarios (Helm release + created namespace; no Secret rows — the webhook cert Secret is chart-owned and travels with the release). The lane caught the namespace-label ownership defect: the chart's post-install hook stamps the exemption label onto the module-owned namespace, so the module must declare it or every post-import plan (and day-2 apply) strips it | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kuberneteskeycloakoperator` | 2026-08-01, both scenarios (the 20-document release-manifest bundle blind-imported across all three ordered resource groups — namespace, workloads, CRDs — via `from_address_key_segment`; the namespaced AND cluster-wide variants both re-imported) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row); the round-trip's reconcile-apply restores them into state before destroy |
| `kuberneteskeycloak` | 2026-08-01, all three scenarios (the namespaced 4-part composed ID `k8s.keycloak.org/v2beta1//Keycloak//<ns>//<name>`; re-imported alongside the live operator + composed-Postgres fixture chain; the operator-generated `<name>-initial-admin` Secret is deliberately outside the module's state — no row exists for it) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetesopenbao` | 2026-08-01, all four scenarios (Helm release + created namespace; dev/standalone/single-replica-Raft shapes all re-imported). The conditional `<name>-seal-credentials` Secret row is offline-validated only — no lane scenario declares a static-credential seal arm (real KMS credentials have no kind stand-in; the deferred-E2E ledger carries the unblock) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesopenfga` | 2026-08-01, both scenarios — including the LIVE exercise of the conditional `<name>-authn-keys` Secret row on the full-surface lane (inline preshared keys materialize the Secret; the memory-arm lane re-imports release + namespace only) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesharbor` | 2026-08-01, all four scenarios (12 resources each on the internal-arm lanes: release + module-owned admin/internal Secrets + SEVEN `random_password` companions imported by VALUE via `from_cluster_secret_key` — six scoped keys on `-internal-auth`, the admin on `-admin-auth`, the internal-database password read from the CHART-created `<name>-database` Secret; the full-surface lane re-imported alongside the live composed-Postgres fixture; the anchor-namespace row conditional on `create_namespace`). The conditional `redis_auth`/`storage_auth` rows are LIVE-proven on the composed-storage lane (declared external-cache password + S3 credentials against live Valkey + SeaweedFS fixtures materialize both Secrets; blind-imported by suffix alongside the release) | `helm_release` install-time attributes + `kubernetes_secret_v1.wait_for_service_account_token` (config-only, see the catalog rows); `internal_auth` `data.REGISTRY_HTPASSWD` — the FIRST `import_normalized` component-map declaration (random_password's `bcrypt_hash` re-salts on import; sibling keys verified identical by the oracle) |
| `kubernetesoteloperator` | 2026-08-01, both scenarios (Helm release + created namespace + the module-owned CRDs — derived from the pinned chart, so the set follows `chart_version`; four at the default pin — as cluster-scoped 3-part composed IDs keyed by each CRD's own name; re-imported over CRDs kept by previous lanes — the adoption path itself) | `helm_release` install-time attributes + `kubectl_manifest` provider-side knobs incl. `apply_only` (config-only, see the catalog rows) |
| `kubernetesotelcollector` | 2026-08-01, all three scenarios (the namespaced 4-part composed ID `opentelemetry.io/v1beta1//OpenTelemetryCollector//<ns>//<name>`; re-imported alongside the live operator fixture chain; everything the operator creates from the CR is operator-owned — no rows exist for it) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetesopensearchoperator` | 2026-08-01, both scenarios (Helm release + created namespace + the module-owned CRDs — derived from the pinned chart, so the set follows `chart_version`; nine at 2.7.0, ten at 2.8.0 — as cluster-scoped 3-part composed IDs keyed by each CRD's own name; re-imported OVER CRDs kept by previous installs — the adoption path itself) | `helm_release` install-time attributes + `kubectl_manifest` provider-side knobs incl. `apply_only` (config-only, see the catalog rows) |
| `kubernetessolroperator` | 2026-08-01, both scenarios (Helm release + created namespace + the module-owned CRDs derived from the pinned chart — three solr.apache.org from the chart's `crds/` directory plus the bundled zookeeper-operator's ZookeeperCluster rendered behind its own switch — as cluster-scoped 3-part composed IDs keyed by each CRD's own name; re-imported over kept CRDs) | `helm_release` install-time attributes + `kubectl_manifest` provider-side knobs incl. `apply_only` (config-only, see the catalog rows) |
| `kubernetesairflow` | 2026-08-02, all three scenarios (Helm release + anchor namespace + the module-owned credential/connection Secret family — the FIRST live `random_bytes` import: the fernet key re-imported by VALUE via `from_cluster_secret_key` alongside the api/webserver/jwt key and admin-credential rows; the full-surface lane's Celery/Redis/PgBouncer arms materialized the redis-password, broker/result-backend/metadata/log-read connection and pgbouncer config/stats rows against the live composed-Postgres fixture) | `helm_release` install-time attributes + `kubernetes_secret_v1.wait_for_service_account_token` (config-only, see the catalog rows); `random_password`/`random_bytes` generation-shape args ignored by module design |
| `kubernetessparkoperator` | 2026-08-02, scenario lanes within the compute-engines wave's blind TF round-trips (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kuberneteskuberayoperator` | 2026-08-02, scenario lanes within the compute-engines wave's blind TF round-trips (Helm release + created namespace) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesraycluster` | 2026-08-02, scenario lanes within the compute-engines wave's blind TF round-trips (the namespaced 4-part composed ID `ray.io/v1//RayCluster//<ns>//<name>` + the anchor namespace; re-imported alongside the live operator fixture) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetesflinkoperator` | 2026-08-02, scenario lanes within the compute-engines wave's blind TF round-trips (Helm release + created namespace + the `for_each` watch-namespace rows via `from_address_key` — the row was ADDED mid-proof when the lanes caught the module-owned watch namespaces missing from the map, then proven on the re-run; the conditional `-webhook-keystore` Secret + its `random_password` companion ride the webhook-enabled lanes) | `helm_release` install-time attributes (config-only, see the catalog row) |
| `kubernetesflinkdeployment` | 2026-08-02, scenario lanes within the compute-engines wave's blind TF round-trips (the namespaced 4-part composed ID `flink.apache.org/v1beta1//FlinkDeployment//<ns>//<name>` + the anchor namespace; the recovery lane's round-trip re-imported alongside the live operator + SeaweedFS fixture chain) | `kubectl_manifest` provider-side knobs (config-only, see the catalog row) |
| `kubernetesjupyterhub` | 2026-08-02, all three scenarios (Helm release + created namespace + the `-auth` shared-password Secret with its `random_password` companion imported by VALUE via `from_cluster_secret_key`; the full-surface round-trip covered the conditional `-hub-secret` and `-read-anchor` rows) | `helm_release` install-time attributes + `kubernetes_secret_v1.wait_for_service_account_token` (config-only, see the catalog rows); `random_password` generation-shape args ignored by module design |
| `kubernetesmlflow` | 2026-08-02, all three scenarios (8–9 resources each on the module's TYPED shape — Deployment + Service + anchor namespace + the data/artifacts PVCs + the composed `-backend-uri`/`-auth-config` Secrets, recomposed identically post-adoption — with the admin password AND the boot-required Flask CSRF key imported by VALUE via `from_cluster_secret_key`; re-imported alongside the live composed Postgres + SeaweedFS fixture chain) | typed-resource config-only knobs (rollout/bind waits, see the catalog rows); `random_password` generation-shape args ignored by module design |
| `kubernetestrino` | 2026-08-02, all three scenarios (5 resources each: Helm release + the module-owned `-auth`/`-internal` Secrets + BOTH `random_password` companions imported by VALUE via `from_cluster_secret_key`; the full-surface lane re-imported alongside the live composed-Postgres fixture) | `helm_release` install-time attributes + `kubernetes_secret_v1.wait_for_service_account_token` (config-only, see the catalog rows); `auth` `data["password.db"]` — the `import_normalized` bcrypt re-salt row, and the FIRST bracket-quoted sub-path segment (the Secret key itself contains a dot, which a plain dotted path cannot express — the grammar gained the segment form on this lane) |
| `kubernetessuperset` | 2026-08-02, all three scenarios (6 resources each: Helm release + the always-owned `-env` Secret — its generated keys recomposed identically after adoption, zero drift — + the conditional `-secret-key`/`-admin-auth` Secrets + BOTH `random_password` companions imported by VALUE via `from_cluster_secret_key`; the full-surface lane re-imported alongside the live composed Postgres + AUTHED Valkey fixture chain) | `helm_release` install-time attributes + `kubernetes_secret_v1.wait_for_service_account_token` (config-only, see the catalog rows); the ws-JWT random's row is offline-validated only (no lane scenario declares the websockets arm — the deferred-E2E ledger carries the unblock) |
| `kuberneteslocust` | 2026-08-04, all three scenarios (6–7 resources each: Helm release + the `-locustfile` script ConfigMap — the `-lib` row joining on the full-surface lane's lib-modules arm — + the `-web-auth` login-backend ConfigMap + the `-auth` Secret + BOTH `random_password` companions imported by VALUE via `from_cluster_secret_key`, the login password and the Flask session-signing key; re-imported alongside the live podinfo target fixture) | `helm_release` install-time attributes + `kubernetes_secret_v1.wait_for_service_account_token` (config-only, see the catalog rows); the anchor-namespace row is offline-validated only (every lane joins the fixture-owned namespace — no scenario declares `create_namespace`); `random_password` generation-shape args ignored by module design |
| `digitaloceanvpc` | 2026-09-15, both scenarios (`{vpc_id}` UUID; the `minimal` lane proves the Optional+Computed `ip_range` reads back with no diff after a blind import) | — |
| `digitaloceanvolume` | 2026-09-15, both scenarios (`{volume_id}` UUID). The `full` lane is the live proof of the modules' `ignore_changes` on the create-only trio: the blind import leaves `snapshot_id` / `initial_filesystem_type` / `initial_filesystem_label` empty, the config re-supplies them, and the plan is clean — a `config_only_attributes` tolerance alone could never pass it, because the provider marks them ForceNew and the oracle refuses replaces | the create-only trio, ignored by module design (declared in the catalog for the raw-resource adopter) |
| `digitaloceancontainerregistry` | 2026-09-15, both scenarios (`{registry_name}` — a NAME-addressed import; the `full` lane also exercised `not_importable_upstream_reason` on `digitalocean_container_registry_docker_credentials`: skipped at import, re-created by the reconcile-apply, which re-mints the credential) | docker-credentials re-create (declared not importable upstream: the passthrough importer's Read requires `registry_name` from prior state at v2.99.1) |
| `digitaloceancertificate` | **offline-validated only** (map conformance-green; `{certificate_name}` — the resource id IS the name). The only live-runnable scenario (`minimal`, a custom certificate) skips the round-trip by annotation: the PEM trio is write-only AND ForceNew, so the only post-import plan is a replace, which is correct behaviour for secret material and must not be papered over. Live proof waits on a PEM-free scenario — Let's Encrypt, once a domain is delegated to the account (`PLANTON_E2E_DIGITALOCEAN_DELEGATED_DOMAIN`) | — |
| `digitaloceandroplet` | 2026-09-16, both scenarios (`{droplet_id}` integer id; the provider's importer seeds `image` from the live slug and `resize_disk` as true). The `full` lane is the live proof of the modules' `ignore_changes` on `ssh_keys` / `user_data` / `droplet_agent`: all three are ForceNew and never read back, so without it the first post-import plan was a droplet REPLACE — a `config_only_attributes` tolerance alone could never pass it | `digitalocean_droplet.graceful_shutdown` (config-only, re-applied in place) and `resize_disk` (write-normalized: the importer seeds true against a configured false); the create-only trio, ignored by module design (declared in the catalog for the raw-resource adopter) |
| `digitaloceanfirewall` | 2026-09-16, both scenarios (`{firewall_id}` UUID; the `full` lane re-imported a firewall with droplet-id AND tag targeting plus a rule-level source tag; the `minimal` lane the rules-rich shape with `port_range: all` and an icmp rule) | — (the declared `inbound_rule.port_range` / `outbound_rule.port_range` write-normalization did not fire: authored `all` and the icmp no-port form read back byte-identical) |
| `digitaloceandnszone` | 2026-09-16, both scenarios (`{domain_name}` — a NAME-addressed import — plus every inline record as `{domain},{record_id}` with the record id derived BLIND via `from_stack_output_keyed_by_address` over the module's `record_ids` map output keyed `<name>-<recIdx>-<valIdx>`; the `full` lane re-imported 9 resources — the domain and eight records across A/CNAME/MX/TXT/SRV/CAA — with a clean plan). The `full` lane is also the live proof of the modules' `ignore_changes` on the ForceNew, never-read-back `ip_address`: without it the first post-import plan was a ZONE replace | `ip_address`, ignored by module design (declared in the catalog for the raw-resource adopter) |
| `digitaloceandnsrecord` | 2026-09-16, all three scenarios (the two-part `{domain},{record_id}`, both halves from stack outputs; A, SRV with the presence trio and a custom TTL, and CAA with `flags: 0` — the explicit zero the provider drops on create and the API defaults to 0 agreed after import) | — (the declared `value` trailing-dot write-normalization did not fire: the scenarios author the dotted FQDN the API reports back, which is also the only form that is idempotent) |
| `digitaloceansshkey` | 2026-09-16, the single scenario (`{ssh_key_id}` — the NUMERIC id only; a fingerprint would fail the provider's `strconv.Atoi` in Read) | — (lossless: name, public_key, and fingerprint read back byte-identical, exactly as the provider's own `ImportStateVerify` test says) |
| `digitaloceanproject` | 2026-09-16, both scenarios (`{project_id}` UUID; the `full` lane re-imported a project with description, purpose, a lowercase `staging` environment, and one droplet member by URN) | — (the declared `environment` write-normalization did not fire on the plan: state holds `Staging` against the lowercase config, but the provider's case-insensitive diff-suppress made the post-import plan a no-op; the record stays because state really does differ from config) |
| `digitaloceanloadbalancer` | 2026-09-16, both scenarios (`{load_balancer_id}` UUID; the `full` lane re-imported a tag-targeted balancer with a real member behind the tag, an LB-level firewall, cookie sticky sessions, tuned health check, PROXY protocol, keepalive, idle-timeout, and `size_unit`; the `minimal` lane the region-default-VPC shape) | — (none of the declared tolerances fired: `size`/`size_unit` read back as the configured unit, and `network`/`network_stack`/`tls_cipher_policy` were unset on both sides) |
| `digitaloceankubernetescluster` | 2026-09-17, both scenarios (`{cluster_id}` UUID; the `hardened` lane re-imported a cluster with a control-plane firewall, maintenance policy, autoscaler tuning, the routing-agent addon, and an autoscaling default pool with labels, a taint, and tags; the `minimal` lane the one-node shape). Lossless: the provider filters its own `terraform:default-node-pool` marker out of the pool tags on read, and its `node_count` diff-suppress against `actual_node_count` means the unpopulated `node_count` after a blind import produces no plan | — (none of the declared config-only tolerances fired: `destroy_all_associated_resources`, `kubeconfig_expire_seconds`, and `registry_integration` were unset on both sides) |
| `digitaloceanuptimecheck` | 2026-09-16, both scenarios (`{check_id}` UUID plus every composed alert row as `{check_id},{alert_id}` with the alert id derived BLIND via `from_stack_output_keyed_by_address` over the module's `alert_ids` map output keyed `<idx>-<alert_name>`; the `alerted` lane re-imported 3 resources — the check, a down row, a latency row — with a clean plan). Lossless BY DESIGN: the modules send the threshold/comparison values DigitalOcean forces on down/down_global/ssl_expiry rows, so nothing normalizes; the earlier `threshold` write-normalized tolerance was retired with that design | — |
| `digitaloceanmonitoralert` | 2026-09-16, both scenarios (`{alert_id}` UUID — the resource id, since the provider's own `uuid` attribute is never written at the pin; the `tag-targeted` lane's `entities` read back as an empty list, so the watched materialization class did not appear) | — |
| `digitaloceanapp` | 2026-09-16, all three scenarios (`{app_id}` UUID; the `alerted` lane re-imported an app carrying an app-level and a component alert rule; the `composed` lane a service plus a worker). The provider's Read flattens the whole App spec, so the only tolerated update is the provider-local `deployment_per_page` pagination knob | `digitalocean_app.deployment_per_page` (config-only, never sent to the API). Alert DESTINATIONS are never read back by the provider and would surface as an update on any import -- the kind records that as a provider-defect deferral and its scenarios do not set them |
| `digitaloceanfunction` | 2026-09-16, the single scenario (`{app_id}` UUID derived from the `function_id` output -- the kind has no resource of its own; it re-imported the App Platform app that hosts the functions component, built from a public git source) | `digitalocean_app.deployment_per_page` (config-only) |
| `digitaloceancdn` | 2026-09-16, both scenarios (`{cdn_id}` UUID; the `configured` lane re-imported an endpoint with an explicit `ttl`, the `minimal` lane the provider-default 3600 read-back) | — (lossless: origin, ttl, and endpoint read back exactly; no certificate on either lane -- the certificate + custom-domain arm is an environmental-gate deferral) |
| `digitaloceanspaceskey` | **offline-validated only** (map conformance-green; `{access_key}`). Both scenarios skip the round-trip by annotation: `digitalocean_spaces_key` ships NO importer at the provider pin, and the `secret_key` exists only in the create response, so no import could ever restore it -- and the `not_importable_upstream_reason` form was deliberately NOT used, because its reconcile-apply re-CREATES the resource, which for a key mints a SECOND access key and orphans the first (the catalog row records the missing importer; re-evaluate on a pin bump that adds one) | — |
| `digitaloceankubernetesnodepool` | 2026-09-17, both scenarios (`{node_pool_id}` bare UUID -- the provider's importer recovers the owning cluster by scanning the account, and it found the pool under the fixture cluster on both lanes; the `full` lane re-imported an autoscaling 1-2 pool with two labels, a `PreferNoSchedule` taint, and two user tags beside the six Planton identity tags). Lossless: every tag read back exactly as sent (the provider filters only its own `k8s:`/`terraform:` machinery tags), labels and the taint read back verbatim, and `node_count` produced no plan because the provider suppresses its diff against `actual_node_count` (1 on both lanes) | — (the map declares no tolerances and none was needed) |
| `digitaloceandatabasecluster` | 2026-09-17, all three scenarios (`{cluster_id}` UUID from the stack output; the `minimal` lane a single-node PostgreSQL 16, the `valkey` lane a Valkey 8 cluster with an eviction policy, the `mysql-full` lane a VPC-attached MySQL 8.4 cluster with custom storage, a maintenance window, sql_mode, and a user tag beside the six Planton identity tags) | `minimal`: nothing -- lossless (every unconditionally read-back attribute agreed: tags exactly as sent, `storage_size_mib`, the Optional+Computed `private_network_uuid`/`project_id` that populate from the region's default VPC and the default project). `valkey`: `eviction_policy` alone. `mysql-full`: `maintenance_window` and `sql_mode` -- exactly the declared write-normalized attributes (the provider's Read fetches each only when already in state, so a blind import leaves them empty and the first plan re-asserts the configured value as a server-side no-op); `backup_restore` and `storage_autoscale` stayed unexercised (deferred arms) |
| `digitaloceandatabasedb` | 2026-09-17, the `minimal` scenario (`{cluster_id},{database_name}` composite, both from the stack outputs; a logical database on the PostgreSQL cluster fixture) | nothing -- lossless (the resource has two create-only arguments and the provider's Read is a bare existence check) |
| `digitaloceandatabaseuser` | 2026-09-17, both scenarios (`{cluster_id},{user_name}` composite from the stack outputs; `minimal` a PostgreSQL user declaring `settings: {}`, `mysql-auth` a MySQL user with `mysql_native_password` and no settings) | `minimal`: `settings` alone -- the declared config-only attribute (the provider stores settings only from the create response and never reads them back, so the blind import left the attribute empty and the first plan re-asserted the configured empty block, a 201 no-op on PostgreSQL). `mysql-auth`: nothing -- lossless (a MySQL manifest declares no settings, and the API would refuse a settings update there; `mysql_auth_plugin` read back as sent) |
| `digitaloceandatabaseconnectionpool` | 2026-09-17, both scenarios (`{cluster_id},{pool_name}` composite from the stack outputs; `minimal` a transaction pool as doadmin, `inbound-user` a session pool with the user omitted) | nothing on either -- lossless (every argument is create-only upstream and reads back exactly as sent; the omitted `user` reads back empty and stable) |
| `digitaloceandatabasefirewall` | 2026-09-17, both scenarios (the BARE `{cluster_id}` from the stack output -- the provider mints a random state id that is never compared; `minimal` three IPv4 shapes, `full` an IPv4 rule plus a Droplet by reference) | nothing on either -- lossless (the rule set reads back as an unordered set with server-side uuid/created_at noise the provider ignores; the declared `port_range`-style normalization does not exist on this resource) |
| `digitaloceandatabasereplica` | 2026-09-17, both scenarios on Terraform (`{cluster_id},{replica_name}` composite from the stack outputs; `minimal` and `sized` on the PostgreSQL cluster fixture -- `size`, `storage_size_mib`, and all eight tags read back exactly as sent) | nothing -- lossless on both |
| `digitaloceanbucket` | 2026-09-17, both scenarios (`{region},{bucket_name}` composite from the `region` and `bucket_id` outputs; the `full` lane re-imported 3 resources -- the bucket plus its CORS configuration and policy satellites, all on the same composite; the `minimal` lane the one-resource, provider-default-region bucket whose `region` output read back `nyc3`) | `force_destroy` (the importer hardcodes it to false; the `full` lane's configured `true` re-planned as the one in-place update) and `acl` (write-only; Read never calls GetBucketAcl) on both lanes -- exactly the two declared config-only tolerances and nothing else; the declared `policy` write-normalization never fired because the configured JSON already matched the provider's normalized read-back |
| `digitaloceandropletautoscalepool` | 2026-09-16, both scenarios on Terraform (`{pool_id}` UUID; blind re-import, then a plan whose ONLY change is the declared image normalization). The kind itself is NOT proven -- its destroy fails on an upstream waiter defect -- but the round-trip evidence stands | `droplet_template.image` (write-normalized: the API reads the configured slug back as a numeric image id, so the first post-import plan updates it back to the slug; re-sending the identical template is a no-op on DigitalOcean's side -- measured, no member roll). `droplet_template.public_networking` (config-only, dead on write) did not fire |
| `digitaloceandatabasekafkatopic` | 2026-09-17, both scenarios on Terraform (`{cluster_id},{topic_name}` composite from the stack outputs; a topic on the Basic-plan Kafka fixture) | `partition_count` on both lanes -- the ONLY tolerated diff (`0 -> 3`): the provider's Read echoes the value already in state, which is empty after an import, and the reconcile-apply's equal-count update is an API no-op. The whole config block, `min_insync_replicas` included, read back exactly as sent, so the previously declared write-normalized tolerance for it was retired from the catalog |
| `digitaloceandatabasekafkaschema` | 2026-09-17, the single scenario on Terraform -- the type is declared `not_importable_upstream_reason` (the importer restores only `cluster_id` at the pin), so the runner skipped the import and the reconcile-apply RE-REGISTERED the subject; the plan after it was clean | the re-create itself, tolerated as not-importable-upstream: safe for this type because an identical re-register answers 201 with the same schema id and version (measured), where the same form would mint a second Spaces key |
| `digitaloceanreservedip` | 2026-09-17, all three scenarios on Terraform (`{reserved_ip_address}` bare address for both families; the `ipv6-assigned` lane also re-imported `digitalocean_reserved_ipv6_assignment` blind by `{ip},{droplet_id}` with `droplet_id` derived from the reference-resolved spec -- 2 resources on that lane) | nothing -- lossless on all three; the declared config-only `droplet_id` never fired because the importer restores an attached droplet |
| `digitaloceanvpcpeering` | 2026-09-17, the single scenario on Terraform (`{peering_id}` UUID passthrough; a dual-VPC-fixture scenario) | nothing -- lossless |

Kinds where an import map is **deliberately not applicable** (recorded so
absence is never mistaken for an oversight):

| Component | Why no import map |
|-----------|-------------------|
| `kubernetesmanifest` | The kind's state is arbitrary user YAML — there is no per-kind resource schema for a blind round-trip oracle to compare against; each deployment's resource set is defined by the manifest itself. Adopting existing raw resources is done by pasting their YAML into `spec.manifest_yaml` and applying (server-side apply takes ownership of unmanaged fields). |
| `kubernetesgatewayapicrds` (and other multi-document CRD-bundle installers) | The module applies a fetched multi-document bundle with one `kubectl_manifest` per document, keyed by the SPLIT INDEX — a positional key, never identity. The composed import ID needs each document's own apiVersion/kind/name triple, and no single derivation can honestly feed three placeholders per positional address. Adopting an existing CRD install is done by applying over it (server-side apply with force_conflicts takes ownership). |
| `kubernetesistiobasecrds` | Same multi-document CRD-bundle class as `kubernetesgatewayapicrds` — the istio/base CRD bundle applies one `kubectl_manifest` per split document, positionally keyed. |

## Adding a kind

1. Add the component's resource types (with import-ID formats, or
   `not_importable_upstream_reason` for importer-less types) to the
   provider catalog if absent; declare any config-only attributes. When a
   type is listable via Cloud Control (`aws cloudcontrol list-resources
   --type-name ...` succeeds without extra parameters -- verify
   empirically, never assume), also declare its `cloud_control_type_name`:
   it is the scan-side correspondence account scanning and the mapping
   eval harness (`pkg/iac/mappingeval`) translate through.
2. Author `{component}/v1/iac/import-map.yaml` naming each placeholder's
   derivations (prefer derivable sources; `where_to_find` is mandatory when
   nothing derives). The file's presence enrolls the component in every
   check — nothing else to register.
3. Run the live round-trip lane
   (`PLANTON_E2E_IMPORT_ROUNDTRIP=1 go test -tags=e2e -run Test<Kind>_Terraform ./e2e/...`)
   before merging, and record the result in the ledger above.
