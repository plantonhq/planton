# Overview

The **SigNoz Kubernetes API Resource** provides a standardized and efficient way to deploy SigNoz observability platform onto Kubernetes clusters. This API resource simplifies the deployment process by encapsulating all necessary configurations, enabling consistent and repeatable SigNoz deployments across various environments.

## Purpose

Deploying SigNoz on Kubernetes involves complex configurations, including managing multiple components (UI, API, OpenTelemetry Collector, ClickHouse, Zookeeper), resource management, storage persistence, networking, and environment settings. The SigNoz Kubernetes API Resource aims to:

- **Standardize Deployments**: Offer a consistent interface for deploying SigNoz, reducing complexity and minimizing errors.
- **Simplify Configuration Management**: Centralize all deployment settings for the entire observability stack, making it easier to manage, update, and replicate configurations.
- **Support Flexible Database Options**: Enable both self-managed ClickHouse deployments and connections to external ClickHouse instances.
- **Enable Production-Grade Deployments**: Support high-availability configurations with distributed ClickHouse clusters and Zookeeper coordination.
- **Streamline Telemetry Ingestion**: Configure OpenTelemetry Collector for scalable data ingestion from instrumented applications.

## What is SigNoz?

SigNoz is an open-source, OpenTelemetry-native observability platform that unifies logs, metrics, and traces into a single application. It provides a comprehensive alternative to proprietary SaaS solutions like DataDog and New Relic, offering:

- **Unified Observability**: Single platform for logs, metrics, and distributed traces
- **OpenTelemetry Native**: Built on open standards, avoiding vendor lock-in
- **High Performance**: Powered by ClickHouse columnar database for fast analytical queries
- **Cost Effective**: Self-hosted deployment with predictable infrastructure costs
- **Seamless Correlation**: Out-of-the-box correlation between telemetry signals using TraceID and SpanID

## Key Features

### SigNoz Main Container

The SigNoz binary consolidates multiple services into a single deployment:
- **UI**: ReactJS-based frontend for visualization, dashboarding, and alerting
- **API Server**: Backend for querying telemetry data from ClickHouse
- **Ruler**: Continuous evaluation of alerting rules
- **Alertmanager**: Alert deduplication, grouping, and routing to notification channels

**Configuration Options**:
- **Replicas**: Number of SigNoz pods for high availability (default: 1, recommended for production: 2+)
- **Resources**: CPU and memory allocations optimized for query workloads
  - Default CPU: 200m (requests), 1000m (limits)
  - Default Memory: 512Mi (requests), 2Gi (limits)
- **Container Image**: Customizable image repository and tag

### OpenTelemetry Collector

The data ingestion gateway for SigNoz, responsible for:
- Receiving telemetry data in multiple formats (OTLP, Jaeger, Zipkin, Prometheus)
- Processing and enriching data (batching, adding Kubernetes metadata, sampling)
- Exporting processed data to ClickHouse

**Configuration Options**:
- **Replicas**: Number of collector pods to handle ingestion load (default: 2)
- **Resources**: CPU and memory allocations for data processing
  - Default CPU: 500m (requests), 2000m (limits)
  - Default Memory: 1Gi (requests), 4Gi (limits)
- **Ingress**: Separate ingress configuration for gRPC (port 4317) and HTTP (port 4318) endpoints

### Database Configuration

SigNoz supports two database deployment modes:

#### Self-Managed ClickHouse (Default)

Deploy and manage ClickHouse within the Kubernetes cluster with support for:
- **Single-Node Deployments**: Simple evaluation and development environments
- **Distributed Clusters**: Production-grade high availability with:
  - **Sharding**: Horizontal data distribution across multiple nodes
  - **Replication**: Data redundancy for fault tolerance (recommended: 2 replicas per shard)
  - **Zookeeper Coordination**: Required for distributed clusters (recommended: 3 or 5 nodes for quorum)

**ClickHouse Container Options**:
- **Replicas**: Number of ClickHouse pods (default: 1)
- **Resources**: CPU and memory for OLAP workloads
  - Default CPU: 500m (requests), 2000m (limits)
  - Default Memory: 1Gi (requests), 4Gi (limits)
- **Persistence**: 
  - Enable/disable data persistence (default: enabled)
  - Configurable disk size (default: 20Gi)
  - **Note**: Disk size cannot be changed after creation due to Kubernetes StatefulSet limitations

**Clustering Options**:
- **Enable Clustering**: Toggle distributed cluster mode (default: disabled)
- **Shard Count**: Number of shards for horizontal scaling (recommended for production: 2+)
- **Replica Count**: Number of replicas per shard (recommended for production: 2)

**Zookeeper Configuration**:
- **Enable Zookeeper**: Required when clustering is enabled
- **Replicas**: Number of Zookeeper nodes (must be odd: 3 or 5 for production quorum)
- **Resources**: CPU and memory allocations
  - Default CPU: 100m (requests), 500m (limits)
  - Default Memory: 256Mi (requests), 512Mi (limits)
- **Disk Size**: Persistent volume size (default: 8Gi)

#### External ClickHouse

Connect to an existing external ClickHouse instance instead of deploying one:
- **Host**: Endpoint of the external ClickHouse instance
- **HTTP Port**: Port for HTTP interface (default: 8123)
- **TCP Port**: Port for native protocol (default: 9000)
- **Cluster Name**: Name of the distributed cluster configuration (default: "cluster")
- **Security**: Toggle TLS connection support
- **Credentials**: Username and password for authentication

**Prerequisites for External ClickHouse**:
- At least one Zookeeper instance must be available for distributed cluster support
- External ClickHouse must have a distributed cluster configured
- Provided credentials must have sufficient privileges to create and manage databases/tables

### Ingress Configuration

Two separate ingress configurations support different access patterns:

#### SigNoz UI Ingress
- Access to the SigNoz web interface and API
- Configure hostname and path routing
- Enable TLS/SSL certificates

#### OpenTelemetry Collector Ingress
- Data ingestion endpoints for instrumented applications
- **Important**: Separate ingress resources required for gRPC and HTTP protocols due to nginx annotation requirements
- gRPC endpoint: Requires `nginx.ingress.kubernetes.io/backend-protocol: "GRPC"` annotation
- HTTP endpoint: Standard HTTP routing

### Helm Chart Customization

Provide additional customization through Helm values:
- Configure alerting integrations (Slack, PagerDuty, email)
- Set data retention policies (TTL) for logs, metrics, and traces
- Customize environment variables
- Configure S3 archiving for long-term data retention
- For detailed options, refer to the [SigNoz Helm Chart documentation](https://github.com/SigNoz/charts)

## Namespace Management

The SigNoz Kubernetes component provides flexible namespace management through the `create_namespace` flag:

- **When `create_namespace` is `true`**: The component will create the specified namespace on the target cluster. This is the recommended approach for new deployments where the namespace doesn't exist yet.

- **When `create_namespace` is `false`**: The component will use an existing namespace on the cluster. The namespace must already exist, or the deployment will fail. This is useful when:
  - The namespace is managed by a separate process or team
  - Multiple components share the same namespace
  - Namespace creation is controlled by organizational policies

**Default Behavior**: There is no default value; `create_namespace` must be explicitly set to either `true` or `false`.

**Best Practice**: For most deployments, set `create_namespace` to `true` to let the component manage its own namespace. Only set it to `false` when you have specific requirements for external namespace management.

## Architecture

SigNoz consists of four main components working together:

1. **SigNoz Binary** (Frontend + API + Ruler + Alertmanager): User-facing services for visualization and alerting
2. **OpenTelemetry Collector**: Data ingestion and processing gateway
3. **ClickHouse**: High-performance columnar database for telemetry data storage
4. **Zookeeper**: Coordination service for distributed ClickHouse clusters (optional, required for HA)

### Data Flow

1. Instrumented applications export telemetry via OpenTelemetry SDK
2. Data is received by the OpenTelemetry Collector service (gRPC/HTTP)
3. Collector processes and writes data to ClickHouse
4. Users access SigNoz UI to query and visualize telemetry data
5. SigNoz API queries ClickHouse and returns results to the frontend

## Benefits

- **Unified Observability**: Single platform for logs, metrics, and traces eliminates tool sprawl
- **OpenTelemetry Native**: Leverage open standards and avoid vendor lock-in
- **Cost Effective**: Self-hosted deployment with predictable infrastructure costs vs. per-GB SaaS pricing
- **High Performance**: ClickHouse columnar database enables fast queries on massive datasets
- **Flexibility**: Support for both simple single-node and production-grade distributed deployments
- **Data Control**: Full control over data location and retention for compliance requirements
- **Seamless Correlation**: Native correlation between logs, metrics, and traces using TraceID/SpanID

## Use Cases

- **Microservices Observability**: Monitor distributed systems with unified logs, metrics, and traces
- **Application Performance Monitoring (APM)**: Track service latencies, error rates, and throughput
- **Infrastructure Monitoring**: Collect and analyze metrics from Kubernetes clusters and applications
- **Log Analytics**: Centralized log aggregation and analysis with powerful query capabilities
- **Distributed Tracing**: Visualize request flows across microservices with flamegraphs and Gantt charts
- **Custom Dashboards**: Create tailored dashboards for specific services and teams
- **Alerting and Incident Response**: Configure alerts based on logs, metrics, or trace data
- **Cost Optimization**: Identify performance bottlenecks and optimize resource utilization

## Deployment Strategies

### Development/Evaluation

Simple single-node deployment with default settings:
- 1 SigNoz pod
- 2 OpenTelemetry Collector pods
- 1 ClickHouse pod with persistence
- No Zookeeper (clustering disabled)

### Production

High-availability deployment with distributed ClickHouse:
- 2+ SigNoz pods for redundancy
- Scale OpenTelemetry Collector based on ingestion volume
- Distributed ClickHouse with 2 shards, 2 replicas per shard
- 3 or 5 Zookeeper nodes for quorum
- Persistent storage with appropriate StorageClass
- Ingress with TLS certificates
- Monitoring of PVC usage and resource consumption

### Hybrid (External ClickHouse)

Use external managed ClickHouse service:
- SigNoz and OpenTelemetry Collector deployed in-cluster
- Connect to external ClickHouse (managed by cloud provider or separate team)
- No Zookeeper deployment needed in SigNoz namespace
- Reduced operational complexity while maintaining observability platform control

## Important Considerations

### Storage Management

- **PVC Monitoring**: Monitor disk usage of persistent volumes to prevent outages
- **Disk Size**: Cannot be modified after creation due to Kubernetes StatefulSet limitations
- **StorageClass**: A default StorageClass must be configured in the cluster for dynamic provisioning
- **Data Retention**: Configure TTL policies to manage storage growth and costs

### High Availability

- **Zookeeper Dependency**: Required for distributed ClickHouse, adds operational complexity
- **Pod Anti-Affinity**: Configure to ensure replicas are scheduled on different nodes
- **Quorum**: Use odd number of Zookeeper nodes (3 or 5) to maintain quorum

### Backup and Restore

- **Current Limitation**: No built-in unified backup/restore mechanism
- **Configuration Backup**: SQLite database file (`/var/lib/signoz/signoz.db`) contains dashboards and alerts
- **Telemetry Data**: Use third-party tools like `clickhouse-backup` or cloud-provider snapshots
- **S3 Archiving**: Configure OTel Collector for long-term archiving (not a backup solution)

### Ingress Complexity

- **Dual Protocol Support**: OpenTelemetry Collector requires separate ingress resources for gRPC and HTTP
- **TLS Configuration**: Use cert-manager for automated SSL/TLS certificate management
- **Network Policies**: Consider implementing NetworkPolicy resources for security

### Resource Planning

- **Monitor Actual Usage**: Use SigNoz's built-in Kubernetes dashboards to track resource consumption
- **Right-size Allocations**: Adjust requests and limits based on observed patterns
- **Scaling**: Independently scale SigNoz, OpenTelemetry Collector, and ClickHouse based on workload

