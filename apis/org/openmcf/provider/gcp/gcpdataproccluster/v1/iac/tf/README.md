# GcpDataprocCluster - Terraform Module

Terraform implementation for the GcpDataprocCluster deployment component.

## Provider

Requires Google Cloud provider `~> 6.0`.

## Usage

This module is designed to be called by the OpenMCF framework. Direct usage requires providing the `spec`, `metadata`, and `provider_config` variables matching the protobuf-defined schema.

## Resources Created

- `google_dataproc_cluster` - Managed Dataproc cluster with master nodes, worker nodes, optional secondary (spot) workers, software configuration, and lifecycle management.
