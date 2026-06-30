# Overview

The GCP Service Account API resource provides a consistent and streamlined interface for creating and managing Google
Cloud service accounts, their optional JSON keys, and IAM role bindings within Google Cloud Platform. By abstracting
the complexities of service account provisioning and permission management, this resource allows you to define identity
and access configurations effortlessly while ensuring consistency and compliance across different environments.

## Why We Created This API Resource

Managing GCP service accounts involves coordinating multiple steps: creating the account, optionally generating keys,
and binding IAM roles at project or organization levels. This can become complex when dealing with multiple environments
and requires careful security practices. To simplify this process and promote secure, standardized patterns, we
developed this API resource. It enables you to:

- **Simplify Service Account Management**: Easily create and configure service accounts without dealing with low-level
  GCP IAM configurations or multiple `gcloud` commands.
- **Ensure Consistency**: Maintain uniform service account configurations across different environments and applications.
- **Enhance Security**: Default to keyless authentication patterns while supporting key generation only when explicitly
  needed, following modern cloud security best practices.
- **Improve Productivity**: Reduce the time and effort required to manage service accounts and their permissions,
  allowing you to focus on application development and deployment.
- **Enable GitOps Workflows**: Define service accounts declaratively in version-controlled manifests that work with both
  Pulumi and Terraform backends.

## Key Features

### Environment Integration

- **Environment Info**: Integrates seamlessly with Planton's environment management system to deploy service
  account configurations within specific environments.
- **Stack Job Settings**: Supports custom stack-update settings for infrastructure-as-code deployments.
- **Multi-Backend Support**: The same manifest works with both Pulumi and Terraform backends, providing flexibility in
  tool choice.

### Service Account Creation

- **Account Provisioning**: Creates a GCP service account with a specified `serviceAccountId` in the target GCP project.
- **Display Name**: Automatically uses the metadata name as the display name for easy identification in the GCP Console.
- **Project Scoping**: Specify the target `projectId` where the service account should be created, or use the provider's
  default project.

### Optional Key Generation

- **Keyless by Default**: The `createKey` field defaults to `false`, encouraging modern keyless authentication patterns
  using Workload Identity, attached service accounts, or federated identity.
- **Legacy Support**: When `createKey` is set to `true`, generates a JSON private key for scenarios requiring
  traditional key-based authentication.
- **Secure Key Handling**: Generated keys are base64-encoded in stack outputs and should be stored in secure secret
  management systems.

### IAM Role Management

- **Project-Level Roles**: Grant IAM roles scoped to the service account's project via the `projectIamRoles` field.
  Supports any valid GCP IAM role (e.g., `roles/logging.logWriter`, `roles/storage.admin`).
- **Organization-Level Roles**: When `orgId` is provided, grant IAM roles at the organization level via the
  `orgIamRoles` field, enabling cross-project permissions.
- **Declarative Bindings**: All role bindings are managed declaratively, ensuring permissions match the desired state
  defined in the manifest.

### Validation and Compliance

- **Input Validation**: Implements strict validation rules to ensure service account configurations meet GCP
  requirements:
    - **Account ID Validation**: Enforces `serviceAccountId` length between 6-30 characters, matching GCP naming rules.
    - **Required Fields**: Ensures essential fields like `serviceAccountId` are always provided.
    - **Org Role Validation**: Prevents configuration of `orgIamRoles` without a corresponding `orgId`.
- **Proto-Defined Schema**: Uses Protocol Buffers with buf.validate rules to provide compile-time type safety and
  runtime validation.

## Benefits

- **Simplified Deployment**: Abstracts the complexities of GCP service account creation, key management, and IAM
  bindings into a single, easy-to-use API.
- **Security by Default**: Encourages keyless authentication patterns by making key creation opt-in rather than
  automatic, reducing the risk of credential leaks.
- **Consistency**: Ensures all service accounts adhere to organizational standards and best practices across
  environments.
- **Scalability**: Allows for efficient management of service account configurations as your infrastructure grows,
  supporting both project-level and organization-level permissions.
- **Auditability**: Declarative YAML manifests stored in Git provide a complete audit trail of service account
  configurations and permission changes.
- **Flexibility**: Supports both modern keyless patterns (recommended) and traditional key-based authentication (when
  necessary) without compromising on best practices.
- **Tool Agnostic**: The same manifest deploys with either Pulumi or Terraform, preventing vendor lock-in and supporting
  gradual IaC tool migrations.
