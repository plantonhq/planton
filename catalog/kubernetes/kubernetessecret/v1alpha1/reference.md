# KubernetesSecret

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesSecretSpec** defines the configuration for creating and managing a Kubernetes Secret.
This spec implements a "Secret-as-a-Service" pattern, providing type-safe configuration for all
standard Kubernetes secret types while keeping values as plain strings matching the Kubernetes API.
It is the confidential mirror of KubernetesConfigMap.

The type-safe oneof design ensures that exactly one secret type is configured per resource,
with type-specific fields and validations for each variant. Every standard secret type is
expressible: Opaque (with UTF-8 and binary entries), TLS, docker-registry, basic-auth,
ssh-auth, and service-account-token.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSecret
metadata:
  name: test-secret
spec:
  name: test-secret
  namespace:
    value: default
  labels:
    team: platform-engineering
    environment: test
    created-by: planton
  annotations:
    description: "Test secret created by Planton"
  opaque:
    data:
      username: "test-user"
      password: "test-password"
      api-key: "test-api-key-12345"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.namespace` | `string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.immutable` | `bool` |  |  |  |
| `spec.opaque` | `KubernetesSecretOpaqueData` |  |  |  |
| `spec.opaque.data` | `map<string, string>` |  |  |  |
| `spec.opaque.binaryData` | `map<string, string>` |  |  |  |
| `spec.tls` | `KubernetesSecretTlsData` |  |  |  |
| `spec.tls.tlsCrt` | `string` | yes |  |  |
| `spec.tls.tlsKey` | `string` | yes |  |  |
| `spec.dockerConfigJson` | `KubernetesSecretDockerConfigJsonData` |  |  |  |
| `spec.dockerConfigJson.registryServer` | `string` | yes |  |  |
| `spec.dockerConfigJson.username` | `string` | yes |  |  |
| `spec.dockerConfigJson.password` | `string` (sensitive) | yes |  |  |
| `spec.dockerConfigJson.email` | `string` |  |  |  |
| `spec.basicAuth` | `KubernetesSecretBasicAuthData` |  |  |  |
| `spec.basicAuth.username` | `string` | yes |  |  |
| `spec.basicAuth.password` | `string` | yes |  |  |
| `spec.sshAuth` | `KubernetesSecretSshAuthData` |  |  |  |
| `spec.sshAuth.sshPrivateKey` | `string` | yes |  |  |
| `spec.serviceAccountToken` | `KubernetesSecretServiceAccountTokenData` |  |  |  |
| `spec.serviceAccountToken.serviceAccountName` | `string \| valueFrom` | yes |  | KubernetesServiceAccount (`spec.name`) |

## Field Details

### spec.name

`string` · required

The name of the Kubernetes Secret.
Must be a valid DNS subdomain name (lowercase alphanumeric, hyphens, and dots).

- rule: Name must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots, no leading/trailing dots or hyphens)
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.namespace

`string | valueFrom`

The namespace in which the Secret is created. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource, so an infra chart can create the namespace
and the Secret in one run. When omitted, the Secret lands in the cluster's `default`
namespace — the same behavior as kubectl without a namespace flag.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Additional labels to apply to the secret.
These are merged with standard Planton labels for resource tracking and governance.

### spec.annotations

`map<string, string>`

Additional annotations to apply to the secret.

### spec.immutable

`bool`

When true, the secret data cannot be updated after creation.
Immutable secrets provide protection against accidental updates and
improve cluster performance by reducing watch load on the API server.

### spec.opaque

`KubernetesSecretOpaqueData`

Opaque secret with arbitrary key-value string data.
Maps to Kubernetes secret type "Opaque".

- rule: at least one entry is required in data or binary_data
- rule: data and binary_data must not contain the same key

### spec.opaque.data

`map<string, string>`

Key-value pairs of secret data.
Values are plain strings (Kubernetes stringData semantics); Kubernetes stores them
base64-encoded at rest. Use `binary_data` for values that are not valid UTF-8.

### spec.opaque.binaryData

`map<string, string>`

Binary secret entries with base64-encoded values — the exact wire form the Kubernetes
API uses for `data` in YAML manifests. Use this for payloads that are not valid UTF-8
(keystores, certificates in binary form, serialized blobs). Keys must not overlap with
`data` keys: both maps merge into the same underlying Secret data.

A value may also be a downstream platform's secret or variable reference token
(`$secret/...`, `$var/...`) in place of the literal: the platform resolves the token to
the stored value before the module runs, and the stored value is then the base64 the
module writes. Without this arm a private key or a CA bundle held in the platform's
secret store could never reach this field, because a reference is the only form a
secret is ever allowed to take in a manifest. The literal arm keeps refusing malformed
base64 before any apply.

- rule: {"map":{"keys":{"string":{"maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}},"values":{"string":{"pattern":"^(?:\\$(?:secret|var)/.+|(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?)$"}}}}

### spec.tls

`KubernetesSecretTlsData`

TLS secret containing a certificate and private key.
Maps to Kubernetes secret type "kubernetes.io/tls".

### spec.tls.tlsCrt

`string` · required

PEM-encoded TLS certificate (or certificate chain).
Stored as the "tls.crt" key in the Kubernetes Secret.

- rule: {"string":{"minLen":"1"}}

### spec.tls.tlsKey

`string` · required

PEM-encoded TLS private key.
Stored as the "tls.key" key in the Kubernetes Secret.

- rule: {"string":{"minLen":"1"}}

### spec.dockerConfigJson

`KubernetesSecretDockerConfigJsonData`

Docker registry credentials for image pulls.
Maps to Kubernetes secret type "kubernetes.io/dockerconfigjson".

### spec.dockerConfigJson.registryServer

`string` · required

Docker registry server URL (e.g., "https://index.docker.io/v1/", "gcr.io", "ghcr.io").

- rule: {"string":{"minLen":"1"}}

### spec.dockerConfigJson.username

`string` · required

Username for registry authentication.

- rule: {"string":{"minLen":"1"}}

### spec.dockerConfigJson.password

`string` · required · sensitive

Password or access token for registry authentication. Never plaintext in a
manifest: only a `$secret/<slug>` reference to an organization secret is
accepted, and the runner resolves it inside your infrastructure at deploy time.

- rule: {"string":{"minLen":"1"}}

### spec.dockerConfigJson.email

`string`

Optional email associated with the registry account.

### spec.basicAuth

`KubernetesSecretBasicAuthData`

Basic authentication credentials (username/password).
Maps to Kubernetes secret type "kubernetes.io/basic-auth".

### spec.basicAuth.username

`string` · required

Username for basic authentication.
Stored as the "username" key in the Kubernetes Secret.

- rule: {"string":{"minLen":"1"}}

### spec.basicAuth.password

`string` · required

Password for basic authentication.
Stored as the "password" key in the Kubernetes Secret.

- rule: {"string":{"minLen":"1"}}

### spec.sshAuth

`KubernetesSecretSshAuthData`

SSH authentication credentials (private key).
Maps to Kubernetes secret type "kubernetes.io/ssh-auth".

### spec.sshAuth.sshPrivateKey

`string` · required

PEM-encoded SSH private key.
Stored as the "ssh-privatekey" key in the Kubernetes Secret.

- rule: {"string":{"minLen":"1"}}

### spec.serviceAccountToken

`KubernetesSecretServiceAccountTokenData`

Long-lived API token bound to a ServiceAccount, populated by the cluster's
token controller. Maps to Kubernetes secret type "kubernetes.io/service-account-token".

### spec.serviceAccountToken.serviceAccountName

`string | valueFrom` · required

The name of the ServiceAccount the token authenticates as. Applied as the
"kubernetes.io/service-account.name" annotation; the ServiceAccount must exist in the
same namespace as the Secret. Accepts a literal name or a reference to a
KubernetesServiceAccount resource, so a chart can create the identity and its
long-lived token in one run.

- references: KubernetesServiceAccount (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesServiceAccount, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

## Validation Rules

- `secret_data_required`: Exactly one secret data type must be provided (opaque, tls, docker_config_json, basic_auth, ssh_auth, or service_account_token)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesSecret, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.secret_name` | `string` | The name of the created Kubernetes Secret. This is the primary identifier for referencing the secret in other resources. |
| `status.outputs.secret_namespace` | `string` | The namespace where the Kubernetes Secret was created. |
| `status.outputs.secret_type` | `string` | The Kubernetes secret type string. Possible values: "Opaque", "kubernetes.io/tls", "kubernetes.io/dockerconfigjson", "kubernetes.io/basic-auth", "kubernetes.io/ssh-auth", "kubernetes.io/service-account-token" |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.serviceAccountToken.serviceAccountName` | KubernetesServiceAccount | `spec.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesCronJob | `spec.jobTemplate.pod.imagePullSecrets` | `spec.name` |
| KubernetesDaemonSet | `spec.pod.imagePullSecrets` | `spec.name` |
| KubernetesDeployment | `spec.pod.imagePullSecrets` | `spec.name` |
| KubernetesGateway | `spec.listeners[].tls.certificateRefs[].name` | `status.outputs.secret_name` |
| KubernetesGateway | `spec.tls.backend.clientCertificateRef.name` | `status.outputs.secret_name` |
| KubernetesIngress | `spec.tls[].secretName` | `status.outputs.secret_name` |
| KubernetesJob | `spec.pod.imagePullSecrets` | `spec.name` |
| KubernetesKarapace | `spec.httpAuthentication.basic.secretName` | `status.outputs.secret_name` |
| KubernetesListenerSet | `spec.listeners[].tls.certificateRefs[].name` | `status.outputs.secret_name` |
| KubernetesMetricsServer | `spec.tls.existingSecretName` | `metadata.name` |
| KubernetesOpenSearch | `spec.security.transportTls.caSecret` | `metadata.name` |
| KubernetesOpenSearch | `spec.security.config.securityConfigSecret` | `metadata.name` |
| KubernetesOpenSearch | `spec.security.config.adminSecret` | `metadata.name` |
| KubernetesOpenSearch | `spec.security.config.adminCredentialsSecret` | `metadata.name` |
| KubernetesOpenSearch | `spec.dashboards.opensearchCredentialsSecret` | `metadata.name` |
| KubernetesOpenSearch | `spec.monitoring.monitoringUserSecret` | `metadata.name` |
| KubernetesOpenSearch | `spec.keystore[].secret` | `metadata.name` |
| KubernetesRabbitMq | `spec.tls.caSecretName` | `metadata.name` |
| KubernetesServiceAccount | `spec.imagePullSecrets` | `spec.name` |
| KubernetesSolr | `spec.security.basicAuthSecret` | `metadata.name` |
| KubernetesSolrOperator | `spec.mtls.clientCertSecret` | `metadata.name` |
| KubernetesSolrOperator | `spec.mtls.caCertSecret` | `metadata.name` |
| KubernetesStatefulSet | `spec.pod.imagePullSecrets` | `spec.name` |

## See Also

- [Overview](../README.md)
