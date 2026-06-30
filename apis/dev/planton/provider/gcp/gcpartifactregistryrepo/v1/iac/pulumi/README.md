# GCP Artifact Registry Pulumi Module

This Pulumi module simplifies the provisioning of Google Cloud Artifact Registry repositories, including Docker, Maven,
NPM, and Python repositories. It enables developers to define repositories and their access configurations through a
standardized API resource, automating the setup process and ensuring consistency across environments.

## Key Features

- **Standardized API Resource**: Utilizes a consistent API structure with `apiVersion`, `kind`, `metadata`, `spec`, and
  `status`, making resource definitions straightforward.

- **Multiple Repository Formats**: Supports the creation of Docker, Maven, NPM, and Python repositories within Google
  Cloud Artifact Registry.

- **Service Account Management**: Automatically creates reader and writer service accounts with appropriate permissions
  for each repository.

- **Access Control**:
    - **Internal Repositories**: Restrict access to authenticated users and service accounts.
    - **External Repositories**: Optionally allow unauthenticated (public) access to repositories, suitable for
      open-source projects.

- **Pulumi Integration**: Written in Golang, leveraging Pulumi for infrastructure as code, enabling seamless integration
  into CI/CD pipelines and existing workflows.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
    - [Sample YAML Configuration](#sample-yaml-configuration)
    - [Deploying with CLI](#deploying-with-cli)
- [Module Components](#module-components)
    - [Service Accounts](#service-accounts)
    - [Repository Creation](#repository-creation)
        - [Docker Repository](#docker-repository)
        - [Maven Repository](#maven-repository)
        - [NPM Repository](#npm-repository)
        - [Python Repository](#python-repository)
- [Outputs](#outputs)
- [Contributing](#contributing)
- [License](#license)

## Prerequisites

- Google Cloud account with necessary permissions.
- Pulumi CLI installed.
- Golang environment set up if modifying the module code.
- Google Cloud SDK installed and configured.

## Installation

Clone the repository containing the Pulumi module:

```bash
git clone https://github.com/your-org/gcp-artifact-registry-pulumi-module.git
```

Install the required dependencies:

```bash
cd gcp-artifact-registry-pulumi-module
go mod download
```

## Usage

Refer to [example](example.md) for usage instructions.

## Module Components

### Service Accounts

The module creates two service accounts:

- **Reader Service Account**: Has read-only access to the repositories.
- **Writer Service Account**: Has read and write access, including administrative permissions.

Each service account is provisioned with a key, and their credentials are exported as outputs for use in CI/CD pipelines
or other services.

### Repository Creation

The module creates repositories for Docker, Maven, NPM, and Python artifacts. For each repository, it sets up the
necessary permissions for the reader and writer service accounts.

#### Docker Repository

- **Name**: `<metadata.id>-docker`
- **Format**: Docker
- **Permissions**:
    - **Reader Service Account**: `roles/artifactregistry.reader`
    - **Writer Service Account**: `roles/artifactregistry.writer`, `roles/artifactregistry.repoAdmin`
- **Public Access**: If `is_external` is `true`, grants `roles/artifactregistry.reader` to `allUsers`.

#### Maven Repository

- **Name**: `<metadata.id>-maven`
- **Format**: Maven
- **Permissions**: Same as Docker repository.

#### NPM Repository

- **Name**: `<metadata.id>-npm`
- **Format**: NPM
- **Permissions**: Same as Docker repository.

#### Python Repository

- **Name**: `<metadata.id>-python`
- **Format**: Python
- **Permissions**: Same as Docker repository.

### Access Control

- **Internal Repositories**: When `is_external` is `false`, only the reader and writer service accounts have access.
- **External Repositories**: When `is_external` is `true`, `allUsers` have read access to the repositories.

## Outputs

After deployment, the module provides several outputs:

- **Reader Service Account Email**: `reader_service_account_email`
- **Reader Service Account Key (Base64 Encoded)**: `reader_service_account_key_base64`
- **Writer Service Account Email**: `writer_service_account_email`
- **Writer Service Account Key (Base64 Encoded)**: `writer_service_account_key_base64`
- **Docker Repository Name**: `docker_repo_name`
- **Docker Repository Hostname**: `docker_repo_hostname`
- **Docker Repository URL**: `docker_repo_url`
- **Maven Repository Name**: `maven_repo_name`
- **Maven Repository URL**: `maven_repo_url`
- **NPM Repository Name**: `npm_repo_name`
- **Python Repository Name**: `python_repo_name`

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the [MIT License](LICENSE).
