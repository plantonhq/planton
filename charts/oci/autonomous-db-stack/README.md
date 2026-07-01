# Autonomous Database Stack InfraChart

This chart provisions a **managed Oracle Autonomous Database with private networking** on OCI:

* Private VCN with NAT and service gateways (no internet-facing database)
* Private subnet for the database endpoint
* Network security group allowing SQL*Net (1522) and HTTPS (443) from within the VCN
* Autonomous Database supporting both ATP (OLTP) and ADW (Data Warehouse) workloads
* Optional customer-managed KMS vault and key for encryption at rest

## Resources Created

| Resource | Kind | Condition |
|----------|------|-----------|
| Virtual Cloud Network | `OciVcn` | Always |
| Private Subnet | `OciSubnet` | Always |
| Database NSG | `OciSecurityGroup` | Always |
| Autonomous Database | `OciAutonomousDatabase` | Always |
| KMS Vault | `OciKmsVault` | `enable_encryption` |
| KMS Key | `OciKmsKey` | `enable_encryption` |

## Parameters

| Name | Description | Default |
|------|-------------|---------|
| `compartment_ocid` | OCI compartment OCID | — |
| `vcn_cidr` | VCN CIDR block | `10.0.0.0/16` |
| `subnet_cidr` | Database subnet CIDR | `10.0.1.0/24` |
| `db_name` | Database name | `mydb` |
| `db_workload` | oltp (ATP) or dw (ADW) | `oltp` |
| `compute_count` | ECPU count | `2` |
| `storage_in_tbs` | Storage in TB | `1` |
| `admin_password` | ADMIN password | — |
| `is_free_tier` | Use Always Free tier | `false` |
| `is_auto_scaling_enabled` | Auto-scale compute | `true` |
| `enable_encryption` | Customer-managed KMS | `false` |

## Workload Types

* **oltp** -- Autonomous Transaction Processing (ATP): designed for OLTP, mixed workloads, and JSON document storage
* **dw** -- Autonomous Data Warehouse (ADW): designed for analytics, data warehousing, and data lake queries
