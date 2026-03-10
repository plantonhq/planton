# Pulumi Module to Deploy AwsRoute53Zone

## CLI usage (OpenMCF pulumi)

```bash
# Preview
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
openmcf pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Debugging

This module includes a `debug.sh` helper. To enable debugging, edit `Pulumi.yaml` and uncomment the `runtime.options.binary` line so Pulumi runs the program via the script:

```yaml
name: aws-module-test-pulumi-project
runtime:
  name: go
#  options:
#    binary: ./debug.sh
```

Then make the script executable and run your command (e.g., `preview` or `update`). See `docs/pages/docs/guide/debug-pulumi-modules.mdx` for full instructions.

```bash
chmod +x debug.sh
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

# AWS Route53 Zone Pulumi Module

This Pulumi module is designed to automate the creation and management of Route53 hosted zones and DNS records on AWS using a Kubernetes-like API resource model. The module is part of Planton’s unified APIs, allowing developers to define infrastructure in a declarative YAML format. By leveraging the Pulumi infrastructure-as-code platform, this module simplifies the provisioning process across AWS environments, ensuring efficient DNS management.

**Important Note:** If the API resource spec is empty, the module is not fully implemented and some features might be missing. Refer to future updates for a complete implementation.

## Key Features

### Kubernetes-Like API Resource Structure
This module adopts a Kubernetes-style API resource model. Each resource includes standard Kubernetes fields like `apiVersion`, `kind`, `metadata`, `spec`, and `status`. This makes it intuitive for developers already familiar with Kubernetes to define infrastructure in a declarative format. 

### Route53 Hosted Zone and DNS Record Management
The primary function of this module is to create and manage AWS Route53 hosted zones and their DNS records. Developers can easily define the DNS zone and its associated records through a simple YAML configuration. The module supports various DNS record types, including A, CNAME, and MX, along with TTL (time-to-live) settings for each record.

### AWS Multi-Provider Support
This module leverages both the AWS Native and AWS Classic Pulumi providers:
- **AWS Native Provider**: Handles the creation of hosted zones, ensuring efficient zone management.
- **AWS Classic Provider**: Manages DNS records within the hosted zone, providing full control over record configurations.

This dual-provider setup ensures that developers can utilize the latest AWS features while maintaining compatibility with older services.

### Seamless AWS Credential Integration
The module integrates AWS credentials securely via the provider configuration. This allows the module to authenticate and interact with AWS services, ensuring that only authorized actions are performed within your AWS account.

### Automated and Scalable DNS Management
By abstracting away the complexity of AWS Route53, the module enables the automatic creation and scaling of DNS zones and records without manual intervention. Once the YAML configuration is defined, the infrastructure is provisioned automatically via Pulumi, saving significant time and effort.

### Declarative DNS Record Configuration
The module allows developers to declaratively configure DNS records using simple YAML definitions. Record types, names, values, and TTL can all be specified in the configuration, and the Pulumi engine handles the creation and maintenance of these records.

### Status and Outputs
The module provides detailed status outputs, including hosted zone names, IDs, and nameservers. These outputs are captured in the `status.outputs`, allowing you to easily reference them in other resources or applications. This ensures that the created infrastructure can be seamlessly integrated with other components within your environment.

## Benefits

- **Declarative Infrastructure Management**: Define AWS Route53 resources as YAML files, making it easier to maintain and scale infrastructure.
- **Cross-Cloud Compatibility**: This module fits into Planton’s multi-cloud infrastructure management platform, making it easy to manage resources across different providers.
- **Comprehensive DNS Record Support**: Configure and manage a wide range of DNS record types, ensuring flexible and detailed control over your DNS setup.
- **AWS Provider Flexibility**: Uses both the AWS Native and Classic providers to ensure all relevant Route53 features are available and managed effectively.
- **Seamless AWS Integration**: Automatically provisions infrastructure using the provided AWS credentials, ensuring that the setup is secure and compliant with AWS best practices.

## Usage

Refer to the example section for usage instructions.

## Future Enhancements

The module is under continuous development, and additional features and improvements will be introduced over time. Stay tuned for updates to expand functionality and enhance usability.
