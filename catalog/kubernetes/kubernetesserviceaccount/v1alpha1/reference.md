# KubernetesServiceAccount

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**KubernetesServiceAccountSpec** defines the configuration for creating and managing a
Kubernetes ServiceAccount — the in-cluster identity that pods run as. A ServiceAccount is
the anchor of three concerns:

  1. **API authentication** — pods authenticate to the kube-apiserver as the ServiceAccount,
     and RBAC grants (KubernetesRbac) attach permissions to it.
  2. **Registry authentication** — `image_pull_secrets` attach docker-registry credentials
     that the kubelet uses when pulling images for pods running as this identity.
  3. **Cloud identity federation** — `workload_identity` binds the ServiceAccount to a cloud
     identity (GCP service account, AWS IAM role, or Azure managed identity) so pods reach
     cloud APIs keylessly, with no long-lived credentials anywhere in the cluster.

The spec covers the complete meaningful ServiceAccount surface. The upstream `secrets` list
is deliberately not modeled: its only remaining behavior (mountable-secrets enforcement) is
deprecated since Kubernetes v1.32, and token secrets are superseded by the TokenRequest API.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesServiceAccount
metadata:
  name: test-service-account
spec:
  name: test-service-account
  namespace:
    value: default
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.namespace` | `string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.imagePullSecrets` | `[]string \| valueFrom` |  |  | KubernetesSecret (`spec.name`) |
| `spec.automountServiceAccountToken` | `bool` |  |  |  |
| `spec.workloadIdentity` | `KubernetesWorkloadIdentity` |  |  |  |
| `spec.workloadIdentity.gke` | `KubernetesWorkloadIdentityGke` |  |  |  |
| `spec.workloadIdentity.gke.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.workloadIdentity.eks` | `KubernetesWorkloadIdentityEksIrsa` |  |  |  |
| `spec.workloadIdentity.eks.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.workloadIdentity.aks` | `KubernetesWorkloadIdentityAks` |  |  |  |
| `spec.workloadIdentity.aks.clientId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.client_id`) |
| `spec.workloadIdentity.aks.tenantId` | `string` |  |  |  |

## Field Details

### spec.name

`string` · required

The name of the ServiceAccount (its `metadata.name` in the cluster).
Must be a valid DNS subdomain: lowercase alphanumeric characters, hyphens, and dots,
at most 253 characters. Workloads reference this name in `spec.serviceAccountName`,
and RBAC subjects reference it as "system:serviceaccount:<namespace>:<name>".

- rule: Name must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots, no leading/trailing dots or hyphens)
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.namespace

`string | valueFrom`

The namespace in which the ServiceAccount is created. Accepts a literal namespace name
or a reference to a KubernetesNamespace resource, so an infra chart can create the
namespace and the identity in one run. When omitted, the ServiceAccount lands in the
cluster's `default` namespace — the same behavior as kubectl without a namespace flag.
The namespace participates in the identity's fully-qualified RBAC name and in every
cloud federation subject, so charts should almost always set it explicitly.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Additional labels to apply to the ServiceAccount.
These are merged with standard Planton labels for resource tracking and governance.

### spec.annotations

`map<string, string>`

Additional annotations to apply to the ServiceAccount. Cloud workload-identity
annotations should be expressed through `workload_identity` rather than written here —
the typed field validates the identity handle and keeps the intent visible to the
resource graph. Annotations set here are merged last and must not collide with the
annotations `workload_identity` generates.

### spec.imagePullSecrets

`[]string | valueFrom`

Docker-registry credentials the kubelet presents when pulling images for pods that run
as this ServiceAccount. Each entry names a `kubernetes.io/dockerconfigjson` secret in the
SAME namespace — accepts a literal secret name or a reference to a KubernetesSecret
resource, so a chart can materialize the registry credential and attach it to the
identity in one run. Attaching pull secrets at the ServiceAccount level frees every pod
spec from repeating `imagePullSecrets`.

- references: KubernetesSecret (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.automountServiceAccountToken

`bool` · optional (explicit presence)

Whether pods running as this ServiceAccount automatically receive a projected API token
mount. Tri-state, mirroring the Kubernetes API: unset defers to the cluster/pod-level
default (mount), `false` hardens pods that never talk to the kube-apiserver (a common
security-baseline requirement), and `true` makes the mount explicit. Individual pods can
override this in their own spec.

### spec.workloadIdentity

`KubernetesWorkloadIdentity`

Binds this ServiceAccount to a cloud identity for keyless access to cloud APIs
(GKE Workload Identity, EKS IRSA, or Azure AD Workload Identity). The module translates
the selected arm into the exact annotations the cloud's webhook expects; the matching
cloud-side trust configuration (IAM binding, trust policy, or federated credential) is
owned by the referenced cloud identity resource. Omit for ServiceAccounts that never
leave the cluster.

### spec.workloadIdentity.gke

`KubernetesWorkloadIdentityGke`

GKE Workload Identity: annotate the ServiceAccount with a GCP service account email.

### spec.workloadIdentity.gke.serviceAccountEmail

`string | valueFrom` · required

GCP service account email, e.g. "dns-manager@my-project.iam.gserviceaccount.com".
Applied as the `iam.gke.io/gcp-service-account` annotation. Accepts a literal
email or a reference to a GcpServiceAccount resource's output.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.workloadIdentity.eks

`KubernetesWorkloadIdentityEksIrsa`

EKS IRSA: annotate the ServiceAccount with an AWS IAM role ARN.

### spec.workloadIdentity.eks.roleArn

`string | valueFrom` · required

AWS IAM role ARN, e.g. "arn:aws:iam::123456789012:role/dns-manager".
Applied as the `eks.amazonaws.com/role-arn` annotation. Accepts a literal ARN
or a reference to an AwsIamRole resource's output.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.workloadIdentity.aks

`KubernetesWorkloadIdentityAks`

Azure AD Workload Identity: annotate the ServiceAccount with an Entra application
(or user-assigned managed identity) client ID.

### spec.workloadIdentity.aks.clientId

`string | valueFrom` · required

Client ID (GUID) of the user-assigned managed identity or Entra application.
Applied as the `azure.workload.identity/client-id` annotation. Accepts a literal
GUID or a reference to an AzureUserAssignedIdentity resource's output.

- references: AzureUserAssignedIdentity (`status.outputs.client_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.workloadIdentity.aks.tenantId

`string` · optional (explicit presence)

Entra tenant ID (GUID). Optional: only needed for cross-tenant scenarios; when
omitted the azure-workload-identity webhook uses its default tenant. Applied as
the `azure.workload.identity/tenant-id` annotation when set.

- rule: {"string":{"uuid":true}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesServiceAccount, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.service_account_name` | `string` | The name of the created ServiceAccount — the value workloads set in `spec.serviceAccountName`. |
| `status.outputs.namespace` | `string` | The namespace in which the ServiceAccount was created. |
| `status.outputs.rbac_subject` | `string` | The fully-qualified RBAC subject string: "system:serviceaccount:<namespace>:<name>". This is the exact value cloud trust configuration matches on (an AWS IAM trust policy condition, an Azure federated credential subject) and the username the kube-apiserver sees — exported so downstream resources never re-assemble it by hand. |
| `status.outputs.workload_identity_handle` | `string` | The cloud identity handle this ServiceAccount is bound to via workload identity: the GCP service account email, AWS IAM role ARN, or Azure client ID — whichever arm was configured. Empty when workload identity is not configured. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.imagePullSecrets` | KubernetesSecret | `spec.name` |
| `spec.workloadIdentity.gke.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.workloadIdentity.eks.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.workloadIdentity.aks.clientId` | AzureUserAssignedIdentity | `status.outputs.client_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesClusterIssuer | `spec.config.acme.solvers[].dns01.route53.serviceAccount.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesClusterIssuer | `spec.config.vault.kubernetesAuth.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesClusterSecretStore | `spec.config.aws.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesClusterSecretStore | `spec.config.gcpSecretManager.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesClusterSecretStore | `spec.config.azureKeyVault.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesClusterSecretStore | `spec.config.vault.kubernetes.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesClusterSecretStore | `spec.config.kubernetes.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesCronJob | `spec.jobTemplate.pod.serviceAccount` | `status.outputs.service_account_name` |
| KubernetesDaemonSet | `spec.pod.serviceAccount` | `status.outputs.service_account_name` |
| KubernetesDeployment | `spec.pod.serviceAccount` | `status.outputs.service_account_name` |
| KubernetesIssuer | `spec.config.acme.solvers[].dns01.route53.serviceAccount.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesIssuer | `spec.config.vault.kubernetesAuth.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesJob | `spec.pod.serviceAccount` | `status.outputs.service_account_name` |
| KubernetesRbac | `spec.subjects[].serviceAccount.name` | `spec.name` |
| KubernetesSecret | `spec.serviceAccountToken.serviceAccountName` | `spec.name` |
| KubernetesSecretStore | `spec.config.aws.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesSecretStore | `spec.config.gcpSecretManager.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesSecretStore | `spec.config.azureKeyVault.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesSecretStore | `spec.config.vault.kubernetes.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesSecretStore | `spec.config.kubernetes.serviceAccountName` | `status.outputs.service_account_name` |
| KubernetesStatefulSet | `spec.pod.serviceAccount` | `status.outputs.service_account_name` |

## See Also

- [Overview](../README.md)
