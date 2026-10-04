# Kubernetes External Secrets Operator

## When NOT to Use This

**If you only need a secret synced, you do not need this component directly.** One External Secrets Operator installation per cluster serves every store and every synced secret on it — check whether your cluster already runs ESO before adding a second (the release name is fixed to `external-secrets` precisely because the CRDs and webhook configuration are cluster-global, and a second installation would fight the first). Which backends exist and which secrets sync are separate first-class kinds: create **KubernetesClusterSecretStore** / **KubernetesSecretStore** for the backend connections and **KubernetesExternalSecret** for each secret to sync. This is the same split as cert-manager vs. issuers/certificates: this component installs machinery ONLY.

## Overview

**KubernetesExternalSecretsOperator** installs the External Secrets Operator (ESO) — the controller that syncs secrets FROM external stores (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, Vault/OpenBao, ...) INTO Kubernetes Secrets — from the official Helm chart (`external-secrets` at `https://charts.external-secrets.io`). The operator runs three components: the **controller** (reconciles stores and external secrets), the **webhook** (validates ESO resources at admission), and the **cert-controller** (bootstraps the webhook's serving certificate). The typed spec covers the chart's meaningful configuration surface, with a `helm_values` escape hatch (merged last, Helm `-f` semantics, identical on both engines) for anything beyond it.

**Key design points:**

- **CRDs install with the release by default** (`crds.install`, default true) and **survive uninstall by default** (`crds.keep_on_uninstall`, true) — deleting ESO's CRDs cascades to every ExternalSecret and SecretStore object cluster-wide, so that destructive act requires an explicit `false`
- **Workload identity is the ambient-identity option**: `workload_identity` binds the CONTROLLER ServiceAccount to a cloud identity (EKS IRSA, GKE Workload Identity, AKS Workload Identity); stores that leave their auth block empty authenticate through it. Per-store identities — finer-grained, recommended for multi-team clusters — live in each store's auth block instead and need nothing here
- **Scaling knobs where they matter**: `concurrent` raises parallel ExternalSecret reconciliation; `replicas` + `leader_elect` add controller redundancy (enforced together — replicas without leader election race); `webhook`/`cert_controller` blocks size the other two components
- **The install waits for real readiness**: both engines wait for all three Deployments to become Available — an operator whose webhook is down rejects every SecretStore/ExternalSecret apply, so a premature "success" would just move the failure downstream

## Environment Injection (where cloud identity flows in)

Kubernetes is a dual provider: the cluster runs IN one environment while the secret stores often live in ANOTHER. The operator install only carries the AMBIENT identity option — the identity stores fall back to when their own auth block is empty:

| Host cluster | Ambient identity (`workload_identity` on this component) | Per-store identity (on the store kinds) | Static credentials (on the store kinds) |
|---|---|---|---|
| EKS | `workload_identity.eks.role_arn` (IRSA) — one role may read every synced secret | store auth references a dedicated ServiceAccount with its own IRSA role | store auth references a credential Secret |
| GKE | `workload_identity.gke.service_account_email` (Workload Identity) | dedicated ServiceAccount per store | credential Secret |
| AKS | `workload_identity.aks.client_id` (Azure AD Workload Identity; the module also stamps the required pod label) | dedicated ServiceAccount per store | credential Secret |
| Self-managed / kind / datacenter | — (no cloud federation) | — | required: static credentials or token-based backends (Vault, ...) |

The cross-cloud combinations are first-class: a GKE cluster syncing from AWS Secrets Manager simply creates a store with static AWS credentials (or an assumable role) — nothing on this component changes. The ambient arm is the simplest posture when ONE cloud identity may read everything the cluster syncs; multi-team clusters should prefer per-store identities and leave `workload_identity` unset.

The controller ServiceAccount name is fixed to `external-secrets` and exported (`status.outputs.controller_service_account`) so the cloud-side half of an ambient binding (IAM trust policy, Workload Identity binding, federated credential) can be composed in the same Infra Chart.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: installation namespace (`external-secrets` by convention) — literal or a KubernetesNamespace reference

### Common

- **`spec.create_namespace`**: create (and own) the namespace with the release
- **`spec.chart_version`**: pinned chart version (defaults to the validated pin; chart and operator versions are aligned upstream — chart 2.8.0 ships operator v2.8.0)
- **`spec.concurrent`**: parallel ExternalSecret reconciliation (chart default 1; raise for clusters with hundreds of ExternalSecrets)
- **`spec.replicas` + `spec.leader_elect`**: controller redundancy — validation enforces leader election with more than one replica
- **`spec.workload_identity`**: the ambient-identity binding for stores without their own auth
- **`spec.webhook` / `spec.cert_controller`**: per-component replicas and resources
- **`spec.prometheus.service_monitor`**: opt-in ServiceMonitor (requires the Prometheus operator CRDs — the release fails without them)
- **`spec.helm_values`**: escape hatch for chart values beyond the typed fields — never the primary interface

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Installation namespace — the home for credential Secrets that cluster-scoped stores read |
| `release_name` | Helm release name (always `external-secrets`) |
| `controller_service_account` | Controller ServiceAccount — the identity to bind cloud-side for ambient keyless store access |

## Composing in Infra Charts

The standard chart wiring: this component first, then stores (KubernetesClusterSecretStore for cluster-wide backends, KubernetesSecretStore for namespace-scoped ones), then KubernetesExternalSecret resources referencing the stores. Catalog kinds (an IAM role for Secrets Manager, a GCP service account for Secret Manager) deploy in the same run and flow their handles into `workload_identity` — or into the stores' own auth blocks for per-team isolation.

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesExternalSecretsOperator
metadata:
  name: external-secrets-operator
spec:
  namespace:
    value: external-secrets
  createNamespace: true
  workloadIdentity:
    eks:
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: external-secrets-role
          fieldPath: status.outputs.role_arn
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
