# Terraform Module to Deploy AliCloudCdnDomain

This Terraform module provisions an Alibaba Cloud CDN accelerated domain from
an Planton manifest. It creates a single `alicloud_cdn_domain_new` resource
with origin sources, optional HTTPS certificate configuration, and
metadata-derived tags.

## CLI Usage

All commands use the `planton tofu` CLI. The `--manifest` flag points to a
YAML manifest, and `--module-dir` specifies the directory containing the
Terraform module.

### Plan

```shell
planton tofu plan \
  --manifest ../hack/manifest.yaml \
  --module-dir .
```

### Apply

```shell
planton tofu apply \
  --manifest ../hack/manifest.yaml \
  --module-dir .
```

### Refresh State

Detect drift between the Terraform state and the actual Alibaba Cloud resources:

```shell
planton tofu plan \
  --manifest ../hack/manifest.yaml \
  --module-dir . \
  --refresh-only
```

### Destroy

Remove all resources managed by this module:

```shell
planton tofu destroy \
  --manifest ../hack/manifest.yaml \
  --module-dir .
```

## Credentials

The module requires Alibaba Cloud credentials. Set the following environment
variables before running any command:

```shell
export ALICLOUD_ACCESS_KEY="<your-access-key>"
export ALICLOUD_SECRET_KEY="<your-secret-key>"
```

The module configures the alicloud provider using `spec.region` from the
manifest, so no additional region configuration is required.

## Module Files

| File | Purpose |
|------|---------|
| `provider.tf` | Configures the `aliyun/alicloud` provider (`~> 1.200`). |
| `variables.tf` | Defines `metadata` and `spec` input variables with validation. |
| `locals.tf` | Computes final tags from metadata fields and user-defined tags. |
| `main.tf` | Creates `alicloud_cdn_domain_new` with dynamic source and certificate blocks. |
| `outputs.tf` | Exports `domain_name`, `cname`, and `status`. |

## Further Reading

- [Examples](./examples.md) — progressive deployment examples.
- [Hack Manifest](../hack/manifest.yaml) — minimal test manifest.
- [AliCloudCdnDomain Overview](../../../README.md) — full field reference.
- [Research Document](../../docs/README.md) — design rationale.
