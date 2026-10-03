# Kubernetes ServiceAccount

## Overview

**KubernetesServiceAccount** is a catalog kind that creates and manages Kubernetes ServiceAccounts — the in-cluster identity that pods run as. Every pod runs as exactly one ServiceAccount (the namespace's `default` one unless told otherwise), and that identity is the anchor for three distinct concerns:

1. **API authentication (RBAC anchor)** — pods authenticate to the kube-apiserver as the ServiceAccount, and RBAC grants (KubernetesRbac) attach permissions to it. Its fully-qualified RBAC name is `system:serviceaccount:<namespace>:<name>`.
2. **Registry authentication** — `imagePullSecrets` attach docker-registry credentials that the kubelet presents when pulling images for pods running as this identity, freeing every pod spec from repeating `imagePullSecrets`.
3. **Cloud identity federation (workload identity)** — `workloadIdentity` binds the ServiceAccount to a cloud identity (GCP service account, AWS IAM role, or Azure managed identity) so pods reach cloud APIs keylessly, with no long-lived credentials anywhere in the cluster.

The spec covers the complete meaningful ServiceAccount surface. The upstream `secrets` list is deliberately not modeled: its only remaining behavior (mountable-secrets enforcement) is deprecated since Kubernetes v1.32, and token secrets are superseded by the TokenRequest API.

## Purpose

A dedicated ServiceAccount per workload is the foundation of least-privilege Kubernetes: RBAC grants, image pull credentials, and cloud trust all attach to the identity, not to individual pods. Yet the raw object is deceptively thin — the real configuration lives in annotations with cloud-specific magic strings that are easy to get subtly wrong.

**Key value over raw manifests:**

- **Typed workload identity**: instead of hand-writing `iam.gke.io/gcp-service-account` or `eks.amazonaws.com/role-arn` annotations, you set a typed field per cloud; the module emits the exact annotation the cloud's webhook expects
- **References, not strings**: the cloud identity handle, image pull secrets, and namespace all accept references to other Planton resources, so a chart can create the namespace, the registry credential, the cloud identity, and the ServiceAccount in one run
- **RBAC subject exported**: the `system:serviceaccount:<namespace>:<name>` string is an output, so cloud trust configuration never re-assembles it by hand
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity

## Cross-Cloud Workload Identity

Workload identity is the mechanism that makes pod-to-cloud access keyless: the cluster's OIDC issuer vouches for the ServiceAccount, and the cloud exchanges that token for cloud credentials. A ServiceAccount can be bound to at most one cloud identity — the `workloadIdentity` field is a oneof with three arms:

| Arm | Field(s) | Annotation(s) produced | Mechanism |
|-----|----------|------------------------|-----------|
| `gke` | `serviceAccountEmail` | `iam.gke.io/gcp-service-account: <email>` | GKE Workload Identity |
| `eks` | `roleArn` | `eks.amazonaws.com/role-arn: <role-arn>` | IRSA (IAM Roles for Service Accounts) |
| `aks` | `clientId` (+ optional `tenantId`) | `azure.workload.identity/client-id: <client-id>` (+ `azure.workload.identity/tenant-id: <tenant-id>` when set) | Azure AD Workload Identity |

The annotation is only half the binding. The cloud-side trust configuration is owned by the referenced cloud identity resource:

- **GKE**: the GCP service account must carry a `roles/iam.workloadIdentityUser` binding for member `serviceAccount:<project>.svc.id.goog[<namespace>/<ksa-name>]`
- **EKS**: the IAM role's trust policy must trust the cluster's OIDC provider with a condition on `system:serviceaccount:<namespace>:<ksa-name>`
- **AKS**: the managed identity (or Entra app) must carry a federated identity credential for subject `system:serviceaccount:<namespace>:<ksa-name>` against the cluster's OIDC issuer

> **AKS note:** pods that use the identity must ALSO carry the `azure.workload.identity/use: "true"` **pod label** — the azure-workload-identity webhook only mutates labeled pods. That half lives on the workload spec, not on the ServiceAccount.

Because the namespace and name participate in every federation subject, renaming or moving a ServiceAccount breaks its cloud trust — treat the pair as part of the identity's contract.

## Essential Configuration Fields

### Required

- **`spec.name`**: The ServiceAccount name (DNS subdomain: lowercase alphanumeric, hyphens, dots, 1–253 chars). Workloads reference it in `spec.serviceAccountName`; RBAC subjects reference it as `system:serviceaccount:<namespace>:<name>`.

### Common

- **`spec.namespace`**: Literal namespace name or reference to a KubernetesNamespace resource. When omitted, the ServiceAccount lands in `default`. Because the namespace participates in the RBAC name and every cloud federation subject, set it explicitly in almost all cases.
- **`spec.imagePullSecrets`**: List of `kubernetes.io/dockerconfigjson` secret names in the SAME namespace — literal names or references to KubernetesSecret resources.
- **`spec.automountServiceAccountToken`**: Tri-state. Unset defers to the cluster/pod default (mount); `false` hardens pods that never talk to the kube-apiserver (a common security-baseline requirement); `true` makes the mount explicit. Individual pods can override.
- **`spec.workloadIdentity`**: One of `gke`, `eks`, or `aks` (see above). Omit for ServiceAccounts that never leave the cluster.
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton labels. Cloud workload-identity annotations should be expressed through `workloadIdentity`, not written here.

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`service_account_name`**: The created ServiceAccount's name — the value workloads set in `spec.serviceAccountName`
- **`namespace`**: The namespace it was created in
- **`rbac_subject`**: The fully-qualified `system:serviceaccount:<namespace>:<name>` string — the exact value cloud trust configuration matches on and the username the kube-apiserver sees
- **`workload_identity_handle`**: The bound cloud identity (GCP email, IAM role ARN, or Azure client ID); empty when workload identity is not configured

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the namespace, image pull secret names, and workload-identity handle (literal values or resolved references)
2. Translate the selected `workloadIdentity` arm into the exact annotation set that cloud's webhook expects, merged with user annotations
3. Create the ServiceAccount with pull secrets and the automount setting
4. Export the name, namespace, RBAC subject string, and cloud identity handle

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesServiceAccount** when you need:

- A dedicated identity per workload (least-privilege baseline) that RBAC grants attach to
- Keyless pod access to cloud APIs on GKE, EKS, or AKS
- Registry pull credentials attached once at the identity level instead of on every pod spec
- Token automount hardening for pods that never call the kube-apiserver

**Do NOT use** when:

- The workload component you deploy already creates and manages its own ServiceAccount — check before doubling up
- You need the deprecated mountable-secrets enforcement (`secrets` list) — deliberately unsupported

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist (or be created in the same chart via a reference)
- **For workload identity**: the cluster feature enabled (GKE Workload Identity / EKS OIDC provider / AKS workload-identity add-on) and the cloud-side trust configured on the referenced identity
- **For image pull secrets**: the `kubernetes.io/dockerconfigjson` secret(s) existing in the same namespace (or created in the same chart)

## Best Practices

1. **One ServiceAccount per workload**: Never let production workloads run as the namespace `default` ServiceAccount — grants to it leak to everything in the namespace
2. **Set the namespace explicitly**: It participates in the RBAC name and every federation subject; implicit `default` placement is almost never what you want
3. **Disable token automount unless needed**: Set `automountServiceAccountToken: false` for pods that never call the kube-apiserver; re-enable per pod where required
4. **Express cloud bindings through `workloadIdentity`**: Hand-written annotations bypass validation and hide intent from the resource graph
5. **Treat name+namespace as a contract**: Cloud trust (IAM bindings, trust policies, federated credentials) matches on them exactly; renames break federation silently

## References

- [Kubernetes ServiceAccounts Documentation](https://kubernetes.io/docs/concepts/security/service-accounts/)
- [Configure Service Accounts for Pods](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
- [GKE Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
- [EKS IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [Azure AD Workload Identity](https://azure.github.io/azure-workload-identity/docs/)
- [Add ImagePullSecrets to a Service Account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
