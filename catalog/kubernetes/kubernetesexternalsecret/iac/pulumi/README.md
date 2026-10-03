# KubernetesExternalSecret Pulumi Module

## Module Behavior

- **Typed-to-CRD rendering**: `spec_builder.go` renders the typed spec into
  the `external-secrets.io/v1` ExternalSecret CRD shape, applied as an
  untyped CustomResource — the same posture as the store kinds and the
  cert-manager family. ESO's validating webhook checks the applied spec
  strictly and the kind-cluster E2E lanes verify the full sync loop live,
  so shape errors fail loudly without typed args.
- **Pinned Secret name**: the CR's `target.name` is always rendered from
  the resolved Secret name (`target.name` when set, else `metadata.name`),
  so the exported `secret_name` output can never drift from what the
  operator creates.
- **No credential materialization**: the sync declaration carries no
  credentials — authentication lives on the store kinds.
- **Never waits for the sync**: the materialized Secret appears when the
  operator reaches the backend, which is not part of applying the resource.
  The E2E verifier (not the module) asserts synced state.

## Usage

```bash
export IAC_INPUT=$(cat ../../e2e/manifest.yaml | base64)
pulumi up
```

## Local Development

```bash
make deps
make build
```

## Debug

```bash
bash debug.sh ../../e2e/manifest.yaml
```
