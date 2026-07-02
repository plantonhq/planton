# AWS Microservices Backend

Provisions a full backend environment with VPC, ALB, ECS (Fargate), Aurora PostgreSQL, ElastiCache Redis, and SQS. This is a composable microservices platform -- enable only the data, cache, and messaging layers your service needs.

This chart extends the [ECS Environment](../ecs-environment/) pattern with database, cache, and messaging tiers, giving you a production-ready backend in a single deployment.

## Architecture

```
                         Internet
                            │
                            ▼
                   ┌─────────────────┐
                   │  AwsRoute53Zone │──── DNS records
                   └────────┬────────┘
                            │
                   ┌────────▼────────┐     ┌───────────────────┐
                   │     AwsAlb      │◄────│ AwsCertManagerCert│
                   │ (load balancer) │     │   (TLS cert)      │
                   └────────┬────────┘     └───────────────────┘
                            │
               ┌────────────┼────────────┐
               ▼            ▼            ▼
     ┌──────────────┐ ┌──────────┐ ┌────────────────┐
     │ AwsEcsCluster│ │AwsEcrRepo│ │AwsCloudwatch   │
     │  (Fargate)   │ │ (images) │ │  LogGroup      │
     └──────┬───────┘ └──────────┘ └────────────────┘
            │
   ┌────────┼────────────────┐
   ▼        ▼                ▼
┌────────┐ ┌──────────────┐ ┌──────────────┐
│AwsRds  │ │AwsRedis      │ │AwsSqsQueue   │
│Cluster │ │Elasticache   │ │ (messaging)  │
│(Aurora)│ │ (cache)      │ └──────────────┘
└────────┘ └──────────────┘

┌──────────────────────┐  ┌───────────────────┐
│ AwsIamRole           │  │ AwsSecurityGroup  │
│ (task execution)     │  │ (HTTP/S ingress)  │
└──────────────────────┘  └───────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  AwsVpc, AwsIamRole, AwsEcsCluster, AwsEcrRepo,
                     AwsCloudwatchLogGroup, AwsSqsQueue
Layer 1 (dep VPC):   AwsSecurityGroup
Layer 2 (dep SG):    AwsAlb, AwsRdsCluster, AwsRedisElasticache
Layer 3 (dep DNS):   AwsRoute53Zone
Layer 4 (dep Zone):  AwsCertManagerCert
```

## Included Cloud Resources

| Resource | Kind | Group | Condition | Purpose |
|----------|------|-------|-----------|---------|
| VPC | `AwsVpc` | network | Always | Isolated network with public/private subnets and NAT |
| Security Group | `AwsSecurityGroup` | network | Always | HTTP/HTTPS ingress, all egress |
| Route 53 Zone | `AwsRoute53Zone` | network | `dnsEnabled` | DNS hosted zone |
| ACM Certificate | `AwsCertManagerCert` | security | `httpsEnabled` | DNS-validated TLS certificate |
| ALB | `AwsAlb` | network | Always | Application load balancer with optional DNS and TLS |
| ECS Cluster | `AwsEcsCluster` | compute | Always | Fargate + Fargate Spot capacity |
| ECR Repository | `AwsEcrRepo` | compute | Always | Container image registry |
| IAM Role | `AwsIamRole` | identity | Always | ECS task execution role |
| CloudWatch Log Group | `AwsCloudwatchLogGroup` | monitoring | Always | Centralized logging (30-day retention) |
| RDS Cluster | `AwsRdsCluster` | database | `databaseEnabled` | Aurora PostgreSQL with encryption |
| ElastiCache Redis | `AwsRedisElasticache` | cache | `cacheEnabled` | Redis with 2-node replication |
| SQS Queue | `AwsSqsQueue` | messaging | `messagingEnabled` | Standard queue for async tasks |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| **Network** | | | |
| `availability_zone_1` | First AZ for subnet pair | `us-east-1a` | Yes |
| `availability_zone_2` | Second AZ for subnet pair | `us-east-1b` | Yes |
| `domain_name` | Route 53 zone domain | `example.com` | Yes |
| `load_balancer_domain_name` | DNS name for the ALB | `app.example.com` | Yes |
| `dnsEnabled` | Create Route53 zone and ALB DNS | `true` | No |
| `httpsEnabled` | Create ACM cert and terminate TLS | `true` | No |
| `alb_idle_timeout_seconds` | ALB idle timeout | `60` | No |
| **Compute** | | | |
| `service_name` | Prefix for ECS cluster, ALB, and related resources | `my-service` | Yes |
| `ecr_repo_name` | ECR repository name | `my-service` | Yes |
| **Database** | | | |
| `databaseEnabled` | Create Aurora PostgreSQL cluster | `true` | No |
| `db_engine_version` | Aurora PostgreSQL version | `15.4` | No |
| `db_name` | Initial database name | `appdb` | No |
| **Cache** | | | |
| `cacheEnabled` | Create ElastiCache Redis cluster | `false` | No |
| `redis_engine_version` | Redis version | `7.1` | No |
| `redis_node_type` | ElastiCache node type | `cache.t3.micro` | No |
| **Messaging** | | | |
| `messagingEnabled` | Create SQS queue | `false` | No |
| `sqs_queue_name` | SQS queue name | `task-queue` | No |

## Common Configurations

### Minimal (VPC + ALB + ECS + RDS)

```yaml
databaseEnabled: true
cacheEnabled: false
messagingEnabled: false
```

### Full Stack (all layers)

```yaml
databaseEnabled: true
cacheEnabled: true
messagingEnabled: true
```

### HTTP-only (no TLS, no DNS)

```yaml
dnsEnabled: false
httpsEnabled: false
```

## Important Notes

- When `httpsEnabled: true`, ACM issues a DNS-validated certificate -- Route 53 records must be publicly resolvable during validation.
- The RDS cluster uses `manageMasterUserPassword: true` -- AWS Secrets Manager stores and rotates the password automatically.
- ElastiCache Redis is deployed with 2 cache clusters and automatic failover for high availability. Set `cacheEnabled: false` for dev environments to save cost.
- SQS uses Standard queue type (at-least-once delivery). Switch to FIFO by editing the template if exactly-once processing is required.
- All cross-resource references are wired with `valueFrom`; you rarely need to touch the templates.

---

© Planton. Licensed under [Apache-2.0](../../../LICENSE).
