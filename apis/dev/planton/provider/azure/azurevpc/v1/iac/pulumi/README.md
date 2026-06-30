# Azure VPC Pulumi Module

## Introduction

This Pulumi module provides a standardized way to manage Azure Virtual Networks (VNets) using our Unified APIs that mimic Kubernetes' resource modeling. It allows developers to define network infrastructure configurations in a YAML file, simplifying the deployment and management of Azure networking resources across multiple environments.

## Key Features

- **Unified API Structure**: Adheres to a standardized API format with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, ensuring consistency across different resources.
- **Multi-Cloud Support**: Designed to work seamlessly in a multi-cloud environment, starting with Azure.
- **Pulumi Integration**: Leverages Pulumi's infrastructure-as-code capabilities to automate resource provisioning.
- **Credential Management**: Securely handles Azure credentials for authenticating with Azure services.
- **Simplified Deployment**: Enables developers to deploy VNets, subnets, and NAT Gateways using a single YAML configuration file.
- **Production-Ready**: Implements Azure networking best practices out of the box.

## Usage

Refer to the example section for usage instructions.

## Module Details

### API Resource Specification

The module expects an `api-resource.yaml` file defining the desired state of the Azure VPC. The key components of this file include:

- **`address_space_cidr`** (required): The CIDR block for the Virtual Network address space.
- **`nodes_subnet_cidr`** (required): The CIDR block for the primary subnet (typically for AKS nodes).
- **`is_nat_gateway_enabled`** (optional): Toggle to enable NAT Gateway for outbound connectivity.
- **`dns_private_zone_links`** (optional): List of Azure Private DNS zone resource IDs to link to the VNet.
- **`tags`** (optional): Map of tags to apply to Azure resources.

### Pulumi Module Functionality

The core functionality of this module revolves around provisioning a complete Azure networking stack including VNet, subnet, optional NAT Gateway, and DNS zone links.

#### Steps Performed:

1. **Azure Provider Initialization**:  
   Initializes the Azure provider in Pulumi using credentials supplied in the `AzureVpcStackInput`. The credentials required are:

   - `ClientId`
   - `ClientSecret`
   - `SubscriptionId`
   - `TenantId`

2. **Resource Group Creation**:  
   Creates an Azure Resource Group to contain all VNet-related resources.

3. **Virtual Network Provisioning**:  
   Creates an Azure Virtual Network with the specified address space CIDR.

4. **Subnet Creation**:  
   Provisions a subnet within the VNet for AKS node placement or other workloads.

5. **NAT Gateway (Optional)**:  
   If enabled, creates a NAT Gateway with a public IP and associates it with the subnet for outbound internet connectivity.

6. **Private DNS Zone Links (Optional)**:  
   Links specified Private DNS zones to the VNet for name resolution of Azure PaaS services.

7. **Output Handling**:  
   Captures and exports the VNet ID and subnet ID for use by other resources (such as AKS clusters).

## Outputs

- **`vnet_id`**: The Azure resource ID of the created Virtual Network
- **`nodes_subnet_id`**: The Azure resource ID of the primary subnet

## NAT Gateway

The NAT Gateway feature provides several benefits:

- **Eliminates SNAT Port Exhaustion**: Prevents connection failures from large-scale workloads
- **Predictable Egress IPs**: Enables firewall allow-listing for external services
- **Scales to 1M+ Ports**: Uses public IP prefixes for massive scale
- **Production Best Practice**: Recommended by Azure for AKS clusters

## Documentation

For detailed API definitions and additional documentation, please refer to our resources available via [buf.build](https://buf.build).

## Contributing

Contributions are welcome! Please open issues or pull requests to help improve this module.

## License

This project is licensed under the [MIT License](LICENSE).
