# GcpDataprocCluster - Terraform Module

This Terraform module provisions a Dataproc cluster (`google_dataproc_cluster`) — Apache Spark/Hadoop on dedicated Compute Engine VMs or as pods on an existing GKE cluster (Dataproc-on-GKE). It is the Terraform-side implementation of the Planton `GcpDataprocCluster` resource kind and has feature parity with the Pulumi module.

## Overview

The spec carries two mutually exclusive deployment arms — `cluster_config` (GCE) and `virtual_cluster_config` (GKE) — validated pre-deploy and re-checked by a variable validation. Omitting both yields GCP's default GCE cluster. User labels merge beneath Planton's platform attribution labels (platform keys win on conflict); the Dataproc API does not support labels on virtual clusters, and the spec validation rejects that combination before the module runs.

`software_config.properties` maps to the provider's `override_properties` (the API's writable properties surface; the provider's `properties` attribute is the computed resolved set). Kerberos secret fields are GCS URIs of KMS-encrypted files — paths, never inline secret material. The module runs on the plain `google` provider — every modeled field is GA on the released 8.x line.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpdataproccluster/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpDataprocCluster spec | — |

The `spec` object includes: identity (`region`, `cluster_name`, `project_id` — empty rides the provider default), `graceful_decommission_timeout`, `labels`, and one of two arms. The `cluster_config` arm carries buckets, `cluster_tier`, `gce_config` (network XOR subnetwork, service account + scopes, zone, internal-only, tags, metadata, shielded/reservation/sole-tenant/confidential blocks), `master_config` / `worker_config` / `secondary_worker_config` (disks with `local_ssd_interface`, accelerators, the secondary `instance_flexibility_policy` with instance selections and the standard/spot mix), `software_config`, `initialization_actions`, `autoscaling_policy_uri` (resolved policy resource name), `encryption_kms_key_name`, `security_config` (kerberos XOR identity), `endpoint_config`, `lifecycle_config`, `metastore_config`, `dataproc_metric_config`, and `auxiliary_node_groups`. The `virtual_cluster_config` arm carries the staging bucket, `kubernetes_cluster_config` (namespace, `gke_cluster_config` with the resolved GKE cluster name and node-pool targets, `kubernetes_software_config`), and `auxiliary_services_config` (metastore + Spark History Server). All `StringValueOrRef` fields arrive as resolved literal strings.

## Outputs

| Name | Description |
|------|-------------|
| `cluster_id` | Fully qualified cluster resource name (`projects/{p}/regions/{r}/clusters/{c}`) — the composition handle downstream resources reference |
| `cluster_name` | Short name of the cluster |
| `staging_bucket` | Cloud Storage bucket used for staging (user-supplied or auto-created), resolved from whichever arm is active |

## Lifecycle Notes

In-place updates: `labels`, primary/secondary worker counts, `min_num_instances`, the autoscaling-policy attachment, and both lifecycle TTLs. Everything else on the GCE arm is ForceNew; the virtual arm is fully immutable — any change replaces the virtual cluster while leaving the underlying GKE cluster and node pools untouched.
