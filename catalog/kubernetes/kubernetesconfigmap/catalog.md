# Kubernetes ConfigMap

Deploys a Kubernetes ConfigMap carrying UTF-8 configuration entries and base64 binary payloads that workloads consume as environment variables or mounted files. Supports the immutable flag for versioned, rollout-friendly configuration. The spec covers the complete upstream ConfigMap surface; for confidential data use Kubernetes Secret instead — the two kinds are deliberate mirrors of each other.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes ConfigMap** -- a single ConfigMap in the specified namespace carrying the `data` (UTF-8) and `binaryData` (base64) entries, with the `immutable` flag applied when set
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- The target namespace must already exist (the module does not create it). Use the Kubernetes Namespace kind to manage namespaces declaratively.

## Deploy

### Console

Open the deployment store, find **Kubernetes ConfigMap**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Application Configuration** preset for flag-style settings or **Immutable Versioned Configuration** for production rollout discipline in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesConfigMap
metadata:
  name: checkout-settings
  org: acme-corp
  env: prod
spec:
  name: checkout-settings
  namespace:
    value: backend-services
  data:
    LOG_LEVEL: "info"
    FEATURE_FLAGS: "fast-search,checkout-v2"
```

```shell
planton apply -f configmap.yaml
```

This creates a mutable ConfigMap in the `backend-services` namespace with two flag-style entries, ready to consume via `envFrom.configMapRef`. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference the namespace so it is created before the ConfigMap:

```yaml
spec:
  name: checkout-settings
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: backend-namespace
      fieldPath: spec.name
  data:
    LOG_LEVEL: "info"
```

The InfraPipeline creates the namespace first, then the ConfigMap inside it in dependency order.

## Key Configuration

These are the most important decisions when configuring a Kubernetes ConfigMap. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Key style follows consumption** -- Env-var-style keys (`LOG_LEVEL`) suit `envFrom` injection; file-style keys (`application.yaml`, `nginx.conf`) suit volume mounts where each key becomes a file. Keys allow alphanumerics, dashes, underscores, and dots (max 253 characters).

**Data vs binary data** -- `data` holds UTF-8 values (from single flags to whole embedded config files); `binaryData` holds base64-encoded payloads for non-UTF-8 content like keystores. A key may live in one map or the other, never both -- inside the ConfigMap they merge into one key space, capped at 1 MiB combined.

**Immutability** -- Set `immutable` to `true` to freeze the data forever (Kubernetes rejects every subsequent edit) and reduce API server watch load. Changing immutable config means creating a NEW ConfigMap with a versioned name and repointing workloads -- an explicit, reviewable rollout.

**Rotation reality for mutable ConfigMaps** -- In-place edits do NOT restart consumers: env vars refresh only on pod restart; volume-mounted files refresh within about a minute.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

The namespace field is optional: when omitted, the ConfigMap lands in the cluster's `default` namespace — the same behavior as kubectl without a namespace flag.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `configmap_name` | The name of the created ConfigMap | Pod `envFrom.configMapRef`, `configMapKeyRef`, and `configMap` volume sources |
| `namespace` | The namespace where the ConfigMap was created | Verifying consumer co-location (consumption is namespace-local) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Application settings** -- Flag-style entries (log levels, feature flags) injected as env vars via `envFrom`. Start from the **Application Configuration** preset.

**Immutable versioned config** -- A versioned name (`app-settings-v1`) with `immutable: true`; every config change becomes a new object and a visible diff in the workload's reference. Start from the **Immutable Versioned Configuration** preset.

**Binary payloads** -- Keystores or binary certificates carried as base64 in `binaryData`, mounted as files. Start from the **Binary Payload** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- reference the namespace so infra charts create it and this ConfigMap in dependency order.
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- workloads consume entries as env vars (`envFrom`, `configMapKeyRef`) or mounted files (`configMap` volume source), from the same namespace only.
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) -- the confidential mirror of this kind; put credentials and keys there, not here.
