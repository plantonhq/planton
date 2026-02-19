# Production HTTPS ALB

This preset creates a production-grade ALB with HTTPS, WAF integration, strict TLS policy, access logging, and session stickiness.

## When to Use

- Production environments serving external HTTPS traffic
- Applications requiring WAF protection at the load balancer level
- APIs that need session affinity for stateful interactions

## Key Configuration Choices

- **StandardWithWaf edition** -- integrated Web Application Firewall
- **HTTPS listener on port 443** -- TLS termination at the ALB
- **TLS 1.2 strict policy** -- disables TLS 1.0 and 1.1, removes weak ciphers
- **Access logging** -- ships all request logs to SLS for observability
- **Sticky sessions** -- Insert-mode cookie with 1-hour timeout
- **Enhanced health checks** -- HTTPS probes with GET method and tuned thresholds

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-alb-name>` | ALB name (2-128 chars) | Choose a descriptive name |
| `<your-org>` | Organization identifier | Your org name |
| `<alibaba-cloud-region>` | Region code | Your deployment region |
| `<your-vpc-resource-name>` | AlicloudVpc resource name | Your VPC manifest |
| `<zone-a>`, `<zone-b>` | Availability zones | Region's available zones |
| `<your-vswitch-resource-a>`, `<your-vswitch-resource-b>` | AlicloudVswitch resource names | Your VSwitch manifests |
| `<your-sls-project>` | SLS log project name | `AlicloudLogProject` stack outputs |
| `<your-sls-logstore>` | SLS log store name | `AlicloudLogProject` stack outputs |
| `<your-certificate-id>` | CAS certificate ID | Alibaba Cloud Certificate Management console |
| `<your-team>`, `<your-cost-center>` | Tag values | Your organization's tagging policy |

## Related Presets

- **01-internet-http** -- Simpler setup without HTTPS
- **03-internal-grpc** -- Internal ALB for GRPC services
