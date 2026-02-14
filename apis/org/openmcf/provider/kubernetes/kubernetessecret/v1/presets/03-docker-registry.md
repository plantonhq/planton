# Docker Registry Secret

This preset creates a Docker registry authentication secret (`kubernetes.io/dockerconfigjson`) for pulling images from private container registries. Referenced by pods via `imagePullSecrets`.

## When to Use

- Pulling container images from private registries (Docker Hub, GCR, ECR, ACR, GHCR, Harbor)
- Kubernetes nodes that do not have pre-configured registry credentials
- Multi-registry environments where different namespaces need different credentials

## Key Configuration Choices

- **Docker config JSON type** -- Kubernetes-native `kubernetes.io/dockerconfigjson` type; automatically used by kubelet for image pulls

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the secret | Your namespace management |
| `<your-registry-server>` | Registry server URL (e.g., `https://index.docker.io/v1/`, `ghcr.io`, `123456789.dkr.ecr.us-east-1.amazonaws.com`) | Your container registry settings |
| `<your-registry-username>` | Registry username or access key | Your registry's credential management |
| `<your-registry-password>` | Registry password or access token | Your registry's credential management |

## Related Presets

- **01-opaque** -- Generic key-value secret for credentials and API keys
- **02-tls** -- TLS certificate and key pair
