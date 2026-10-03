# Kubernetes Kyverno

## When NOT to Use This

**One resource is ONE Kyverno install** — the Kubernetes-native policy
engine: validation, mutation, generation, and cleanup driven by
policies written as Kubernetes resources.

Not the right kind when:

- **You want the policies themselves** — this kind installs the ENGINE.
  ClusterPolicy / Policy resources (and the policies.kyverno.io v1
  families) are applied separately: `KubernetesManifest` resources or
  GitOps, once the engine is running.
- **Your policy team lives in Rego/OPA** — that is
  `KubernetesGatekeeper`. Kyverno's edge is policies AS Kubernetes
  YAML (no new language) plus mutation, generation, image
  verification, and cleanup; Gatekeeper's is the CNCF OPA constraint
  framework.
- **You want one engine per team** — webhook registration and the
  policy CRDs are cluster-global: one Kyverno per cluster.

## The webhook lifecycle (what everything follows from)

The chart templates NO webhook configurations. The admission
controller REGISTERS them at runtime and keeps them scoped to the
installed policies. Uninstall runs the chart's pre-delete cleanup
hook (on by default — leave it on) and a module-owned cleanup that
deletes `kyverno-*` webhook configurations by label after the release
is gone — needed because the chart helper at the pinned Kyverno
release does not reliably remove validating webhook configurations.
A release force-deleted without either path strands those configs
with no backing service, and since policy rules default to
fail-CLOSED, everything they match stops admitting until they are
deleted by hand.

## CRDs and destroy

The ~23 policy CRDs are chart-templated: installed and DELETED with
the release, cascade-deleting every policy and report on the cluster.
`crds.keep_on_uninstall` preserves them (a later install must then set
`crds.install: false` — kept CRDs carry this release's Helm ownership
metadata).

## The four controllers

Admission (the webhook server — the only one you cannot disable),
background (mutate-existing/generate), cleanup (CleanupPolicy), and
reports (PolicyReport aggregation). Each is sized and scheduled
independently; the admission controller also takes an HPA. Run 3
admission replicas in production — it sits on the cluster's write
path.

## Certificates

Default: Kyverno generates and rotates its own webhook certificates —
zero prerequisites. The `certificates.cert_manager` arm delegates
issuance to cert-manager (compose `KubernetesCertManager` first) and
applies to BOTH webhook servers (admission and cleanup).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
