# Terraform Module to Deploy AwsS3Bucket

This module provisions an Amazon S3 bucket and its complete behavioral surface: the bucket root resource plus one satellite resource per configured spec block — public-access block, ownership controls, optional canned ACL, bucket policy, versioning, default encryption, lifecycle configuration, replication, website hosting, access logging, CORS, event notifications, Object Lock default retention, transfer acceleration, requester pays, and per-name Intelligent-Tiering archive configurations.

The security posture is explicit: the public-access block and ownership controls satellites are always created (fully-private / ACLs-disabled unless the spec relaxes them), so the bucket's posture is visible in state rather than implied by absence.

`variables.tf` is generated from the proto schema for `AwsS3Bucket` — do not edit it by hand.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For a working example, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml) and the presets in [`../../presets/`](../../presets/).
