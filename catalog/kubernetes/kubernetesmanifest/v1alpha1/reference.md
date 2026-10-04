# KubernetesManifest

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesManifestSpec** deploys raw Kubernetes YAML — the catalog's
bring-your-own-manifest escape hatch. Hand it any valid manifest (single or
multi-document, core kinds or custom resources) and both engines apply it
to the cluster exactly as written.

WHEN NOT TO USE THIS: a first-class catalog kind always wins. Typed
kinds validate configuration before deploy, export composable outputs
other resources can reference, and document their trade-offs field by
field — raw YAML does none of that. Reach for KubernetesManifest only when
the catalog has no kind for what you need to apply (a vendor's
install manifest, a CRD bundle, an exotic custom resource).

NAMESPACE SEMANTICS (identical on both engines): documents that declare
their own `metadata.namespace` keep it; namespaced documents that declare
none land in `spec.namespace`. Cluster-scoped documents (CRDs,
ClusterRoles, ...) are applied as-is — the anchor namespace never distorts
them. The manifest content itself is never otherwise mutated: no injected
labels, no rewritten fields.

## Example

```yaml
# Full-surface test manifest for the offline plan proofs. Exercises the
# anchor namespace with creation, multi-document YAML (a namespaced document
# WITHOUT an explicit namespace that must land in the anchor, one WITH its
# own namespace that must keep it), and the skip_await knob.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesManifest
metadata:
  name: test-manifest
spec:
  namespace:
    value: test-manifest-ns
  create_namespace: true
  skip_await: true
  manifest_yaml: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: anchored-config
    data:
      environment: test
      log_level: debug
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: pinned-config
      namespace: test-manifest-ns
    data:
      pinned: "true"
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: test-secret
    type: Opaque
    stringData:
      api-key: test-api-key-value
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.manifestYaml` | `string` | yes |  |  |
| `spec.skipAwait` | `bool` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

The anchor namespace for this manifest: namespaced documents that do not
declare their own `metadata.namespace` are applied here. Accepts a
literal namespace name or a reference to a KubernetesNamespace resource.
Documents with an explicit namespace, and cluster-scoped documents, are
unaffected.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the anchor namespace is created (with the standard Planton
governance labels) before the manifest is applied, and deleted with the
resource. When false, the namespace must already exist.

### spec.manifestYaml

`string` · required

The raw Kubernetes manifest YAML to apply. A single document or multiple
documents separated by `---`. Every valid Kubernetes resource is
accepted: core kinds, CRDs, and custom resources — including a CRD and
its custom resources in the same manifest (both engines order the CRD
install first).

Example multi-document manifest:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  key: value
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  ...
```

- rule: Manifest YAML must contain at least one Kubernetes document (a non-empty YAML body)
- rule: {"required":true}

### spec.skipAwait

`bool`

When true, neither engine waits for the applied resources to become
ready — the deploy returns as soon as the API server accepts every
document. When false (the default), both engines block until readiness:
Deployments/DaemonSets/StatefulSets complete their rollout, and other
kinds pass their engine's readiness checks. Skip the await for manifests
whose readiness depends on something deployed later (e.g. a webhook
configuration waiting on its service) or that intentionally stay
not-ready at install time.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesManifest, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | The anchor namespace: where namespaced documents without an explicit metadata.namespace were applied. |
| `status.outputs.applied_resources` | `[]string` | One entry per applied document, in manifest order, formatted as "<apiVersion>/<Kind>/<name>" (e.g. "apps/v1/Deployment/app", "v1/ConfigMap/app-config"). Documents keep the name they declare; the entry does not include the namespace (see `namespace` for the anchor). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
