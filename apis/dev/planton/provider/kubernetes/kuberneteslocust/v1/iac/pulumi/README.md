# Locust Kubernetes Pulumi Module

## Key Features

- **Distributed Locust Clusters:** The module supports deploying distributed Locust clusters by configuring both Locust master and worker containers. You can specify the number of replicas for each, allowing for highly scalable load testing environments.
  
- **Custom Load Test Scripts:** You can define custom load testing scenarios by providing Python scripts directly via the `LocustKubernetes` API resource. The module allows you to include multiple scripts and libraries, as well as install additional Python packages via pip.

- **Kubernetes-Native Integration:** The module follows a Kubernetes-native approach, using the `LocustKubernetes` API resource to manage all aspects of the deployment, including resource allocation, scaling, and ingress. It simplifies complex Kubernetes operations such as service creation, namespace management, and Helm-based deployments.

- **Resource Management:** The module allows fine-grained control over CPU and memory resources for both master and worker Locust containers. This ensures that the Locust cluster is appropriately scaled for the desired load test and performance requirements.

- **Ingress and TLS Support:** The module provides optional ingress configuration, allowing you to expose Locust’s web UI and services either internally within the Kubernetes cluster or externally to users. TLS can be enabled for secure communication by integrating with Kubernetes’ certificate management features, making it suitable for production environments.

- **Helm Chart Customization:** The module allows advanced customization of the Locust Helm chart through `helm_values`. This includes configuring resource limits, environment variables, and other advanced deployment settings.

- **Load Test Configuration:** The load testing logic is customizable, with options to include custom Python scripts for defining user behavior, additional library files, and any pip packages necessary for test execution. This ensures flexibility for testing different application behaviors.

- **Outputs and Easy Access:** After deployment, the module exports key output information like:
  - Kubernetes namespace in which the Locust cluster is deployed.
  - Service name and port-forward commands for accessing the Locust UI locally when ingress is disabled.
  - Internal and external endpoints for accessing the Locust UI and APIs.
  - TLS-enabled access if applicable.

  These outputs are captured in `status.outputs`, making them easily retrievable for further integration or automation.

## Usage

Refer to the **example section** for detailed usage instructions on how to configure the API resource and use this Pulumi module.

## Inputs

The following key inputs are supported by the module from the `LocustKubernetes` API resource:

- **kubernetes_credential_id**: (Required) The Kubernetes cluster credentials used to authenticate and deploy resources on the target cluster.
  
- **master_container**: (Required) Defines the resource configuration for the Locust master container, including CPU and memory limits, as well as the number of replicas.

- **worker_container**: (Required) Defines the resource configuration for Locust worker containers, including CPU and memory limits, as well as the number of replicas.

- **ingress**: (Optional) Configures ingress rules to expose the Locust web UI and APIs externally. This includes options for enabling TLS and routing settings for accessing Locust over the web.

- **load_test**: (Optional) A comprehensive specification for the load test, including the main Python script (`main_py_content`), additional library files (`lib_files_content`), and any required pip packages for the Locust environment.

- **helm_values**: (Optional) Allows for advanced customization of the Locust Helm chart deployment. This includes settings such as resource limits, environment variables, and other Helm chart values for fine-tuning the deployment.

## Outputs

The module provides several key outputs for managing and accessing the deployed Locust cluster:

- **namespace**: The Kubernetes namespace where the Locust cluster is deployed.
- **service**: The name of the Kubernetes service for the Locust cluster.
- **port_forward_command**: A command to set up port forwarding, enabling local access to the Locust web UI when ingress is not enabled.
- **kube_endpoint**: The internal Kubernetes endpoint to access the Locust cluster.
- **external_hostname**: The public endpoint for accessing Locust from external clients (if ingress is enabled).
- **internal_hostname**: The internal endpoint for accessing Locust within the Kubernetes cluster.

## Benefits

This Pulumi module is a comprehensive solution for managing load testing infrastructure using Locust on Kubernetes. By providing a declarative interface to define Locust clusters, the module automates the setup and management of Kubernetes resources, allowing teams to focus on creating and executing load tests. With its built-in support for scaling, ingress, and custom load test configurations, the module provides flexibility and efficiency for testing various applications under heavy user load.

The module is highly configurable, allowing developers and testers to tailor the Locust deployment to their specific needs, from resource allocation to custom Python scripts. Additionally, it provides essential output information for easy access to Locust’s web UI and APIs, ensuring that teams can monitor and analyze load testing results effectively.
