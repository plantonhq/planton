# GCP Secrets Manager Pulumi Module

## Overview

The GCP Secrets Manager Pulumi module enables developers to declaratively manage secrets in Google Cloud's Secret Manager using Kubernetes-like API resource definitions. This module is designed to integrate seamlessly with Planton's unified API resources, where developers define infrastructure through YAML specifications. The module takes an API resource as input and provisions Google Cloud secrets based on the defined specifications. Key outputs, such as secret IDs, are captured in `status.outputs`, making it easy to reference and use these secrets across different cloud services or components.

This module allows you to automate the creation of secrets in Google Cloud, initializing each secret with a placeholder version that can be updated with real values post-deployment. The developer can run the module using the `planton pulumi up --stack-input <api-resource.yaml>` command, which automatically sets up the GCP provider using the credentials provided in the API resource, then proceeds to create secrets in the defined Google Cloud project.

## Key Features

- **Unified API Resource Model:** The module follows a standardized Kubernetes-like structure, making it easy to define and deploy secrets across environments.
- **Automated GCP Provider Setup:** The module configures Google Cloud credentials automatically based on the API resource's input, ensuring seamless integration with the GCP Secrets Manager service.
- **Declarative Secret Creation:** Secrets are defined in the specification, allowing you to manage secret infrastructure consistently using the API resource.
- **Placeholder Secret Versions:** Each secret is initialized with a placeholder value, making it easy to later update the secret's actual data.
- **Captured Outputs in `status.outputs`:** The module captures the secret IDs and other metadata, storing them in the stack outputs for easy retrieval and referencing.
- **Scalable and Flexible:** This module is ideal for teams looking to manage secrets at scale across multiple projects and environments with minimal manual intervention.

## Usage

Refer to the example section for usage instructions.

## Important Note

If the API resource specification is empty, this module is not completely implemented. Ensure the specification is populated correctly to take full advantage of the module's capabilities.
