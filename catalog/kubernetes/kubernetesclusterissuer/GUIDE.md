# KubernetesClusterIssuer Guide

The judgment this guide carries: an architecture with a public HTTPS
endpoint is INCOMPLETE without a certificate issuer — and proposing the
endpoint without the issuer chain is the single most common omission in
agent-composed cluster architectures, because nothing in the endpoint's own
manifest fails without it. Certificates just never get signed.

## The composition checklist for public TLS

When an architecture exposes anything over public HTTPS, it needs this
chain, in this order — typically in the shared-cluster chart, not the
application environment:

1. **[KubernetesCertManager](../kubernetescertmanager/GUIDE.md)** —
   the controller. Cluster-scoped, once per cluster. Without it, a
   ClusterIssuer is a custom resource nothing reads.
2. **KubernetesClusterIssuer** (this component) — the certificate authority
   front-end, e.g. ACME/Let's Encrypt for public endpoints. Its
   `spec.certManagerNamespace` is a foreign key defaulting to the
   KubernetesCertManager resource's output — wire it with `valueFrom` so
   the dependency is explicit and ordered.
3. The endpoint's certificate — a
   [KubernetesCertificate](../kubernetescertificate/GUIDE.md) or the
   `cert-manager.io/cluster-issuer` ingress annotation (the
   [Ingress guide](../kubernetesingress/GUIDE.md) carries that lane;
   [Gateway listeners](../kubernetesgateway/GUIDE.md) consume the
   explicit kind's Secret), selecting this resource by name (the
   ClusterIssuer is named after `metadata.name`).

If the user asked for "a standard cluster my app's public endpoint runs
on", steps 1 and 2 belong in the proposal even though the user never said
"certificate" — that is what makes the endpoint actually serve HTTPS.

## Choosing this kind vs KubernetesIssuer

Same signing capabilities, different scope (the spec doc on
[reference.md](v1alpha1/reference.md) carries the full comparison): one platform-wide
"letsencrypt-production" ClusterIssuer serving every namespace is the
default shape; reach for the namespace-scoped KubernetesIssuer only when a
namespace needs its own CA or its own credential blast radius.

## On the diagram

The cert chain renders as its own visible layer: cert-manager and the
cluster issuer as shared-cluster nodes, with the issuer's `valueFrom` edge
into cert-manager. An architecture whose diagram shows the endpoint but no
issuer chain is showing a promise it cannot keep.

## Pairs well with

- KubernetesCertManager — required, once per cluster, wired via
  `spec.certManagerNamespace`.
- KubernetesCertificate — explicit certificates selecting this issuer.
- Ingress/gateway-bearing workloads whose public endpoints the issuer
  signs for.
