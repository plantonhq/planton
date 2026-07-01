# AliCloud Web App Environment

Traditional web application stack with Application Load Balancer, ECS compute, RDS database, and optional Redis cache, OSS storage, and CDN acceleration.

## Resources

| Resource | Kind | Purpose | Conditional |
|----------|------|---------|-------------|
| VPC | `AliCloudVpc` | Network isolation | Always |
| VSwitch (×2) | `AliCloudVswitch` | Multi-AZ subnets | Always |
| Security Group | `AliCloudSecurityGroup` | HTTP/HTTPS firewall | Always |
| ALB | `AliCloudApplicationLoadBalancer` | L7 load balancing | Always |
| ECS Instance | `AliCloudEcsInstance` | Application compute | Always |
| RDS Instance | `AliCloudRdsInstance` | Relational database | Always |
| Redis | `AliCloudRedisInstance` | In-memory cache | `redisEnabled` |
| OSS Bucket | `AliCloudStorageBucket` | Static asset storage | `ossEnabled` |
| CDN Domain | `AliCloudCdnDomain` | Content acceleration | `cdnEnabled` |

## Dependency Graph

```
               ┌─────────┐
               │   VPC   │
               └────┬────┘
        ┌───────────┼───────────┐
        ▼           ▼           ▼
  ┌──────────┐ ┌──────────┐ ┌────────┐
  │ VSwitch  │ │ VSwitch  │ │   SG   │
  │   AZ-1   │ │   AZ-2   │ │        │
  └────┬─────┘ └────┬─────┘ └───┬────┘
       │            │            │
  ┌────┴────────────┴────────────┴────┐
  │        Application Load Balancer  │
  └───────────────────────────────────┘
       │            │
  ┌────┴───┐   ┌────┴───┐   ┌───────┐   ┌──────┐
  │  ECS   │   │  RDS   │   │ Redis │   │ OSS  │
  │        │   │        │   │ (opt) │   │(opt) │
  └────────┘   └────────┘   └───────┘   └──┬───┘
                                            │
                                       ┌────┴───┐
                                       │  CDN   │
                                       │ (opt)  │
                                       └────────┘
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Alibaba Cloud region | `cn-hangzhou` |
| `availability_zone_1/2` | Availability zones | `cn-hangzhou-h/i` |
| `vpc_cidr` | VPC CIDR | `10.0.0.0/16` |
| `vswitch_cidr_1/2` | VSwitch CIDRs | `10.0.0.0/20`, `10.0.16.0/20` |
| `alb_address_type` | ALB type (Internet/Intranet) | `Internet` |
| `ecs_instance_type` | ECS class | `ecs.g7.large` |
| `ecs_image_id` | ECS AMI | Alibaba Linux 3 |
| `ecs_system_disk_size` | System disk (GB) | `40` |
| `rds_engine` | Database engine | `PostgreSQL` |
| `rds_engine_version` | Engine version | `16.0` |
| `rds_instance_type` | RDS class | `rds.pg.s2.large` |
| `rds_storage_gb` | RDS storage (GB) | `50` |
| `redisEnabled` | Deploy Redis | `false` |
| `redis_instance_class` | Redis class | `redis.master.small.default` |
| `ossEnabled` | Create OSS bucket | `false` |
| `cdnEnabled` | Create CDN domain | `false` |
| `cdn_domain_name` | CDN domain | `cdn.example.com` |
| `cdn_origin_domain` | CDN origin | `origin.example.com` |

## Deployment Time

| Phase | Duration |
|-------|----------|
| VPC + VSwitches + SG | ~30 seconds |
| ALB + ECS + RDS | ~12-15 minutes |
| Redis (if enabled) | ~5-8 minutes |
| OSS + CDN (if enabled) | ~2 minutes |
| **Total (core)** | **~15 minutes** |
