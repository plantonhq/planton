# Solr Kubernetes Pulumi Module

## Key Features

### 1. **Kubernetes Provider Integration**

The module automatically creates a Kubernetes provider using the Kubernetes cluster credentials specified in the stack input. This ensures secure and authenticated communication with the target Kubernetes cluster.

### 2. **Namespace Creation**

A new namespace is automatically created within the Kubernetes cluster, isolating the Solr instance and its associated components. This ensures that the resources deployed for Solr are logically separated from other workloads running in the cluster.

### 3. **Solr Cluster Deployment**

The module deploys a Solr cluster based on the configuration provided in the API resource spec. Key settings, such as the number of replicas, container image, and resource limits (CPU and memory), can be customized. These configurations ensure that the Solr cluster is both scalable and tuned for performance.

### 4. **Zookeeper Integration**

The Zookeeper instance required by the Solr cluster is automatically provisioned. Zookeeper pods are configured with appropriate resource limits and persistent volumes, ensuring the reliable operation of the Solr cluster.

### 5. **Persistent Storage Configuration**

Each Solr and Zookeeper pod is provisioned with persistent storage volumes. This ensures that data is retained across pod restarts and is available for high-availability configurations.

### 6. **Ingress Management**

If ingress is enabled in the API resource spec, the module provisions ingress resources, including Istio-based ingress for routing external traffic to the Solr service. This makes it easier to expose Solr services to clients both inside and outside the Kubernetes cluster.

### 7. **Resource Customization**

The module allows fine-grained control over the deployment by enabling developers to specify resource configurations, such as CPU and memory limits for Solr and Zookeeper pods. This ensures optimal resource usage and performance tuning for both services.

### 8. **Pulumi Stack Outputs**

Upon successful deployment, the module generates several useful outputs, including:

- **Namespace**: The Kubernetes namespace in which Solr is deployed.
- **Service Name**: The service name for the Solr dashboard, which allows easy access to the Solr management UI.
- **Port Forward Command**: A command to set up port forwarding for local access to Solr if ingress is disabled.
- **Internal and External Endpoints**: URLs for accessing Solr within and outside the Kubernetes cluster.

These outputs provide essential information for developers to monitor, manage, and interact with their deployed Solr clusters.

## Usage

Refer to the example section for usage instructions.

## Benefits

1. **Standardized API Resource Structure**: The module leverages a standardized YAML specification to configure and deploy Solr clusters, making it easy for developers to replicate configurations across different environments.
2. **Seamless Kubernetes Integration**: Designed to work natively with Kubernetes, this module automates the creation and management of critical Kubernetes resources such as namespaces, deployments, and services.
3. **Infrastructure-as-Code**: By utilizing Pulumi, this module brings the benefits of infrastructure-as-code, such as version control, repeatable deployments, and rollback capabilities, to Solr deployments.
4. **Customizable Resource Definitions**: The module supports the customization of key Solr configurations, such as the number of pods, memory settings, and persistent storage allocations, allowing for flexibility based on specific use cases.
5. **Scalability and High Availability**: With support for replica configurations and resource tuning, this module ensures that Solr clusters can scale to meet performance demands while maintaining high availability and data durability.

## Prerequisites

- **Pulumi Setup**: Ensure that Pulumi is installed and configured for the cloud provider being used.
- **Kubernetes Cluster**: A Kubernetes cluster must be available, with the appropriate credentials provided in the stack input.
- **Planton CLI**: The Planton CLI should be set up to run the `planton pulumi up --stack-input <api-resource.yaml>` command.

## Pulumi Outputs

Once the module is deployed, several outputs are generated to provide critical information for managing the Solr cluster:

1. **Namespace**: The Kubernetes namespace where Solr and Zookeeper resources are deployed.
2. **Service Name**: The name of the service for accessing the Solr dashboard.
3. **Port Forward Command**: A command to enable local port forwarding to access Solr when ingress is disabled.
4. **Kube Endpoint**: The internal endpoint for accessing Solr from within the Kubernetes cluster.
5. **External Hostname**: The public endpoint for accessing Solr from outside the cluster.
6. **Internal Hostname**: The internal hostname for accessing Solr within the cluster.

These outputs ensure easy access and management of the Solr cluster post-deployment.

## Deletion Behavior

### Background Deletion Propagation

This module uses **background deletion propagation** for the SolrCloud custom resource. This is configured via the `pulumi.com/deletionPropagationPolicy: "background"` annotation and is critical for reliable `pulumi destroy` operations.

### Why This Matters

The Apache Solr Operator creates multiple child resources (ZookeeperCluster, StatefulSets, Services, ConfigMaps, PVCs) with owner references pointing to the SolrCloud CR. By default, Pulumi uses "foreground" cascading deletion, which causes a race condition:

1. Pulumi issues DELETE with `propagationPolicy: Foreground`
2. Kubernetes adds `foregroundDeletion` finalizer to SolrCloud CR
3. Garbage Collector starts deleting child resources
4. **Solr Operator sees SolrCloud still exists and recreates deleted children**
5. GC deletes them again, operator recreates them again
6. This loop continues until the 10-minute timeout

With **background deletion**:

1. Pulumi issues DELETE with `propagationPolicy: Background`
2. SolrCloud CR is removed immediately
3. Solr Operator stops reconciling (CR is gone)
4. Kubernetes GC cleans up child resources asynchronously
5. Destroy completes in seconds

### Testing Destroy Operations

When testing this module, always verify the full lifecycle:

```bash
# Create
planton pulumi up --stack-input solr.yaml

# Destroy (should complete in < 1 minute, not 10 minutes)
planton pulumi destroy --stack-input solr.yaml

# Recreate (should succeed without conflicts)
planton pulumi up --stack-input solr.yaml
```

If destroy operations timeout, check the operator logs for repeated "Creating..." messages, which indicate the race condition is occurring.

## Conclusion

The Solr Kubernetes Pulumi module provides a powerful and flexible solution for automating the deployment of Solr clusters on Kubernetes. With features like customizable resource configurations, seamless Kubernetes integration, and robust infrastructure management via Pulumi, this module greatly simplifies the operational overhead of managing Solr in cloud-native environments. By following a standardized YAML-based API resource approach, it empowers developers to deploy complex infrastructure with minimal effort, making Solr deployment more accessible and scalable.
