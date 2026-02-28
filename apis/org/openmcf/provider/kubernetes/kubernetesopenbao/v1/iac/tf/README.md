# KubernetesOpenBao Terraform Module

This Terraform module deploys OpenBao on Kubernetes using the official OpenBao Helm chart.

## Overview

OpenBao is an open-source secrets management solution forked from HashiCorp Vault. This module supports:

- Standalone and High Availability (HA) deployment modes
- Raft integrated storage for HA deployments
- Auto-unseal via GCP Cloud KMS, AWS KMS, Azure Key Vault, or Transit seal
- Optional Agent Injector for automatic secret injection
- Ingress configuration for external access
- UI enablement

## Usage

```hcl
module "openbao" {
  source = "path/to/module"

  metadata = {
    name = "my-openbao"
  }

  spec = {
    namespace        = "openbao"
    create_namespace = true

    server_container = {
      replicas          = 1
      data_storage_size = "10Gi"
      resources = {
        limits = {
          cpu    = "500m"
          memory = "256Mi"
        }
        requests = {
          cpu    = "100m"
          memory = "128Mi"
        }
      }
    }

    ui_enabled = true
  }
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| kubernetes | >= 2.0 |
| helm | >= 2.0 |

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata (name, id, org, env, labels) | object | yes |
| spec | OpenBao deployment specification | object | yes |

### spec Object

| Attribute | Description | Type | Default |
|-----------|-------------|------|---------|
| namespace | Kubernetes namespace | string | - |
| create_namespace | Create namespace if not exists | bool | - |
| helm_chart_version | Helm chart version override | string | "0.23.3" |
| server_container | Server container configuration | object | - |
| high_availability | HA configuration | object | null |
| ingress | Ingress configuration | object | null |
| ui_enabled | Enable OpenBao UI | bool | true |
| injector | Agent Injector configuration | object | null |
| tls_enabled | Enable TLS | bool | false |
| auto_unseal | Auto-unseal configuration | object | null |
| auto_unseal.gcp_kms | GCP Cloud KMS seal (project, region, key_ring, crypto_key, workload_identity_service_account) | object | null |
| auto_unseal.aws_kms | AWS KMS seal (region, kms_key_id, credentials_secret_name) | object | null |
| auto_unseal.azure_key_vault | Azure Key Vault seal (vault_name, key_name, tenant_id, credentials_secret_name) | object | null |
| auto_unseal.transit | Transit seal (address, key_name, mount_path, token_secret_name) | object | null |

## Outputs

| Name | Description |
|------|-------------|
| namespace | Kubernetes namespace |
| service | Kubernetes service name |
| port_forward_command | Port-forward command for local access |
| kube_endpoint | Internal Kubernetes endpoint |
| external_hostname | External hostname (when ingress enabled) |
| api_address | Full API address |
| cluster_address | Cluster communication address (HA mode) |
| ha_enabled | HA mode status |

## Post-Deployment

After deploying OpenBao, you need to:

1. **Initialize OpenBao**:
   ```bash
   kubectl exec -it <pod-name> -n <namespace> -- bao operator init
   ```

2. **Unseal OpenBao** (Shamir only -- repeat with threshold keys). If `auto_unseal` is configured, this step is handled automatically by the KMS provider on every pod startup:
   ```bash
   kubectl exec -it <pod-name> -n <namespace> -- bao operator unseal
   ```

3. **Login with root token**:
   ```bash
   kubectl exec -it <pod-name> -n <namespace> -- bao login <root-token>
   ```

## License

MIT
