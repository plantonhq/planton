# Kubernetes networking primitives at full configuration depth: Service rebuilt, Ingress and NetworkPolicy forged

**Date**: 2026-07-21
**Scope**: `apis/dev/planton/provider/kubernetes` (kubernetesservice rebuilt; kubernetesingress + kubernetesnetworkpolicy new), `apis/dev/planton/iac/catalogkindimportmap`, `apis/dev/planton/iac/providerimportcatalog`, `pkg/iac/importmap`, `pkg/outputs`, `aa_import`, `aa_e2e`, `e2e/framework/runner`, Makefile E2E tiers, site catalog, `_rules/catalog-kind/forge`

## What changed

### KubernetesService rebuilt to the complete core/v1 surface

The previous spec self-described as "80/20". The rebuilt spec covers the
complete meaningful core/v1 ServiceSpec, designed field-by-field from the
upstream API types:

- Dual-stack: `ip_families` (order-preserving) + `ip_family_policy`
  (SingleStack / PreferDualStack / RequireDualStack).
- Traffic shaping: `internal_traffic_policy`, `traffic_distribution`
  (PreferSameZone / PreferSameNode), `external_traffic_policy` (unchanged).
- LoadBalancer tuning: `load_balancer_class`,
  `allocate_load_balancer_node_ports` (tri-state),
  `health_check_node_port`, source ranges (now CIDR-validated).
- Session affinity timeout (`session_affinity_timeout_seconds`),
  `publish_not_ready_addresses`, `external_ips`, static
  `cluster_ip_address`, per-port `app_protocol`.
- The namespace became a `StringValueOrRef` foreign key to
  KubernetesNamespace, matching every current-anatomy kind.
- Twelve cross-field CEL rules mirror live kube-apiserver rejections
  (ExternalName exclusions, headless incompatibilities, LB-only knobs,
  health-check gating, dual-stack arity) so misconfigurations fail at
  validation instead of apply. The Service name rule is the stricter
  DNS-1035 label (must start with a letter) — verified against a live
  API server, which rejects leading digits for Services.
- Outputs converge on the catalog's handle vocabulary: `kube_endpoint`,
  `port_forward_command`, plus separate `load_balancer_ip` /
  `load_balancer_hostname` handles for DNS automation.

The single deliberate omission is deprecated `loadBalancerIP` (upstream
deprecated it as non-portable; provider annotations are the portable path).
One documented parity exception: `traffic_distribution` is not exposed by
the Terraform kubernetes provider (v3.2.x), so the Terraform module fails
the plan loudly via a precondition when it is set, instead of silently
dropping it; the Pulumi engine applies it.

### KubernetesIngress (new, enum 813) — first-class HTTP(S) exposure

Complete networking/v1 IngressSpec: `ingress_class_name`, default backend,
TLS blocks (SNI-only or named secret), host rules with wildcard-host
validation, all three path types with per-type path requirements, and both
backend port forms (number XOR name, enforced by CEL as the API does).

Composition is the design center: backend `service_name` is a
`StringValueOrRef` foreign key (default kind KubernetesService) — the wiring
point for a workload's exported `service` output — and TLS `secret_name` is
a foreign key to KubernetesSecret, the seam cert-manager fills. Backend
references are deliberately Planton FKs unlike the Gateway API projection
kinds (which keep plain upstream refs for 100% CRD fidelity): Ingress is a
core-API kind where the catalog owns the shape; the spec comments teach the
distinction.

Both engines deliberately create the Ingress WITHOUT blocking on an ingress
controller (Terraform `wait_for_load_balancer = false`, Pulumi `skipAwait`):
an Ingress is a valid object with no controller installed, and infra charts
routinely deploy workload + exposure before the controller wave. The
load-balancer address handles (`load_balancer_ip`, `load_balancer_hostname`)
surface through outputs as soon as a controller reconciles the object;
`first_host` exports the primary public FQDN.

### KubernetesNetworkPolicy (new, enum 814) — the in-cluster firewall

Complete networking/v1 NetworkPolicySpec: label-selector pod selection
(match_labels + match_expressions with the operator/values contract
enforced), explicit or inferred `policy_types`, ingress/egress allow rules
with all three peer forms (pod selector, namespace selector, IP block with
`except` carve-outs), the AND'd pod+namespace peer semantics, named ports,
and `end_port` ranges (numeric-anchor and ordering rules enforced at
validation, as the API does).

Both engines ALWAYS submit policy_types explicitly — when the spec omits
them, the modules apply the API server's own inference (ingress always,
egress only with egress rules) — so both engines submit byte-identical
direction sets and the `policy_types` output always states the deployed
truth. CEL rules reject rules in an ungoverned direction (silently ignored
upstream) and empty peers (match nothing).

The spec and docs are explicit that enforcement requires a
NetworkPolicy-implementing CNI; on clusters whose CNI ignores them the
object exists but traffic flows.

### Import recipes proven end-to-end (and the machinery they needed)

All three kinds ship import maps, and the live blind re-import round-trip
is green for them AND for all five workload kinds (their maps had never
been round-trip-proven):

- The kubernetes provider catalog now declares the provider's
  configuration-only lifecycle knobs (`wait_for_rollout`,
  `wait_for_load_balancer`, `wait_for_service_account_token`) so
  post-import plans that touch only those are recognized as the documented
  shape.
- `write_normalized_attributes`/`config_only_attributes` entries may now be
  a dotted sub-path (e.g. `"spec.update_strategy"`, which the provider's
  StatefulSet importer does not read back): the round-trip oracle prunes
  exactly that sub-path from both plan sides and requires the remainder
  identical, so sibling drift inside the same attribute still fails.
- New `from_metadata_name_suffix` derivation in the catalog kind import map
  vocabulary: convention-named satellites (a workload's
  `<name>-env-secrets` Secret) are now blind-derivable without exporting
  their names as outputs. The deployment map's HPA recipe was also
  corrected: the HPA shares the workload's name (both engines), not a
  `-hpa` suffix as the map's guidance claimed.
- The live-proven ledger records all eight kinds.

### E2E

- 21 scenarios across the three kinds run the full six-phase runner on both
  engines against the kind cluster — including a composed Ingress scenario
  that deploys a real KubernetesService fixture through the
  e2e-prerequisites annotation and resolves the backend through the
  foreign-key reference, proving the exposure composition end to end. A
  `prerequisite.yaml` install profile ships with kubernetesservice for
  consumers.
- Tier-1 Makefile targets now include ConfigMap, ServiceAccount, Rbac,
  Service, Ingress, and NetworkPolicy.
- Verifier cases added for ingress and networkpolicy (existence-based: an
  Ingress without a controller and a NetworkPolicy without an enforcing
  CNI are both valid objects).

### Docs, presets, catalog

Full four-doc set per kind (kind README, research doc, per-engine
module READMEs), catalog pages, and four presets each (Service: ClusterIP
app / public LB / headless StatefulSet / ExternalName; Ingress: single host
/ TLS with cert-manager / fanout / default backend; NetworkPolicy:
default-deny-all / allow-same-namespace / allow-from-namespace /
allow-dns-egress). Site catalog regenerated. Workload spec comments now
name KubernetesIngress alongside the Gateway API kinds as the exposure
composition path.

### Forge workflow

Two timeless lessons folded into the forge rule: declare a provider's
`wait_for_*`-class knobs as config-only when the resource type first enters
the import catalog (each consumer pays the round-trip failure otherwise),
and the profile-honesty mechanics for new kinds (the harness skips
deferred profiles, so the profile flips green immediately before the lanes
run, and reverts in the same commit if they fail).

## Validation

- Spec tests green for all three kinds (protovalidate suites; positive and
  negative cases per CEL rule).
- Offline gates green: outputs conformance (three new cases), import-map
  conformance, secret-coverage, validate-refs, kind-map regen, e2e-matrix,
  `make build-go`.
- Offline tofu plan proofs per kind with optionals present AND absent, plus
  a negative proof that the traffic_distribution precondition fails the
  Terraform plan.
- Live dual-engine E2E on the persistent kind cluster: Service 13 scenarios,
  Ingress 4, NetworkPolicy 4 — both engines, zero orphaned resources.
- Live import round-trip green for kubernetesservice, kubernetesingress,
  kubernetesnetworkpolicy, kubernetesdeployment, kubernetesstatefulset,
  kubernetesdaemonset, kubernetesjob, kubernetescronjob.
- Not verified: real-cluster (EKS/GKE/AKS) lanes and LoadBalancer-type
  provisioning (kind has no cloud LB controller); NetworkPolicy behavioral
  traffic-blocking (kind's default CNI does not enforce policies).
