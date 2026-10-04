# KubernetesExternalSecret

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesExternalSecretSpec** declares ONE secret sync: the External
Secrets Operator reads the referenced entries from a store's backend (AWS
Secrets Manager, GCP Secret Manager, Azure Key Vault, Vault/OpenBao, ...)
and materializes them as a Kubernetes Secret in this namespace, refreshing
on the configured interval. Workloads then consume the materialized Secret
exactly like any other (env valueFrom, volume mounts) — the external system
stays the single source of truth and the cluster never stores the value
anywhere else.

The store connection is a separate first-class resource
(KubernetesSecretStore / KubernetesClusterSecretStore); this resource picks
WHAT to sync — explicit entries via `data`, and/or bulk pulls via
`data_from`. Requires the External Secrets Operator on the cluster
(KubernetesExternalSecretsOperator).

## Example

```yaml
# Full-surface offline-proof manifest: exercises the cluster-store
# reference, both sync forms (explicit data with property/version/decoding
# and dataFrom with find + rewrite), the templated target with lifecycle
# policies, and refresh tuning — so the offline tofu plan and pulumi
# preview proofs cover every rendering arm. Placeholder values; never
# applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesExternalSecret
metadata:
  name: hack-external-secret
spec:
  namespace:
    value: team-a
  storeRef:
    clusterSecretStore:
      name:
        value: platform-aws
  refreshInterval: 15m
  refreshPolicy: Periodic
  target:
    name: app-credentials
    creationPolicy: Owner
    deletionPolicy: Delete
    immutable: false
    template:
      type: Opaque
      mergePolicy: Merge
      labels:
        app: payments
      annotations:
        example.org/rendered-by: external-secrets
      data:
        connection-string: "postgres://{{ .username }}:{{ .password }}@db:5432/app"
  data:
    - secretKey: password
      remoteRef:
        key: prod/app/database
        property: password
        version: "2"
        decodingStrategy: Auto
    - secretKey: username
      remoteRef:
        key: prod/app/database
        property: username
  dataFrom:
    - extract:
        key: prod/app/all
      rewrite:
        - source: "^prod/app/(.*)$"
          target: "$1"
    - find:
        path: prod/
        nameRegexp: "^app-.*"
        tags:
          env: prod
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.storeRef` | `KubernetesExternalSecretStoreRef` | yes |  |  |
| `spec.storeRef.secretStore` | `KubernetesExternalSecretSecretStoreRef` |  |  |  |
| `spec.storeRef.secretStore.name` | `string \| valueFrom` | yes |  | KubernetesSecretStore (`status.outputs.store_name`) |
| `spec.storeRef.clusterSecretStore` | `KubernetesExternalSecretClusterSecretStoreRef` |  |  |  |
| `spec.storeRef.clusterSecretStore.name` | `string \| valueFrom` | yes |  | KubernetesClusterSecretStore (`status.outputs.store_name`) |
| `spec.refreshInterval` | `string` |  | `1h` |  |
| `spec.refreshPolicy` | `string` |  |  |  |
| `spec.target` | `KubernetesExternalSecretTarget` |  |  |  |
| `spec.target.name` | `string` |  |  |  |
| `spec.target.creationPolicy` | `string` |  | `Owner` |  |
| `spec.target.deletionPolicy` | `string` |  | `Retain` |  |
| `spec.target.immutable` | `bool` |  |  |  |
| `spec.target.template` | `KubernetesExternalSecretTemplate` |  |  |  |
| `spec.target.template.type` | `string` |  |  |  |
| `spec.target.template.mergePolicy` | `string` |  | `Replace` |  |
| `spec.target.template.labels` | `map<string, string>` |  |  |  |
| `spec.target.template.annotations` | `map<string, string>` |  |  |  |
| `spec.target.template.data` | `map<string, string>` |  |  |  |
| `spec.data` | `[]KubernetesExternalSecretData` |  |  |  |
| `spec.data[].secretKey` | `string` | yes |  |  |
| `spec.data[].remoteRef` | `KubernetesExternalSecretRemoteRef` | yes |  |  |
| `spec.data[].remoteRef.key` | `string` | yes |  |  |
| `spec.data[].remoteRef.property` | `string` |  |  |  |
| `spec.data[].remoteRef.version` | `string` |  |  |  |
| `spec.data[].remoteRef.decodingStrategy` | `string` |  | `None` |  |
| `spec.dataFrom` | `[]KubernetesExternalSecretDataFrom` |  |  |  |
| `spec.dataFrom[].extract` | `KubernetesExternalSecretRemoteRef` |  |  |  |
| `spec.dataFrom[].extract.key` | `string` | yes |  |  |
| `spec.dataFrom[].extract.property` | `string` |  |  |  |
| `spec.dataFrom[].extract.version` | `string` |  |  |  |
| `spec.dataFrom[].extract.decodingStrategy` | `string` |  | `None` |  |
| `spec.dataFrom[].find` | `KubernetesExternalSecretFind` |  |  |  |
| `spec.dataFrom[].find.path` | `string` |  |  |  |
| `spec.dataFrom[].find.nameRegexp` | `string` |  |  |  |
| `spec.dataFrom[].find.tags` | `map<string, string>` |  |  |  |
| `spec.dataFrom[].rewrite` | `[]KubernetesExternalSecretRewrite` |  |  |  |
| `spec.dataFrom[].rewrite[].source` | `string` | yes |  |  |
| `spec.dataFrom[].rewrite[].target` | `string` | yes |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace the ExternalSecret (and the Secret it materializes) lives in.
Accepts a literal namespace name or a reference to a KubernetesNamespace
resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.storeRef

`KubernetesExternalSecretStoreRef` · required

The store this secret syncs from.

- rule: {"required":true}
- rule: A store is required -- choose 'secret_store' (a namespaced SecretStore in this secret's namespace) or 'cluster_secret_store' (a cluster-scoped ClusterSecretStore any namespace can read)

### spec.storeRef.secretStore

`KubernetesExternalSecretSecretStoreRef`

A namespaced SecretStore in this secret's namespace, by name.

### spec.storeRef.secretStore.name

`string | valueFrom` · required

SecretStore name. Accepts a literal name or a reference to a
KubernetesSecretStore resource.

- references: KubernetesSecretStore (`status.outputs.store_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecretStore, name: <that resource's name>, fieldPath: status.outputs.store_name}} -- a bare string does not parse

### spec.storeRef.clusterSecretStore

`KubernetesExternalSecretClusterSecretStoreRef`

A cluster-scoped ClusterSecretStore, by name.

### spec.storeRef.clusterSecretStore.name

`string | valueFrom` · required

ClusterSecretStore name. Accepts a literal name or a reference to a
KubernetesClusterSecretStore resource.

- references: KubernetesClusterSecretStore (`status.outputs.store_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClusterSecretStore, name: <that resource's name>, fieldPath: status.outputs.store_name}} -- a bare string does not parse

### spec.refreshInterval

`string` · optional (explicit presence)

How often ESO re-reads the backend and refreshes the Secret (Go
duration). Upstream default: "1h". "0s" = fetch exactly once (pair with
refresh_policy CreatedOnce for immutable bootstrap secrets).

- default: `1h`
- rule: {"string":{"pattern":"^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.refreshPolicy

`string` · optional (explicit presence)

When the Secret is (re)synced. Empty = upstream behavior (periodic per
refresh_interval).
"CreatedOnce": create once, never touch again (drift stays).
"Periodic": sync on the interval (the standard posture).
"OnChange": sync only when this ExternalSecret's spec changes.

- rule: {"string":{"in":["","CreatedOnce","Periodic","OnChange"]}}

### spec.target

`KubernetesExternalSecretTarget`

The Kubernetes Secret to materialize. Empty = a Secret named after this
resource with default policies (owned, retained on delete).

### spec.target.name

`string`

Name of the materialized Secret. Empty = this resource's metadata.name.

### spec.target.creationPolicy

`string` · optional (explicit presence)

Who owns the Secret's lifecycle.
"Owner" (default): ESO creates it, owns it (ownerReference), and fails
  if an unowned Secret of that name already exists.
"Orphan": ESO creates it but never garbage-collects it.
"Merge": ESO merges synced keys into an EXISTING Secret it does not own
  (fails when absent).
"None": ESO never writes the Secret (paired with generic manifest
  targets via helm-level tooling; niche).

- default: `Owner`
- rule: {"string":{"in":["Owner","Orphan","Merge","None"]}}

### spec.target.deletionPolicy

`string` · optional (explicit presence)

What happens to the Secret when this ExternalSecret is deleted or a
synced key disappears from the backend.
"Retain" (upstream default): keep the Secret — the safe posture.
"Delete": delete the Secret with the ExternalSecret (and prune keys
  that vanish from the backend).
"Merge": remove only the keys this ExternalSecret owns.

- default: `Retain`
- rule: {"string":{"in":["Retain","Delete","Merge"]}}

### spec.target.immutable

`bool`

When true, the materialized Secret is marked immutable — consumers get
kubelet-cache stability, but every refresh that would change data FAILS
(immutable Secrets cannot be updated). Pair with refresh_policy
CreatedOnce.

### spec.target.template

`KubernetesExternalSecretTemplate`

Template applied when materializing the Secret — set its type, stamp
metadata, or reshape values.

### spec.target.template.type

`string`

Secret type (e.g. "kubernetes.io/dockerconfigjson",
"kubernetes.io/tls"). Empty = "Opaque".

### spec.target.template.mergePolicy

`string` · optional (explicit presence)

How templated data combines with synced data.
"Replace" (upstream default): the template's data REPLACES the synced
  keys — only templated keys land in the Secret.
"Merge": templated keys merge over the synced keys.

- default: `Replace`
- rule: {"string":{"in":["Replace","Merge"]}}

### spec.target.template.labels

`map<string, string>`

Labels stamped on the materialized Secret.

### spec.target.template.annotations

`map<string, string>`

Annotations stamped on the materialized Secret.

### spec.target.template.data

`map<string, string>`

Templated data entries. Values are Go templates over the synced keys —
e.g. reshape a username/password pair into a connection string:
`postgres://{{ .username }}:{{ .password }}@db:5432/app`.

### spec.data

`[]KubernetesExternalSecretData`

Explicit entries: each maps one backend key (or one property within it)
to one key of the materialized Secret. The precise, reviewable form —
prefer it for application credentials.

### spec.data[].secretKey

`string` · required

Key in the materialized Kubernetes Secret this entry lands in.

- rule: {"required":true}

### spec.data[].remoteRef

`KubernetesExternalSecretRemoteRef` · required

Backend entry to read.

- rule: {"required":true}

### spec.data[].remoteRef.key

`string` · required

Backend key — the secret's name/path in the external system (e.g.
"prod/app/database" in AWS Secrets Manager, "secret/data/app" in
Vault v2, the secret name in GCP/Azure).

- rule: {"required":true}

### spec.data[].remoteRef.property

`string`

Property WITHIN the backend entry when it holds structured data (a JSON
document): e.g. "password" extracts just that field. Empty = the whole
value.

### spec.data[].remoteRef.version

`string`

Backend version of the entry (e.g. an AWS version stage, a GCP version
number, a Vault v2 version). Empty = latest.

### spec.data[].remoteRef.decodingStrategy

`string` · optional (explicit presence)

Decode the backend value before storing it.
"None" (upstream default): store as-is.
"Base64"/"Base64URL": decode base64(url) — for backends that store
  binary content encoded.
"Auto": decode when the value looks base64-encoded, else store as-is.

- default: `None`
- rule: {"string":{"in":["None","Base64","Base64URL","Auto"]}}

### spec.dataFrom

`[]KubernetesExternalSecretDataFrom`

Bulk pulls: each extracts ALL properties of one backend entry — or finds
every entry matching a name pattern / tags — into the materialized
Secret. For structured entries (a JSON document of related credentials)
and fleet patterns; `rewrite` reshapes the resulting keys.

- rule: Each data_from pull needs its source — set extract (one structured entry) or find (entries by pattern)

### spec.dataFrom[].extract

`KubernetesExternalSecretRemoteRef`

Pull ALL properties of one structured backend entry (a JSON document)
— each property becomes a Secret key.

### spec.dataFrom[].extract.key

`string` · required

Backend key — the secret's name/path in the external system (e.g.
"prod/app/database" in AWS Secrets Manager, "secret/data/app" in
Vault v2, the secret name in GCP/Azure).

- rule: {"required":true}

### spec.dataFrom[].extract.property

`string`

Property WITHIN the backend entry when it holds structured data (a JSON
document): e.g. "password" extracts just that field. Empty = the whole
value.

### spec.dataFrom[].extract.version

`string`

Backend version of the entry (e.g. an AWS version stage, a GCP version
number, a Vault v2 version). Empty = latest.

### spec.dataFrom[].extract.decodingStrategy

`string` · optional (explicit presence)

Decode the backend value before storing it.
"None" (upstream default): store as-is.
"Base64"/"Base64URL": decode base64(url) — for backends that store
  binary content encoded.
"Auto": decode when the value looks base64-encoded, else store as-is.

- default: `None`
- rule: {"string":{"in":["None","Base64","Base64URL","Auto"]}}

### spec.dataFrom[].find

`KubernetesExternalSecretFind`

Pull EVERY backend entry matching a name pattern and/or tags — each
matched entry becomes a Secret key.

- rule: find needs a criterion — set name_regexp and/or tags (path alone only scopes the search)

### spec.dataFrom[].find.path

`string`

Path/prefix to search under (backend-specific semantics).

### spec.dataFrom[].find.nameRegexp

`string`

Regular expression matched against backend entry names (e.g.
"^prod-app-.*").

### spec.dataFrom[].find.tags

`map<string, string>`

Backend tags/labels the entries must carry (all must match).

### spec.dataFrom[].rewrite

`[]KubernetesExternalSecretRewrite`

Key rewrites applied to the pulled entries, in order — e.g. strip a
"prod/app/" path prefix so Secret keys are bare names.

### spec.dataFrom[].rewrite[].source

`string` · required

Regular expression applied to each pulled key (e.g. "^prod/app/(.*)$").

- rule: {"required":true}

### spec.dataFrom[].rewrite[].target

`string` · required

Replacement, with capture groups (e.g. "$1").

- rule: {"required":true}

## Validation Rules

- `xsec.data_or_data_from`: Declare something to sync — at least one data entry or one data_from pull

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesExternalSecret, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.external_secret_name` | `string` | Name of the created ExternalSecret (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the ExternalSecret and its materialized Secret live in. |
| `status.outputs.secret_name` | `string` | Name of the Kubernetes Secret the operator materializes (target.name, defaulting to metadata.name). Workloads consume THIS Secret — wire env valueFrom / volume secretName references to it. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.storeRef.secretStore.name` | KubernetesSecretStore | `status.outputs.store_name` |
| `spec.storeRef.clusterSecretStore.name` | KubernetesClusterSecretStore | `status.outputs.store_name` |

## See Also

- [Overview](../README.md)
