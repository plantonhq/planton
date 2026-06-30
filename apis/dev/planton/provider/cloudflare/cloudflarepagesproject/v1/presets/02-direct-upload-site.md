# Preset: Direct-Upload Site

A Pages project with no git connection. You build the site in your own CI and
push it with `wrangler pages deploy`, while the project, its bindings, and its
domains are managed declaratively here.

## When to use

- You already build in your own pipeline and want to upload the artifact.
- You want Planton to own the project/bindings while your CI owns the upload.

## Key choices

- No `source`: this is a direct-upload project. Deploy versions with
  `wrangler pages deploy ./dist --project-name=direct-upload-site`.
- `deploymentConfigs.production`: runtime config and bindings (mirrored to
  preview unless you also set `preview`). The KV binding shows the `valueFrom`
  composition pattern.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
