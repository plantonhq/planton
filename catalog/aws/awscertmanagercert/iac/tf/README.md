# Terraform Module to Deploy AwsCertManagerCert

This module provisions an AWS Certificate Manager (ACM) certificate in any of
ACM's three creation modes — requested (Amazon-issued, DNS or EMAIL validated),
imported (bring-your-own PEM material), or private (issued by an ACM-PCA
authority). For DNS-validated certificates with a managed Route53 zone it also
creates the validation CNAME records and (by default) waits for issuance.

Generated `variables.tf` reflects the proto schema for `AwsCertManagerCert`
(generator-owned; regenerate with the variables.tf drift test, never hand-edit).

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For more examples, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml) and the
kind's presets.
