# KubernetesHelmRelease Guide

The judgment this guide carries: this is the catalog's last resort, and
the mistake it protects against is reaching for it second. When a user
asks for software by name, the answer is a first-class kind or a verified
compatible alternative — the search-first workflow lives in the
[catalog guide](../../_docs/GUIDE.md); run it before declaring any chart here.

## What a chart install costs the architecture

A typed kind validates configuration before deploy, exports outputs
other resources reference by `valueFrom`, and documents its trade-offs
field by field. A Helm release does none of that: values are opaque to
validation until Helm renders them, and nothing it creates is referenceable
— no output edges in, none out. Whatever the chart deploys, the platform
cannot wire the rest of the architecture to it.

## When it is legitimately the right call

The catalog has no kind for the software AND no compatible alternative —
a vendor's proprietary chart, or your own in-house chart. Then this kind
is the honest answer, and its spec is a faithful Helm surface: pinned
`version` (required — reproducibility is the point), Helm's own values
precedence, real hooks and release history (details on
[reference.md](v1alpha1/reference.md)). Say plainly in the proposal that the
catalog has no first-class kind yet.

## Namespace ownership

`spec.namespace` is a required foreign key targeting KubernetesNamespace,
and `createNamespace: true` follows the same ownership contract as every
workload kind — created before the release, owned in IaC state, deleted
with the resource. Charts commonly share namespaces with their consumers;
the judgment and wiring:
[namespace-ownership pattern](../../_patterns/namespace-ownership.md).

## On the diagram

One node, regardless of how many resources the chart deploys — the
internals are invisible and nothing can draw an edge to them. An
architecture built from typed kinds shows its actual shape; one built
from chart installs shows a row of black boxes.

## Pairs well with

- KubernetesNamespace — the namespace owner (pattern above).
- KubernetesManifest — the raw-YAML sibling escape hatch: prefer THIS
  kind when upstream ships a chart (hooks, values, release lifecycle);
  prefer KubernetesManifest for plain YAML documents with no chart.
