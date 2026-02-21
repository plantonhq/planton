# AliCloudContainerRegistry

Deploys an Alibaba Cloud Container Registry (ACR) Enterprise Edition instance with optional namespaces.

## Overview

AliCloudContainerRegistry provisions a managed container image registry on Alibaba Cloud. The Enterprise Edition offers three tiers -- Basic, Standard, and Advanced -- providing scalable storage, image scanning, and multi-region replication capabilities.

Namespaces are bundled into the component because a registry instance without namespaces cannot store images. Namespaces organize repositories by team or application (e.g., `platform`, `frontend`, `backend`).

## Provider Resources

| Resource | Terraform | Pulumi |
|----------|-----------|--------|
| Registry Instance | `alicloud_cr_ee_instance` | `cr.RegistryEnterpriseInstance` |
| Namespace | `alicloud_cr_ee_namespace` | `cs.RegistryEnterpriseNamespace` |

## Key Features

- **Three tier options**: Basic (individuals/small teams), Standard (SMEs), Advanced (large enterprises)
- **Bundled namespaces**: Create organizational namespaces in the same deployment
- **Auto-create repositories**: Optionally auto-create repos when pushing new images
- **VPC endpoint**: Pull images from within VPC without internet egress
- **Flexible billing**: Subscription (pre-paid) or PayAsYouGo (post-paid)

## Outputs

| Output | Description |
|--------|-------------|
| `instance_id` | ACR instance ID |
| `instance_name` | Instance name |
| `public_endpoint` | Internet-facing registry domain for `docker login` |
| `vpc_endpoint` | VPC-internal registry domain for in-VPC image pulls |
| `namespace_ids` | Map of namespace names to IDs |

## Quick Start

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudContainerRegistry
metadata:
  name: my-registry
spec:
  region: cn-hangzhou
  instanceName: my-acr
  instanceType: Standard
  paymentType: Subscription
  period: 1
  namespaces:
    - name: platform
      autoCreate: true
      defaultVisibility: PRIVATE
```

## Notes

- ACR Enterprise Edition instances do not support tags (provider limitation for BSS-provisioned resources)
- The `instance_type` and `payment_type` fields are immutable after creation
- The `public_endpoint` and `vpc_endpoint` outputs are extracted from the computed `instance_endpoints` attribute
