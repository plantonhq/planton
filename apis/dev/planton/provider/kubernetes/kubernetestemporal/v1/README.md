# Deploy Temporal on Kubernetes

The `TemporalKubernetes` component enables users of Planton to deploy and manage [Temporal](https://temporal.io/)
—a robust, distributed, and highly scalable workflow orchestration platform—within Kubernetes environments. It leverages
Planton's standardized, protobuf-defined API resource model to ensure consistent and simplified deployment across
any Kubernetes cluster, regardless of the underlying cloud provider.

## Why We Created the TemporalKubernetes Component

Deploying Temporal involves significant complexity, particularly when managing external dependencies like databases (
PostgreSQL, Cassandra, MySQL) and Elasticsearch for visibility and observability. To streamline this complexity and
standardize the deployment process, the TemporalKubernetes component:

- **Simplifies Deployment**: Reduces configuration complexity, making it easy to set up Temporal with minimal inputs.
- **Standardizes Across Clouds**: Abstracts the differences among Kubernetes environments across AWS, GCP, Azure, or any
  other platform.
- **Ensures Flexibility**: Offers built-in support for common external database backends and Elasticsearch, allowing
  integration with existing infrastructure.
- **Improves Observability**: Provides straightforward integration with external Elasticsearch to leverage advanced
  monitoring and troubleshooting capabilities.

## Key Features

### Database Flexibility

- **Multiple Backend Support**: Easily configure Temporal with your choice of Cassandra, PostgreSQL, or MySQL.
- **External Database Integration**: Optionally connect Temporal to external databases, simplifying integration into
  existing database infrastructure.
- **Automated Schema Setup**: Ability to toggle automatic schema creation, ensuring database structures align with
  organizational compliance and governance standards.

### Observability and Monitoring

- **External Elasticsearch Integration**: Supports external Elasticsearch clusters, enabling robust, scalable
  observability and monitoring features critical for production deployments.

### Search Attributes

- **Custom Search Attributes**: Define indexed fields for filtering and querying workflows in the Temporal UI and SDK.
- **Type Safety**: Support for all Temporal search attribute types (Keyword, Text, Int, Double, Bool, Datetime, KeywordList).
- **Backend Compatibility**: Works with both SQL and Elasticsearch visibility stores (note: Text type requires Elasticsearch).

### Deployment Simplicity

- **Automatic UI and Frontend Exposure**: If ingress is enabled, the component automatically configures frontend
  services with a load balancer and optionally exposes Temporal's Web UI via Kubernetes ingress, ensuring seamless and
  secure external access.
- **Customizable Web UI**: Easily enable or disable the Temporal Web UI based on organizational security policies or
  deployment preferences.

### Consistent Validation

- **Protobuf-based Validation**: Employs Protocol Buffers (Protobuf) for field-level validation, ensuring correctness of
  configuration before deployment.

## Benefits

- **Reduced Complexity**: Dramatically simplifies the complex setup and management of Temporal deployments.
- **Cross-Cloud Compatibility**: Ensures identical deployment processes and configuration standards regardless of cloud
  providers.
- **Enhanced Scalability**: Designed to scale seamlessly within Kubernetes, handling workflow orchestration from simple
  tasks to complex, large-scale operations.
- **Secure Integrations**: Securely integrates with existing infrastructure, enhancing compliance, governance, and
  security.

## Usage Examples

### Example: Deploying Temporal with External PostgreSQL and Elasticsearch

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: TemporalKubernetes
metadata:
  name: temporal-production
spec:
  database:
    backend: postgresql
    externalDatabase:
      host: "postgres.prod.example.com"
      port: 5432
      user: "temporal_user"
      password: "secure_password"
    databaseName: "temporal"
    visibilityName: "temporal_visibility"
    disableAutoSchemaSetup: false

  externalElasticsearch:
    host: "elasticsearch.prod.example.com"
    port: 9200
    user: "elastic_user"
    password: "elastic_password"

  ingress:
    frontend:
      enabled: true
      hostname: temporal-frontend.example.com
    webUi:
      enabled: true
      hostname: temporal-ui.example.com

  disableWebUi: false
```

### Minimal Example: In-cluster Cassandra and no external dependencies

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: TemporalKubernetes
metadata:
  name: temporal-dev
spec:
  database:
    backend: cassandra

  disableWebUi: true
```

## Verifying Deployment

Once deployed, verify Temporal using the following commands:

```bash
kubectl get pods -n temporal-production
kubectl port-forward svc/temporal-frontend -n temporal-production 7233:7233
```

Access the Temporal Web UI (if enabled):

```bash
kubectl port-forward svc/temporal-web-ui -n temporal-production 8080:8080
```

Then visit [http://localhost:8080](http://localhost:8080).

## Conclusion

The `TemporalKubernetes` component simplifies the deployment and management of Temporal within Kubernetes environments,
seamlessly integrating with your existing infrastructure and observability tools. Its standardized, flexible design
ensures efficiency and reliability for your workflow orchestration needs within the Planton ecosystem.
