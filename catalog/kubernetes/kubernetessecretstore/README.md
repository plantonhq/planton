# Kubernetes Secret Store

## When NOT to Use This

**When every team should share one backend connection, use KubernetesClusterSecretStore instead.** A SecretStore is namespaced: only ExternalSecrets in the SAME namespace can sync from it, and its credential Secrets live in that namespace. A ClusterSecretStore serves every namespace (optionally fenced with conditions) and anchors its credentials in one shared namespace. The backend surface is identical — scope is the only difference, and the deciding factor.

## Overview

**KubernetesSecretStore** creates one External Secrets Operator (ESO) SecretStore — a namespaced connection to an external secret backend. The store is named after the resource (`metadata.name`); ExternalSecrets in the same namespace select it by that name (`kind: SecretStore`, the upstream default).

The External Secrets family divides cleanly into three roles:

- **KubernetesExternalSecretsOperator** — the machinery: the ESO controller and its CRDs, installed once per cluster
- **KubernetesSecretStore / KubernetesClusterSecretStore** — backend connections: WHERE secrets live and HOW the operator authenticates to read them
- **KubernetesExternalSecret** — one sync declaration: WHAT to read and which Kubernetes Secret to materialize it into

Six backends cover the connection surface (the config is shared with KubernetesClusterSecretStore, so the two kinds can never drift):

| Backend | Connects to | Typical use |
|---|---|---|
| `aws` | AWS Secrets Manager / SSM Parameter Store / ACM export | Secrets hosted in AWS |
| `gcp_secret_manager` | GCP Secret Manager | Secrets hosted in GCP |
| `azure_key_vault` | Azure Key Vault | Secrets hosted in Azure |
| `vault` | HashiCorp Vault or OpenBao KV engine | Self-hosted centralized secrets |
| `kubernetes` | Another Kubernetes cluster's Secrets | Cluster-to-cluster sync |
| `fake` | Literal entries declared in the spec | Pipelines, tests, sandboxes — never real secrets |

## The Per-Team Grain

The namespaced grain is the blast-radius choice: a team's store lives in the team's namespace, its credential Secrets are readable only there, and nothing outside the namespace can sync through it. Different teams connect to different backends — or the same backend with DIFFERENT identities — without touching cluster-scoped resources. Platform-wide backends every team shares belong in a KubernetesClusterSecretStore instead.

## Any Cluster, Any Backend

The store is the environment-injection point: the cluster the operator runs on and the backend the secrets live in are independent choices. An EKS cluster can sync from GCP Secret Manager, a GKE cluster from Vault, a self-managed cluster from Azure Key Vault — cross-cloud pairs are first-class, not workarounds. The only coupling is authentication: keyless identity federation requires the cluster and backend clouds to match (IRSA is EKS-to-AWS, Workload Identity is GKE-to-GCP), while declared static credentials and Vault auth work from anywhere.

## Authentication Model

Every cloud backend supports the same three postures:

1. **Keyless via a referenced ServiceAccount** (the production posture): `service_account_name` names a ServiceAccount — a foreign key to a KubernetesServiceAccount resource, whose workload-identity arms (IRSA / GKE Workload Identity / AKS Workload Identity) carry the cloud binding. ESO exchanges that ServiceAccount's token with the cloud, so different stores can carry DIFFERENT identities — the multi-team posture. On a SecretStore the ServiceAccount defaults to the store's own namespace.
2. **Keyless via the operator's ambient identity**: leave the auth fields empty entirely; the controller's own identity (KubernetesExternalSecretsOperator's `workload_identity`, or the node identity) authorizes the read.
3. **Declared static credentials** (the fallback for hosts with no cloud identity federation): sensitive values declared in the spec (AWS access keys, a GCP service-account key, an Azure client secret, a Vault token) are materialized by the IaC modules as a Kubernetes Secret named `<metadata.name>-credentials` in the store's namespace, and the store's CR references that Secret. The credential never appears inside the CR itself.

Vault/OpenBao has its own three auth methods (`token` / `app_role` / `kubernetes`); the `kubernetes` method is the keyless-equivalent posture there — the cluster's ServiceAccount token is exchanged for a Vault token.

## Deploys never block on readiness

Store readiness depends on external reachability (the cloud secrets API, Vault) that is not part of applying the resource. Neither engine waits for Ready — the same posture as the cert-manager issuers. Check `kubectl get secretstore -n <namespace>` for live status.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: where the SecretStore (and its credential Secrets) live — the only namespace whose ExternalSecrets can use it
- **`spec.config`**: exactly one backend

### Common

- **`spec.config.controller_class`**: shard stores across multiple operator installations
- **`spec.config.refresh_interval` / `retry`**: connection re-validation and retry tuning

## Outputs

| Output | Purpose |
|---|---|
| `store_name` | The handle ExternalSecrets in the same namespace reference (`store_ref.name`, kind SecretStore) |
| `namespace` | Where the store and its credential Secrets live |

## Composing in Infra Charts

`KubernetesExternalSecretsOperator → KubernetesSecretStore → KubernetesExternalSecret` deploys in one chart run: the store lives in the team's namespace, ExternalSecrets in that namespace reference `status.outputs.store_name`, and workloads reference the ExternalSecret's `secret_name` output in env/volume Secret references. The cross-cloud pattern (cluster in EKS, secrets in GCP Secret Manager) is one store with a declared GCP service-account key — no cloud identity federation required.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
