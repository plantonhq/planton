# Terraform Module to Deploy AwsNetworkLoadBalancer

This module provisions an AWS Network Load Balancer with listeners, target
groups, and optional Route53 DNS records. The NLB operates at Layer 4
(TCP/UDP/TLS) and supports static IP addresses via Elastic IP allocation.

Generated `variables.tf` reflects the proto schema for `AwsNetworkLoadBalancer`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest hack/manifest.yaml
planton tofu plan --manifest hack/manifest.yaml
planton tofu apply --manifest hack/manifest.yaml --auto-approve
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For more examples, see [`examples.md`](../../examples.md) and [`hack/manifest.yaml`](../hack/manifest.yaml).
