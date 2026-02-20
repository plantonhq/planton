# VPC-Connected API Function

This preset creates a Node.js 20 function configured as a backend API handler with access to VPC-internal resources (databases, caches, internal services). The function runs inside a VPC with security group isolation, assumes a RAM execution role for downstream API access, and has full SLS observability including instance-level and request-level metrics. Compute sizing is tuned for API workloads with moderate concurrency.

## When to Use

- API backends that need to query databases, caches, or other services inside a VPC
- Functions behind API Gateway or HTTP triggers serving synchronous request-response traffic
- Workloads that need a RAM execution role to access Alibaba Cloud services (OSS, RDS, Redis)
- Production functions requiring full observability with both instance and request metrics

## Key Configuration Choices

- **Node.js 20** (`runtime: nodejs20`) -- Latest LTS Node.js runtime with excellent async I/O performance for API workloads. Change to `python3.12` or `java11` based on your team's language preference.
- **API-tuned compute** (`cpu: 1.0`, `memorySize: 2048`, `timeout: 30`) -- 1 vCPU and 2 GB RAM handles typical API payloads with database queries. The 30-second timeout is appropriate for synchronous HTTP APIs; increase for long-running operations.
- **Instance concurrency 10** (`instanceConcurrency: 10`) -- Each function instance handles up to 10 concurrent requests, reducing cold starts and improving throughput. Requires the function code to be safe for concurrent execution (no shared mutable state).
- **VPC networking** (`vpcConfig`) -- The function's ENIs are placed in the specified VSwitches with the given security group. Two VSwitches across AZs provide resilience. The function can reach any private endpoint within the VPC (RDS, Redis, NAS, internal ALBs).
- **Internet access enabled** (`internetAccess: true`) -- Allows the function to call external APIs even when running inside a VPC. Requires a NAT gateway in the VPC for outbound traffic.
- **Execution role** (`role`) -- The RAM role the function assumes at runtime. The role must trust the FC service principal (`fc.aliyuncs.com`) and have policies granting access to downstream resources.
- **Full logging** (`logConfig`) -- DefaultRegex log parsing, instance metrics (CPU/memory per instance), and request metrics (latency/status per invocation). Both metric types are essential for production API monitoring.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-function-name>` | Function name (1-128 chars) | Your naming convention |
| `<your-code-bucket>` | OSS bucket for function code | `AlicloudStorageBucket` stack outputs |
| `<your-code-object-key>` | Code ZIP object key (e.g., `functions/api-v2.zip`) | Your CI/CD pipeline |
| `<your-execution-role-arn>` | RAM role ARN (e.g., `acs:ram::123456:role/fc-api-role`) | `AlicloudRamRole` stack outputs |
| `<your-vpc-id>` | VPC ID | `AlicloudVpc` stack outputs |
| `<vswitch-id-zone-a>` | VSwitch in first AZ | `AlicloudVswitch` stack outputs |
| `<vswitch-id-zone-b>` | VSwitch in second AZ | `AlicloudVswitch` stack outputs |
| `<your-security-group-id>` | Security group for function ENIs | `AlicloudSecurityGroup` stack outputs |
| `<your-log-project-name>` | SLS project for function logs | `AlicloudLogProject` stack outputs |
| `<your-logstore-name>` | SLS logstore | Your SLS project configuration |
| `<your-database-endpoint>` | Database connection endpoint | `AlicloudRdsInstance` or `AlicloudRedisInstance` stack outputs |
| `<your-team>` | Team or business unit | Your organizational structure |
| `<your-service-name>` | Logical service name for tagging | Your service catalog |

## Related Presets

- **01-event-handler** -- Use for simpler event-driven functions that do not need VPC access
- **03-custom-container** -- Use when the function requires a custom Docker image with specialized dependencies
