# Kubernetes ServiceAccount

Deploys a Kubernetes ServiceAccount -- the identity pods run as -- with inherited image pull secrets, a tri-state automatic API token mount, and optional keyless federation to a GCP service account (GKE Workload Identity), AWS IAM role (EKS IRSA), or Azure managed identity (AKS Workload Identity). One identity per app is the posture that keeps least privilege auditable: workloads reference it, RBAC grants bind to it, and cloud federation attaches to it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes ServiceAccount** -- a single ServiceAccount in the specified namespace, with `imagePullSecrets` attached and `automountServiceAccountToken` applied only when a position was taken
- **Cloud identity annotation** -- written automatically when a workload-identity arm is configured: `iam.gke.io/gcp-service-account` (GKE), `eks.amazonaws.com/role-arn` (EKS), or `azure.workload.identity/client-id` (+ optional tenant) for AKS
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- The target namespace must already exist (the module does not create it). Use the Kubernetes Namespace kind to manage namespaces declaratively.
- For workload identity, the CLOUD side must trust this exact namespace/ServiceAccount pair: a `workloadIdentityUser` IAM binding (GKE), an OIDC trust policy on the role (EKS), or a federated credential on the managed identity (AKS).

## Deploy

### Console

Open the deployment store, find **Kubernetes ServiceAccount**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic ServiceAccount** preset for a plain per-app identity, or **GKE Workload Identity** / **EKS IRSA (IAM Roles for Service Accounts)** for keyless cloud access, in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesServiceAccount
metadata:
  name: checkout-identity
  org: acme-corp
  env: prod
spec:
  name: checkout-identity
  namespace:
    value: backend-services
  automountServiceAccountToken: false
```

```shell
planton apply -f service-account.yaml
```

This creates a hardened per-app identity in the `backend-services` namespace: pods running as it get no automatic API server token. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the identity to its namespace and cloud principal:

```yaml
spec:
  name: checkout-identity
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: backend-namespace
      fieldPath: spec.name
  workloadIdentity:
    eks:
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: checkout-role
          fieldPath: status.outputs.role_arn
```

The InfraPipeline deploys the namespace and the IAM role first, then creates the federated identity against them.

## Key Configuration

These are the most important decisions when configuring a Kubernetes ServiceAccount. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One identity per app** -- Workloads reference it via `serviceAccountName`, RBAC grants bind permissions to it, and cloud federation attaches to it. Separate identities per app keep least privilege auditable.

**Image pull secrets are inherited** -- Secret NAMES listed in `imagePullSecrets` (never the credential material) are attached to every pod running as this identity -- private-registry access configured once instead of per workload.

**API token mount is a tri-state** -- Unset means the cluster default (the token mounts); `false` hardens every pod running as this identity (most app workloads never call the Kubernetes API); `true` forces the mount. The unset state is a real position -- the module writes the field only when one was taken.

**Workload identity is keyless** -- Exactly one arm (GKE / EKS / AKS) federates this Kubernetes identity to a cloud identity: pods call cloud APIs with its permissions and zero stored keys. The AKS arm optionally carries a tenant ID for cross-tenant identities.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesSecret** | `imagePullSecrets[]` | `spec.name` |
| **GcpServiceAccount** | `workloadIdentity.gke.serviceAccountEmail` | `status.outputs.email` |
| **AwsIamRole** | `workloadIdentity.eks.roleArn` | `status.outputs.role_arn` |
| **AzureUserAssignedIdentity** | `workloadIdentity.aks.clientId` | `status.outputs.client_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_account_name` | The name of the created ServiceAccount | Workload `serviceAccountName` references |
| `namespace` | The namespace the identity lives in | Verifying workload co-location |
| `rbac_subject` | The ready-to-use subject string (`system:serviceaccount:<namespace>:<name>`) | RBAC bindings inside and outside the platform |
| `workload_identity_handle` | The federated cloud identity (GSA email, role ARN, or client ID); empty when not configured | Cloud-side trust verification and audit |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Per-app identity** -- A plain ServiceAccount named for the app, referenced by its workload and its RBAC grant. Start from the **Basic ServiceAccount** preset.

**Keyless cloud access** -- The GKE / EKS-IRSA / AKS arm federates the identity to a cloud principal; pods call cloud APIs with zero stored keys. Start from the **GKE Workload Identity** or **EKS IRSA (IAM Roles for Service Accounts)** preset.

**Private registry access** -- Registry credentials attached once via `imagePullSecrets`; every pod running as this identity pulls without per-pod configuration. Start from the **Image Pull Secrets with Automount Hardening** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- reference the namespace so Infra Charts create it and this identity in dependency order.
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) -- docker-registry Secrets attach via `imagePullSecrets`; the Secret kind's `serviceAccountToken` variant mints long-lived tokens FOR this identity.
- [**Kubernetes RBAC**](/infra-catalog/kubernetes-rbac) -- grants bind Kubernetes permissions to this identity as a ServiceAccount subject.
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) and the other workload kinds -- run as this identity via `serviceAccountName`.
