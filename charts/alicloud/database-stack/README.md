# AliCloud Database Stack

Self-contained database environment with VPC networking, private DNS, and toggleable database engines for Alibaba Cloud.

## Resources

| Resource | Kind | Purpose | Conditional |
|----------|------|---------|-------------|
| VPC | `AliCloudVpc` | Isolated database network | Always |
| VSwitch (×2) | `AliCloudVswitch` | Subnets across two AZs for HA | Always |
| Private DNS Zone | `AliCloudPrivateDnsZone` | Internal DNS resolution for databases | Always |
| RDS Instance | `AliCloudRdsInstance` | Managed relational database (MySQL/PostgreSQL/SQLServer/MariaDB) | `rdsEnabled` |
| PolarDB Cluster | `AliCloudPolardbCluster` | Cloud-native relational database | `polardbEnabled` |
| Redis Instance | `AliCloudRedisInstance` | In-memory cache and data store | `redisEnabled` |
| MongoDB Instance | `AliCloudMongodbInstance` | Document database | `mongodbEnabled` |

## Dependency Graph

```
              ┌─────────┐
              │   VPC   │
              └────┬────┘
       ┌───────────┼───────────┐
       ▼           ▼           ▼
┌──────────┐ ┌──────────┐ ┌─────────────┐
│ VSwitch  │ │ VSwitch  │ │ Private DNS │
│   AZ-1   │ │   AZ-2   │ │    Zone     │
└────┬─────┘ └──────────┘ └─────────────┘
     │
     ├────────────────────────────────────┐
     ▼            ▼          ▼            ▼
┌─────────┐ ┌─────────┐ ┌────────┐ ┌─────────┐
│   RDS   │ │ PolarDB │ │ Redis  │ │ MongoDB │
│  (opt)  │ │  (opt)  │ │ (opt)  │ │  (opt)  │
└─────────┘ └─────────┘ └────────┘ └─────────┘
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Alibaba Cloud region | `cn-hangzhou` |
| `availability_zone_1` | Primary AZ | `cn-hangzhou-h` |
| `availability_zone_2` | Standby AZ | `cn-hangzhou-i` |
| `vpc_cidr` | Database VPC CIDR | `10.1.0.0/16` |
| `vswitch_cidr_1` | VSwitch 1 CIDR | `10.1.0.0/24` |
| `vswitch_cidr_2` | VSwitch 2 CIDR | `10.1.1.0/24` |
| `private_zone_name` | Private DNS zone | `db.internal` |
| `rdsEnabled` | Deploy RDS | `true` |
| `rds_engine` | RDS engine | `PostgreSQL` |
| `rds_engine_version` | RDS version | `16.0` |
| `rds_instance_type` | RDS class | `rds.pg.s2.large` |
| `rds_storage_gb` | RDS storage (GB) | `50` |
| `polardbEnabled` | Deploy PolarDB | `false` |
| `polardb_db_type` | PolarDB engine | `PostgreSQL` |
| `polardb_db_version` | PolarDB version | `14` |
| `polardb_node_class` | PolarDB node class | `polar.pg.x4.large` |
| `redisEnabled` | Deploy Redis | `false` |
| `redis_instance_class` | Redis class | `redis.master.small.default` |
| `redis_engine_version` | Redis version | `7.0` |
| `mongodbEnabled` | Deploy MongoDB | `false` |
| `mongodb_engine_version` | MongoDB version | `7.0` |
| `mongodb_instance_class` | MongoDB class | `dds.mongo.mid` |
| `mongodb_storage_gb` | MongoDB storage (GB) | `20` |

## Deployment Time

| Phase | Duration |
|-------|----------|
| VPC + VSwitches | ~30 seconds |
| Private DNS Zone | ~15 seconds |
| RDS Instance | ~10-15 minutes |
| PolarDB Cluster | ~10-15 minutes |
| Redis Instance | ~5-8 minutes |
| MongoDB Instance | ~8-12 minutes |
| **Total (RDS only)** | **~12 minutes** |
| **Total (all engines)** | **~15 minutes** (parallel) |
