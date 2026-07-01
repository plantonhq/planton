# GCP Cloud Run Environment

The **GCP Cloud Run Environment** InfraChart provisions all cloud resources required to run containerized services on Google Cloud Run—with optional support for VPC networking, Cloud SQL database (PostgreSQL or MySQL), storage bucket, Docker repository, service account, and DNS zone.

Like the AWS ECS chart, it leverages **Jinja-based conditionals**, so you can turn features on or off with boolean flags, making it flexible for different environment needs.

Chart manifests live in the [`templates`](templates) directory; every tunable value is documented in [`values.yaml`](values.yaml).

---

## Included Cloud Resources (conditional)

| Resource                       | Always created | Controlled by boolean flag   |
|--------------------------------|----------------|------------------------------|
| **Frontend Cloud Run Service** | Yes            | —                            |
| **Backend Cloud Run Service**  | *No*           | `backendServiceEnabled`      |
| **VPC Network**                | *No*           | `networkingEnabled`          |
| **Subnetwork**                 | *No*           | `networkingEnabled`          |
| **Router NAT**                 | *No*           | `networkingEnabled`          |
| **DNS Zone**                   | *No*           | `dnsZoneEnabled`             |
| **Docker Repository**          | *No*           | `dockerRepoEnabled`          |
| **Cloud SQL Database**         | *No*           | `databaseEnabled`            |
| **Storage Bucket**             | *No*           | `storageBucketEnabled`       |
| **Service Account**            | *No*           | `serviceAccountEnabled`      |

### How the Boolean Flags Work

Each optional resource is controlled by a boolean flag in `values.yaml`:

* **`networkingEnabled: true`** → Creates VPC, Subnetwork, and Router NAT for private networking
* **`backendServiceEnabled: true`** → Creates a second Cloud Run service for backend applications
* **`dnsZoneEnabled: true`** → Creates a GCP DNS Zone for the specified domain
* **`dockerRepoEnabled: true`** → Creates an Artifact Registry Docker repository for container images
* **`databaseEnabled: true`** → Creates a Cloud SQL database instance (PostgreSQL or MySQL, based on `database_engine`)
* **`storageBucketEnabled: true`** → Creates a GCS bucket for object storage
* **`serviceAccountEnabled: true`** → Creates a service account with a JSON key

Set any flag to `false` to skip that resource entirely.

---

## Chart Input Values

Booleans are shown as **unquoted YAML booleans** (`true`/`false`) to avoid string/boolean casting issues.

### GCP Configuration

| Parameter           | Description                  | Example / Default          | Required / Default         |
|---------------------|------------------------------|----------------------------|----------------------------|
| **gcp_project_id**  | GCP Project ID               | `planton-cloud-testing`    | Required                   |
| **gcp_region**      | GCP region for all resources | `us-central1`              | Default `us-central1`      |

### Optional Networking

| Parameter              | Description                                | Example / Default      | Required / Default    |
|------------------------|--------------------------------------------|------------------------|-----------------------|
| **networkingEnabled**  | Create VPC, Subnetwork, and Router NAT     | `true` / `false`       | **Default:** `true`   |
| **vpc_name**           | Name of the VPC network                    | `cloud-run-vpc`        | Default `cloud-run-vpc` |
| **subnet_cidr**        | Primary CIDR range for the subnet          | `10.0.0.0/24`          | Default `10.0.0.0/24` |

When `networkingEnabled: true`:
- A custom-mode VPC is created with regional routing
- A subnetwork is created with private Google access enabled
- A Cloud Router with NAT gateway is created for outbound internet access
- PostgreSQL database (if enabled) uses private IP via the VPC

### Service Configuration

| Parameter                    | Description                           | Example / Default | Required / Default   |
|------------------------------|---------------------------------------|-------------------|----------------------|
| **frontend_service_name**    | Name of the frontend Cloud Run service| `frontend`        | Default `frontend`   |
| **frontend_service_port**    | Port for the frontend service         | `8080`            | Default `8080`       |
| **backend_service_name**     | Name of the backend Cloud Run service | `backend`         | Default `backend`    |
| **backend_service_port**     | Port for the backend service          | `8080`            | Default `8080`       |
| **backendServiceEnabled**    | Create backend Cloud Run service      | `true` / `false`  | **Default:** `true`  |

### Optional DNS Zone

| Parameter           | Description                    | Example / Default | Required / Default   |
|---------------------|--------------------------------|-------------------|----------------------|
| **domain_name**     | Custom domain name for DNS zone| `example.com`     | Default `example.com`|
| **dnsZoneEnabled**  | Create GCP DNS Zone            | `true` / `false`  | **Default:** `true` |

### Optional Docker Repository

| Parameter             | Description                  | Example / Default | Required / Default   |
|-----------------------|------------------------------|-------------------|----------------------|
| **dockerRepoEnabled** | Create Docker repository     | `true` / `false`  | **Default:** `true` |
| **docker_repo_name**  | Name of the Docker repository| `docker-repo`     | Default `docker-repo`|

### Optional Cloud SQL Database

| Parameter                         | Description                                              | Example / Default          | Required / Default              |
|-----------------------------------|----------------------------------------------------------|----------------------------|---------------------------------|
| **databaseEnabled**               | Create Cloud SQL database instance                       | `true` / `false`           | **Default:** `true`            |
| **database_engine**               | Database engine type                                     | `POSTGRESQL` / `MYSQL`     | **Default:** `POSTGRESQL`       |
| **database_instance_name**        | Name of the database instance                            | `database`                 | Default `database`              |
| **database_tier**                 | Cloud SQL machine tier                                   | `db-f1-micro`              | Default `db-f1-micro`           |
| **database_storage_gb**           | Storage size in GB                                       | `10`                       | Default `10`                    |
| **database_version**              | Database version (e.g., POSTGRES_15 or MYSQL_8_0)        | `POSTGRES_15`              | Default `POSTGRES_15`           |
| **database_root_password**        | Root password (rotate after deploy)                      | `change-me-immediately`    | Default `change-me-immediately` |
| **database_authorized_networks**  | List of CIDR ranges allowed to connect publicly          | `["1.2.3.4/32"]`           | Default `[]` (empty)            |

**Note:** The `database_engine` parameter is a **string enum** with two allowed values: `POSTGRESQL` and `MYSQL`. When selecting an engine, ensure the `database_version` matches the engine type (e.g., `POSTGRES_15` for PostgreSQL, `MYSQL_8_0` for MySQL).

### Optional Storage Bucket

| Parameter                 | Description              | Example / Default  | Required / Default   |
|---------------------------|--------------------------|--------------------|----------------------|
| **storageBucketEnabled**  | Create storage bucket    | `true` / `false`   | **Default:** `true` |
| **storage_bucket_name**   | Name of the storage bucket| `storage-bucket`  | Default `storage-bucket`|

### Optional Service Account

| Parameter                  | Description                       | Example / Default      | Required / Default   |
|----------------------------|-----------------------------------|------------------------|----------------------|
| **serviceAccountEnabled**  | Create Service Account with JSON key| `true` / `false`     | **Default:** `true` |
| **service_account_id**     | Service account ID                | `app-service-account`  | Default `app-service-account`|

> **Tip:** All resources are enabled by default for a complete environment setup. Toggle feature flags to `false` per environment if you don't need certain resources (e.g., disable the database for frontend-only deployments).

---

## Resource Dependencies and Deployment Order

This chart uses **synthetic relationships** to ensure resources are created in the correct order. The platform automatically builds a dependency graph (DAG) and orchestrates deployment accordingly.

### Dependency Flow

```
VPC (if networkingEnabled)
  ↓
Subnetwork + Router NAT (if networkingEnabled, depends on VPC)
  ↓
Cloud SQL Database (if enabled, depends on Subnetwork when networkingEnabled)
  ↓
Frontend Service (depends on Database, Docker Repo, DNS Zone)

Service Account (if enabled)
  ↓
Storage Bucket (if enabled, depends on Service Account)
  ↓
Backend Service (if enabled, depends on Storage Bucket, Service Account, Database, Docker Repo, DNS Zone)
```

### How It Works

- **Subnetwork and Router NAT** wait for:
  - VPC network (if `networkingEnabled: true`)

- **Cloud SQL Database** waits for:
  - Subnetwork (if `networkingEnabled: true`)

- **Frontend Service** waits for:
  - Cloud SQL database (if `databaseEnabled: true`)
  - Docker repository (if `dockerRepoEnabled: true`)
  - DNS zone (if `dnsZoneEnabled: true`)

- **Backend Service** waits for:
  - Cloud SQL database (if `databaseEnabled: true`)
  - Docker repository (if `dockerRepoEnabled: true`)
  - Storage bucket (if `storageBucketEnabled: true`)
  - Service account (if `serviceAccountEnabled: true`)
  - DNS zone (if `dnsZoneEnabled: true`)

- **Storage Bucket** waits for:
  - Service account (if `serviceAccountEnabled: true`)

**Key Benefits:**
- Resources deploy in parallel when no dependencies exist
- Services don't deploy until infrastructure (database, storage, DNS) is ready
- Conditional relationships mean disabled resources don't block deployment
- All services are grouped together for visualization

---

## Service Configuration Details

Both the frontend and backend Cloud Run services are configured with:

- **Default Image**: `nginx:latest` (replace with your actual container images after deployment)
- **Resources**: 1 CPU, 512MB memory
- **Scaling**: Min 0, Max 1 replica (adjust as needed)
- **Concurrency**: 80 requests per container
- **Authentication**: Public (unauthenticated access allowed)

### Environment Variables

Each service includes placeholder environment variables that you should customize:

- **`SERVICE_NAME`**: Set to the service name (frontend or backend)
- **`ENV`**: Set to the environment slug (dev, staging, prod)

You can add additional environment variables or replace these after deployment by updating the Cloud Run service configuration.

---

## IAM Permissions Note

When `serviceAccountEnabled: true`, the chart creates a service account with a JSON key but **does not assign IAM permissions**.

**Required Manual Steps:**

1. After deployment, retrieve the service account email from the outputs
2. Grant necessary IAM roles using GCP Console or `gcloud` CLI
3. Example: To grant Storage Admin permissions:
   ```bash
   gcloud projects add-iam-policy-binding PROJECT_ID \
     --member="serviceAccount:SERVICE_ACCOUNT_EMAIL" \
     --role="roles/storage.admin"
   ```
4. Download the JSON key from the GCP Console or retrieve it via API
5. Share the key securely with developers for application authentication

---

## Cloud SQL Database Configuration

When `databaseEnabled: true`, a Cloud SQL instance is created with the engine specified by `database_engine`. The chart supports both **PostgreSQL** and **MySQL**.

### Choosing a Database Engine

| Engine       | `database_engine` | Example `database_version` |
|--------------|-------------------|----------------------------|
| PostgreSQL   | `POSTGRESQL`      | `POSTGRES_15`, `POSTGRES_14`, `POSTGRES_13` |
| MySQL        | `MYSQL`           | `MYSQL_8_0`, `MYSQL_5_7` |

> **Important:** Ensure `database_version` matches the selected `database_engine`. For example, use `POSTGRES_15` with `POSTGRESQL` or `MYSQL_8_0` with `MYSQL`.

### Network Configuration

Network configuration depends on the `networkingEnabled` flag:

#### With Networking Enabled (`networkingEnabled: true`)

- **Private IP**: Enabled via VPC peering (most secure)
- **Public IP**: Only enabled if `database_authorized_networks` is not empty
- **Authorized Networks**: Uses the list from `database_authorized_networks` (can be empty for private-only access)

#### Without Networking (`networkingEnabled: false`)

- **Public IP**: Enabled by default
- **Authorized Networks**: Uses `database_authorized_networks` if provided, otherwise defaults to `0.0.0.0/0` (open to all IPs)

### Important Security Steps

1. **Rotate the root password immediately** after deployment
2. **Enable networking** (`networkingEnabled: true`) for production environments to use private IP
3. **Restrict authorized networks** by providing specific CIDR ranges in `database_authorized_networks`
4. **Create application-specific database users** instead of using root
5. **Use Cloud SQL Proxy** or VPC connector for secure connections from Cloud Run services

---

## Docker Repository (Artifact Registry)

When `dockerRepoEnabled: true`, an Artifact Registry Docker repository is created for storing container images.

**Usage:**

1. Authenticate Docker to the registry:
   ```bash
   gcloud auth configure-docker REGION-docker.pkg.dev
   ```
2. Tag your images:
   ```bash
   docker tag my-app:latest REGION-docker.pkg.dev/PROJECT_ID/REPO_NAME/my-app:latest
   ```
3. Push images:
   ```bash
   docker push REGION-docker.pkg.dev/PROJECT_ID/REPO_NAME/my-app:latest
   ```
4. Update the Cloud Run service to use your pushed images

---

## Customization & Management

* Toggle feature flags per environment (dev vs prod) in a higher-priority `values.yaml`
* Change service ports to expose different container ports
* All resources are independent (no cross-resource dependencies), so they deploy in parallel
* Update container images, CPU, memory, and scaling parameters post-deployment as needed
* Add custom domains to Cloud Run services by enabling the `dns` block in the service templates

---

## Important Notes

* **Networking**: Enable `networkingEnabled: true` for production to use private IP for the database. The VPC includes a Router NAT for outbound internet access from private resources
* **DNS Zone**: Ensure your domain is registered and delegated to GCP before enabling `dnsZoneEnabled`
* **Database Engine**: Use `database_engine` to choose between `POSTGRESQL` and `MYSQL`, and ensure `database_version` matches the selected engine
* **Database Password**: The default password `change-me-immediately` should be rotated immediately after deployment for security
* **Authorized Networks**: When networking is disabled, the default `0.0.0.0/0` allows connections from anywhere. Use `database_authorized_networks` to restrict access or enable networking for private IP
* **Service Account**: IAM permissions must be granted manually after the service account is created
* **Container Images**: Both services default to `nginx:latest`. Replace with your actual application images after deployment
* **Environment Variables**: Customize the placeholder `SERVICE_NAME` and `ENV` variables for your applications

---

© 2025 Planton. All rights reserved.
