---
title: "How to Deploy PostgreSQL on Google Cloud SQL"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "gcp"
  - "postgresql"
  - "cloud-sql"
  - "database"
  - "infra-catalog"
category: "gcp"
excerpt: "Deploy a production-ready PostgreSQL database on Google Cloud SQL using a single YAML manifest and the Planton CLI."
---

# How to Deploy PostgreSQL on Google Cloud SQL

This tutorial walks you through deploying a fully managed PostgreSQL database on Google Cloud SQL through Planton. You will write a YAML manifest describing the database you want, deploy it with a single CLI command, and verify the outputs you need to connect your applications. By the end, you will have a running Cloud SQL instance with automated backups, high availability, and private networking -- or a lightweight development instance, depending on your needs.

> **Note**: The Planton web console provides a guided creation wizard for Cloud SQL
> and other Infra Components. This tutorial uses the CLI/YAML approach for stability
> and reproducibility. The console UI evolves frequently — always check it for the
> latest experience.

## What You Will Learn

- What Infra Components are and how they relate to the Infra Catalog
- How to write a `GcpCloudSql` manifest that deploys a PostgreSQL instance
- How to deploy infrastructure with `planton apply` and monitor progress in real time
- How to retrieve deployment outputs (connection name, IP addresses) for application use
- How production and development configurations differ and when to use each

## Prerequisites

- [ ] A GCP provider connection configured and set as the default for your target environment (see [How to Connect Your GCP Project to Planton](/tutorials/how-to-connect-your-gcp-project-to-planton))
- [ ] A Planton organization and at least one environment created
- [ ] The `planton` CLI installed and authenticated (`planton auth login`)

If you plan to use private IP networking (recommended for production), your GCP VPC must have [Private Services Access](https://cloud.google.com/vpc/docs/configure-private-services-access) configured for `servicenetworking.googleapis.com`. Without this, private IP assignment will fail during deployment.

## What Is an Infra Component?

An Infra Component is a Planton API resource that represents a piece of cloud infrastructure -- a database, a cluster, a storage bucket, or any of the 270+ resource types in the Infra Catalog. You define it in a YAML manifest, apply it with `planton apply`, and Planton provisions it in your cloud account using your provider connection. For more on Infra Components and the Infra Catalog, see the [infrastructure documentation](/docs/infrastructure/infra-components).

## Step 1: Write the Cloud SQL Manifest

Create a file named `cloud-sql.yaml` with the following content. This manifest describes a production-grade PostgreSQL instance with private networking, high availability, and automated backups.

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSql
metadata:
  name: app-database
  org: your-org
  env: production
spec:
  project_id:
    value: "your-gcp-project-id"
  region: us-central1
  database_engine: POSTGRESQL
  database_version: POSTGRES_15
  tier: db-custom-2-8192
  storage_gb: 20
  network:
    private_ip_enabled: true
    vpc_id:
      value: "projects/your-gcp-project-id/global/networks/your-vpc"
  high_availability:
    enabled: true
    zone: us-central1-b
  backup:
    enabled: true
    start_time: "02:00"
    retention_days: 7
```

Replace these placeholder values with your own:

- **`metadata.name`**: A human-readable name for the database instance. Planton generates a URL-safe slug from it.
- **`metadata.org`**: Your Planton organization slug.
- **`metadata.env`**: The environment this database belongs to (e.g., `production`, `staging`, `development`).
- **`spec.project_id.value`**: Your GCP project ID where the Cloud SQL instance will be created.
- **`spec.network.vpc_id.value`**: The full self-link of the VPC network for private IP connectivity. The format is `projects/{project}/global/networks/{network}`.
- **`spec.high_availability.zone`**: The GCP zone for the standby instance (must be in the same region).

The key fields in this manifest:

- **`database_engine` / `database_version`**: The Cloud SQL engine (`POSTGRESQL` or `MYSQL`) and version string (`POSTGRES_15`, `POSTGRES_14`, etc.) -- these match the GCP Cloud SQL API identifiers.
- **`tier`**: The compute tier. Use `db-f1-micro` for development, `db-custom-{vCPUs}-{memoryMB}` for production (e.g., `db-custom-2-8192` = 2 vCPUs, 8 GB RAM).
- **`storage_gb`**: SSD storage in gigabytes (minimum 10, maximum 65,536).
- **`network.private_ip_enabled`**: When `true`, the instance gets a private IP within your VPC (requires Private Services Access). When `false` or omitted, the instance gets a public IP -- use `authorized_networks` to restrict access.
- **`network.vpc_id`**: Required when private IP is enabled. The full VPC self-link: `projects/{project}/global/networks/{network}`.
- **`high_availability.enabled`**: Creates a standby instance in a different zone for automatic failover. Roughly doubles compute cost -- omit for development.
- **`backup.enabled` / `start_time` / `retention_days`**: Automated daily backups with point-in-time recovery. `start_time` is UTC HH:MM format. Retention range: 1-365 days.

The `project_id` and `vpc_id` fields use a nested `value` key because they also support a `variable` reference to a Planton variable or a `value_from` reference to another Infra Component's outputs.

## Step 2: Deploy with `planton apply`

Run the following command to deploy the Cloud SQL instance. The `-t` flag tells the CLI to stream the deployment progress to your terminal in real time.

```bash
planton apply -f cloud-sql.yaml -t
```

Planton validates the manifest, creates a deployment job, and begins provisioning the Cloud SQL instance on GCP. The terminal output shows four phases:

1. **init**: Configures the GCP provider and state backend (a few seconds)
2. **refresh**: Checks for any existing state (a few seconds)
3. **preview**: Plans the changes -- shows which GCP resources will be created (several seconds)
4. **update**: Creates the Cloud SQL instance on GCP (typically 5-15 minutes for Cloud SQL)

Cloud SQL instances take longer to provision than many other resource types because GCP needs to set up the managed database engine, configure replication (if HA is enabled), and assign networking. Expect the update phase to take 5-15 minutes.

If you prefer to deploy without streaming, omit the `-t` flag:

```bash
planton apply -f cloud-sql.yaml
```

The CLI prints the deployment job ID immediately. You can check on it later with:

```bash
planton follow <infra-job-id>
```

## Step 3: Verify the Deployment

After the deployment completes, retrieve the Infra Component to see its status and outputs:

```bash
planton get GcpCloudSql app-database -o yaml
```

The `status.outputs` section contains the values you need to connect your applications to the database (`planton get` prints camelCase keys, like every manifest):

| Output | Description | Example |
|---|---|---|
| `instanceName` | The GCP Cloud SQL instance name | `app-database` |
| `connectionName` | Full connection identifier in `project:region:instance` format | `your-project:us-central1:app-database` |
| `privateIp` | Private IP address (when private IP is enabled) | `10.0.0.5` |
| `publicIp` | Public IP address (when public IP is enabled) | `34.123.45.67` |
| `selfLink` | GCP resource self-link URL | `https://sqladmin.googleapis.com/...` |

The `connectionName` is the most important output. It is the identifier you use with the Cloud SQL Proxy, with Cloud SQL language connectors, and in connection strings for applications running on GCP services like Cloud Run, GKE, or Compute Engine.

To list all deployment jobs for this resource:

```bash
planton infra-job list <infra-component-id>
```

## Development Configuration

For development and testing environments where cost, speed, and direct connectivity matter more than resilience, use a lighter configuration:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSql
metadata:
  name: app-database-dev
  org: your-org
  env: development
spec:
  project_id:
    value: "your-gcp-project-id"
  region: us-central1
  database_engine: POSTGRESQL
  database_version: POSTGRES_15
  tier: db-f1-micro
  storage_gb: 10
  network:
    authorized_networks:
      - 0.0.0.0/0
  backup:
    enabled: true
    start_time: "03:00"
    retention_days: 3
```

Here is what changed from the production configuration and why:

- **`tier: db-f1-micro`**: The smallest available Cloud SQL tier (shared-core, 0.6 GB RAM). Adequate for development workloads and significantly cheaper than dedicated-core tiers.
- **No `high_availability`**: Omitted entirely. A single-zone instance is sufficient for development -- there is no need to pay for a standby replica.
- **No `private_ip_enabled`**: The instance gets a public IP instead. This makes it accessible from your local machine and CI environments without VPN or VPC peering setup.
- **`authorized_networks: 0.0.0.0/0`**: Allows connections from any IP address. **Use this only for development databases with no real data.** For staging or shared environments, restrict this to your team's IP ranges or office CIDR blocks.
- **`retention_days: 3`**: Shorter backup retention to reduce storage costs.
- **No VPC configuration**: Without private IP, no VPC peering is needed, which simplifies setup.

Deploy the development configuration the same way:

```bash
planton apply -f cloud-sql-dev.yaml -t
```

## Common Patterns and Tips

### Tuning with database flags

The `database_flags` field accepts key-value pairs that map directly to Cloud SQL database flags. For PostgreSQL, common flags include connection limits and memory settings:

```yaml
database_flags:
  max_connections: "200"
  shared_buffers: "256MB"
  work_mem: "4MB"
```

These flags are passed directly to the PostgreSQL configuration. Refer to the [GCP Cloud SQL flags documentation](https://cloud.google.com/sql/docs/postgres/flags) for available flags and valid values.

### Cost awareness

The two largest cost drivers for a Cloud SQL instance are the **machine tier** and **high availability**. An `db-f1-micro` instance costs a fraction of a `db-custom-4-16384` instance, and enabling HA roughly doubles the compute cost because GCP runs a second instance. Storage costs are proportional to allocated size. For non-production environments, start with the smallest tier and no HA -- you can resize later without recreating the instance.

## What to Do Next

Your PostgreSQL database is now running on Google Cloud SQL. From here:

- **Connect your application** using the `connectionName` output. For applications running on GCP, the [Cloud SQL Auth Proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy) provides secure, IAM-authenticated connections without managing SSL certificates or authorized networks. For applications outside GCP, connect directly using the public or private IP with the appropriate credentials.
- **Set the root password** for the `postgres` user. Cloud SQL creates the user without a password by default. Use `gcloud sql users set-password postgres --instance=app-database --password=your-secure-password` or set it through the GCP Console.
- **Explore other GCP resources** in the Infra Catalog. The same `planton apply` workflow you used here works for GKE clusters, GCS buckets, Cloud Run services, VPCs, and dozens of other GCP resource types.
