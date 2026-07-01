# AliCloud Microservices Environment

Self-contained ACK-based microservices platform with ALB ingress, container registry, RocketMQ messaging, Redis caching, NAS shared storage, and centralized logging.

## Resources

| Resource | Kind | Purpose | Conditional |
|----------|------|---------|-------------|
| VPC | `AliCloudVpc` | Network isolation | Always |
| VSwitch (×2) | `AliCloudVswitch` | Multi-AZ subnets | Always |
| Security Group | `AliCloudSecurityGroup` | HTTP/HTTPS firewall | Always |
| EIP | `AliCloudEipAddress` | NAT Gateway public IP | Always |
| NAT Gateway | `AliCloudNatGateway` | Outbound internet for nodes | Always |
| Log Project | `AliCloudLogProject` | Audit and app logging | Always |
| ACR Enterprise | `AliCloudContainerRegistry` | Container registry | `acrEnabled` |
| ACK Cluster | `AliCloudKubernetesCluster` | Managed Kubernetes | Always |
| Node Pool | `AliCloudKubernetesNodePool` | Autoscaled workers | Always |
| ALB | `AliCloudApplicationLoadBalancer` | L7 ingress | Always |
| RocketMQ | `AliCloudRocketmqInstance` | Async messaging | `rocketmqEnabled` |
| Redis | `AliCloudRedisInstance` | In-memory cache | `redisEnabled` |
| NAS | `AliCloudNasFileSystem` | Shared file storage | `nasEnabled` |

## Dependency Graph

```
Layer 0:  ┌─────┐  ┌──────────┐  ┌──────┐
          │ VPC │  │ Log Proj │  │ ACR  │ (opt)
          └──┬──┘  └────┬─────┘  └──────┘
             │          │
Layer 1:  ┌──┴───────┐  │  ┌──────┐  ┌──────┐
          │ VSwitches │  │  │  SG  │  │ EIP  │
          └──┬───────┘  │  └──┬───┘  └──┬───┘
             │          │     │         │
Layer 2:  ┌──┴──────────┴─────┴─────────┴──┐
          │          NAT Gateway            │
          └────────────┬───────────────────-┘
                       │
Layer 3:  ┌────────────┴────────────┐  ┌─────────────┐
          │      ACK Cluster        │  │     ALB     │
          └────────────┬────────────┘  └─────────────┘
                       │
Layer 4:  ┌────────────┴────────────┐
          │       Node Pool         │
          └─────────────────────────┘

Layer 2:  ┌────────────┐  ┌────────┐  ┌──────┐  (parallel, on VSwitch)
(svc)     │  RocketMQ  │  │ Redis  │  │ NAS  │
          │   (opt)    │  │ (opt)  │  │(opt) │
          └────────────┘  └────────┘  └──────┘
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Alibaba Cloud region | `cn-hangzhou` |
| `availability_zone_1/2` | Availability zones | `cn-hangzhou-h/i` |
| `vpc_cidr` | VPC CIDR | `10.0.0.0/16` |
| `vswitch_cidr_1/2` | VSwitch CIDRs | `10.0.0.0/20`, `10.0.16.0/20` |
| `eip_bandwidth` | NAT EIP bandwidth (Mbps) | `10` |
| `cluster_name` | Cluster name suffix | `microservices` |
| `pod_cidr` | Pod CIDR | `172.20.0.0/16` |
| `service_cidr` | Service CIDR | `172.21.0.0/20` |
| `node_instance_types` | Worker types | `ecs.g7.xlarge` |
| `node_desired_size` | Desired workers | `3` |
| `node_min_size` / `node_max_size` | Autoscaling range | `2` / `10` |
| `alb_address_type` | ALB type | `Internet` |
| `log_retention_days` | Log retention | `30` |
| `acrEnabled` | Create ACR | `true` |
| `acr_instance_type` | ACR tier | `Basic` |
| `rocketmqEnabled` | Deploy RocketMQ | `true` |
| `rocketmq_spec` | RocketMQ spec | `rmq.s1.micro` |
| `redisEnabled` | Deploy Redis | `true` |
| `redis_instance_class` | Redis class | `redis.master.small.default` |
| `nasEnabled` | Deploy NAS | `false` |
| `nas_storage_type` | NAS type | `Performance` |

## Deployment Time

| Phase | Duration |
|-------|----------|
| VPC + Log + ACR | ~1 minute |
| VSwitches + SG + EIP | ~30 seconds |
| NAT Gateway | ~2 minutes |
| ACK Cluster + ALB | ~10-12 minutes |
| Node Pool | ~5-8 minutes |
| RocketMQ + Redis + NAS | ~5-8 minutes (parallel) |
| **Total** | **~18-25 minutes** |
