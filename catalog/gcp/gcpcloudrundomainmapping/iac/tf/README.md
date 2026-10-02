# GcpCloudRunDomainMapping - Terraform Module

This Terraform module provisions a Cloud Run domain mapping (`google_cloud_run_domain_mapping`) — a verified custom domain pointed directly at a Cloud Run service, with Cloud Run itself provisioning the TLS certificate. It is the Terraform-side implementation of the Planton `GcpCloudRunDomainMapping` resource kind and has feature parity with the Pulumi module.

## Overview

The mapping is fully IMMUTABLE at the provider: every argument is create-only, so any spec change replaces the mapping (cheap — the object is free and re-creates in seconds, with a brief serving gap while the managed certificate re-issues). The domain MUST already be verified by the provisioning identity (Search Console / `gcloud domains verify`) — GCP rejects the create otherwise. After apply, the `resource_records` output carries the DNS records the domain's zone must publish before the domain serves. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpcloudrundomainmapping/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpCloudRunDomainMapping spec | — |

The `spec` object includes: `domain` (the verified domain — the mapping's name in GCP), `region` (must match the target service's region), `route` (the Cloud Run service name; accepts references to a GcpCloudRun resource's `service_name` output), `certificate_mode` (AUTOMATIC/NONE; empty leaves the provider default AUTOMATIC), `force_override` (sent only when true, preserving GCP's safe conflict error), `namespace` (empty falls back to the project ID; GCP requires project ID or number), `labels`/`annotations` (non-authoritative Knative metadata), `project_id` (empty falls back to the provider default project), and `deletion_policy` (DELETE/PREVENT/ABANDON — the one field that updates in place).

## Outputs

| Name | Description |
|------|-------------|
| `domain` | The mapped domain (the mapping's name in GCP) |
| `region` | GCP region the mapping lives in |
| `resource_records` | DNS records the domain's zone must publish (A/AAAA for a root domain, one CNAME for a subdomain) |
| `mapped_route_name` | The Cloud Run route (service) the mapping currently points to |
