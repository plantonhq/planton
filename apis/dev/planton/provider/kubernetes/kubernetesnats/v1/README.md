# NATS on Kubernetes

The **NatsKubernetes** component in Planton provides a streamlined and reliable way to deploy and manage NATS
clusters on Kubernetes environments. By simplifying NATS deployment complexities, it integrates seamlessly into
Planton’s multi-cloud deployment ecosystem, offering consistent operations across AWS, GCP, Azure, and other
Kubernetes-supported platforms.

## Purpose and Functionality

* **Robust NATS Cluster Deployment**: Deploy scalable and highly available NATS clusters effortlessly using standardized
  Kubernetes manifests.
* **Built-in JetStream Support**: Optionally enable JetStream for reliable streaming and message persistence, with easy
  configuration for resource allocation and storage.
* **Flexible Authentication Schemes**: Secure your NATS clusters using configurable authentication mechanisms, including
  Bearer Token and Basic Authentication.
* **Integrated TLS Encryption**: Enable TLS encryption with minimal configuration to ensure secure communication across
  your messaging infrastructure.
* **External Access via Ingress**: Seamlessly expose NATS clusters externally through Kubernetes ingress controllers or
  load balancers, allowing secure, managed external client connectivity.

## Key Benefits

* **Simplified Management**: Consolidates NATS cluster operations into easy-to-use YAML manifests, validated by
  Planton’s Protobuf schemas.
* **Scalable by Default**: Provides sensible default configurations optimized for production use, with easy adjustments
  for specific requirements.
* **Unified Multi-Cloud Experience**: Utilizes Planton’s standardized APIs and CLI workflows, ensuring repeatable
  and consistent deployments across various cloud environments.
* **Enhanced Security**: Built-in options for authentication and TLS encryption make securing your message
  infrastructure straightforward and reliable.

Below is a minimal YAML example demonstrating a basic deployment of a NATS cluster with JetStream enabled, secure
authentication, and external ingress access (note the use of **camel-case** keys):

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: NatsKubernetes
metadata:
  name: exampleNats
spec:
  target_cluster:
    cluster_name: "my-gke-cluster"
  namespace:
    value: "nats-system"
  create_namespace: true
  serverContainer:
    replicas: 3
    resources:
      requests:
        cpu: "100m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "2Gi"
    diskSize: "10Gi"
  disableJetStream: false
  auth:
    enabled: true
    scheme: bearerToken
  tlsEnabled: true
  ingress:
    enabled: true
    hostname: nats.example.com
  disableNatsBox: false
```

Leverage the **NatsKubernetes** component to rapidly deploy robust, secure, and highly available NATS clusters within
your Planton multi-cloud strategy, enabling efficient and secure message-driven architectures.
