# GCP Secrets Manager Pulumi Module

## Overview

The GCP Secrets Manager Pulumi module enables developers to declaratively manage secrets in Google Cloud's Secret Manager using Kubernetes-like API resource definitions. This module is designed to integrate seamlessly with Planton's unified API resources, where developers define infrastructure through YAML specifications. The module takes an API resource as input and provisions Google Cloud secrets based on the defined specifications. Key outputs, such as secret IDs, are captured in `status.outputs`, making it easy to reference and use these secrets across different cloud services or components.

This module allows you to automate the creation of secrets in Google Cloud, initializing each secret with a placeholder version that can be updated with real values post-deployment. The developer can run the module using the `planton pulumi up --stack-input <api-resource.yaml>` command, which automatically sets up the GCP provider using the credentials provided in the API resource, then proceeds to create secrets in the defined Google Cloud project.

## Cross-Resource References

The `project_id` field supports the `StringValueOrRef` type, enabling both literal values and cross-resource references:

### Literal Value
```yaml
spec:
  projectId:
    value: "my-gcp-project-123456"
```

### Reference to Another Resource
```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: main-project
      fieldPath: "status.outputs.project_id"
```

**Current limitation**: Reference resolution (`value_from`) is not yet fully implemented. Only literal `value` is currently used. References will need to be resolved by an external orchestrator or future CLI implementation.

**Future work**: Implement reference resolution in a shared library that all modules can use.

