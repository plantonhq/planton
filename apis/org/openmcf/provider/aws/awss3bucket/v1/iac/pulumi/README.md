# Pulumi Module to Deploy AwsS3Bucket

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

# AWS AWS S3 Bucket Pulumi Module

## Introduction

The AWS AWS S3 Bucket Pulumi Module provides a standardized and efficient way to define and deploy Amazon S3 buckets on AWS using a Kubernetes-like API resource model. By leveraging our unified APIs, developers can specify their S3 bucket configurations in simple YAML files, which the module then uses to create and manage AWS S3 resources through Pulumi. This approach abstracts the complexity of AWS interactions and streamlines the deployment process, enabling consistent infrastructure management across multi-cloud environments.

## Key Features

- **Kubernetes-Like API Resource Model**: Utilizes a familiar structure with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, making it intuitive for developers accustomed to Kubernetes to define AWS S3 resources.

- **Unified API Structure**: Ensures consistency across different resources and cloud providers by adhering to a standardized API resource model.

- **Pulumi Integration**: Employs Pulumi for infrastructure provisioning, enabling the use of real programming languages and providing robust state management and automation capabilities.

- **Customizable AWS S3 Bucket Configuration**: Supports detailed specification of S3 bucket attributes, including public access settings and region configuration.

- **Public Access Control**: Allows the bucket to be configured as public or private by setting the `isPublic` field in the `spec`, providing control over bucket accessibility.

- **Region Specification**: Enables deployment of the S3 bucket in any valid AWS region by specifying the `region` field, offering flexibility in geographical placement.

- **Credential Management**: Securely handles AWS credentials via the `provider_config` in the stack input, creating an explicit Pulumi AWS provider with the supplied access key, secret key, region, and optional session token.

- **Status Reporting**: Captures and stores outputs such as the bucket ID in `status.outputs`. This facilitates easy reference and integration with other systems or automation tools.

## Architecture

The module operates by accepting an AWS AWS S3 Bucket API resource definition as input. It interprets the resource definition and uses Pulumi to interact with AWS, creating the specified S3 bucket. The main components involved are:

- **API Resource Definition**: A YAML file that includes all necessary information to define an S3 bucket, following the standard API structure. Developers specify the bucket's desired state in this file, including the public access setting and AWS region.

- **Pulumi Module**: Written in Go, the module reads the API resource and uses Pulumi's AWS SDK to provision S3 resources based on the provided specifications. It abstracts the complexity of resource creation, update, and deletion.

- **AWS Provider Initialization**: The module creates an explicit AWS provider within Pulumi using the credentials from `provider_config` in the stack input (AccessKeyId, SecretAccessKey, Region, SessionToken). This ensures that all AWS resource operations are authenticated, authorized, and targeted at the correct region.

- **Resource Creation**: Provisions the S3 bucket as defined in the `spec`, applying configurations such as public access settings and regional placement.

- **Public Access Configuration**: Controls the bucket's public accessibility by setting the appropriate policies and permissions based on the `isPublic` field.

- **Status Outputs**: Outputs from the Pulumi deployment, such as the bucket ID, are captured and stored in `status.outputs`. This information is crucial for accessing the bucket and integrating with other systems.

## Usage

Refer to the example section for usage instructions.

## Limitations

- **Advanced Bucket Configurations**: Certain advanced features of S3 buckets, such as versioning, lifecycle policies, encryption settings, or replication configurations, may not be supported in the current version of the module. Future updates may include additional capabilities based on user needs.

## Contributing

We welcome contributions to enhance the functionality of this module. Please submit pull requests or open issues to help improve the module and its documentation.
