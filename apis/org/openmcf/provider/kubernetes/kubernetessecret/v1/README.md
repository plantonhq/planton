# Kubernetes Secret

## Overview

**KubernetesSecret** is an OpenMCF deployment component that implements a "Secret-as-a-Service" pattern for creating and managing Kubernetes Secrets as first-class, declaratively managed resources. It provides type-safe configuration for all common Kubernetes secret types -- Opaque, TLS, DockerConfigJson, BasicAuth, and SSHAuth -- with per-type validation and a clean, structured API.

## Purpose

Kubernetes Secrets are foundational to every cluster, yet creating and managing them declaratively with proper typing and validation requires boilerplate across different IaC tools. This component abstracts that complexity into a single, type-safe API that follows the 80/20 principle: supporting the five most common secret types that cover the vast majority of production use cases.

**Key value over raw manifests:**

- **Type-safe data variants**: Each secret type has its own structured message with required fields, preventing mistyped keys (e.g., forgetting `tls.key` in a TLS secret)
- **Validation**: DNS name validation, required field enforcement, and per-type constraints at the schema level
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity
- **Lifecycle management**: Integrated with OpenMCF's deployment lifecycle for status tracking and outputs
- **Immutable secrets**: First-class support for Kubernetes immutable secrets (1.21+)

## Relationship to Other Components

- **KubernetesExternalSecrets** (enum 829): Syncs secrets _from external stores_ (AWS Secrets Manager, Vault, etc.) into Kubernetes. Use when secrets originate in an external provider.
- **KubernetesSecret** (this component): Creates secrets _directly_ with literal values provided at deploy time. Use when secret data is available in your CI/CD pipeline, environment config, or IaC variables.

These are complementary components, not overlapping.

## Supported Secret Types

### 1. Opaque

The most common secret type, used for arbitrary key-value pairs. Maps to Kubernetes type `Opaque`.

### 2. TLS

For storing TLS certificates and private keys. Maps to Kubernetes type `kubernetes.io/tls`. Requires both `tls_crt` and `tls_key` fields.

### 3. DockerConfigJson

For authenticating with container registries during image pulls. Maps to Kubernetes type `kubernetes.io/dockerconfigjson`. Requires `registry_server`, `username`, and `password`.

### 4. BasicAuth

For username/password authentication. Maps to Kubernetes type `kubernetes.io/basic-auth`. Requires both `username` and `password`.

### 5. SSHAuth

For SSH key-based authentication. Maps to Kubernetes type `kubernetes.io/ssh-auth`. Requires `ssh_private_key`.

## Essential Configuration Fields

### Required

- **`spec.name`**: The secret name (DNS subdomain: lowercase alphanumeric, hyphens, dots, 1-253 chars)
- **`spec.secret_data`**: Exactly one of `opaque`, `tls`, `docker_config_json`, `basic_auth`, or `ssh_auth`

### Common

- **`spec.namespace`**: Namespace where the secret is created (defaults to `"default"`)
- **`spec.labels`**: Additional labels for governance and resource tracking
- **`spec.annotations`**: Annotations for custom metadata
- **`spec.immutable`**: When true, prevents updates to secret data after creation

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`secret_name`**: The name of the created Kubernetes Secret
- **`secret_namespace`**: The namespace where the secret was created
- **`secret_type`**: The Kubernetes secret type string (e.g., `"Opaque"`, `"kubernetes.io/tls"`)

## How It Works

This component includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Determine the Kubernetes secret type from the `oneof secret_data` variant
2. Map the type-safe fields to the corresponding Kubernetes Secret `stringData` keys
3. Create the Kubernetes Secret with the appropriate type, data, labels, and annotations
4. Apply the immutable flag if specified
5. Export observable outputs for downstream reference

Both IaC implementations have feature parity and follow identical logic, ensuring consistent behavior regardless of which tool you use.

## When to Use

Use **KubernetesSecret** when you need:

- Declarative management of Kubernetes Secrets as infrastructure
- Type-safe validation of secret data (TLS cert+key pairs, registry credentials, etc.)
- Secrets whose values are available at deploy time (from CI/CD, env vars, or IaC config)
- Immutable secrets for production workloads
- Consistent secret management across Pulumi and Terraform

**Do NOT use** when:

- Secret values originate in an external store (use **KubernetesExternalSecrets** instead)
- You need dynamic secret rotation (consider Vault or External Secrets Operator)

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist before creating the secret (unless deploying to `default`)

## Best Practices

1. **Use immutable for production secrets**: Set `immutable: true` to prevent accidental modifications and improve API server performance
2. **Use type-specific variants**: Prefer `tls`, `docker_config_json`, etc. over `opaque` when the secret type is known -- this gives you schema-level validation
3. **Avoid storing secrets in version control**: The manifest will contain literal secret values; pass them via CI/CD variables or environment config
4. **Use meaningful names**: Follow the pattern `{app}-{purpose}` (e.g., `myapp-db-credentials`, `myapp-tls-cert`)
5. **Label for governance**: Add `team`, `environment`, and `purpose` labels for cost tracking and auditing

## References

- [Kubernetes Secrets Documentation](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Secret Types](https://kubernetes.io/docs/concepts/configuration/secret/#secret-types)
- [Immutable Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#secret-immutable)
- [Managing Secrets with kubectl](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kubectl/)
- [Good Practices for Kubernetes Secrets](https://kubernetes.io/docs/concepts/security/secrets-good-practices/)
