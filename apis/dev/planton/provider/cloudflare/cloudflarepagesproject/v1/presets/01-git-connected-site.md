# Preset: Git-Connected Site

A Pages project connected to a GitHub (or GitLab) repository. Cloudflare runs
your build and creates a new deployment on every push — Cloudflare is the CI.

## When to use

- You want push-to-deploy without running builds yourself.
- You want automatic preview deployments for branches/PRs.

## Key choices

- `source.type` / `source.config`: the connected repo, production branch, and
  preview policy. **Prerequisite:** authorize Cloudflare with your git provider
  once in the Cloudflare dashboard (the provider can't bootstrap the OAuth/App
  connection).
- `buildConfig`: the build command and output directory Cloudflare runs.
- `deploymentConfigs.production`: runtime config (mirrored to preview unless you
  also set `preview`).

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<github-owner>` | Repository owner (user or org) |
| `<repo-name>` | Repository name |
