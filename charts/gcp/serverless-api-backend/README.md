# GCP Serverless API Backend

Provisions a production-ready serverless API backend on Cloud Run with private VPC networking, Cloud SQL database, and optional Redis cache, Pub/Sub messaging, Cloud Tasks async processing, and Secret Manager. This is a composable microservices infrastructure pattern -- enable only the components your API needs.

This chart differs from the [Cloud Run Environment](../cloud-run-environment/) by focusing on backend API infrastructure patterns: caching layers, async task queues, event-driven messaging, and secrets management. The Cloud Run Environment is oriented toward frontend/backend web applications with DNS and Docker repositories.

## Architecture

```
                    ┌──────────────────────────────────────────┐
                    │  Network                                  │
                    │                                          │
                    │  ┌──────────┐  ┌────────────┐           │
                    │  │  GcpVpc  │─▶│ GcpSubnet  │           │
                    │  │  (PSA)   │  └─────┬──────┘           │
                    │  └────┬─────┘        │                  │
                    │       │        ┌─────┴──────┐           │
                    │       │        │GcpRouterNat│           │
                    │       │        └────────────┘           │
                    └───────│──────────────────────────────────┘
                            │
              ┌─────────────┼─────────────────────────┐
              ▼             ▼                         ▼
     ┌──────────────┐ ┌──────────┐          ┌──────────────┐
     │ GcpCloudSql  │ │GcpRedis  │          │ GcpCloudRun  │
     │ (PostgreSQL) │ │Instance  │          │ (API service)│
     │              │ │ (cache)  │          │              │
     └──────────────┘ └──────────┘          └──────┬───────┘
                                                   │ uses
                                    ┌──────────────┼──────────────┐
                                    ▼              ▼              ▼
                            ┌────────────┐ ┌────────────┐ ┌─────────────┐
                            │GcpPubSub   │ │GcpCloud    │ │GcpSecrets   │
                            │  Topic     │ │  Tasks     │ │  Manager    │
                            │(messaging) │ │  Queue     │ │             │
                            └────────────┘ └────────────┘ └─────────────┘

                            ┌────────────────────────────┐
                            │   GcpServiceAccount        │
                            │   (Cloud Run identity)     │
                            └────────────────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  GcpVpc, GcpServiceAccount, GcpPubSubTopic, GcpCloudTasksQueue, GcpSecretsManager
Layer 1 (dep VPC):   GcpSubnetwork, GcpRouterNat, GcpCloudSql, GcpRedisInstance
Layer 2 (dep all):   GcpCloudRun
```

## Included Cloud Resources

| Resource | Kind | Group | Condition | Purpose |
|----------|------|-------|-----------|---------|
| VPC Network | `GcpVpc` | network | Always | Private networking with Private Services Access |
| Subnetwork | `GcpSubnetwork` | network | Always | Subnet with Private Google Access |
| Router NAT | `GcpRouterNat` | network | Always | Outbound internet for Cloud Run |
| Service Account | `GcpServiceAccount` | identity | Always | Cloud Run service identity |
| Cloud Run | `GcpCloudRun` | compute | Always | The API service |
| Cloud SQL | `GcpCloudSql` | database | `databaseEnabled` | PostgreSQL database |
| Redis | `GcpRedisInstance` | cache | `cacheEnabled` | In-memory cache |
| Pub/Sub Topic | `GcpPubSubTopic` | messaging | `messagingEnabled` | Event-driven messaging |
| Cloud Tasks Queue | `GcpCloudTasksQueue` | async | `tasksEnabled` | Async task processing |
| Secret Manager | `GcpSecretsManager` | secrets | `secretsEnabled` | Secret storage |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `gcp_project_id` | GCP project ID | `my-gcp-project` | Yes |
| `region` | GCP region | `us-central1` | Yes |
| `vpc_name` | VPC name | `api-backend-vpc` | Yes |
| `subnet_cidr` | Subnet CIDR | `10.0.0.0/24` | Yes |
| `service_account_id` | Service account ID | `api-backend-sa` | Yes |
| `service_name` | Cloud Run service name | `api-backend` | Yes |
| `container_image` | Container image | `us-docker.pkg.dev/cloudrun/container/hello` | Yes |
| `container_port` | Container port | `8080` | Yes |
| **Database** | | | |
| `databaseEnabled` | Create Cloud SQL | `true` | No |
| `database_instance_name` | Instance name | `api-database` | No |
| `database_tier` | Machine tier | `db-f1-micro` | No |
| `database_version` | DB version | `POSTGRES_15` | No |
| `database_root_password` | Root password | `change-me-immediately` | No |
| **Cache** | | | |
| `cacheEnabled` | Create Redis | `false` | No |
| `redis_instance_name` | Redis name | `api-cache` | No |
| `redis_memory_size_gb` | Memory in GiB | `1` | No |
| **Messaging** | | | |
| `messagingEnabled` | Create Pub/Sub topic | `false` | No |
| `pubsub_topic_name` | Topic name | `api-events` | No |
| **Tasks** | | | |
| `tasksEnabled` | Create Cloud Tasks queue | `false` | No |
| `tasks_queue_name` | Queue name | `api-tasks` | No |
| **Secrets** | | | |
| `secretsEnabled` | Create Secret Manager secrets | `false` | No |
| `secret_names` | Comma-separated secret names | `db-password,api-key` | No |

## Common Configurations

### Minimal (Cloud Run + Database)

```yaml
databaseEnabled: true
cacheEnabled: false
messagingEnabled: false
tasksEnabled: false
secretsEnabled: false
```

### Full Microservices Stack

```yaml
databaseEnabled: true
cacheEnabled: true
messagingEnabled: true
tasksEnabled: true
secretsEnabled: true
```

### API with Cache and Async Tasks

```yaml
databaseEnabled: true
cacheEnabled: true
tasksEnabled: true
messagingEnabled: false
secretsEnabled: false
```

## Networking

The VPC is configured with:
- **Private Services Access (PSA)**: Enables private IP connectivity to Cloud SQL and Redis (no public IPs)
- **Private Google Access**: Cloud Run can access Google APIs without egress through the public internet
- **Cloud NAT**: Provides outbound internet for Cloud Run when calling external services

## Service Account Roles

The service account dynamically receives roles based on enabled components:

| Role | Condition |
|------|-----------|
| `roles/run.invoker` | Always |
| `roles/cloudsql.client` | Always (Cloud SQL proxy) |
| `roles/redis.editor` | When `cacheEnabled` |
| `roles/pubsub.publisher` | When `messagingEnabled` |
| `roles/cloudtasks.enqueuer` | When `tasksEnabled` |
| `roles/secretmanager.secretAccessor` | When `secretsEnabled` |

## Important Notes

- The `container_image` defaults to Google's hello-world image. **Replace this** with your actual API image after initial deployment.
- **Rotate the `database_root_password`** immediately after deployment. The default value is a placeholder.
- Cloud SQL uses **private IP** only (no public access). Connect from Cloud Run via the Cloud SQL Auth Proxy or direct VPC connection.
- Redis uses **BASIC** tier (no HA). For production, modify the Redis resource to use `STANDARD_HA` after deployment.
- Secret Manager creates **empty secrets**. Populate secret values through the GCP console or `gcloud secrets versions add` after deployment.
