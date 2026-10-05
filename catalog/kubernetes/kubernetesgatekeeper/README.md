# Kubernetes Gatekeeper

## When NOT to Use This

**One resource is ONE OPA Gatekeeper install** — the Open Policy
Agent's Kubernetes admission controller: constraint-based validation
and mutation with Rego or CEL, plus a continuous audit loop.

Not the right kind when:

- **You want the constraints themselves** — this kind installs the
  ENGINE. ConstraintTemplates and Constraints are applied separately:
  `KubernetesManifest` resources or GitOps, once the engine runs.
- **Your team wants policy as plain Kubernetes YAML** — that is
  `KubernetesKyverno`. Gatekeeper's edge is the OPA constraint
  framework (Rego's expressiveness, the template/constraint split,
  external data providers); Kyverno's is no-new-language policies with
  mutation/generation/cleanup.
- **You want one engine per team** — the webhook configurations and
  engine CRDs are cluster-global, and the chart HARDCODES its resource
  names: one Gatekeeper per cluster, by construction.

## Fail-open by default (and the one fail-closed piece)

The policy webhook ships `failurePolicy: Ignore`: an engine outage
never blocks cluster admissions — requests during one simply go
unevaluated. Flip `validating_webhook.failure_policy` to `Fail` for
enforcement without that gap, with three replicas and a short timeout
(an outage then blocks every matched admission). The exception is the
namespace-label check webhook (default `Fail`): it guards Gatekeeper's
own exemption label so nobody exempts a namespace during an outage;
its blast radius is namespace label edits only.

## Audit

The audit controller re-evaluates EXISTING resources on an interval
and records violations in each constraint's status — how you adopt
policy on a brownfield cluster (audit first, enforce when boring).
`match_kind_only` and `chunk_size` are the API-load levers at scale.

## CRDs and destroy

Engine CRDs ship in the chart's `crds/` directory: installed once,
never chart-upgraded (the chart's own upgrade-CRDs hook compensates on
upgrades), and KEPT on uninstall — as are the per-template constraint
CRDs Gatekeeper creates at runtime. Destroying the engine does not
delete your policy library; the webhook configurations DO delete with
the release (chart-owned).

## Certificates

Default: the embedded cert-controller generates and rotates the
webhook certificate — zero prerequisites. The `external_cert` arm
serves a cert-manager-issued Secret instead (compose
`KubernetesCertManager` + `KubernetesCertificate`); the module
disables the embedded rotator on both deployments (the chart only
auto-disables it on one — chart-truth the module compensates for).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
