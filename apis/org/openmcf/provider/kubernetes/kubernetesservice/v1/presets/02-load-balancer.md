# LoadBalancer Service

This preset creates a LoadBalancer service that provisions a cloud load balancer with a public IP. Suitable for services that need direct external access without an ingress controller.

## When to Use

- Services that need a dedicated external IP (e.g., gRPC services, non-HTTP protocols)
- When an ingress controller is not available or not appropriate
- TCP/UDP services that cannot be routed through HTTP ingress

## Key Configuration Choices

- **LoadBalancer type** -- provisions a cloud-provider load balancer (NLB on AWS, L4 LB on GCP, Standard LB on Azure)
- **Dual ports** (80 and 443) -- serves both HTTP and HTTPS; adjust based on your protocol needs
- **Label selector** -- routes traffic to pods matching the `app` label

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the service | Your namespace management |
| `<your-app-label>` | Label value matching your deployment's pods | Your deployment manifest's `metadata.labels.app` |

## Related Presets

- **01-cluster-ip** -- Internal-only service without a load balancer
- **03-headless** -- Headless service for StatefulSet DNS resolution
