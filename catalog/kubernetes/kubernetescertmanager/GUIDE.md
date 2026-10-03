# KubernetesCertManager Guide

The judgment this guide carries: cert-manager is the root of the
certificate chain and the easiest layer to forget — it does nothing
visible by itself, so an architecture missing it fails silently later,
when certificates never get signed.

## Install once, in the shared-cluster chart

One installation serves the whole cluster: every KubernetesClusterIssuer,
KubernetesIssuer, and KubernetesCertificate on the cluster depends on this
single controller machinery (the component split is on
[reference.md](v1alpha1/reference.md)). It belongs in the shared-cluster chart
beside the ingress controller — never in an application environment.
The full public-HTTPS composition checklist lives in the
[KubernetesClusterIssuer guide](../kubernetesclusterissuer/GUIDE.md);
this kind is its step one.

## The namespace is effectively permanent

`createNamespace: true` with a dedicated "cert-manager" namespace is the
normal single-tenant shape (the
[namespace-ownership pattern](../../_patterns/namespace-ownership.md)'s
sole-tenant case). But note the reference page's CRD-retention warning
before ever planning to move or rebuild the installation: with kept CRDs,
the namespace choice outlives the release, and unwinding it cascades to
certificate data cluster-wide. Choose the namespace once, conventionally,
and leave it.

## On the diagram

cert-manager renders as a shared-cluster node with the cluster issuer's
reference edge into it — the visible root of the certificate layer. Its
absence is invisible in any single application's manifests, which is
exactly why the proposal step must add it.

## Pairs well with

- KubernetesClusterIssuer / KubernetesIssuer — the signing authorities;
  wired to this component via `certManagerNamespace`.
- KubernetesCertificate — explicit certificates, including the wildcard
  default the ingress controller's `defaultTlsCertificate` can reference.
- KubernetesIngressNginx — the entry point whose TLS this machinery
  ultimately serves.
