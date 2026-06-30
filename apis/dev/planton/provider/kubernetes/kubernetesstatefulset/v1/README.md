# Overview

The **Kubernetes StatefulSet** API resource provides a standardized and streamlined way to deploy stateful applications onto Kubernetes clusters. This deployment module is designed for applications that require stable, unique network identifiers, stable persistent storage, ordered graceful deployment and scaling, and ordered automated rolling updates.

## Purpose

Deploying stateful applications to Kubernetes requires careful consideration of persistent storage, network identity, and ordering guarantees. The Kubernetes StatefulSet API resource aims to:

- **Standardize Stateful Deployments**: Offer a consistent interface for deploying stateful applications like databases, distributed systems, and message queues.
- **Simplify Configuration Management**: Consolidate all deployment-related settings including persistent volumes, network identity, and scaling policies into one place.
- **Provide Stable Identities**: Guarantee that each pod gets a stable, predictable hostname and persistent storage that survives pod rescheduling.

## Key Features

### Stable Network Identity

- **Headless Service**: Automatically creates a headless service for stable DNS-based pod discovery.
- **Predictable Pod Names**: Pods are named `<statefulset-name>-0`, `<statefulset-name>-1`, etc.
- **Pod DNS**: Each pod gets a DNS name: `<pod-name>.<headless-service>.<namespace>.svc.cluster.local`

### Persistent Storage

- **Volume Claim Templates**: Define PVC templates that create unique persistent volumes for each pod.
- **Storage Class Support**: Specify storage classes for different performance tiers.
- **Access Modes**: Configure ReadWriteOnce, ReadOnlyMany, ReadWriteMany, or ReadWriteOncePod.

### Ordered Deployment

- **Pod Management Policy**: Choose between `OrderedReady` (default) or `Parallel`.
  - **OrderedReady**: Pods are created/deleted one at a time, waiting for readiness.
  - **Parallel**: All pods are created/deleted simultaneously.

### Namespace Management

- **Namespace Configuration**: Specify the Kubernetes namespace where the StatefulSet will be deployed.
- **Namespace Creation Control**: Use the `create_namespace` flag to control whether the module should create the namespace or use an existing one.

### Container Specification

- **App Container Configuration**: Define the main application container, including:
  - **Container Image**: Set the container image with repository and tag.
  - **Resources**: Allocate CPU and memory resources.
  - **Environment Variables and Secrets**: Manage configuration data and sensitive information.
  - **Ports**: Configure container and service ports.
  - **Volume Mounts**: Mount persistent volumes into the container.
  - **Health Probes**: Configure liveness, readiness, and startup probes.

### Networking and Ingress

- **Headless Service**: Automatically created for stable network identity.
- **ClusterIP Service**: Optional service for load-balanced client access.
- **Ingress Configuration**: Set up external access with hostname routing.

### Availability

- **Replicas Management**: Define the number of pod replicas.
- **Pod Disruption Budgets**: Ensure minimum availability during voluntary disruptions.

## Benefits

- **Data Persistence**: Pods maintain their data across restarts and rescheduling.
- **Stable Identity**: Each pod has a consistent identity for clustering and leader election.
- **Ordered Operations**: Deployments and scaling happen in a controlled, ordered manner.
- **Security**: Securely manage sensitive information like credentials and secrets.

## Use Cases

- **Databases**: PostgreSQL, MySQL, MongoDB, Redis with data persistence.
- **Distributed Systems**: Kafka, ZooKeeper, Consul, etcd clusters.
- **Message Queues**: RabbitMQ, NATS with persistent message storage.
- **Caching Systems**: Redis, Memcached with persistence.
- **Search Engines**: Elasticsearch, Solr clusters.
- **Any Application Requiring**: Stable network identity or persistent storage.
