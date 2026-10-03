# KubernetesConfigMap

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**KubernetesConfigMapSpec** defines the configuration for creating and managing a Kubernetes
ConfigMap. A ConfigMap holds non-confidential configuration data — file-like text values,
property settings, or binary payloads — that pods consume as environment variables, command-line
arguments, or mounted files. For confidential data use KubernetesSecret instead; the two kinds
are deliberate mirrors of each other.

The spec covers the complete Kubernetes ConfigMap surface: UTF-8 `data`, base64 `binary_data`,
and `immutable`. There is nothing an upstream ConfigMap can express that this spec cannot.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesConfigMap
metadata:
  name: test-configmap
spec:
  name: test-configmap
  namespace:
    value: default
  labels:
    created-by: planton
  data:
    LOG_LEVEL: info
    GREETING: hello-from-planton
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.namespace` | `string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.data` | `map<string, string>` |  |  |  |
| `spec.binaryData` | `map<string, string>` |  |  |  |
| `spec.immutable` | `bool` |  |  |  |

## Field Details

### spec.name

`string` · required

The name of the ConfigMap (its `metadata.name` in the cluster).
Must be a valid DNS subdomain: lowercase alphanumeric characters, hyphens, and dots,
at most 253 characters. Pods reference this name in `configMapRef` / `configMapKeyRef`
env sources and in `configMap` volume definitions.

- rule: Name must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots, no leading/trailing dots or hyphens)
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.namespace

`string | valueFrom`

The namespace in which the ConfigMap is created. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource, so an infra chart can create the namespace
and the ConfigMap in one run. When omitted, the ConfigMap lands in the cluster's
`default` namespace — the same behavior as kubectl without a namespace flag.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Additional labels to apply to the ConfigMap.
These are merged with standard Planton labels for resource tracking and governance.

### spec.annotations

`map<string, string>`

Additional annotations to apply to the ConfigMap.

### spec.data

`map<string, string>`

UTF-8 configuration entries. Each key becomes a file name (when mounted as a volume) or
an environment variable name (when consumed via `envFrom`/`configMapKeyRef`), so keys are
restricted to alphanumeric characters, '-', '_', and '.'. Values with non-UTF-8 byte
sequences belong in `binary_data` instead. A ConfigMap with neither data nor binary_data
is valid — Kubernetes allows empty ConfigMaps (commonly used as name reservations or
coordination markers).

Note: the Kubernetes API server caps the combined size of data and binary_data at 1MiB;
oversized ConfigMaps are rejected at apply time.

- rule: {"map":{"keys":{"string":{"maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}}}

### spec.binaryData

`map<string, string>`

Binary configuration entries with base64-encoded values — the exact wire form the
Kubernetes API uses for `binaryData` in YAML manifests. Use this for payloads that are
not valid UTF-8 (compiled files, images, serialized blobs). Keys share the same character
rules as `data` keys and must not overlap with them (mirroring the API server's own
validation). Binary keys are only consumable as mounted files, not as environment variables.

- rule: {"map":{"keys":{"string":{"maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}},"values":{"string":{"pattern":"^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$"}}}}

### spec.immutable

`bool`

When true, the ConfigMap's data cannot be updated after creation (only metadata can
change); updating requires deleting and recreating the ConfigMap. Immutable ConfigMaps
protect against accidental config drift and materially reduce kube-apiserver watch load
in clusters with many ConfigMap consumers. Recommended for versioned, roll-forward
configuration (e.g. "app-config-v42").

## Validation Rules

- `data_keys_no_overlap`: data and binary_data must not contain the same key

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesConfigMap, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.configmap_name` | `string` | The name of the created ConfigMap — the handle workloads use in `configMapRef`, `configMapKeyRef`, and `configMap` volume definitions. |
| `status.outputs.namespace` | `string` | The namespace in which the ConfigMap was created. Consumers must live in the same namespace — ConfigMap references never cross namespaces in Kubernetes. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesBackendTlsPolicy | `spec.validation.caCertificateRefs[].name` | `status.outputs.configmap_name` |
| KubernetesGateway | `spec.tls.frontend.default.validation.caCertificateRefs[].name` | `status.outputs.configmap_name` |
| KubernetesGateway | `spec.tls.frontend.perPort[].tls.validation.caCertificateRefs[].name` | `status.outputs.configmap_name` |
| KubernetesGhaRunnerScaleSet | `spec.githubServerTls.configMapName` | `status.outputs.configmap_name` |
| KubernetesKafkaConnector | `spec.listOffsets.toConfigMap` | `status.outputs.configmap_name` |
| KubernetesKafkaConnector | `spec.alterOffsets.fromConfigMap` | `status.outputs.configmap_name` |
| KubernetesPodMonitor | `spec.podMetricsEndpoints[].tlsConfig.ca.configMap.name` | `status.outputs.configmap_name` |
| KubernetesPodMonitor | `spec.podMetricsEndpoints[].tlsConfig.cert.configMap.name` | `status.outputs.configmap_name` |
| KubernetesPodMonitor | `spec.podMetricsEndpoints[].oauth2.clientId.configMap.name` | `status.outputs.configmap_name` |
| KubernetesPodMonitor | `spec.podMetricsEndpoints[].oauth2.tlsConfig.ca.configMap.name` | `status.outputs.configmap_name` |
| KubernetesPodMonitor | `spec.podMetricsEndpoints[].oauth2.tlsConfig.cert.configMap.name` | `status.outputs.configmap_name` |
| KubernetesServiceMonitor | `spec.endpoints[].tlsConfig.ca.configMap.name` | `status.outputs.configmap_name` |
| KubernetesServiceMonitor | `spec.endpoints[].tlsConfig.cert.configMap.name` | `status.outputs.configmap_name` |
| KubernetesServiceMonitor | `spec.endpoints[].oauth2.clientId.configMap.name` | `status.outputs.configmap_name` |
| KubernetesServiceMonitor | `spec.endpoints[].oauth2.tlsConfig.ca.configMap.name` | `status.outputs.configmap_name` |
| KubernetesServiceMonitor | `spec.endpoints[].oauth2.tlsConfig.cert.configMap.name` | `status.outputs.configmap_name` |

## See Also

- [Overview](../README.md)
