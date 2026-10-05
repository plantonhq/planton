# Kubernetes Charts: GitOps Delivery, Software Supply Chain, Identity and Access

**Date**: August 5, 2026
**Type**: Feature
**Provider**: Kubernetes
**Chart(s)**: `charts/kubernetes/gitops-delivery-platform`, `charts/kubernetes/software-supply-chain`, `charts/kubernetes/identity-and-access-platform`

## Summary

The catalog's developer-platform trio lands, completing the sixteen-chart
Kubernetes-era catalog. The **GitOps Delivery Platform** makes Git the
deployment truth from day one: Argo CD continuously syncing the cluster
from repositories, Argo Workflows running the pipelines beside it, with
workflow artifacts and archived logs on an in-cluster S3 store and run
history in a CloudNativePG-managed PostgreSQL — both wired entirely by
reference. The **Software Supply Chain** puts a Harbor registry with
Trivy scanning at the center, every data arm composed from first-class
kinds (PostgreSQL metadata, an authenticated Valkey cache, S3 blob
storage — the shape that lets the registry scale out), with Tekton
pipelines and an optional scale-to-zero GitHub Actions runner fleet
beside it. The **Identity and Access Platform** is the self-hosted
security triad — Keycloak for single sign-on, OpenFGA for fine-grained
authorization, OpenBao for secrets — over one PostgreSQL bootstrapped
with both databases under one least-privilege owner role.

## Kind fixes the compositions surfaced

Charts are the composition proof of the kind catalog, and this trio
surfaced two seam gaps in `KubernetesArgoWorkflows`, both fixed at the
root:

- **The S3 artifact credentials Secret now composes by reference.** The
  field was a plain string with fixed `accesskey`/`secretkey` key names,
  and its comment wrongly claimed a KubernetesSeaweedFs credentials
  Secret already had that shape — the documented composition deployed
  cleanly and failed artifact authentication at runtime. The credentials
  are now a Secret reference (defaulting to the SeaweedFs
  `s3_credentials_secret_name` output) with key-name fields defaulting
  to that store's generated-Secret convention.
- **The archive credentials Secret now composes by reference** the same
  way — defaulting to the KubernetesPostgres application Secret, whose
  `username`/`password` keys are the selector defaults.

## Notable design decisions

- **Declared-on-both-sides credentials where fixed key contracts meet**:
  Harbor's cache and S3 storage arms render the same chart-controlled
  values the producing stores enforce — the shape proven live by the
  Harbor composed-storage lane — with letters-only credential parameters
  (config parsers crash intermittently on structural characters).
- **One toggle per arm whose halves are useless apart**: Tekton's
  operator and its TektonConfig declaration ride a single toggle (the
  config never reconciles without the operator; both are one-per-cluster),
  as do the GitHub Actions runner controller and scale set (whose chart
  versions GitHub supports only in lockstep).
- **The runner fleet defaults off**: it needs a GitHub URL and a
  pre-created credential Secret that nothing in the chart's graph
  produces — external credentials never ride chart parameters.
- **Keycloak ships behind-proxy correct**: plain pod HTTP paired with
  X-Forwarded header trust and a full-URL hostname parameter — the shape
  that satisfies the server's hostname pairing rules from day one.
- **OpenBao's seal lifecycle is taught, never faked**: pods report
  NotReady by design until `bao operator init` and unseal — runtime
  operations the chart deliberately leaves to the operator's runbook.

## Validation

All three charts green on `planton chart validate` — defaults plus every
bool flip, feature combinations beyond the flips, and a hyphenated
environment probe; the full-catalog sweep passes 16/16; icon URLs verified
live; offline Terraform plan and Pulumi preview proofs cover both new
kind reference arms.
