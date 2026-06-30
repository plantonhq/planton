# Overview

The **Harbor Kubernetes API Resource** provides a standardized and efficient way to deploy Harbor cloud-native registry onto Kubernetes clusters. This API resource simplifies the deployment process by encapsulating all necessary configurations, enabling consistent and repeatable Harbor deployments across various environments.

## Purpose

Deploying Harbor on Kubernetes involves complex configurations, including managing multiple components (Core, Portal, Registry, Jobservice), stateful dependencies (PostgreSQL, Redis), object storage, and networking. The Harbor Kubernetes API Resource aims to:

- **Standardize Deployments**: Offer a consistent interface for deploying Harbor, reducing complexity and minimizing errors.
- **Simplify Configuration Management**: Centralize all deployment settings for the entire registry stack, making it easier to manage, update, and replicate configurations.
- **Support Flexible Database Options**: Enable both self-managed and external PostgreSQL and Redis deployments.
- **Enable Production-Grade Deployments**: Support high-availability configurations with external managed databases, Redis Sentinel, and object storage.
- **Secure Artifact Management**: Configure role-based access control, vulnerability scanning, and content signing.

## Namespace Management

Harbor Kubernetes provides flexible namespace management through the `create_namespace` flag:

- **Managed Namespace** (`create_namespace: true`): The component creates and manages the namespace with appropriate resource labels for tracking and organization
- **Existing Namespace** (`create_namespace: false`): Deploy into a pre-existing namespace that must be created before deployment

### Example - Managed Namespace

```yaml
spec:
  namespace:
    value: "harbor-prod"
  create_namespace: true  # Component creates the namespace
```

### Example - Existing Namespace

```yaml
spec:
  namespace:
    value: "container-registry"
  create_namespace: false  # Must exist before deployment
```

## What is Harbor?

Harbor is an open source cloud-native registry that secures artifacts with policies and role-based access control, ensures images are scanned and free from vulnerabilities, and signs images as trusted. Harbor delivers compliance, performance, and interoperability to help consistently and securely manage artifacts across cloud-native compute platforms like Kubernetes.

Key capabilities include:
- **Container Image Registry**: Store and distribute Docker/OCI container images
- **Helm Chart Repository**: Host and manage Helm charts for Kubernetes applications
- **Vulnerability Scanning**: Integrate with Trivy, Clair, or Anchore for automated security scanning
- **Content Trust**: Sign and verify images using Docker Content Trust (Notary)
- **Policy-Based Replication**: Replicate images across registries for disaster recovery and multi-region deployments
- **Role-Based Access Control (RBAC)**: Fine-grained permissions and project-based isolation
- **Garbage Collection**: Automated cleanup of unreferenced artifacts to save storage costs

## Key Features

### Harbor Core Container

The Core service is the central API and authentication component:
- **API Server**: RESTful API for all registry operations
- **Authentication**: Supports local users, LDAP, OIDC, and UAA
- **Authorization**: Project-based RBAC with fine-grained permissions
- **Webhook**: Event notifications for image push, pull, delete, and scanning results

**Configuration Options**:
- **Replicas**: Number of Core pods for high availability (default: 1, recommended for production: 2+)
- **Resources**: CPU and memory allocations
  - Default CPU: 200m (requests), 1000m (limits)
  - Default Memory: 512Mi (requests), 2Gi (limits)
- **Container Image**: Customizable image repository and tag

### Harbor Portal Container

The Portal provides the web-based user interface:
- **Web UI**: Manage projects, users, replication policies, and view scan results
- **Dashboard**: Overview of repository statistics and recent activities
- **Project Management**: Create and configure projects with specific policies

**Configuration Options**:
- **Replicas**: Number of Portal pods (default: 1)
- **Resources**: CPU and memory allocations
  - Default CPU: 100m (requests), 500m (limits)
  - Default Memory: 256Mi (requests), 512Mi (limits)

### Harbor Registry Container

The Registry component handles the actual storage and retrieval of artifacts:
- **OCI Distribution**: Compliant with OCI Distribution Specification
- **Docker Registry V2**: Full compatibility with Docker clients
- **Layer Deduplication**: Efficient storage through content-addressable layers
- **Chunked Upload**: Resumable uploads for large images

**Configuration Options**:
- **Replicas**: Number of Registry pods for HA (default: 1, production: 2+)
- **Resources**: CPU and memory allocations for image transfer workloads
  - Default CPU: 200m (requests), 1000m (limits)
  - Default Memory: 512Mi (requests), 2Gi (limits)

### Harbor Jobservice Container

Jobservice handles asynchronous tasks:
- **Vulnerability Scanning**: Schedule and execute image scans
- **Replication**: Execute replication jobs across registries
- **Garbage Collection**: Clean up deleted artifacts and free storage
- **Retention**: Apply retention policies to automatically delete old images

**Configuration Options**:
- **Replicas**: Number of Jobservice pods (default: 1)
- **Resources**: CPU and memory for background job processing
  - Default CPU: 100m (requests), 1000m (limits)
  - Default Memory: 256Mi (requests), 1Gi (limits)

### Database Configuration (PostgreSQL)

Harbor requires PostgreSQL for storing metadata. Two deployment modes are supported:

#### Self-Managed PostgreSQL (Default)

Deploy PostgreSQL within the Kubernetes cluster:
- **Single-Node Deployment**: Suitable for development and testing
- **Persistence**: Configurable persistent storage (default: 20Gi)
- **Resource Management**: Customizable CPU and memory allocations

**PostgreSQL Container Options**:
- **Replicas**: Number of PostgreSQL pods (default: 1)
- **Resources**: CPU and memory allocations
  - Default CPU: 200m (requests), 1000m (limits)
  - Default Memory: 512Mi (requests), 2Gi (limits)
- **Persistence**:
  - Enable/disable data persistence (default: enabled)
  - Configurable disk size (default: 20Gi)
  - **Note**: Disk size cannot be changed after creation due to Kubernetes StatefulSet limitations

**Database Names**:
Harbor uses multiple databases for different components:
- **registry**: Core Harbor metadata
- **clair**: Vulnerability scanner data (if Clair is enabled)
- **notary_server**: Notary server database for content trust
- **notary_signer**: Notary signer database for content trust

#### External PostgreSQL

Connect to an existing external PostgreSQL instance:
- **Host**: Endpoint of the external PostgreSQL server
- **Port**: PostgreSQL port (default: 5432)
- **Credentials**: Username and password for authentication
- **SSL Support**: Enable TLS/SSL connections
- **Multiple Databases**: Configure separate database names for Core, Clair, and Notary components

**Prerequisites for External PostgreSQL**:
- PostgreSQL 12+ recommended
- Databases must be created before Harbor deployment
- User must have CREATE and DROP privileges on the databases
- For HA production deployments, use managed PostgreSQL services (AWS RDS, Google Cloud SQL, Azure Database)

### Cache Configuration (Redis)

Harbor uses Redis for caching and as a session store. Two deployment modes are supported:

#### Self-Managed Redis (Default)

Deploy Redis within the Kubernetes cluster:
- **Single-Node Deployment**: Suitable for development and testing
- **Persistence**: Configurable persistent storage (default: 8Gi)
- **Resource Management**: Customizable CPU and memory allocations

**Redis Container Options**:
- **Replicas**: Number of Redis pods (default: 1)
- **Resources**: CPU and memory allocations
  - Default CPU: 100m (requests), 500m (limits)
  - Default Memory: 256Mi (requests), 512Mi (limits)
- **Persistence**:
  - Enable/disable data persistence (default: enabled)
  - Configurable disk size (default: 8Gi)

#### External Redis

Connect to an existing external Redis instance:
- **Host**: Endpoint of the external Redis server
- **Port**: Redis port (default: 6379)
- **Credentials**: Username (for ACLs) and password
- **Database Index**: Redis database number (default: 0)
- **Sentinel Support**: High-availability Redis Sentinel configuration with master set name

**Prerequisites for External Redis**:
- Redis 5.0+ recommended
- For HA production deployments, use Redis Sentinel or managed Redis services (AWS ElastiCache, Google Memorystore, Azure Cache for Redis)

### Object Storage Configuration

For production high-availability deployments, external object storage is **required**. Harbor supports multiple storage backends:

#### Filesystem (Development Only)

- **Use Case**: Local development and testing only
- **Limitation**: Not suitable for production HA (ReadWriteOnce PVC limits to single Registry pod)
- **Configuration**: Disk size and optional storage class

#### AWS S3 (Recommended for Production)

- **Bucket**: S3 bucket name
- **Region**: AWS region (e.g., "us-west-2")
- **Credentials**: Access key ID and secret access key
- **Encryption**: Server-side encryption support
- **S3-Compatible**: Also works with MinIO, Ceph, and other S3-compatible services

#### Google Cloud Storage (GCS)

- **Bucket**: GCS bucket name
- **Credentials**: Service account key (base64-encoded JSON)
- **Chunk Size**: Upload chunk size (default: 5MB)

#### Azure Blob Storage

- **Account**: Azure storage account name
- **Container**: Blob container name
- **Credentials**: Account key for authentication

#### Alibaba Cloud OSS

- **Bucket**: OSS bucket name
- **Endpoint**: OSS endpoint URL
- **Credentials**: Access key ID and secret
- **HTTPS**: Secure connection support

### Ingress Configuration

Two separate ingress configurations for different Harbor endpoints:

#### Core/Portal Ingress
- Access to Harbor web UI and Docker registry API
- Configure hostname for external access
- Automatically configured with TLS via cert-manager integration

#### Notary Ingress
- Content trust and image signing endpoint
- Separate hostname configuration
- Required only if using Docker Content Trust

### Helm Chart Customization

Provide additional customization through Helm values:
- Enable and configure Trivy vulnerability scanner
- Configure Notary for image signing
- Set up replication policies
- Configure authentication providers (LDAP, OIDC)
- Customize retention and garbage collection policies
- For detailed options, refer to the [Harbor Helm Chart documentation](https://github.com/goharbor/harbor-helm)

## Architecture

Harbor consists of multiple microservices working together:

1. **Core**: API server, authentication, and webhook management
2. **Portal**: Web-based user interface
3. **Registry**: OCI/Docker artifact storage and distribution
4. **Jobservice**: Asynchronous job processing (scanning, replication, GC)
5. **PostgreSQL**: Metadata storage (projects, users, policies, scan results)
6. **Redis**: Session cache and job queue
7. **Object Storage**: Artifact blob storage (images, charts)

### Data Flow

**Image Push**:
1. User authenticates with Harbor Core via Docker CLI
2. Docker client uploads image layers to Registry service
3. Registry stores blobs in object storage backend
4. Registry stores manifest metadata in PostgreSQL via Core API
5. Jobservice triggers vulnerability scan (if enabled)

**Image Pull**:
1. User authenticates with Harbor Core
2. Docker client requests manifest from Registry
3. Registry retrieves manifest from PostgreSQL and blob references
4. Registry streams image layers from object storage to client

## Benefits

- **Enterprise-Ready**: Production-grade artifact registry with RBAC, vulnerability scanning, and audit logs
- **Cloud-Native**: Kubernetes-native deployment with support for HA and multi-region replication
- **Cost Effective**: Self-hosted registry eliminates per-pull pricing from commercial registries
- **Compliance**: Built-in vulnerability scanning and policy enforcement for security compliance
- **Flexibility**: Support for both simple single-node and production-grade distributed deployments
- **Multi-Tenancy**: Project-based isolation with fine-grained access control
- **Extensibility**: Webhook integration for CI/CD pipelines and custom workflows

## Use Cases

- **Private Container Registry**: Secure, self-hosted Docker/OCI image registry for organizations
- **Helm Chart Repository**: Centralized Helm chart storage and distribution
- **Multi-Cluster Deployments**: Replicate images across regions for reduced latency
- **Air-Gapped Environments**: On-premises registry for disconnected environments
- **CI/CD Integration**: Automated image scanning and policy enforcement in pipelines
- **Multi-Tenancy**: Separate projects for different teams with isolated access control
- **Compliance and Governance**: Vulnerability scanning, content trust, and audit trails
- **Hybrid Cloud**: Single registry interface across on-premises and cloud Kubernetes clusters

## Deployment Strategies

### Development/Evaluation

Simple single-node deployment with default settings:
- 1 Core, Portal, Registry, Jobservice pod each
- Self-managed PostgreSQL (single pod, 20Gi storage)
- Self-managed Redis (single pod, 8Gi storage)
- Filesystem storage (PVC-backed)
- No ingress (use port-forward for local access)

### Production (AWS Example)

High-availability deployment with managed services:
- 2+ replicas for Core, Portal, Registry, Jobservice
- External managed PostgreSQL (AWS RDS)
- External managed Redis (AWS ElastiCache with Sentinel)
- S3 object storage for artifacts
- Ingress with TLS certificates via cert-manager
- Trivy scanner enabled
- Automated backup and replication policies

### Enterprise Multi-Region

Geo-distributed deployment:
- Harbor deployed in multiple Kubernetes clusters (different regions)
- External object storage in each region (S3, GCS, etc.)
- Replication policies to sync artifacts across regions
- Global load balancing for read traffic
- Single source of truth in primary region

## Important Considerations

### Storage Management

- **Object Storage First**: Always use external object storage (S3, GCS, Azure) for production deployments
- **Filesystem Limitation**: ReadWriteOnce PVCs prevent multi-replica Registry deployments
- **Garbage Collection**: Schedule regular GC jobs to reclaim storage from deleted artifacts
- **Retention Policies**: Configure automated retention to control storage growth

### High Availability

- **Multiple Replicas**: Deploy 2+ replicas of all stateless components (Core, Portal, Registry, Jobservice)
- **External Databases**: Use managed PostgreSQL and Redis with built-in HA
- **Object Storage Redundancy**: S3 and GCS provide automatic redundancy
- **Health Checks**: Configure liveness and readiness probes for automatic pod recovery

### Security Considerations

- **HTTPS Required**: Always enable TLS for production registries
- **Credentials Management**: Use Kubernetes secrets for database passwords and storage credentials
- **Vulnerability Scanning**: Enable Trivy scanner for automated security checks
- **Content Trust**: Configure Notary for image signing in sensitive environments
- **RBAC**: Implement project-based access control with principle of least privilege
- **Network Policies**: Restrict network access to Harbor components

### Backup and Disaster Recovery

- **Database Backups**: Automated backups of PostgreSQL (RDS automatic backups, Cloud SQL snapshots)
- **Object Storage**: Enable versioning and cross-region replication on S3/GCS buckets
- **Configuration Export**: Backup Harbor configuration through API or database dumps
- **Replication**: Use Harbor's built-in replication for disaster recovery

### Performance Optimization

- **Object Storage Regions**: Deploy storage in same region as Kubernetes cluster
- **Redis Sizing**: Size Redis based on number of concurrent users and job queue depth
- **PostgreSQL Connection Pooling**: Configure appropriate connection limits
- **Image Layer Caching**: Leverage Harbor's content-addressable storage for deduplication
- **CDN Integration**: Use CloudFront, Cloud CDN for global image distribution


