# Jenkins Kubernetes Pulumi Module

## Features

- **API Resource-Based Deployment:** The module utilizes Kubernetes-style API resources (`JenkinsKubernetes`) as input, making it easy to standardize and replicate deployments across various environments. The API resource follows a familiar structure with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, aligning with Kubernetes best practices.
  
- **Helm Chart Integration:** Jenkins is deployed using the official Helm chart. The module allows you to customize the deployment using `helm_values`, which can include adjustments to resource limits, environment variables, version tags, and other advanced configurations. For further details on the available Helm values, refer to the [Jenkins Helm chart documentation](https://github.com/jenkinsci/helm-charts/blob/main/charts/jenkins/values.yaml).

- **Namespace and Kubernetes Resource Management:** The module provides flexible namespace management through the `create_namespace` flag. When set to `true` (default), it creates a dedicated Kubernetes namespace for the Jenkins instance, ensuring proper isolation and organization of resources. When set to `false`, it uses an existing namespace, allowing for shared namespace deployments or centralized namespace management. The module also manages other key resources like Kubernetes secrets (for admin credentials), services, and ingress controllers (if enabled).

- **Kubernetes Provider Setup:** Based on the `kubernetes_credential_id` provided in the spec, the module automatically sets up the required Kubernetes provider, allowing seamless interaction with the target Kubernetes cluster. The credentials are securely handled and used to authenticate with the cluster.

- **Ingress Support:** If ingress is enabled in the `JenkinsKubernetes` resource, the module sets up an ingress controller to expose Jenkins externally, allowing users to access Jenkins via a public or private URL. If ingress is disabled, the module generates a `port_forward_command` that allows developers to access Jenkins from their local machine using port forwarding.

- **Automated Credential Management:** The module generates and manages Jenkins admin credentials as Kubernetes secrets, simplifying the authentication process. The username and the secret key for the password are captured and exported as outputs, making them easily retrievable when needed.

- **Detailed Output Handling:** After deployment, the module provides several useful outputs, including:
  - The Kubernetes namespace in which Jenkins is deployed.
  - The service name and port-forward command for accessing Jenkins when ingress is disabled.
  - The internal and external hostnames for accessing Jenkins.
  - Jenkins admin credentials (username and secret key).
  
  These outputs are stored in the `status.outputs`, ensuring that the necessary connection details are always available for future use.

## Usage

Refer to the **example section** for detailed usage instructions on how to configure the API resource and use this Pulumi module.

## Inputs

The Pulumi module accepts the following key input parameters from the `JenkinsKubernetes` API resource:

- **kubernetes_credential_id**: (Required) The ID of the Kubernetes cluster credentials used to authenticate and deploy resources on the target cluster.
- **namespace**: (Required) The Kubernetes namespace where Jenkins will be deployed.
- **create_namespace**: (Optional, default: true) Controls whether the module should create the namespace or use an existing one. Set to `false` when using a pre-existing namespace.
- **container resources**: (Required) Specifies the CPU and memory limits for the Jenkins container. This ensures that Jenkins runs with the appropriate resource allocation on the Kubernetes cluster.
- **helm_values**: (Optional) A map of key-value pairs for customizing the Jenkins Helm chart deployment. These values can be used to modify resource configurations, environment variables, and other deployment settings.
- **ingress**: (Optional) A detailed specification for setting up Kubernetes ingress to expose Jenkins externally.

## Outputs

The module exports several important output values after successful deployment:

- **namespace**: The name of the Kubernetes namespace where Jenkins is deployed.
- **service**: The service name used for internal communication with Jenkins on the Kubernetes cluster.
- **port_forward_command**: A command to set up port forwarding when ingress is disabled. This allows local access to Jenkins from a developer's machine.
- **kube_endpoint**: The internal Kubernetes endpoint for accessing Jenkins within the cluster.
- **external_hostname**: The external hostname for accessing Jenkins from outside the Kubernetes cluster (if ingress is enabled).
- **internal_hostname**: The internal hostname for accessing Jenkins in case there are internal-facing services.
- **username**: The Jenkins admin username.
- **password_secret**: The secret key that stores the Jenkins admin password in the Kubernetes secret.

## Benefits

This Pulumi module provides a consistent and repeatable way to deploy Jenkins on Kubernetes, integrating with Helm for application deployment and Kubernetes for infrastructure management. The module abstracts much of the complexity associated with setting up Jenkins in a cloud-native environment, allowing developers and DevOps teams to focus on application development rather than infrastructure.

By encapsulating the infrastructure setup in a modular form, it ensures that deployments are standardized, secure, and easy to manage. The module is designed to scale with the needs of the development team, enabling quick iteration, testing, and production rollouts of Jenkins environments across multiple cloud providers.

## Documentation

For more detailed information about the API resources and Pulumi module, including proto specifications, visit the official documentation hosted on [buf.build](https://buf.build).

