# tofuannotationkeys

This package builds the manifest **annotation** keys that carry Terraform/OpenTofu
backend state configuration in Planton resource manifests.

## Why annotations, not labels

`metadata.labels` are derived into cloud-provider tags by planton IaC modules, so a
platform key there would leak internal configuration onto the user's real provider
resources. Platform-behavior signals — including backend configuration — therefore
live in `metadata.annotations`, which never touch the cloud.

## Provisioner-aware keys

Backend annotation keys are provisioner-aware. Each function takes the provisioner
("terraform" or "tofu") and returns the fully-prefixed key:

| Function | Example ("tofu") |
|----------|------------------|
| `BackendTypeAnnotationKey(p)` | `tofu.planton.dev/backend.type` |
| `BackendBucketAnnotationKey(p)` | `tofu.planton.dev/backend.bucket` |
| `BackendKeyAnnotationKey(p)` | `tofu.planton.dev/backend.key` |
| `BackendRegionAnnotationKey(p)` | `tofu.planton.dev/backend.region` |
| `BackendEndpointAnnotationKey(p)` | `tofu.planton.dev/backend.endpoint` |

There is no cross-prefix fallback: a manifest deployed with the `tofu` provisioner
must use `tofu.*` keys, and one deployed with `terraform` must use `terraform.*` keys.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsVpc
metadata:
  name: my-vpc
  annotations:
    planton.dev/provisioner: tofu
    tofu.planton.dev/backend.type: s3
    tofu.planton.dev/backend.bucket: my-state-bucket
    tofu.planton.dev/backend.key: aws-vpc/prod/terraform.tfstate
    tofu.planton.dev/backend.region: us-west-2
spec:
  region: us-west-2
```

For S3-compatible backends (Cloudflare R2, MinIO), set `backend.region: auto` and
provide `backend.endpoint`; the consumer (`pkg/iac/tofu/backendconfig`) detects the
combination and applies the required skip flags automatically.
