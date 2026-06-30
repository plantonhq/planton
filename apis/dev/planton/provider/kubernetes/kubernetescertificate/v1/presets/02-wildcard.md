# Wildcard TLS Certificate

This preset creates a wildcard certificate (`*.example.com`) via a ClusterIssuer. Wildcard certificates cover all subdomains under a domain, reducing the number of certificates needed for multi-service deployments.

## When to Use

- You have multiple services on subdomains of the same domain (e.g., `app.example.com`, `api.example.com`, `admin.example.com`)
- You want a single certificate shared across services
- Your ClusterIssuer supports DNS-01 challenges (required for wildcards -- HTTP-01 cannot validate wildcards)

## Key Configuration Choices

- **Two DNS names** -- Both `*.example.com` and `example.com` are included. The wildcard covers subdomains but not the bare domain itself.
- **DNS-01 required** -- Wildcard certificates can only be validated via DNS-01 challenges, not HTTP-01
- **Shared Secret** -- Multiple Ingress/Gateway resources in the same namespace can reference the same TLS Secret

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Namespace for the Certificate and TLS Secret | Your application namespace or a shared namespace |
| `*.example.com` | Wildcard domain for the certificate | Replace `example.com` with your actual domain |
| `example.com` | Bare domain (add alongside wildcard) | Same domain as the wildcard |
| `<your-domain>-wildcard-tls` | Secret name | Convention: domain with `-wildcard-tls` suffix |
| `<your-cluster-issuer>` | ClusterIssuer name | KubernetesClusterIssuer's `cluster_issuer_name` output |

## Related Presets

- **01-cluster-issuer** -- Use for single-hostname certificates
- **03-root-ca-bootstrap** -- Use for internal PKI with self-signed root CA
