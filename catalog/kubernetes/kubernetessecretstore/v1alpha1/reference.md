# KubernetesSecretStore

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesSecretStoreSpec** creates a NAMESPACED External Secrets
Operator store — one backend connection (AWS Secrets Manager, GCP Secret
Manager, Azure Key Vault, Vault/OpenBao, ...) that only ExternalSecret
resources in the SAME namespace may sync from. The SecretStore is named
after `metadata.name`; ExternalSecrets in the namespace reference it by
that name (kind SecretStore, the upstream default).

Use the namespaced grain for a connection that belongs to ONE
namespace/team — its credential Secrets live in that namespace and its
blast radius ends at the namespace boundary. Use
KubernetesClusterSecretStore for platform-wide backends every team
shares. Requires the External Secrets Operator on the cluster
(KubernetesExternalSecretsOperator).

## Example

```yaml
# Full-surface offline-proof manifest: exercises the Vault/OpenBao backend
# with AppRole auth (the declared-credential materialization path), KV
# engine addressing, namespace + CA trust, and the common store tuning — so
# the offline tofu plan and pulumi preview proofs cover the credential arm
# the live fake-backend lane excludes. Placeholder values; never applied to
# a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSecretStore
metadata:
  name: hack-secret-store
spec:
  namespace:
    value: team-a
  config:
    vault:
      server: https://openbao.example.internal:8200
      path: secret
      version: v2
      namespace: platform
      caBundle: hack-placeholder-pem
      appRole:
        path: approle
        roleId: hack-role-id
        secretId: hack-placeholder-secret-id
    refreshInterval: 5m
    retry:
      maxRetries: 3
      retryInterval: 10s
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.config` | `ExternalSecretsStoreConfig` | yes |  |  |
| `spec.config.aws` | `ExternalSecretsStoreAws` |  |  |  |
| `spec.config.aws.service` | `string` |  | `SecretsManager` |  |
| `spec.config.aws.region` | `string` | yes |  |  |
| `spec.config.aws.role` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.config.aws.prefix` | `string` |  |  |  |
| `spec.config.aws.serviceAccountName` | `string \| valueFrom` |  |  | KubernetesServiceAccount (`status.outputs.service_account_name`) |
| `spec.config.aws.serviceAccountNamespace` | `string` |  |  |  |
| `spec.config.aws.accessKeyId` | `string` |  |  |  |
| `spec.config.aws.secretAccessKey` | `string` (sensitive) |  |  |  |
| `spec.config.gcpSecretManager` | `ExternalSecretsStoreGcp` |  |  |  |
| `spec.config.gcpSecretManager.projectId` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.config.gcpSecretManager.location` | `string` |  |  |  |
| `spec.config.gcpSecretManager.serviceAccountName` | `string \| valueFrom` |  |  | KubernetesServiceAccount (`status.outputs.service_account_name`) |
| `spec.config.gcpSecretManager.serviceAccountNamespace` | `string` |  |  |  |
| `spec.config.gcpSecretManager.serviceAccountKeyJson` | `string` (sensitive) |  |  |  |
| `spec.config.azureKeyVault` | `ExternalSecretsStoreAzure` |  |  |  |
| `spec.config.azureKeyVault.vaultUrl` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.vault_uri`) |
| `spec.config.azureKeyVault.tenantId` | `string` |  |  |  |
| `spec.config.azureKeyVault.authType` | `string` |  | `WorkloadIdentity` |  |
| `spec.config.azureKeyVault.identityId` | `string` |  |  |  |
| `spec.config.azureKeyVault.serviceAccountName` | `string \| valueFrom` |  |  | KubernetesServiceAccount (`status.outputs.service_account_name`) |
| `spec.config.azureKeyVault.serviceAccountNamespace` | `string` |  |  |  |
| `spec.config.azureKeyVault.clientId` | `string` |  |  |  |
| `spec.config.azureKeyVault.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.config.vault` | `ExternalSecretsStoreVault` |  |  |  |
| `spec.config.vault.server` | `string` | yes |  |  |
| `spec.config.vault.path` | `string` |  |  |  |
| `spec.config.vault.version` | `string` |  | `v2` |  |
| `spec.config.vault.namespace` | `string` |  |  |  |
| `spec.config.vault.caBundle` | `string` |  |  |  |
| `spec.config.vault.token` | `ExternalSecretsStoreVaultTokenAuth` |  |  |  |
| `spec.config.vault.token.token` | `string` (sensitive) | yes |  |  |
| `spec.config.vault.appRole` | `ExternalSecretsStoreVaultAppRoleAuth` |  |  |  |
| `spec.config.vault.appRole.path` | `string` |  |  |  |
| `spec.config.vault.appRole.roleId` | `string` | yes |  |  |
| `spec.config.vault.appRole.secretId` | `string` (sensitive) | yes |  |  |
| `spec.config.vault.kubernetes` | `ExternalSecretsStoreVaultKubernetesAuth` |  |  |  |
| `spec.config.vault.kubernetes.mountPath` | `string` |  |  |  |
| `spec.config.vault.kubernetes.role` | `string` | yes |  |  |
| `spec.config.vault.kubernetes.serviceAccountName` | `string \| valueFrom` |  |  | KubernetesServiceAccount (`status.outputs.service_account_name`) |
| `spec.config.kubernetes` | `ExternalSecretsStoreKubernetes` |  |  |  |
| `spec.config.kubernetes.serverUrl` | `string` |  |  |  |
| `spec.config.kubernetes.caBundle` | `string` |  |  |  |
| `spec.config.kubernetes.remoteNamespace` | `string` |  |  |  |
| `spec.config.kubernetes.token` | `string` (sensitive) |  |  |  |
| `spec.config.kubernetes.serviceAccountName` | `string \| valueFrom` |  |  | KubernetesServiceAccount (`status.outputs.service_account_name`) |
| `spec.config.fake` | `ExternalSecretsStoreFake` |  |  |  |
| `spec.config.fake.data` | `[]ExternalSecretsStoreFakeEntry` | yes |  |  |
| `spec.config.fake.data[].key` | `string` | yes |  |  |
| `spec.config.fake.data[].value` | `string` | yes |  |  |
| `spec.config.fake.data[].version` | `string` |  |  |  |
| `spec.config.controllerClass` | `string` |  |  |  |
| `spec.config.refreshInterval` | `string` |  |  |  |
| `spec.config.retry` | `ExternalSecretsStoreRetry` |  |  |  |
| `spec.config.retry.maxRetries` | `int32` |  |  |  |
| `spec.config.retry.retryInterval` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace the SecretStore lives in — also where credential Secrets this
store declares (static keys, tokens) are materialized, and the only
namespace whose ExternalSecrets can use it. Accepts a literal namespace
name or a reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.config

`ExternalSecretsStoreConfig` · required

The backend connection and its configuration — which external system
holds the secrets and how ESO authenticates to it. Identical surface to
KubernetesClusterSecretStore: the grain (who may sync from it) is the
only difference between the two kinds.

- rule: {"required":true}
- rule: Select the secret backend this store connects to — set exactly one of aws, gcp_secret_manager, azure_key_vault, vault, kubernetes, or fake

### spec.config.aws

`ExternalSecretsStoreAws`

AWS Secrets Manager / SSM Parameter Store.

- rule: access_key_id and secret_access_key form one credential — set both or neither
- rule: Choose ONE authentication mode: a ServiceAccount reference (keyless) or static access keys — not both

### spec.config.aws.service

`string` · optional (explicit presence)

Which AWS secrets service this store reads.
"SecretsManager": AWS Secrets Manager (the default posture).
"ParameterStore": SSM Parameter Store (cheaper, no automatic rotation).
"CertificateManager": ACM certificate export (niche; reads certificates).

- default: `SecretsManager`
- rule: {"string":{"in":["SecretsManager","ParameterStore","CertificateManager"]}}

### spec.config.aws.region

`string` · required

AWS region the secrets live in (e.g. "us-east-1").

- rule: {"required":true}

### spec.config.aws.role

`string | valueFrom`

IAM role to assume for the reads (full ARN) — the cross-account pattern,
and how a single identity fans out to per-store roles. Accepts a literal
ARN or a reference to an AwsIamRole resource's output.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.config.aws.prefix

`string`

Prefix prepended to every key this store fetches (e.g. "prod/") —
scopes the store to a subtree of the secrets namespace.

### spec.config.aws.serviceAccountName

`string | valueFrom`

Keyless auth: ServiceAccount whose IRSA binding authorizes the reads.
Accepts a literal name or a reference to a KubernetesServiceAccount
resource's output. Leave everything empty for the operator's ambient
identity.

- references: KubernetesServiceAccount (`status.outputs.service_account_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesServiceAccount, name: <that resource's name>, fieldPath: status.outputs.service_account_name}} -- a bare string does not parse

### spec.config.aws.serviceAccountNamespace

`string`

Namespace of service_account_name. Required on a ClusterSecretStore
(cluster scope has no home namespace to default to); on a SecretStore
it defaults to the store's own namespace.

### spec.config.aws.accessKeyId

`string`

Static AWS access key ID. Prefer keyless — set static keys only when no
cloud identity federation exists. Both halves must be set together; the
modules materialize them as a Kubernetes Secret the store references.

### spec.config.aws.secretAccessKey

`string` · sensitive

Static AWS secret access key (the secret half of the key pair).

### spec.config.gcpSecretManager

`ExternalSecretsStoreGcp`

GCP Secret Manager.

- rule: Choose ONE authentication mode: a ServiceAccount reference (keyless) or a service-account key — not both

### spec.config.gcpSecretManager.projectId

`string | valueFrom` · required

GCP project the secrets live in. Accepts a literal project ID or a
reference to a GcpProject resource's output.

containment_exempt: names the project the store READS secrets from —
the store itself lives in the cluster, not the project.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.config.gcpSecretManager.location

`string`

Regional Secret Manager endpoint (e.g. "us-central1") for regional
secrets. Empty = the global endpoint.

### spec.config.gcpSecretManager.serviceAccountName

`string | valueFrom`

Keyless auth: ServiceAccount whose GKE Workload Identity binding
authorizes the reads. Accepts a literal name or a reference to a
KubernetesServiceAccount resource's output. Leave everything empty for
the operator's ambient identity.

- references: KubernetesServiceAccount (`status.outputs.service_account_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesServiceAccount, name: <that resource's name>, fieldPath: status.outputs.service_account_name}} -- a bare string does not parse

### spec.config.gcpSecretManager.serviceAccountNamespace

`string`

Namespace of service_account_name. Required on a ClusterSecretStore;
defaults to the store's own namespace on a SecretStore.

### spec.config.gcpSecretManager.serviceAccountKeyJson

`string` · sensitive

Static GCP service-account key (JSON). Prefer keyless — the modules
materialize it as a Kubernetes Secret the store references.

### spec.config.azureKeyVault

`ExternalSecretsStoreAzure`

Azure Key Vault.

- rule: client_id and client_secret form one service-principal credential — set both or neither
- rule: ServicePrincipal auth needs tenant_id, client_id, and client_secret

### spec.config.azureKeyVault.vaultUrl

`string | valueFrom` · required

Key Vault data-plane URL, e.g. "https://my-vault.vault.azure.net".
Reference an AzureKeyVault's vault_uri output -- in a composed
environment the reference is also the deploy-ordering edge, so the
vault exists before the store starts validating against it -- or pass
the literal URL of an existing vault. The store READS from the vault
(access, not containment), which is why the reference does not nest
the store inside it.

- references: AzureKeyVault (`status.outputs.vault_uri`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.vault_uri}} -- a bare string does not parse

### spec.config.azureKeyVault.tenantId

`string`

Entra (Azure AD) tenant the identity lives in. Required for
service-principal and workload-identity auth.

### spec.config.azureKeyVault.authType

`string` · optional (explicit presence)

How ESO authenticates to the vault.
"WorkloadIdentity": federate via the referenced ServiceAccount (or the
  operator's ambient identity) — the keyless AKS posture.
"ManagedIdentity": the cluster's (or identity_id's) managed identity —
  AKS without Workload Identity.
"ServicePrincipal": client_id + client_secret below — the fallback for
  clusters outside Azure.

- default: `WorkloadIdentity`
- rule: {"string":{"in":["ServicePrincipal","ManagedIdentity","WorkloadIdentity"]}}

### spec.config.azureKeyVault.identityId

`string`

Client ID of a specific user-assigned managed identity (auth_type
ManagedIdentity) or of the Entra app to federate with (auth_type
WorkloadIdentity, when not carried by the ServiceAccount's annotation).

### spec.config.azureKeyVault.serviceAccountName

`string | valueFrom`

Keyless auth: ServiceAccount whose AKS Workload Identity binding
authorizes the reads (auth_type WorkloadIdentity). Accepts a literal
name or a reference to a KubernetesServiceAccount resource's output.

- references: KubernetesServiceAccount (`status.outputs.service_account_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesServiceAccount, name: <that resource's name>, fieldPath: status.outputs.service_account_name}} -- a bare string does not parse

### spec.config.azureKeyVault.serviceAccountNamespace

`string`

Namespace of service_account_name. Required on a ClusterSecretStore;
defaults to the store's own namespace on a SecretStore.

### spec.config.azureKeyVault.clientId

`string`

Service-principal application (client) ID (auth_type ServicePrincipal).
Both halves must be set together; the modules materialize the secret as
a Kubernetes Secret the store references.

### spec.config.azureKeyVault.clientSecret

`string` · sensitive

Service-principal client secret (the secret half).

### spec.config.vault

`ExternalSecretsStoreVault`

HashiCorp Vault or OpenBao KV engine (OpenBao speaks the Vault API —
point `server` at the OpenBao endpoint).

- rule: Select how ESO authenticates to Vault/OpenBao — set exactly one of token, app_role, or kubernetes

### spec.config.vault.server

`string` · required

Vault/OpenBao server URL, e.g. "https://vault.example.com:8200".

- rule: {"required":true,"string":{"uri":true}}

### spec.config.vault.path

`string`

Mount path of the KV secrets engine (e.g. "secret"). Empty = upstream
default ("secret").

### spec.config.vault.version

`string` · optional (explicit presence)

KV engine version. "v2" (versioned secrets, the modern default) or "v1".

- default: `v2`
- rule: {"string":{"in":["v1","v2"]}}

### spec.config.vault.namespace

`string`

Vault Enterprise / OpenBao namespace. Empty = the root namespace.

### spec.config.vault.caBundle

`string`

PEM CA bundle to trust the server's TLS certificate (private CAs).

### spec.config.vault.token

`ExternalSecretsStoreVaultTokenAuth`

Static Vault token. The modules materialize it as a Kubernetes Secret
the store references. Tokens expire — prefer kubernetes auth for
anything long-lived.

### spec.config.vault.token.token

`string` · required · sensitive

The Vault token.

- rule: {"required":true}

### spec.config.vault.appRole

`ExternalSecretsStoreVaultAppRoleAuth`

AppRole: machine identity via role_id + secret_id.

### spec.config.vault.appRole.path

`string`

Mount path of the AppRole auth method. Empty = upstream default
("approle").

### spec.config.vault.appRole.roleId

`string` · required

The AppRole role ID (public half).

- rule: {"required":true}

### spec.config.vault.appRole.secretId

`string` · required · sensitive

The AppRole secret ID (secret half). The modules materialize it as a
Kubernetes Secret the store references.

- rule: {"required":true}

### spec.config.vault.kubernetes

`ExternalSecretsStoreVaultKubernetesAuth`

Kubernetes auth: the cluster's ServiceAccount token is exchanged for
a Vault token — keyless from the cluster's perspective, the
production posture for in-cluster Vault/OpenBao.

### spec.config.vault.kubernetes.mountPath

`string`

Mount path of the Kubernetes auth method. Empty = upstream default
("kubernetes").

### spec.config.vault.kubernetes.role

`string` · required

Vault role to authenticate as (bound to ServiceAccounts server-side).

- rule: {"required":true}

### spec.config.vault.kubernetes.serviceAccountName

`string | valueFrom`

ServiceAccount whose token is presented to Vault. Accepts a literal
name or a reference to a KubernetesServiceAccount resource's output.
Empty = the operator's own ServiceAccount.

- references: KubernetesServiceAccount (`status.outputs.service_account_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesServiceAccount, name: <that resource's name>, fieldPath: status.outputs.service_account_name}} -- a bare string does not parse

### spec.config.kubernetes

`ExternalSecretsStoreKubernetes`

Another Kubernetes cluster's Secrets as the backend — the
cluster-to-cluster sync arm.

- rule: Choose ONE authentication mode: a bearer token or a ServiceAccount reference — not both

### spec.config.kubernetes.serverUrl

`string`

Remote API server URL. Empty = upstream default ("kubernetes.default",
i.e. the local cluster — useful for cross-namespace sync through a
ClusterSecretStore).

### spec.config.kubernetes.caBundle

`string`

PEM CA bundle to trust the remote API server's certificate.

### spec.config.kubernetes.remoteNamespace

`string`

Remote namespace to read Secrets from. Empty = upstream default
("default").

### spec.config.kubernetes.token

`string` · sensitive

Bearer token authenticating to the remote cluster (a remote
ServiceAccount's token). The modules materialize it as a Kubernetes
Secret the store references. Empty = authenticate with the local
ServiceAccount named below.

### spec.config.kubernetes.serviceAccountName

`string | valueFrom`

Local ServiceAccount whose token authenticates to the remote cluster
(the remote cluster must trust this cluster's OIDC issuer). Accepts a
literal name or a reference to a KubernetesServiceAccount resource's
output.

- references: KubernetesServiceAccount (`status.outputs.service_account_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesServiceAccount, name: <that resource's name>, fieldPath: status.outputs.service_account_name}} -- a bare string does not parse

### spec.config.fake

`ExternalSecretsStoreFake`

ESO's built-in fake backend: serves the literal key/value entries
declared below. For pipelines, tests, and evaluating the sync
machinery without any external account; never for real secrets.

### spec.config.fake.data

`[]ExternalSecretsStoreFakeEntry` · required

The key/value entries this store serves.

- rule: {"repeated":{"minItems":"1"}}

### spec.config.fake.data[].key

`string` · required

Remote key an ExternalSecret's remoteRef.key addresses.

- rule: {"required":true}

### spec.config.fake.data[].value

`string` · required

Value served for the key.

- rule: {"required":true}

### spec.config.fake.data[].version

`string`

Optional version gate: when set, only a remoteRef with the SAME version
string finds this entry (empty matches empty).

### spec.config.controllerClass

`string`

Only the operator installation whose controller_class matches reconciles
this store — the sharding knob. Empty = the default operator.

### spec.config.refreshInterval

`string`

How often ESO re-validates the store connection (Go duration, e.g.
"5m"). Empty = upstream default.

- rule: {"string":{"pattern":"^$|^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.config.retry

`ExternalSecretsStoreRetry`

Retry posture for failed backend reads.

### spec.config.retry.maxRetries

`int32` · optional (explicit presence)

Maximum retry attempts for a failed backend call.

- rule: {"int32":{"gte":0}}

### spec.config.retry.retryInterval

`string`

Interval between retries (Go duration, e.g. "10s").

- rule: {"string":{"pattern":"^$|^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesSecretStore, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.store_name` | `string` | Name of the created SecretStore (equals metadata.name). Use it in an ExternalSecret's secretStoreRef.name (kind SecretStore) in the same namespace. |
| `status.outputs.namespace` | `string` | Namespace the SecretStore (and its credential Secrets) live in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.config.aws.role` | AwsIamRole | `status.outputs.role_arn` |
| `spec.config.aws.serviceAccountName` | KubernetesServiceAccount | `status.outputs.service_account_name` |
| `spec.config.gcpSecretManager.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.config.gcpSecretManager.serviceAccountName` | KubernetesServiceAccount | `status.outputs.service_account_name` |
| `spec.config.azureKeyVault.vaultUrl` | AzureKeyVault | `status.outputs.vault_uri` |
| `spec.config.azureKeyVault.serviceAccountName` | KubernetesServiceAccount | `status.outputs.service_account_name` |
| `spec.config.vault.kubernetes.serviceAccountName` | KubernetesServiceAccount | `status.outputs.service_account_name` |
| `spec.config.kubernetes.serviceAccountName` | KubernetesServiceAccount | `status.outputs.service_account_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesExternalSecret | `spec.storeRef.secretStore.name` | `status.outputs.store_name` |

## See Also

- [Overview](../README.md)
