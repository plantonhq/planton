# GcpManagedSslCertificate - Terraform Module

This Terraform module provisions a Google-managed SSL certificate (`google_compute_managed_ssl_certificate`). It is the Terraform-side implementation of the Planton `GcpManagedSslCertificate` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one global Google-managed SSL certificate for the domains you specify. Google provisions and renews the certificate once DNS for each domain points at the load balancer's IP. Reference the certificate's `self_link` from a target HTTPS proxy to terminate TLS.

The whole resource is immutable (name and domains are ForceNew). Because a cert attached to a proxy cannot be deleted, rotate create-before-destroy. The module runs on the plain `google` provider — every modeled field is GA on the released 6.x line.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpmanagedsslcertificate/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpManagedSslCertificate spec | — |

The `spec` object includes: `domains` (required, 1-100 FQDNs), `project_id` (empty falls back to the provider default project), `certificate_name` (empty defaults to `metadata.name`), and optional `description`.

## Outputs

| Name | Description |
|------|-------------|
| `self_link` | Self-link URI — the value a target HTTPS proxy references in `ssl_certificates` |
| `certificate_name` | Name of the SSL certificate in GCP |
| `certificate_id` | Server-assigned numeric ID of the certificate |
| `expire_time` | Expiry time in RFC3339 format; empty until provisioning completes |
