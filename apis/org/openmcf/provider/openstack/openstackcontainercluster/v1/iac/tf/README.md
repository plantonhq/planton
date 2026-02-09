# OpenStackContainerCluster Terraform Module

This directory contains the Terraform HCL module for provisioning OpenStack Magnum container clusters.

## Structure

```
iac/tf/
├── variables.tf    # Input variables (metadata + spec)
├── locals.tf       # FK extraction and label computation
├── main.tf         # Magnum cluster resource
├── outputs.tf      # Stack outputs
├── provider.tf     # OpenStack provider config
└── README.md       # This file
```

## Resource Mapping

| Terraform Resource | Count | Description |
|---|---|---|
| `openstack_containerinfra_cluster_v1` | 1 | Magnum container cluster |

## Sensitive Outputs

The following outputs are marked as `sensitive = true`:
- `kubeconfig_raw` — Full kubeconfig YAML
- `kubeconfig_cluster_ca_cert` — Cluster CA certificate
- `kubeconfig_client_cert` — Client certificate
- `kubeconfig_client_key` — Client private key

## Key Design Notes

- **Single resource**: Creates one `openstack_containerinfra_cluster_v1` (Magnum cluster).
- **FK extraction**: `cluster_template` is a required FK resolved in `locals.tf`; `keypair` is optional.
- **ForceNew fields**: Almost all fields are ForceNew. Only `node_count` (scale) and `cluster_template_id` (upgrade) can be updated after creation.
- **Sensitive kubeconfig**: Kubeconfig-related outputs are marked as sensitive in Terraform.

## Usage

This module is invoked by the OpenMCF CLI's Terraform runner. It is not intended for standalone use.

```bash
# Variables are passed as a JSON file by the runner
terraform apply -var-file=terraform.tfvars.json
```
