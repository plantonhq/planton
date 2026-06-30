# Kubernetes Secret: Research Documentation

## Introduction

Kubernetes Secrets are the standard mechanism for storing and distributing sensitive data -- passwords, tokens, keys, and certificates -- to pods and other workloads running in a cluster. Despite their name, Kubernetes Secrets are not encrypted by default in etcd; they are merely base64-encoded. Their value lies in the lifecycle management, access control (via RBAC), and integration patterns they enable across the Kubernetes ecosystem.

Every production Kubernetes deployment depends on Secrets. They are consumed by pods as environment variables or volume mounts, by Ingress controllers for TLS termination, by kubelet for image pulling, and by operators for connecting to external services. Managing them declaratively with proper typing and validation is essential for platform engineering at scale.

Planton's **KubernetesSecret** component brings structure, type safety, and dual-IaC support to this fundamental primitive. Rather than relying on raw YAML with untyped `stringData` maps, Planton provides distinct message types for each Kubernetes secret variant, with schema-level validation that catches errors before they reach the cluster.

## Evolution and Historical Context

### Early Kubernetes (pre-1.7)

In early Kubernetes, secrets were created imperatively with `kubectl create secret` or via raw YAML manifests. There was no type distinction beyond the `type` field -- the burden of structuring data correctly (e.g., putting `tls.crt` and `tls.key` for TLS secrets) fell entirely on the user.

### Typed Secrets (1.7+)

Kubernetes introduced typed secrets with well-known types (`kubernetes.io/tls`, `kubernetes.io/dockerconfigjson`, etc.) and corresponding validation. The API server would reject TLS secrets missing required keys, adding a safety net. However, this validation only happens at the API server -- IaC tools like Terraform and Pulumi don't validate secret structure at plan time.

### Immutable Secrets (1.21+)

Kubernetes 1.21 graduated immutable secrets to stable. Immutable secrets cannot be modified after creation, providing:
- Protection against accidental or malicious modifications
- Performance improvement by eliminating kubelet watches on immutable secrets
- Consistency guarantee for workloads that depend on secret values not changing

### External Secret Management (2020+)

The External Secrets Operator (ESO) and similar projects emerged to bridge external secret stores (AWS Secrets Manager, HashiCorp Vault, GCP Secret Manager) with Kubernetes Secrets. These handle a fundamentally different use case: secrets that _originate_ in external stores and need to be _synced_ into Kubernetes.

Planton's KubernetesSecret component addresses the complementary use case: secrets whose values are known at deployment time and need to be _created directly_ in the cluster.

## Deployment Methods Landscape

### Level 0: Manual (kubectl)

The most common ad-hoc approach:

```bash
# Opaque secret
kubectl create secret generic my-secret \
  --from-literal=username=admin \
  --from-literal=password=s3cret

# TLS secret
kubectl create secret tls my-tls \
  --cert=tls.crt \
  --key=tls.key

# Docker registry secret
kubectl create secret docker-registry my-registry \
  --docker-server=gcr.io \
  --docker-username=_json_key \
  --docker-password="$(cat sa-key.json)"
```

**Pros:**
- Immediate and intuitive
- Built into every Kubernetes installation
- Type-aware (the `create secret tls` subcommand validates inputs)

**Cons:**
- Imperative, not declarative -- no drift detection
- No state tracking -- can't easily audit what's deployed
- Secret values visible in shell history
- No validation before execution
- Difficult to reproduce across environments

**Verdict:** Suitable for debugging and one-off operations. Not acceptable for production infrastructure management.

### Level 1: Declarative YAML Manifests

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: production
type: Opaque
stringData:
  username: admin
  password: s3cret
```

Applied with `kubectl apply -f secret.yaml`.

**Pros:**
- Declarative and reproducible
- Can be version-controlled (with appropriate secret management)
- Kubernetes API server validates typed secrets

**Cons:**
- No plan/preview -- `kubectl apply` is immediate
- Untyped `stringData` maps -- no schema enforcement for type-specific fields
- Secret values in plain text in YAML files
- No state management beyond what Kubernetes tracks
- Manual tracking of what's deployed where

**Verdict:** Better than imperative, but lacks the safety and lifecycle management of IaC tools.

### Level 2: Terraform

```hcl
resource "kubernetes_secret_v1" "example" {
  metadata {
    name      = "my-secret"
    namespace = "production"
  }

  type = "Opaque"

  data = {
    username = "admin"
    password = "s3cret"
  }
}

resource "kubernetes_secret_v1" "tls" {
  metadata {
    name      = "my-tls"
    namespace = "production"
  }

  type = "kubernetes.io/tls"

  data = {
    "tls.crt" = file("tls.crt")
    "tls.key" = file("tls.key")
  }
}
```

**Pros:**
- Full IaC lifecycle (plan, apply, destroy, import)
- State management with drift detection
- Sensitive values marked in state (though not encrypted by default)
- Can reference other Terraform resources

**Cons:**
- Untyped `data` map -- no schema enforcement per secret type
- Must manually set `type` field and match data keys
- Terraform state contains secret values (requires encrypted backend)
- No compile-time validation of data structure

**Verdict:** Production-grade IaC but lacks type safety for secret data.

### Level 3: Pulumi

```go
secret, err := corev1.NewSecret(ctx, "my-secret", &corev1.SecretArgs{
    Metadata: &metav1.ObjectMetaArgs{
        Name:      pulumi.String("my-secret"),
        Namespace: pulumi.String("production"),
    },
    Type: pulumi.String("Opaque"),
    StringData: pulumi.StringMap{
        "username": pulumi.String("admin"),
        "password": pulumi.String("s3cret"),
    },
})
```

**Pros:**
- Full programming language (type safety, IDE support, testing)
- Automatic secret encryption in Pulumi state
- Strong ecosystem integration
- Plan/preview before apply

**Cons:**
- Still uses untyped `StringData` map
- `Type` is a plain string -- no enum enforcement
- Must manually match data keys to type requirements
- Requires Pulumi runtime and SDK

**Verdict:** Excellent IaC choice with better secret handling than Terraform, but still lacks schema-level type safety.

### Other Methods

**Helm:** Uses `templates/secret.yaml` with `{{ .Values }}` substitution. Common in Helm charts but inherits all the problems of raw YAML plus template complexity.

**Kustomize:** Uses `secretGenerator` which can read from files or literals. Decent for simple cases but limited type support.

**External Secrets Operator:** Syncs from external stores. Different use case entirely -- complementary to direct secret creation.

## Comparative Analysis

| Aspect | kubectl | YAML | Terraform | Pulumi | Planton |
|--------|---------|------|-----------|--------|---------|
| Type Safety | Subcommand-level | None | None | None | Per-type messages |
| Validation | At creation | API server | Plan time (basic) | Preview time (basic) | Schema + CEL |
| State Management | None | None | Full | Full | Full (via IaC) |
| Drift Detection | No | No | Yes | Yes | Yes |
| Immutable Support | Flag only | Field only | Field only | Field only | First-class |
| Dual IaC | N/A | N/A | TF only | Pulumi only | Both |
| Reproducible | No | Partially | Yes | Yes | Yes |

## The Planton Approach

### Type-Safe Secret Data via Protobuf `oneof`

The core innovation is modeling each Kubernetes secret type as a distinct protobuf message within a `oneof`:

```protobuf
oneof secret_data {
  KubernetesSecretOpaqueData opaque = 7;
  KubernetesSecretTlsData tls = 8;
  KubernetesSecretDockerConfigJsonData docker_config_json = 9;
  KubernetesSecretBasicAuthData basic_auth = 10;
  KubernetesSecretSshAuthData ssh_auth = 11;
}
```

This design provides:

1. **Mutual exclusion**: Exactly one secret type per resource (enforced by protobuf `oneof`)
2. **Required field enforcement**: `tls_crt` and `tls_key` are both required for TLS; the schema won't accept a TLS secret without both
3. **IDE autocompletion**: Users see the exact fields available for each type
4. **Generated SDK support**: Every language SDK gets strongly-typed classes per variant
5. **No string-matching errors**: Users can't misspell `tls.crt` because they use the `tls_crt` field directly

### 80/20 Scoping

The five supported types cover the vast majority of production use cases:

- **Opaque (~70%)**: Generic secrets, API keys, database credentials, config values
- **TLS (~15%)**: Certificates for Ingress, webhooks, mTLS
- **DockerConfigJson (~10%)**: Private registry authentication
- **BasicAuth (~3%)**: Service authentication
- **SSHAuth (~2%)**: Git operations, SSH tunnels

Excluded types:
- `kubernetes.io/service-account-token`: Auto-managed by Kubernetes (deprecated in 1.24+)
- `bootstrap.kubernetes.io/token`: Cluster bootstrapping (extremely niche)

### Default Handling

The `namespace` field defaults to `"default"` via the `(dev.planton.shared.options.default)` mechanism. This means:
- Users deploying to the default namespace don't need to specify it
- The default is applied by Planton middleware before IaC modules run
- IaC modules never need defensive default logic

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi module (`iac/pulumi/module/`) follows the standard Planton pattern:

- **`main.go`**: Orchestrates resource creation -- creates the Kubernetes provider and the Secret resource
- **`locals.go`**: Computes labels, annotations, secret type, and data map from the `oneof` variant
- **`secret.go`**: Creates `kubernetes.core.v1.Secret` with the computed type and data
- **`outputs.go`**: Exports `secret_name`, `secret_namespace`, and `secret_type`

The `locals.go` contains the critical mapping logic: inspecting which `oneof` variant is set and translating it to the appropriate Kubernetes secret `type` string and `stringData` map.

### Terraform Module Architecture

The Terraform module (`iac/tf/`) mirrors the Pulumi logic:

- **`variables.tf`**: Mirrors `spec.proto` fields as Terraform variables
- **`locals.tf`**: Computes secret type and data map based on which variable block is populated
- **`main.tf`**: Creates `kubernetes_secret_v1` resource
- **`outputs.tf`**: Exports secret name, namespace, and type
- **`provider.tf`**: Kubernetes provider configuration

### Resource Count

This is a lean component -- it creates exactly **one Kubernetes resource**: the Secret itself. There are no sub-resources, no operators, no custom resource definitions. The complexity is in the spec design and type mapping, not in resource orchestration.

## Production Best Practices

### Secret Storage Security

1. **Enable etcd encryption**: Kubernetes secrets are base64-encoded in etcd by default. Enable encryption at rest with `EncryptionConfiguration`
2. **Use encrypted IaC state**: Pulumi encrypts secrets in state by default. For Terraform, use encrypted backends (S3 + KMS, GCS + CMEK)
3. **Avoid version-controlling secret values**: Pass values via CI/CD variables, not committed manifest files
4. **Rotate regularly**: Establish rotation policies for all credentials

### Immutability

1. **Use immutable secrets for production**: Prevents accidental modifications and improves API server performance
2. **Plan for rotation**: Immutable secrets require delete-and-recreate for updates. Design your deployment pipeline accordingly
3. **Version secret names**: Use names like `myapp-db-creds-v2` to enable zero-downtime rotation

### RBAC

1. **Restrict secret access**: Use Kubernetes RBAC to limit which service accounts and users can read secrets
2. **Namespace isolation**: Create secrets in the namespace where they're consumed. Avoid cross-namespace references
3. **Audit access**: Enable audit logging for Secret read operations

### Performance

1. **Use immutable secrets**: Eliminates kubelet watch overhead for secrets that don't change
2. **Limit secret size**: Kubernetes secrets are limited to 1MiB. Keep individual secrets small
3. **Avoid mounting all keys**: Mount only the specific keys your application needs

## Conclusion

KubernetesSecret brings type safety, schema validation, and dual-IaC support to the most fundamental sensitive data primitive in Kubernetes. By modeling each secret type as a distinct protobuf message, it eliminates entire categories of misconfiguration errors that plague raw YAML, Terraform, and Pulumi approaches.

The component is intentionally lean -- one resource, five types, clear boundaries. For secrets originating in external stores, use KubernetesExternalSecrets. For secrets provided at deploy time, KubernetesSecret is the right choice.

## References

- [Kubernetes Secrets Documentation](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Secret Types](https://kubernetes.io/docs/concepts/configuration/secret/#secret-types)
- [Immutable Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#secret-immutable)
- [Encrypting Secret Data at Rest](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/)
- [Good Practices for Kubernetes Secrets](https://kubernetes.io/docs/concepts/security/secrets-good-practices/)
- [External Secrets Operator](https://external-secrets.io/)
- [Pulumi Kubernetes Secret](https://www.pulumi.com/registry/packages/kubernetes/api-docs/core/v1/secret/)
- [Terraform kubernetes_secret_v1](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/secret_v1)
