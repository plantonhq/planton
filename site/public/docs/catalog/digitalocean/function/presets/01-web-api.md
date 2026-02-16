---
title: "Web API Function"
description: "This preset creates a serverless HTTP function on DigitalOcean, deployed via App Platform from a GitHub repository. The function is exposed as a web endpoint with auto-deploy on push, suitable for..."
type: "preset"
rank: "01"
presetSlug: "01-web-api"
componentSlug: "function"
componentTitle: "Function"
provider: "digitalocean"
icon: "package"
order: 1
---

# Web API Function

This preset creates a serverless HTTP function on DigitalOcean, deployed via App Platform from a GitHub repository. The function is exposed as a web endpoint with auto-deploy on push, suitable for API handlers, webhooks, and lightweight HTTP services.

## When to Use

- Lightweight API endpoints or webhook handlers
- Event-driven HTTP services that don't need always-on compute
- Rapid prototyping with automatic deployment from GitHub

## Key Configuration Choices

- **Node.js 20 runtime** (`runtime: nodejs_20`) -- current LTS. Change to `python_311`, `go_121`, or `php_82` as needed.
- **Web endpoint** (`isWeb: true`) -- exposed as an HTTP endpoint via App Platform's built-in routing.
- **Auto-deploy** (`deployOnPush: true`) -- pushes to the configured branch trigger automatic redeployment.
- **256 MB memory** (`memoryMb: 256`) -- recommended default, sufficient for most API handlers. Scale to 512 or 1024 for heavier workloads.
- **3-second timeout** (`timeoutMs: 3000`) -- suitable for fast API responses. Increase up to 300,000 ms (5 min) for longer operations.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<your-org>/<your-repo>` | GitHub repository in `owner/repo` format | Your GitHub repository |
| `/functions/api` | Path within the repo containing function code | Your repository structure |
| `nyc1` | Target DigitalOcean region slug | [App Platform regions](https://docs.digitalocean.com/products/app-platform/) |

## Related Presets

- **02-scheduled-job** -- Use instead for background jobs triggered by a cron schedule rather than HTTP requests
