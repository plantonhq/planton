# Terraform Module to Deploy AwsCloudFront

This module provisions an Amazon CloudFront distribution at the full provider
surface — origins with S3/custom/VPC arms, primary/failover origin groups,
default + path-matched ordered cache behaviors across both caching
generations, custom domains, error pages, geo restrictions, and access logs —
plus the folded satellites: per-origin Origin Access Controls (for S3 origins
that ask for one) and the CloudWatch additional-metrics monitoring
subscription.

Generated `variables.tf` reflects the proto schema for `AwsCloudFront`
(generator-owned; regenerate with the variables.tf drift test, never
hand-edit).

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

Deploys propagate to every CloudFront edge location, so expect apply/destroy
to take 5-15 minutes each while the distribution converges (and destroy first
disables the distribution, which is itself a propagation).

For more examples, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml) and the
component presets.
