# Overview

The GCP Project API resource offers a consistent and streamlined interface for creating and managing DNS zones and records within Google Cloud DNS, Google's scalable, reliable, and managed authoritative Domain Name System (DNS) service. By abstracting the complexities of DNS configurations, this resource allows you to define your DNS settings effortlessly while ensuring consistency and compliance across different environments.

## Why We Created This API Resource

Managing DNS zones and records can be intricate due to the complexities of DNS protocols, record types, and best practices. To simplify this process and promote a standardized approach, we developed this API resource. It enables you to:

- **Simplify DNS Management**: Easily create and manage DNS zones and records without dealing with low-level GCP DNS configurations.
- **Ensure Consistency**: Maintain uniform DNS settings across different environments and applications.
- **Enhance Security**: Control access to DNS records by specifying IAM service accounts with the necessary permissions.
- **Improve Productivity**: Reduce the time and effort required to manage DNS configurations, allowing you to focus on application development and deployment.

## Key Features

### Environment Integration

- **Environment Info**: Integrates seamlessly with our environment management system to deploy DNS configurations within specific environments.
- **Stack Job Settings**: Supports custom stack-update settings for infrastructure-as-code deployments.

### GCP Credential Management

- **GCP Credential ID**: Utilizes specified GCP credentials to ensure secure and authorized operations within Google Cloud DNS.
- **Project ID**: Automatically computes and uses the GCP project ID where the managed DNS zone will be created, ensuring resources are correctly organized.

### Simplified DNS Zone and Record Management

- **IAM Service Accounts**: Specify a list of GCP service accounts (`iam_service_accounts`) to be granted permissions to manage DNS records within the managed zone. This is particularly useful for workload identities like cert-manager.
- **DNS Records Management**: Define DNS records within the zone, specifying record types, names, values, and TTLs.
    - **Record Types Supported**: Supports various DNS record types as defined in the `DnsRecordType` enum, such as `A`, `AAAA`, `CNAME`, `MX`, `TXT`, etc.
    - **Record Names**: Specify the fully qualified domain name (FQDN) for each record. The name should end with a dot (e.g., `example.com.`).
    - **Record Values**: Provide the values for each DNS record. For `CNAME` records, each value should also end with a dot.
    - **TTL Configuration**: Set the Time-To-Live (TTL) for each DNS record in seconds, controlling how long the record is cached by DNS resolvers.

### Validation and Compliance

- **Input Validation**: Implements validation rules to ensure that DNS names and record values conform to DNS standards.
    - **DNS Name Validation**: Ensures that the domain names provided are valid DNS domain names using regular expressions.
    - **Required Fields**: Enforces the presence of essential fields like `record_type` and `name`.

## Benefits

- **Simplified Deployment**: Abstracts the complexities of Google Cloud DNS configurations into an easy-to-use API.
- **Consistency**: Ensures all DNS zones and records adhere to organizational standards and best practices.
- **Scalability**: Allows for efficient management of DNS settings as your application and infrastructure grow.
- **Security**: Manages DNS configurations securely using specified GCP credentials and IAM service accounts, reducing the risk of unauthorized changes.
- **Flexibility**: Customize DNS records extensively to meet specific application requirements without compromising on best practices.
- **Compliance**: Helps maintain compliance with DNS standards and organizational policies through input validation and enforced field requirements.
