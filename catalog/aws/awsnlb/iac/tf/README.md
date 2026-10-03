# Terraform Module to Deploy AwsNlb

This module provisions an AWS Network Load Balancer and, when DNS is
enabled, Route53 alias A records for each hostname. It creates no listeners
or target groups -- routing attaches to the NLB by ARN through the
`AwsLbListener` and `AwsLbTargetGroup` components.

The module owns what is load-balancer-wide: subnet mappings (each pinning
one node to a subnet, optionally with an Elastic IP for internet-facing
NLBs or a fixed private IPv4 address for internal ones), optional security
groups (a one-way door -- the last group can never be removed), and the NLB
attributes (cross-zone load balancing, DNS client routing policy, zonal
shift, PrivateLink security-group enforcement, and TLS-only access logs to
S3). Only explicitly set attributes are sent to AWS, so everything left
unset keeps its AWS default. Names longer than AWS's 32-character limit are
truncated deterministically.

Generated `variables.tf` reflects the proto schema for `AwsNlb`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For a working example, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml).
