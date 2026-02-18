# AWS ECR Repo Pulumi Module

## Introduction

The AWS ECR Repo Pulumi Module provides a standardized and efficient way to define and deploy Virtual Private Clouds (ECRs) on AWS using a Kubernetes-like API resource model. By leveraging our unified APIs, developers can specify their ECR configurations in simple YAML files, which the module then uses to create and manage AWS networking resources through Pulumi. This approach abstracts the complexity of AWS networking interactions and streamlines the deployment process, enabling consistent infrastructure management across multi-cloud environments.

## Key Features

- **Kubernetes-Like API Resource Model**: Utilizes a familiar structure with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, making it intuitive for developers accustomed to Kubernetes to define AWS ECR Repo resources.

- **Unified API Structure**: Ensures consistency across different resources and cloud providers by adhering to a standardized API resource model.

- **Pulumi Integration**: Employs Pulumi for infrastructure provisioning, enabling the use of real programming languages and providing robust state management and automation capabilities.

- **Comprehensive ECR Configuration**: Supports detailed specification of ECR attributes, including CIDR blocks, availability zones, subnet configurations, NAT gateways, and DNS settings.

- **Subnet Management**: Allows for the creation of multiple subnets per availability zone, specifying the number of hosts in each subnet to tailor network segmentation.

- **NAT Gateway Support**: Enables the configuration of NAT gateways for private subnets, allowing instances in private subnets to access the internet securely.

- **DNS Settings**: Provides options to enable or disable DNS hostnames and DNS support within the ECR, offering control over name resolution behavior.

- **Credential Management**: Securely handles AWS credentials via provider configuration, ensuring authenticated and authorized resource deployments.

- **Status Reporting**: Captures and stores outputs such as ECR IDs, Internet Gateway IDs, and subnet details in `status.outputs` for easy reference and further automation.

## Architecture

The module operates by accepting an AWS ECR Repo API resource definition as input. It interprets the resource definition and uses Pulumi to interact with AWS, creating the specified networking resources. The main components involved are:

- **API Resource Definition**: A YAML file that includes all necessary information to define a ECR, following the standard API structure.

- **Pulumi Module**: Written in Go, the module reads the API resource and uses Pulumi's AWS SDK to provision ECR resources based on the provided specifications.

- **AWS Provider Initialization**: The module initializes the AWS provider within Pulumi using the region from `spec.region` and optional provider credentials.

- **Resource Creation**: Provisions the ECR and associated resources as defined in the `spec`, including subnets, Internet Gateways, NAT Gateways, and route tables.

- **Subnet Configuration**: Creates public and private subnets across specified availability zones, with options to define the number of subnets and their sizes.

- **NAT Gateway Setup**: If enabled, deploys NAT Gateways in public subnets to allow outbound internet access from private subnets.

- **DNS Configuration**: Configures DNS settings within the ECR, such as enabling DNS hostnames and DNS support.

- **Status Outputs**: Outputs from the Pulumi deployment, such as ECR IDs, subnet IDs, and gateway IDs, are captured and stored in `status.outputs` for easy access and integration with other systems.

## Usage

Refer to the example section for usage instructions.

## Limitations

- **Advanced Networking Features**: Certain advanced networking features of AWS ECR Repos that are not specified in the current API resource definition may not be supported. Future updates may include additional capabilities based on user needs.

## Contributing

We welcome contributions to enhance the functionality of this module. Please submit pull requests or open issues to help improve the module and its documentation.
