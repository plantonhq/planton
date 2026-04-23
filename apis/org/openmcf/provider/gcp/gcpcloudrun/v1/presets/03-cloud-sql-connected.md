# Cloud Run with Cloud SQL Native Connection

This preset deploys a Cloud Run service connected to a Cloud SQL instance via the native GCP-managed volume mount. GCP automatically creates a Unix socket at `/cloudsql/<connection_name>` for each configured instance. No VPC, Private Services Access, or sidecar container is needed.

## When to Use

- Backend services that need PostgreSQL or MySQL database access
- Any Cloud Run service connecting to Cloud SQL (this is the recommended approach)
- Services where you want the simplest, most reliable Cloud SQL connectivity

## Key Configuration Choices

- **Native Cloud SQL connection** (`cloudSql.connection`) -- GCP manages the proxy internally; creates a Unix socket at `/cloudsql/<connection_name>`
- **No VPC required** -- the connection uses Google's internal control plane, not VPC peering
- **IAM-authenticated and TLS-encrypted** -- handled automatically by GCP
- **Always-warm** (`replicas.min: 1`) -- avoids cold starts for database-backed services
- **1 GiB memory** -- more memory for services doing database queries and result processing
- **Gen 2 execution** (`executionEnvironment: EXECUTION_ENVIRONMENT_GEN2`) -- full Linux compatibility

## DATABASE_URL Format

The application must use the Unix socket path in its connection string:

```
postgresql://user:password@localhost/dbname?host=/cloudsql/<project>:<region>:<instance>
```

For MySQL:
```
mysql://user:password@localhost/dbname?socket=/cloudsql/<project>:<region>:<instance>
```

## Alternatives

- **Auth Proxy sidecar** (`cloudSql.authProxy`) -- use when your application or connection pooler requires TCP connectivity (`localhost:<port>`) instead of Unix sockets. Adds a sidecar container with its own CPU/memory allocation.
- **Direct VPC Egress** (`vpcAccess`) -- use when you need to reach Cloud SQL via its private IP through VPC peering. Requires VPC, subnet, and Private Services Access to be configured.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |
| `<container-image-repo>` | Container image repository | `GcpArtifactRegistryRepo` outputs or container registry |
| `<image-tag>` | Image tag | Your CI/CD pipeline |
| `<cloud-sql-instance-name>` | Cloud SQL instance name | `GcpCloudSql` metadata name |

## Related Presets

- **01-public-service** -- Use for public services without database connectivity
- **02-private-vpc-connected** -- Use for internal services that need VPC access to private resources beyond Cloud SQL
