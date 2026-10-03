# Kubernetes ConfigMap

## Overview

**KubernetesConfigMap** is a catalog kind that creates and manages Kubernetes ConfigMaps as first-class, declaratively managed resources. A ConfigMap holds non-confidential configuration data — file-like text values, property settings, or binary payloads — that pods consume as environment variables, command-line arguments, or mounted files.

The kind covers the complete Kubernetes ConfigMap surface: UTF-8 `data`, base64-encoded `binary_data`, and the `immutable` flag. There is nothing an upstream ConfigMap can express that this spec cannot.

## Purpose

ConfigMaps are the standard Kubernetes mechanism for decoupling configuration from container images. Every non-trivial workload consumes at least one: environment settings, feature flags, property files, nginx configs, dashboards. Managing them declaratively — with validation, drift detection, and cross-resource references — is what this kind provides.

**Key value over raw manifests:**

- **Schema-level validation**: Key character rules, base64 validation for `binary_data` values, and a cross-field rule rejecting keys that appear in both `data` and `binary_data` — all caught before anything reaches the cluster
- **Namespace by value or reference**: `spec.namespace` accepts a literal name or a reference to a `KubernetesNamespace` resource, so an infra chart can create the namespace and the ConfigMap in one run
- **Immutable ConfigMaps**: First-class support for the immutability flag (stable since Kubernetes 1.21)
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity
- **Lifecycle management**: Integrated with Planton's deployment lifecycle for status tracking and outputs

## Relationship to Other Components

- **KubernetesSecret**: The confidential mirror of this component. The two kinds are deliberate mirrors — same namespace handling, same immutability semantics, same key rules. Use KubernetesSecret for passwords, tokens, keys, and certificates; use KubernetesConfigMap for everything that is safe to read.
- **KubernetesNamespace**: Provides the target namespace. Reference it from `spec.namespace` to deploy both in one chart.
- **Workload components** (KubernetesDeployment and friends): Consume the ConfigMap by name via `envFrom`, `configMapKeyRef`, or `configMap` volumes. The created name and namespace are exported as outputs for exactly this composition.

## Data Model

### `data` — UTF-8 entries

Each key becomes a file name (when mounted as a volume) or an environment variable name (when consumed via `envFrom`/`configMapKeyRef`). Keys are restricted to alphanumeric characters, `-`, `_`, and `.` (max 253 characters). Values must be valid UTF-8 — anything else belongs in `binary_data`.

A ConfigMap with neither `data` nor `binary_data` is valid; Kubernetes allows empty ConfigMaps, commonly used as name reservations or coordination markers.

### `binary_data` — base64-encoded entries

For payloads that are not valid UTF-8: compiled files, images, serialized blobs. Values are supplied base64-encoded — the exact wire form the Kubernetes API uses for `binaryData` in YAML manifests. Keys share the same character rules as `data` keys and must not overlap with them; the schema enforces this at validation time, mirroring the API server's own rule.

Binary keys are only consumable as mounted files, never as environment variables.

### Size limit

The Kubernetes API server caps the combined size of `data` and `binary_data` at 1MiB. Oversized ConfigMaps are rejected at apply time.

## Immutability

When `spec.immutable` is `true`, the ConfigMap's data cannot be updated after creation (only metadata can change); updating requires deleting and recreating the ConfigMap. Immutable ConfigMaps:

- Protect against accidental configuration drift
- Materially reduce kube-apiserver watch load in clusters with many ConfigMap consumers (the kubelet stops watching immutable objects)

The recommended pattern for immutable configuration is versioned, roll-forward names (`app-config-v1`, `app-config-v2`, ...): each config change is a new ConfigMap, and the rollout is a workload update pointing at the new name — giving atomic, revertible config changes.

## Essential Configuration Fields

### Required

- **`spec.name`**: The ConfigMap name (DNS subdomain: lowercase alphanumeric, hyphens, dots, 1–253 chars). This is the handle pods reference in `configMapRef`, `configMapKeyRef`, and `configMap` volume definitions.

### Common

- **`spec.namespace`**: Literal namespace name or reference to a KubernetesNamespace resource. When omitted, the ConfigMap lands in the cluster's `default` namespace — the same behavior as kubectl without a namespace flag.
- **`spec.data`**: UTF-8 key-value entries.
- **`spec.binaryData`**: Base64-encoded binary entries.
- **`spec.immutable`**: Locks data after creation.
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton labels for tracking and governance.

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`configmap_name`**: The name of the created ConfigMap — the handle workloads use in `configMapRef`, `configMapKeyRef`, and volume definitions
- **`namespace`**: The namespace in which the ConfigMap was created. Consumers must live in the same namespace — ConfigMap references never cross namespaces in Kubernetes

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace (literal value or resolved reference)
2. Merge user labels and annotations with standard Planton tracking labels
3. Create the Kubernetes ConfigMap with `data`, `binaryData`, and the `immutable` flag
4. Export the ConfigMap name and namespace for downstream composition

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesConfigMap** when you need:

- Declarative management of non-confidential configuration as infrastructure
- Property files, environment settings, or feature flags consumed by workloads
- Binary configuration payloads (up to the 1MiB limit)
- Immutable, versioned configuration for production rollouts
- Namespace-and-config created together in one infra chart

**Do NOT use** when:

- The data is confidential — passwords, tokens, keys, certificates belong in **KubernetesSecret**. ConfigMap contents are readable by anyone with `get` access on ConfigMaps and are never encrypted at rest by the secrets machinery
- The payload exceeds 1MiB — use a volume, an object store, or an image layer instead

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist before creating the ConfigMap (unless deploying to `default`, or creating the namespace in the same chart via a reference)

## Best Practices

1. **Never put secrets in ConfigMaps**: There is no access-control or encryption benefit; use KubernetesSecret
2. **Use immutable + versioned names for production config**: `app-config-v2` style names give atomic rollouts, easy rollback, and lower API server load
3. **Prefer file-shaped keys for file-shaped config**: A key like `application.properties` mounted as a volume is clearer than dozens of individual scalar keys
4. **Keep consumers in the same namespace**: ConfigMap references never cross namespaces; plan placement accordingly
5. **Label for governance**: Add `team`, `environment`, and `purpose` labels for auditing and cost attribution

## References

- [Kubernetes ConfigMaps Documentation](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [Configure a Pod to Use a ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/)
- [Immutable ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/#configmap-immutable)
- [ConfigMap API Reference](https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/config-map-v1/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
