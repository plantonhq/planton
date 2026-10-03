# Kubernetes Secret

Deploys a type-safe Kubernetes Secret supporting Opaque, TLS, Docker registry, Basic Auth, SSH Auth, and Service Account Token secret types. Exactly one variant is declared per resource, each with type-specific fields validated at creation time, so the correct Kubernetes secret type is produced instead of a mistyped Opaque blob.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Secret** -- a single Secret resource in the specified namespace with the correct `type` field set automatically based on the chosen variant (Opaque, `kubernetes.io/tls`, `kubernetes.io/dockerconfigjson`, `kubernetes.io/basic-auth`, `kubernetes.io/ssh-auth`, or `kubernetes.io/service-account-token`). UTF-8 values are written using Kubernetes `stringData` semantics; Opaque `binaryData` entries are written pre-encoded as base64.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- The target namespace must already exist (the module does not create it). Use the Kubernetes Namespace kind to manage namespaces declaratively.

## Deploy

### Console

Open the deployment store, find **Kubernetes Secret**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Opaque Secret** preset for general-purpose secrets or **TLS Secret** for certificate key pairs in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSecret
metadata:
  name: app-credentials
  org: acme-corp
  env: prod
spec:
  name: app-credentials
  namespace:
    value: backend-services
  opaque:
    data:
      DB_PASSWORD: your-db-password
      API_KEY: your-api-token
```

```shell
planton apply -f secret.yaml
```

This creates an Opaque secret in the `backend-services` namespace with two key-value pairs. The secret type is set to `Opaque` automatically. Immutability and additional labels are not configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the Secret to its Planton-managed namespace:

```yaml
spec:
  name: app-credentials
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: backend-namespace
      fieldPath: spec.name
  opaque:
    data:
      DB_PASSWORD: your-db-password
```

The InfraPipeline deploys the namespace first, then creates the Secret inside it.

## Key Configuration

These are the most important decisions when configuring a Kubernetes Secret. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Secret type selection** -- Exactly one variant must be provided: `opaque` for arbitrary key-value data (with a `binaryData` map for base64 payloads), `tls` for certificate and private key pairs, `dockerConfigJson` for container registry credentials, `basicAuth` for username/password pairs, `sshAuth` for SSH private keys, or `serviceAccountToken` for a long-lived API token the cluster mints for a ServiceAccount. The variant determines the Kubernetes secret type automatically.

**Immutability** -- Set `immutable` to `true` to prevent updates after creation. Immutable secrets reduce API server watch load and protect against accidental overwrites. Once set, the secret data cannot be changed -- you must delete and recreate it.

**Namespace targeting** -- The `namespace` field defaults to `default` if omitted. Ensure the target namespace exists before deploying.

**TLS certificate chain** -- When using the `tls` variant, provide the full PEM-encoded certificate chain in `tlsCrt` (leaf + intermediates) and the corresponding private key in `tlsKey`. Ingress controllers and service meshes automatically validate this type.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesServiceAccount** | `serviceAccountToken.serviceAccountName` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `secret_name` | The name of the created Kubernetes Secret | Pod `envFrom.secretRef` or volume mount references |
| `secret_namespace` | The namespace where the secret was created | Cross-namespace secret references |
| `secret_type` | The Kubernetes secret type string (e.g., Opaque, kubernetes.io/tls) | Conditional logic in downstream configurations |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Application credentials** -- Opaque secret storing database passwords, API keys, and connection strings consumed by application pods via `envFrom` or mounted volumes. Start from the **Opaque Secret** preset.

**Ingress TLS termination** -- TLS secret containing a certificate and private key pair referenced by Ingress resources for HTTPS termination. Use when cert-manager is not available and certificates are managed externally. Start from the **TLS Secret** preset.

**Private registry access** -- Docker registry secret providing authentication for pulling images from private container registries (Docker Hub, GCR, ECR, ACR, GHCR). Referenced by pods via `imagePullSecrets`. Start from the **Docker Registry Secret** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- reference the namespace so infra charts create it and this Secret in dependency order.
- [**Kubernetes ServiceAccount**](/infra-catalog/kubernetes-service-account) -- the `serviceAccountToken` variant references the identity its token belongs to; docker-registry Secrets are attached to ServiceAccounts as `imagePullSecrets`.
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) and the other workload kinds -- consume secrets as env vars, mounted files, or registry credentials.