# KubernetesIssuer Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials)
   ↓
Locals (namespace, issuer_name, labels, issuer_type flags, ca_secret_name)
   ↓
Resources (namespace-scoped Issuer CR: CA or SelfSigned)
   ↓
Outputs (namespace, issuer_name)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: single Issuer CR (CA or SelfSigned) |
| `locals.go` | Extracts and computes values from spec: namespace, name, labels, issuer type detection |
| `outputs.go` | Defines output constant names matching stack_outputs.proto |

## Key Design Decisions

- **Namespace-scoped** -- Unlike ClusterIssuer, an Issuer lives in a specific namespace; Certificate resources must be in the same namespace
- **No namespace creation** -- The target namespace must already exist on the cluster
- **CA vs SelfSigned only** -- ACME issuance is handled by KubernetesClusterIssuer; keeping Issuer simple covers the 80/20 use case
- **Issuer name = metadata.name** -- Unlike ClusterIssuer (named after dns_domain), the Issuer uses the resource's metadata.name directly
