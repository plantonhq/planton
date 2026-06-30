# Pulumi Module to Deploy AliCloudCdnDomain

This Pulumi module provisions an Alibaba Cloud CDN accelerated domain from an
Planton manifest. It creates the CDN domain resource with origin sources,
optional HTTPS certificate configuration, and metadata-derived tags.

## CLI Usage

All commands use the `planton pulumi` CLI. The `--manifest` flag points to a
YAML manifest, `--stack` identifies the Pulumi stack, and `--module-dir`
specifies the directory containing the Pulumi program.

### Preview Changes

```shell
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

### Deploy

```shell
planton pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

### Refresh State

Detect drift between the Pulumi state and the actual Alibaba Cloud resources:

```shell
planton pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

### Destroy

Remove all resources managed by this stack:

```shell
planton pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Credentials

The module requires Alibaba Cloud credentials. Set the following environment
variables before running any command:

```shell
export ALICLOUD_ACCESS_KEY="<your-access-key>"
export ALICLOUD_SECRET_KEY="<your-secret-key>"
```

The module creates an explicit alicloud provider using `spec.region` from the
manifest, so no additional region configuration is needed.

## Debugging

If a deployment fails:

1. Run `preview` first to see the planned changes without modifying resources.
2. Check the Pulumi output for error messages from the Alibaba Cloud API.
3. Common errors:
   - **InvalidDomainName** — the domain name format is invalid or exceeds 63 characters.
   - **DomainAlreadyExist** — the domain is already registered in another CDN account.
   - **InvalidCertificate** — the certificate ID does not exist or the PEM content is malformed.
   - **IcpBlack** — the domain lacks a valid ICP filing for `domestic` or `global` scope.
4. Use `refresh` to sync state if manual changes were made in the console.

## Module Overview

| File | Purpose |
|------|---------|
| `main.go` | Entrypoint — loads stack input and calls the module. |
| `module/main.go` | Creates the alicloud provider and CDN domain resource. |
| `module/locals.go` | Computes tags from metadata and user-defined tags. |
| `module/outputs.go` | Defines output key constants (`domain_name`, `cname`, `status`). |
| `Pulumi.yaml` | Pulumi project configuration (Go runtime). |

## Further Reading

- [AliCloudCdnDomain Overview](../../../README.md)
- [Examples](./examples.md)
- [Module Architecture](./overview.md)
- [Research Document](../../docs/README.md)
- [Hack Manifest](../hack/manifest.yaml)
