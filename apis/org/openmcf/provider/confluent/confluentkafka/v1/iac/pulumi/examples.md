# Microservice Kubernetes Pulumi Module


### Pulumi Module Features

- **Automated Kubernetes Provider Setup**: Utilizes provided credentials and configurations to set up the Pulumi Kubernetes provider automatically, enabling seamless and secure interactions with Kubernetes clusters across different cloud environments.
  
- **Microservice Deployment**: Streamlines the creation and management of microservices by interpreting the `MicroserviceKubernetes` API resource specifications. This includes setting up Kubernetes resources such as Deployments, Services, and Ingresses based on the defined configurations.

- **Resource Management**:
  - **Container Management**: Handles the creation and configuration of containerized applications, ensuring that specified images, ports, and resource allocations are correctly applied.
  - **Networking Configuration**: Manages the setup of network protocols and port mappings, facilitating proper communication between services and external clients.
  - **Environment Variable Injection**: Integrates environment variables and secrets into the deployed containers, enabling dynamic configuration and secure handling of sensitive information.

- **Exported Stack Outputs**: Captures essential outputs including deployment statuses, service endpoints, and resource identifiers in `status.outputs`. These outputs facilitate integration with other infrastructure components, enabling effective monitoring, management, and automation workflows.

- **Scalability and Flexibility**: Designed to accommodate a wide range of microservice architectures, the module supports varying levels of complexity and can be easily extended to meet evolving infrastructure demands, ensuring long-term adaptability.

- **Error Handling**: Implements robust error handling mechanisms to promptly identify and report issues encountered during deployment or configuration processes, aiding in swift troubleshooting and maintaining infrastructure integrity.

## Installation

To integrate the Microservice Kubernetes Pulumi Module into your project, retrieve it from the [GitHub repository](https://github.com/your-repo/microservice-kubernetes-pulumi-module). Ensure that you have both Pulumi and Go installed and properly configured in your development environment.

```shell
git clone https://github.com/your-repo/microservice-kubernetes-pulumi-module.git
cd microservice-kubernetes-pulumi-module
```

## Usage

Refer to the [example section](#examples) for usage instructions.

## Examples

### Basic Example

This example demonstrates a basic setup of a microservice with minimal configuration.

```yaml
apiVersion: confluent.openmcf.org/v1
kind: MicroserviceKubernetes
metadata:
  name: todo-list-api
spec:
  version: main
  container:
    app:
      image:
        repo: nginx
        tag: latest
      ports:
        - appProtocol: http
          containerPort: 8080
          isIngressPort: true
          servicePort: 80
      resources:
        requests:
          cpu: 100m
          memory: 100Mi
        limits:
          cpu: 2000m
          memory: 2Gi
```

### Example with Environment Variables

This example includes environment variables to configure the application.

```yaml
apiVersion: confluent.openmcf.org/v1
kind: MicroserviceKubernetes
metadata:
  name: todo-list-api
spec:
  version: main
  container:
    app:
      env:
        variables:
          DATABASE_NAME: todo
      image:
        repo: nginx
        tag: latest
      ports:
        - appProtocol: http
          containerPort: 8080
          isIngressPort: true
          name: rest-api
          networkProtocol: TCP
          servicePort: 80
      resources:
        requests:
          cpu: 100m
          memory: 100Mi
        limits:
          cpu: 2000m
          memory: 2Gi
```

### Example with Environment Secrets

This example integrates environment secrets managed by Planton's GCP Secrets Manager.

```yaml
apiVersion: confluent.openmcf.org/v1
kind: MicroserviceKubernetes
metadata:
  name: todo-list-api
spec:
  version: main
  container:
    app:
      env:
        secrets:
          # value before dot 'gcpsm-my-org-prod-gcp-secrets' is the id of the gcp-secret-manager resource on planton
          # value after dot 'database-password' is one of the secrets list in 'gcpsm-my-org-prod-gcp-secrets' is the id of the gcp-secret-manager resource on planton
          DATABASE_PASSWORD: ${gcpsm-my-org-prod-gcp-secrets.database-password}
        variables:
          DATABASE_NAME: todo
      image:
        repo: nginx
        tag: latest
      ports:
        - appProtocol: http
          containerPort: 8080
          isIngressPort: true
          name: rest-api
          networkProtocol: TCP
          servicePort: 80
      resources:
        requests:
          cpu: 100m
          memory: 100Mi
        limits:
          cpu: 2000m
          memory: 2Gi
```

### Example with Empty Spec

If the `spec` field is empty, the module is not completely implemented. Below are example configurations, though they may not function as expected.

```yaml
apiVersion: confluent.openmcf.org/v1
kind: MicroserviceKubernetes
metadata:
  name: incomplete-service
spec: {}
```

```yaml
apiVersion: confluent.openmcf.org/v1
kind: MicroserviceKubernetes
metadata:
  name: another-incomplete-service
spec: {}
```

## Module Details

### Input Configuration

The module expects a `MicroserviceKubernetesStackInput` which includes:

- **Pulumi Input**: Configuration details required by Pulumi for managing the stack, such as stack names, project settings, and any necessary Pulumi configurations.
- **Target API Resource**: The `MicroserviceKubernetes` resource defining the desired microservice configuration, including environment, version, container specifications, and resource allocations.
- **Kubernetes Credentials**: Specifications for the Kubernetes credentials used to authenticate and authorize Pulumi operations, ensuring secure interactions with Kubernetes clusters.

### Exported Outputs

Upon successful execution, the module exports the following outputs to `status.outputs`:

- **Deployment Status**: Information regarding the deployment status of the microservice, enabling users to monitor the progress and health of their applications.
- **Service Endpoints**: The endpoints exposed by the microservice, facilitating access and integration with other services or external clients.
- **Resource Identifiers**: Unique identifiers for the deployed Kubernetes resources, enabling precise management and monitoring within the cluster.

These outputs facilitate integration with other infrastructure components, provide essential information for monitoring and management, and enable automation workflows to utilize the deployed microservices effectively.

## Contributing

We welcome contributions to enhance the Microservice Kubernetes Pulumi Module. Please refer to our [contribution guidelines](CONTRIBUTING.md) for more information on how to get involved, including coding standards, submission processes, and best practices.

## License

This project is licensed under the [MIT License](LICENSE), granting you the freedom to use, modify, and distribute the software with minimal restrictions. Please review the LICENSE file for more details.

## Support

For support, please contact our [support team](mailto:support@planton.cloud). We are here to help you with any issues, questions, or feedback you may have regarding the Microservice Kubernetes Pulumi Module.

## Acknowledgements

Special thanks to all contributors and the Planton community for their ongoing support and feedback. Your efforts and dedication are instrumental in making this module robust and effective.

## Changelog

A detailed changelog is available in the [CHANGELOG.md](CHANGELOG.md) file. It documents all significant changes, enhancements, bug fixes, and updates made to the Microservice Kubernetes Pulumi Module over time.

## Roadmap

We are continuously working to enhance the Microservice Kubernetes Pulumi Module. Upcoming features include:

- **Advanced IAM Configurations**: Implementing more granular permission controls for Kubernetes resources to enhance security and compliance.
- **Enhanced Monitoring Integrations**: Integrating with monitoring and logging tools such as Prometheus, Grafana, and the ELK stack for better observability and performance tracking.
- **Support for Additional Cloud Providers**: Extending support to more cloud platforms to increase flexibility and accommodate diverse infrastructure requirements.
- **Automated Scaling and Optimization**: Introducing automated scaling capabilities based on workload demands and performance metrics to optimize resource utilization.
- **Comprehensive Documentation and Tutorials**: Expanding documentation and providing step-by-step tutorials to assist users in effectively leveraging the module's capabilities.

Stay tuned for more updates as we continue to develop and refine the module to meet your infrastructure management needs.

## Contact

For any inquiries or feedback, please reach out to us at [contact@planton.cloud](mailto:contact@planton.cloud). We value your input and are committed to providing the support you need to effectively manage your microservices infrastructure.

## Disclaimer

*This project is maintained by Planton and is not affiliated with any third-party services unless explicitly stated. While we strive to ensure the accuracy and reliability of this module, users are encouraged to review and test configurations in their environments.*

## Security

If you discover any security vulnerabilities, please report them responsibly by contacting our security team at [security@planton.cloud](mailto:security@planton.cloud). We take security seriously and are committed to addressing any issues promptly to protect our users and their infrastructure.

## Code of Conduct

Please adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) when interacting with the project. We are committed to fostering an inclusive and respectful community where all contributors feel welcome and valued.

## References

- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Planton APIs](https://buf.build/openmcf/apis/docs)
