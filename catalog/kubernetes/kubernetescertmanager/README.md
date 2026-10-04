# Kubernetes Cert Manager

## When NOT to Use This

**If you only need a certificate, you do not need this component directly.** One cert-manager installation per cluster serves every issuer and certificate on it — check whether your cluster already runs cert-manager before adding a second (a second installation will fight the first over the same cluster-scoped CRDs and webhooks). Who signs certificates is a separate concern: create KubernetesClusterIssuer / KubernetesIssuer for signing authorities and KubernetesCertificate for the certificates themselves.

## Overview

**KubernetesCertManager** installs cert-manager — the cluster's certificate machinery — from the official Helm chart (`cert-manager` at `https://charts.jetstack.io`). cert-manager runs three components: the **controller** (watches Certificates and drives issuance), the **webhook** (validates cert-manager resources at admission), and the **cainjector** (injects CA bundles into webhook configurations). This component installs and configures that machinery; the typed spec covers the chart's meaningful configuration surface, with a `helm_values` escape hatch (merged last, Helm `-f` semantics, identical on both engines) for anything beyond it.

**Key design points:**

- **CRDs install with the release by default** (`crds.install`, Planton default true) and **survive uninstall by default** (`crds.keep_on_uninstall`, true) — deleting CRDs cascades to every Certificate and Issuer object cluster-wide, so that destructive act requires an explicit `false`
- **Workload identity is first-class**: `workload_identity` binds the controller ServiceAccount to a cloud identity (GKE Workload Identity, EKS IRSA, AKS Workload Identity) for **keyless DNS-01** — issuers whose DNS providers leave static credentials empty authenticate through this identity
- **Split-horizon DNS is solvable in the spec**: `dns01_self_check` points cert-manager's pre-flight TXT lookups at public resolvers, the standard fix when in-cluster DNS serves a private view and DNS-01 issuance hangs
- **The install waits for real readiness**: both engines wait for the chart's post-install check Job that proves the webhook actually serves — a cert-manager whose webhook is down rejects every Issuer/Certificate apply, so a premature "success" would just move the failure downstream

## Environment Injection (where cloud configuration flows in)

cert-manager talks ACROSS clouds: it runs IN one environment and its DNS-01 solvers write records wherever the DNS lives. Cloud identity reaches it through the ServiceAccount annotation the chart stamps from `workload_identity`:

| Host environment | Identity mechanism | Spec field | Cloud-side half |
|---|---|---|---|
| GKE | Workload Identity | `workload_identity.gke.service_account_email` | `roles/iam.workloadIdentityUser` binding for `<project>.svc.id.goog[<namespace>/cert-manager]` |
| EKS | IRSA | `workload_identity.eks.role_arn` | IAM role trust policy on the cluster OIDC provider for `system:serviceaccount:<namespace>:cert-manager` |
| AKS | Azure AD Workload Identity | `workload_identity.aks.client_id` | Federated credential for `system:serviceaccount:<namespace>:cert-manager` (the chart also stamps the required pod label) |
| Any (token-based DNS) | none needed | — | Cloudflare/DigitalOcean/RFC2136 tokens ride the issuer, not the controller |
| kind / datacenter / self-managed | none | — | Use token-based DNS providers or HTTP-01 |

The controller ServiceAccount name is exported (`status.outputs.service_account_name`) precisely so the cloud-side half can be composed in the same Infra Chart.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: installation namespace (`cert-manager` by convention) — literal or a KubernetesNamespace reference

### Common

- **`spec.create_namespace`**: create (and own) the namespace with the release
- **`spec.chart_version`**: pinned chart/app version (defaults to the validated pin)
- **`spec.cluster_resource_namespace`**: where cert-manager reads Secrets for CLUSTER-scoped resources (ClusterIssuer credentials); defaults to the installation namespace — exported as `status.outputs.cluster_resource_namespace`
- **`spec.dns01_self_check`**: public resolvers for DNS-01 pre-flight checks (split-horizon fix)
- **`spec.webhook.host_network` + `spec.webhook.secure_port`**: the EKS-custom-CNI fix when the control plane cannot reach pod IPs
- **`spec.prometheus.service_monitor`**: opt-in ServiceMonitor (requires the Prometheus operator CRDs — the release fails without them)
- **`spec.helm_values`**: escape hatch for chart values beyond the typed fields — never the primary interface

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Installation namespace |
| `release_name` | Helm release name (always `cert-manager`) |
| `service_account_name` | Controller ServiceAccount — the identity to bind cloud-side for keyless DNS-01 |
| `cluster_resource_namespace` | Where ClusterIssuer credential Secrets belong (KubernetesClusterIssuer's FK default) |

## Composing in Infra Charts

The standard chart wiring: this component first, then issuers referencing `status.outputs.cluster_resource_namespace`, then certificates referencing the issuers. Catalog kinds (an IAM role for Route53, a GCP service account for Cloud DNS) deploy in the same run and flow their handles into `workload_identity` — the cross-cloud composition (e.g. EKS cluster + Cloudflare DNS) needs no identity at all, since the Cloudflare token rides the issuer.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
