# GcpSslCertificate - Terraform Module

This Terraform module provisions a self-managed Compute Engine SSL certificate (`google_compute_ssl_certificate` when `spec.region` is empty, `google_compute_region_ssl_certificate` when set). It is the Terraform-side implementation of the Planton `GcpSslCertificate` resource kind and has feature parity with the Pulumi module.

## Overview

The module uploads your PEM certificate chain and private key as one SSL certificate that target HTTPS (and SSL) proxies present to clients. Self-managed and Google-managed certificates share one API collection and name namespace — proxies attach both identically — but nothing here renews itself: track the `expire_time` output and rotate.

Every argument is immutable (ForceNew), and GCP blocks deleting a certificate a proxy still references — rotation is create-before-destroy. The private key is write-only in GCP and never surfaced in outputs. The module runs on the plain `google` provider — every modeled field is GA on the released 6.x line.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml` (carries a throwaway self-signed test pair — replace for real use).

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpsslcertificate/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpSslCertificate spec | — |

The `spec` object includes: `certificate` (required PEM chain, leaf first then intermediates), `private_key` (required unencrypted PEM key; secret material), `region` (empty means global), `project_id` (empty falls back to the provider default project), `certificate_name` (empty defaults to `metadata.name`; shared namespace with managed certificates), and optional `description`.

## Outputs

| Name | Description |
|------|-------------|
| `self_link` | Self-link URI — the value a target HTTPS proxy references in `ssl_certificates` |
| `certificate_name` | Name of the SSL certificate in GCP |
| `certificate_id` | Server-assigned numeric ID of the certificate |
| `expire_time` | Expiry in RFC3339, parsed from the uploaded chain |
| `region` | Region of a regional certificate; empty for global |
