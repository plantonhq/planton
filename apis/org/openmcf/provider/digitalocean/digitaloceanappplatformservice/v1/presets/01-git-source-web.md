# Git Source Web Service

This preset deploys a web service on DigitalOcean App Platform from a GitHub repository. App Platform automatically builds and deploys the application from source, providing HTTPS, health checks, and zero-downtime deployments out of the box. This is the fastest path from code to production on DigitalOcean.

## When to Use

- Web applications or APIs with source code in a GitHub repository
- Projects where App Platform's auto-build (Buildpacks or Dockerfile) is sufficient
- Teams wanting managed deployment without maintaining CI/CD pipelines

## Key Configuration Choices

- **Git source** (`gitSource`) -- App Platform clones the repo, detects the language/framework, and builds automatically.
- **Web service type** (`serviceType: web_service`) -- receives HTTP traffic via App Platform's built-in load balancer with automatic HTTPS.
- **Basic XXS instance** (`instanceSizeSlug: basic_xxs`) -- smallest tier, suitable for low-traffic applications. Scale up to `professional_xs` or higher for production traffic.
- **Single instance** (`instanceCount: 1`) -- suitable for getting started. Increase for redundancy and load handling.
- **Environment variables** (`env`) -- set runtime configuration. Use App Platform's encrypted secrets for sensitive values.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<your-github-repo-url>` | HTTPS URL of your GitHub repository | Your GitHub repository settings |
| `nyc1` | Target DigitalOcean region slug | [App Platform regions](https://docs.digitalocean.com/products/app-platform/) |

## Related Presets

- **02-container-image** -- Use instead when deploying a pre-built container image from DigitalOcean Container Registry
