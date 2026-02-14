# TLS Secret

This preset creates a Kubernetes TLS secret containing a certificate and private key pair. Used by ingress controllers, service meshes, and any workload that needs TLS termination.

## When to Use

- Ingress TLS termination with a manually managed certificate
- Application-level TLS where the app reads the cert/key from a mounted secret
- Environments where cert-manager is not available and certificates are managed externally

## Key Configuration Choices

- **TLS type** -- Kubernetes-native `kubernetes.io/tls` secret type; automatically validated by ingress controllers

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the secret | Your namespace management |
| `<your-tls-certificate-pem>` | PEM-encoded TLS certificate (including intermediate chain) | Your certificate authority or cert management tool |
| `<your-tls-private-key-pem>` | PEM-encoded TLS private key | Generated alongside the certificate |

## Related Presets

- **01-opaque** -- Generic key-value secret for credentials and API keys
- **03-docker-registry** -- Docker registry authentication credentials
