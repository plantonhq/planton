---
title: "Secret"
description: "Secret deployment documentation"
icon: "package"
order: 100
componentName: "kubernetessecret"
---

# Kubernetes Secret

Deploys a type-safe Kubernetes Secret to a target cluster, supporting Opaque, TLS, Docker registry, basic-auth, and SSH-auth secret types through a single declarative manifest. The IaC module handles data encoding, label merging, and `.dockerconfigjson` construction automatically.

## What Gets Created

When you deploy a KubernetesSecret resource, OpenMCF provisions:

- **Secret** — a Kubernetes Secret with the appropriate type (`Opaque`, `kubernetes.io/tls`, `kubernetes.io/dockerconfigjson`, `kubernetes.io/basic-auth`, or `kubernetes.io/ssh-auth`), populated via `stringData` so values can be supplied as plain strings
- **Labels** — standard OpenMCF tracking labels (`managed-by`, `resource`, `resource-kind`) merged with any user-provided labels
- **Annotations** — user-provided annotations applied to the Secret metadata

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists (the module does not create namespaces)
- **Secret values** ready to supply — PEM-encoded certificates for TLS, registry credentials for Docker, SSH private keys for SSH auth, or plain key-value pairs for Opaque secrets

## Quick Start

Create a file `secret.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: my-secret
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesSecret.my-secret
spec:
  name: my-secret
  opaque:
    data:
      API_KEY: "sk-example-key-value"
```

Deploy:

```shell
openmcf apply -f secret.yaml
```

This creates an Opaque Kubernetes Secret named `my-secret` in the `default` namespace with a single key `API_KEY`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `spec.name` | `string` | Name of the Kubernetes Secret. Must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots). | 1–253 characters, matches `^[a-z0-9]([a-z0-9.-]{0,251}[a-z0-9])?$` |
| `spec.<secretData>` | `oneof` | Exactly one secret data variant must be provided: `opaque`, `tls`, `dockerConfigJson`, `basicAuth`, or `sshAuth`. | One variant required |
| `spec.opaque.data` | `map<string, string>` | Key-value pairs of secret data (when using the `opaque` variant). | At least 1 entry |
| `spec.tls.tlsCrt` | `string` | PEM-encoded TLS certificate or certificate chain (when using the `tls` variant). | Non-empty |
| `spec.tls.tlsKey` | `string` | PEM-encoded TLS private key (when using the `tls` variant). | Non-empty |
| `spec.dockerConfigJson.registryServer` | `string` | Docker registry server URL, e.g., `https://index.docker.io/v1/` (when using the `dockerConfigJson` variant). | Non-empty |
| `spec.dockerConfigJson.username` | `string` | Registry authentication username (when using the `dockerConfigJson` variant). | Non-empty |
| `spec.dockerConfigJson.password` | `string` | Registry authentication password or access token (when using the `dockerConfigJson` variant). | Non-empty |
| `spec.basicAuth.username` | `string` | Username for basic authentication (when using the `basicAuth` variant). | Non-empty |
| `spec.basicAuth.password` | `string` | Password for basic authentication (when using the `basicAuth` variant). | Non-empty |
| `spec.sshAuth.sshPrivateKey` | `string` | PEM-encoded SSH private key (when using the `sshAuth` variant). | Non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `spec.targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `spec.targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `spec.namespace` | `string` | `default` | Namespace where the secret will be created. Max 63 characters, valid DNS label. |
| `spec.labels` | `map<string, string>` | `{}` | Additional labels merged with standard OpenMCF labels. |
| `spec.annotations` | `map<string, string>` | `{}` | Additional annotations applied to the secret. |
| `spec.immutable` | `bool` | `false` | When `true`, the secret data cannot be updated after creation. Immutable secrets reduce API server watch load. |
| `spec.dockerConfigJson.email` | `string` | — | Optional email associated with the registry account. |

## Examples

### Basic Opaque Secret

A simple key-value secret for storing application credentials:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: app-credentials
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesSecret.app-credentials
spec:
  name: app-credentials
  namespace: backend
  opaque:
    data:
      DB_PASSWORD: "s3cret-passw0rd"
      API_TOKEN: "tok_abc123"
```

### TLS Certificate Secret

A TLS secret for use with Ingress controllers or services that terminate TLS:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: api-tls
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesSecret.api-tls
spec:
  name: api-tls
  namespace: ingress
  immutable: true
  tls:
    tlsCrt: |
      -----BEGIN CERTIFICATE-----
      MIIBkTCB+wIJAL...
      -----END CERTIFICATE-----
    tlsKey: |
      -----BEGIN EC PRIVATE KEY-----
      MHQCAQEEIBk2...
      -----END EC PRIVATE KEY-----
```

### Docker Registry Credentials with Labels

A Docker registry pull secret for authenticating with a private container registry, with custom labels and annotations:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: ghcr-pull-secret
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesSecret.ghcr-pull-secret
spec:
  name: ghcr-pull-secret
  namespace: production
  labels:
    team: platform
    cost-center: infra
  annotations:
    description: "GitHub Container Registry pull credentials"
  immutable: true
  dockerConfigJson:
    registryServer: ghcr.io
    username: my-bot
    password: "ghp_xxxxxxxxxxxxxxxxxxxx"
    email: bot@example.com
```

The module constructs the `.dockerconfigjson` JSON automatically from the structured fields, including base64-encoded auth.

### Basic Auth Secret

A basic authentication secret for services that require username/password credentials:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: monitoring-auth
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesSecret.monitoring-auth
spec:
  name: monitoring-auth
  namespace: monitoring
  basicAuth:
    username: admin
    password: "mon!t0r-pass"
```

### SSH Auth Secret

An SSH authentication secret for Git operations or SSH-based access:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: deploy-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesSecret.deploy-key
spec:
  name: deploy-key
  namespace: flux-system
  immutable: true
  sshAuth:
    sshPrivateKey: |
      -----BEGIN OPENSSH PRIVATE KEY-----
      b3BlbnNzaC1rZXktdjEA...
      -----END OPENSSH PRIVATE KEY-----
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `secretName` | `string` | Name of the created Kubernetes Secret |
| `secretNamespace` | `string` | Namespace where the Kubernetes Secret was created |
| `secretType` | `string` | Kubernetes secret type string (`Opaque`, `kubernetes.io/tls`, `kubernetes.io/dockerconfigjson`, `kubernetes.io/basic-auth`, or `kubernetes.io/ssh-auth`) |

## Related Components

- [KubernetesDeployment](/docs/catalog/kubernetes/deployment) — references secrets as environment variables via `container.app.env.secrets` or as image pull secrets
- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace for the secret
