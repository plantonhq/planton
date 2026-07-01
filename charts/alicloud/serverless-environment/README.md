# AliCloud Serverless Environment

Serverless compute environment with Function Compute (FC), Serverless App Engine (SAE), OSS code storage, centralized logging, and IAM roles.

## Resources

| Resource | Kind | Purpose | Conditional |
|----------|------|---------|-------------|
| VPC | `AliCloudVpc` | Network for VPC-connected workloads | Always |
| VSwitch | `AliCloudVswitch` | Subnet for SAE and VPC FC | Always |
| Security Group | `AliCloudSecurityGroup` | Outbound-only firewall | Always |
| Log Project | `AliCloudLogProject` | Centralized function logging | Always |
| RAM Role | `AliCloudRamRole` | FC execution role | Always |
| OSS Bucket | `AliCloudStorageBucket` | Function code and artifacts | `ossEnabled` |
| FC Function | `AliCloudFunction` | Serverless function | `fcEnabled` |
| SAE Application | `AliCloudSaeApplication` | Container-serverless app | `saeEnabled` |

## Dependency Graph

```
Layer 0:  ┌─────┐  ┌──────────┐  ┌──────────┐
          │ VPC │  │ RAM Role │  │ Log Proj │
          └──┬──┘  └────┬─────┘  └────┬─────┘
             │          │              │
Layer 1:  ┌──┴───────┐  │              │     ┌──────┐
          │ VSwitch  │  │              │     │ OSS  │ (opt)
          └──┬───────┘  │              │     └──────┘
             │          │              │
          ┌──┴──┐       │              │
          │ SG  │       │              │
          └──┬──┘       │              │
             │          │              │
Layer 2:  ┌──┴──────────┴──────────────┴──┐
          │        FC Function (opt)       │
          │  (VPC + Role + Log optional)   │
          └────────────────────────────────┘
          ┌────────────────────────────────┐
          │     SAE Application (opt)      │
          │     (VPC + VSwitch + SG)       │
          └────────────────────────────────┘
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Alibaba Cloud region | `cn-hangzhou` |
| `availability_zone_1` | AZ for VSwitch | `cn-hangzhou-h` |
| `vpc_cidr` | VPC CIDR | `10.2.0.0/16` |
| `vswitch_cidr` | VSwitch CIDR | `10.2.0.0/20` |
| `log_retention_days` | Log retention | `7` |
| `fc_role_policies` | IAM policies for FC role | `AliyunLogFullAccess,AliyunOSSReadOnlyAccess` |
| `ossEnabled` | Create OSS bucket | `true` |
| `fcEnabled` | Deploy FC function | `true` |
| `fc_function_name` | Function name | `hello-function` |
| `fc_runtime` | Runtime | `python3.10` |
| `fc_handler` | Handler | `index.handler` |
| `fc_memory_size` | Memory (MB) | `256` |
| `fc_timeout` | Timeout (s) | `60` |
| `fcVpcEnabled` | Connect FC to VPC | `false` |
| `saeEnabled` | Deploy SAE app | `false` |
| `sae_app_name` | SAE app name | `hello-app` |
| `sae_package_type` | Package type | `Image` |
| `sae_image_url` | Container image | `nginx:latest` |
| `sae_replicas` | Replica count | `1` |
| `sae_cpu` | CPU (millicores) | `1000` |
| `sae_memory` | Memory (MB) | `2048` |

## Deployment Time

| Phase | Duration |
|-------|----------|
| VPC + VSwitch + SG + Log + RAM + OSS | ~1 minute |
| FC Function | ~30 seconds |
| SAE Application | ~3-5 minutes |
| **Total** | **~5 minutes** |
