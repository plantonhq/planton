# Kubernetes External Secret

## When NOT to Use This

**If the secret's value is known at deploy time, use KubernetesSecret instead.** A KubernetesExternalSecret asks the External Secrets Operator to keep a Kubernetes Secret SYNCED from an external backend for as long as the resource exists — the external system stays the single source of truth. Values available in your CI/CD pipeline or IaC config belong in a KubernetesSecret; consumers reference a Kubernetes Secret either way.

## Overview

**KubernetesExternalSecret** declares ONE secret sync: the External Secrets Operator reads the referenced entries from a store's backend (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, Vault/OpenBao, ...) and materializes them as a Kubernetes Secret in this namespace, refreshing on the configured interval. The materialized Secret (`secret_name`) is the handle every consumer references: env `valueFrom`, volume `secretName`, image pull secrets.

This is the sync-declaration member of the External Secrets family. The machinery (KubernetesExternalSecretsOperator) and the backend connections (KubernetesSecretStore / KubernetesClusterSecretStore) are separate first-class resources; this resource picks WHAT to sync:

- **`data`** — explicit entries: each maps one backend key (or one `property` within a structured entry) to one key of the materialized Secret, with per-entry `version` pinning and decoding strategy. The precise, reviewable form — prefer it for application credentials
- **`data_from.extract`** — pull ALL properties of one structured backend entry (a JSON document of related credentials); each property becomes a Secret key
- **`data_from.find`** — pull EVERY backend entry matching a name regex and/or tags; the fleet pattern
- **`rewrite`** — regex key rewrites on bulk pulls (e.g. strip a `prod/app/` path prefix so Secret keys are bare names)
- **`target`** — the materialized Secret's name, lifecycle policies (`creation_policy` Owner/Orphan/Merge/None, `deletion_policy` Retain/Delete/Merge), immutability, and a `template` that sets the Secret's type, stamps metadata, or reshapes values with Go templates (e.g. render a `kubernetes.io/dockerconfigjson` from synced registry credentials)
- **Refresh** — `refresh_interval` (upstream default `1h`; `0s` = fetch exactly once) and `refresh_policy` (CreatedOnce / Periodic / OnChange)

Because the store carries the connection and this resource carries only WHAT to sync, the same declaration works against any backend on any cluster — an EKS cluster syncing from GCP Secret Manager or a GKE cluster syncing from Vault reads identically; only the store differs.

## Deploys never block on the sync

The materialized Secret appears when the operator reaches the backend — not as part of applying the resource. Neither engine waits for the sync to complete; `kubectl get externalsecret -n <namespace>` shows sync status, and composition (not waiting) is how consumers express the dependency.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: where the ExternalSecret and its materialized Secret live
- **`spec.store_ref`**: the store to sync from — `secret_store.name` (a namespaced store: FK to a KubernetesSecretStore's `store_name` output) or `cluster_secret_store.name` (a cluster-scoped store: FK to a KubernetesClusterSecretStore's `store_name` output)
- At least one `data` entry or one `data_from` pull

### Common

- **`spec.target.name`**: the materialized Secret's name (defaults to `metadata.name`)
- **`spec.refresh_interval`**: how fresh the Secret stays (default `1h`)

## Outputs

| Output | Purpose |
|---|---|
| `external_secret_name` | The ExternalSecret resource name |
| `namespace` | Where the ExternalSecret and its Secret live |
| `secret_name` | The materialized Secret — the handle workloads wire env `valueFrom` / volume `secretName` references to |

## Composing in Infra Charts

`KubernetesExternalSecretsOperator → store kind → KubernetesExternalSecret → workload` deploys in one chart run: the ExternalSecret references the store's `store_name` output, and workloads (e.g. KubernetesDeployment) reference `status.outputs.secret_name` in env/volume Secret references. The external backend stays the single source of truth; the cluster never stores the value anywhere else.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
