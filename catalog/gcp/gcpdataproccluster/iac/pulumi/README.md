# GcpDataprocCluster - Pulumi Module

This Pulumi (Go) module provisions a Dataproc cluster (`dataproc.Cluster`) — Apache Spark/Hadoop on dedicated Compute Engine VMs or as pods on an existing GKE cluster (Dataproc-on-GKE). It is the Pulumi-side implementation of the Planton `GcpDataprocCluster` resource kind and has feature parity with the Terraform module.

## Overview

The spec carries two mutually exclusive deployment arms — `clusterConfig` (GCE) and `virtualClusterConfig` (GKE) — validated pre-deploy. Omitting both yields GCP's default GCE cluster. User labels merge beneath Planton's platform attribution labels (platform keys win on conflict — identical merge order to the Terraform module); the Dataproc API does not support labels on virtual clusters, and the spec validation rejects that combination before the module runs.

Software properties map to the SDK's `OverrideProperties` (the API's writable properties surface). Kerberos secret fields are GCS URIs of KMS-encrypted files — paths, never inline secret material. All `StringValueOrRef` fields arrive resolved to literal strings.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpdataproccluster/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — metadata-derived values and the label merge
- `module/dataproc_cluster.go` — both deployment arms mapped to `dataproc.Cluster` args + the staging-bucket export resolved from whichever arm is active
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `cluster_id` | Fully qualified cluster resource name (`projects/{p}/regions/{r}/clusters/{c}`) — the composition handle downstream resources reference |
| `cluster_name` | Short name of the cluster |
| `staging_bucket` | Cloud Storage bucket used for staging (user-supplied or auto-created), resolved from whichever arm is active |

## Lifecycle Notes

In-place updates: `labels`, primary/secondary worker counts, `minNumInstances`, the autoscaling-policy attachment, and both lifecycle TTLs. Everything else on the GCE arm is ForceNew; the virtual arm is fully immutable — any change replaces the virtual cluster while leaving the underlying GKE cluster and node pools untouched.
