# KubernetesClusterIssuer Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials)
   ↓
Locals (cert_manager_namespace, dns_domain, secret name, acme key name)
   ↓
Resources (optional Cloudflare Secret + ClusterIssuer CR)
   ↓
Outputs (cluster_issuer_name, acme_account_key_secret_name)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: Cloudflare Secret (conditional) + ClusterIssuer CR |
| `locals.go` | Extracts and computes values from spec: namespace, domain, secret names |
| `outputs.go` | Defines output constant names matching stack_outputs.proto |

## Key Design Decisions

- **No namespace creation** -- ClusterIssuers are cluster-scoped Kubernetes resources
- **Conditional Secret** -- Cloudflare API token Secret is only created when the Cloudflare provider is configured
- **ClusterIssuer name = dns_domain** -- Preserves the convention that ingress components use to derive issuer names from hostnames
- **solver config is a pure function** -- `buildSolverConfig` maps the proto oneof to the cert-manager ACME solver YAML structure with no side effects
