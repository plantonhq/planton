# AliCloud ACK Environment

Production-ready ACK Managed Kubernetes cluster with multi-AZ networking, NAT gateway, centralized logging, IAM roles, optional container registry, and autoscaled node pool.

## Resources

| Resource | Kind | Purpose | Conditional |
|----------|------|---------|-------------|
| VPC | `AliCloudVpc` | Network isolation | Always |
| VSwitch (×2) | `AliCloudVswitch` | Multi-AZ subnets | Always |
| Security Group | `AliCloudSecurityGroup` | Cluster firewall | Always |
| EIP | `AliCloudEipAddress` | Public IP for NAT | Always |
| NAT Gateway | `AliCloudNatGateway` | Outbound internet for nodes | Always |
| RAM Role | `AliCloudRamRole` | Worker node IAM permissions | Always |
| Log Project | `AliCloudLogProject` | Audit and event logging | Always |
| ACR Enterprise | `AliCloudContainerRegistry` | Container image registry | `acrEnabled` |
| ACK Cluster | `AliCloudKubernetesCluster` | Managed Kubernetes | Always |
| Node Pool | `AliCloudKubernetesNodePool` | Autoscaled worker nodes | Always |

## Dependency Graph

```
Layer 0:  ┌─────┐  ┌──────────┐  ┌──────────┐  ┌──────┐
          │ VPC │  │ RAM Role │  │ Log Proj │  │ ACR  │ (opt)
          └──┬──┘  └──────────┘  └─────┬────┘  └──────┘
             │                         │
Layer 1:  ┌──┴───────┐ ┌──────┐ ┌─────┘ ┌──────┐
          │ VSwitches │ │  SG  │ │       │ EIP  │
          └──┬───────┘ └──┬───┘ │       └──┬───┘
             │            │     │          │
Layer 2:  ┌──┴────────────┴─────┴──────────┴──┐
          │           NAT Gateway              │
          └───────────────────────────────────-┘
                         │
Layer 3:  ┌──────────────┴──────────────┐
          │      ACK Cluster            │
          │  (VSwitches + SG + Logging) │
          └──────────────┬──────────────┘
                         │
Layer 4:  ┌──────────────┴──────────────┐
          │        Node Pool            │
          │   (Cluster + VSwitches)     │
          └─────────────────────────────┘
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Alibaba Cloud region | `cn-hangzhou` |
| `availability_zone_1` | First AZ | `cn-hangzhou-h` |
| `availability_zone_2` | Second AZ | `cn-hangzhou-i` |
| `vpc_cidr` | VPC CIDR | `10.0.0.0/16` |
| `vswitch_cidr_1` | VSwitch 1 CIDR | `10.0.0.0/20` |
| `vswitch_cidr_2` | VSwitch 2 CIDR | `10.0.16.0/20` |
| `eip_bandwidth` | NAT EIP bandwidth (Mbps) | `10` |
| `cluster_name` | Cluster name suffix | `ack` |
| `kubernetes_version` | K8s version (empty = latest) | ` ` |
| `pod_cidr` | Pod network CIDR | `172.20.0.0/16` |
| `service_cidr` | Service network CIDR | `172.21.0.0/20` |
| `slb_internet_enabled` | Public API server | `true` |
| `enable_rrsa` | Pod-level IAM via OIDC | `true` |
| `node_instance_types` | Worker instance types (comma-separated) | `ecs.g7.xlarge` |
| `node_desired_size` | Desired worker count | `3` |
| `node_min_size` | Min workers (autoscale) | `2` |
| `node_max_size` | Max workers (autoscale) | `10` |
| `node_system_disk_size` | System disk (GB) | `120` |
| `log_retention_days` | Log retention | `30` |
| `acrEnabled` | Create ACR registry | `false` |
| `acr_instance_type` | ACR tier | `Basic` |

## Deployment Time

| Phase | Duration |
|-------|----------|
| VPC + RAM + Log Project + ACR | ~1 minute |
| VSwitches + SG + EIP | ~30 seconds |
| NAT Gateway | ~2 minutes |
| ACK Cluster | ~8-12 minutes |
| Node Pool | ~5-8 minutes |
| **Total** | **~15-20 minutes** |
