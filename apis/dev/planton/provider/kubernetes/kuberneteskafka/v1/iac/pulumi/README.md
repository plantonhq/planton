# Kafka Kubernetes Pulumi Module

This Pulumi module simplifies the deployment and management of a Kafka cluster on Kubernetes using the Strimzi operator. By leveraging a standardized API resource definition, it enables developers to configure and deploy complex Kafka infrastructures with minimal effort. The module supports additional components like Schema Registry and Kafka UI (Kowl), providing a comprehensive solution for event streaming platforms.

## Key Features

- **Standardized API Resource**: Utilizes a consistent API structure with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, making it easy to define and manage resources.

- **Customizable Kafka Cluster**: Configure broker and zookeeper replicas, resource allocations, and storage options to meet specific requirements.

- **Kafka Topics Management**: Define multiple Kafka topics with custom configurations, partitions, and replicas directly in the API resource.

- **Optional Components**:
    - **Schema Registry**: Enable Schema Registry deployment for managing Avro schemas.
    - **Kafka UI (Kowl)**: Optionally deploy Kowl for an intuitive web-based Kafka management interface.

- **Ingress Configuration**: Supports external access via ingress resources, with automatic TLS certificate management using Cert-Manager.

- **Pulumi Integration**: Written in Golang, the module leverages Pulumi for infrastructure as code, enabling seamless integration into existing workflows.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [API Resource Specification](#api-resource-specification)
    - [KafkaKubernetesSpec](#kafkakubernetesspec)
    - [KafkaKubernetesBrokerContainer](#kafkakubernetesbrokercontainer)
    - [KafkaKubernetesZookeeperContainer](#kafkakuberneteszookeepercontainer)
    - [KafkaKubernetesSchemaRegistryContainer](#kafkakubernetesschemaregistrycontainer)
    - [KafkaTopic](#kafkatopic)
- [Usage](#usage)
    - [Sample YAML Configuration](#sample-yaml-configuration)
    - [Deploying with CLI](#deploying-with-cli)
- [Module Components](#module-components)
    - [Namespace Creation](#namespace-creation)
    - [Kafka Cluster Deployment](#kafka-cluster-deployment)
    - [Kafka Topics Creation](#kafka-topics-creation)
    - [Schema Registry Deployment](#schema-registry-deployment)
    - [Kafka UI (Kowl) Deployment](#kafka-ui-kowl-deployment)
- [Outputs](#outputs)
- [Contributing](#contributing)
- [License](#license)

## Prerequisites

- Kubernetes cluster with appropriate permissions.
- Strimzi Kafka Operator installed in the cluster.
- Cert-Manager installed for certificate management.
- Pulumi CLI installed.
- Golang environment set up if modifying the module code.

## Installation

Clone the repository containing the Pulumi module:

```bash
git clone https://github.com/your-org/kafka-kubernetes-pulumi-module.git
```

Install the required dependencies:

```bash
cd kafka-kubernetes-pulumi-module
go mod download
```
## Usage

Refer to [example](example.md) for usage instructions.

## Module Components

### Namespace Management

The module provides flexible namespace management through the `create_namespace` flag:

- **Automatic Creation** (`create_namespace: true`): The module creates a dedicated Kubernetes namespace with appropriate labels, ensuring resource isolation and lifecycle management.
- **External Management** (`create_namespace: false`): Use an existing namespace managed outside this module (e.g., via KubernetesNamespace component or manual creation). Resources will be deployed into the specified namespace without attempting to create it.

**When to use `create_namespace: true`:**
- New deployments where namespace doesn't exist
- Full lifecycle management by the module
- Default and recommended approach

**When to use `create_namespace: false`:**
- Namespace is managed by a centralized governance system
- Pre-existing namespace with specific configurations
- Multi-tenant environments with namespace policies

### Kafka Cluster Deployment

- **Brokers**: Deploys the specified number of Kafka broker pods with the provided resource allocations and storage configurations.
- **Zookeeper**: Sets up Zookeeper nodes required for Kafka coordination.

### Kafka Topics Creation

Automatically creates the defined Kafka topics using the Strimzi KafkaTopic custom resource.

### Schema Registry Deployment

If enabled, deploys the Confluent Schema Registry:

- Configures connections to the Kafka cluster.
- Sets up authentication using SASL/SCRAM.

### Kafka UI (Kowl) Deployment

Optionally deploys Kowl, a web-based UI for Kafka:

- Connects to the Kafka cluster using provided credentials.
- Provides an interface to monitor topics, consumer groups, and more.

### Ingress Configuration

- Sets up ingress resources for external access to Kafka brokers, Schema Registry, and Kowl.
- Manages TLS certificates using Cert-Manager and Kubernetes secrets.

## Outputs

After deployment, the module provides several outputs:

- **Namespace**: The Kubernetes namespace where resources are deployed.
- **Kafka Admin Credentials**: Username and secret references for the admin user.
- **Bootstrap Servers**: Internal and external hostnames for connecting to the Kafka cluster.
- **Schema Registry URLs**: URLs for accessing the Schema Registry.
- **Kafka UI URL**: External URL for accessing Kowl.


## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the [MIT License](LICENSE).
