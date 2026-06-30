# Internet-Facing HTTP ALB

This preset creates a public-facing ALB with a single HTTP listener and one server group. This is the quickest way to get an L7 load balancer running.

## When to Use

- Development or staging environments
- Applications that don't require TLS termination at the ALB
- Quick setup for proof-of-concept deployments

## Key Configuration Choices

- **Internet address type** (default) -- public DNS name resolvable from the internet
- **Standard edition** (default) -- supports all standard L7 features
- **Single server group** -- one backend pool with HTTP health checks
- **HTTP listener on port 80** -- forwards all traffic to the server group

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-alb-name>` | ALB name (2-128 chars) | Choose a descriptive name |
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vpc-id>` | VPC ID | `AliCloudVpc` stack outputs |
| `<zone-a>` | First AZ (e.g., `cn-hangzhou-a`) | Region's available zones |
| `<zone-b>` | Second AZ (e.g., `cn-hangzhou-b`) | Region's available zones |
| `<your-vswitch-id-a>` | VSwitch in zone A | `AliCloudVswitch` stack outputs |
| `<your-vswitch-id-b>` | VSwitch in zone B | `AliCloudVswitch` stack outputs |

## Related Presets

- **02-https-production** -- Add HTTPS with certificate and strict TLS
- **03-internal-grpc** -- Internal ALB for service-to-service GRPC
