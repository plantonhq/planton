# Container Image Service

This preset deploys a web service on DigitalOcean App Platform from a pre-built container image stored in DigitalOcean Container Registry (DOCR). It uses a professional-tier instance with autoscaling, suitable for production workloads built in an external CI/CD pipeline.

## When to Use

- Applications with their own CI/CD pipeline that produces container images
- Teams pushing images to DigitalOcean Container Registry
- Production workloads requiring autoscaling and professional-tier compute

## Key Configuration Choices

- **Container image source** (`imageSource`) -- deploys from DOCR rather than building from source. Gives full control over the build process.
- **Professional XS instance** (`instanceSizeSlug: professional_xs`) -- dedicated CPU and memory suitable for production traffic.
- **Autoscaling** (`enableAutoscale: true`, 2-5 instances) -- automatically scales based on traffic. Minimum of 2 for redundancy.
- **Registry reference** (`registry`) -- uses `StringValueOrRef` to reference the DOCR URL. The system resolves credentials automatically.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<registry-url>` | DigitalOcean Container Registry URL | `DigitalOceanContainerRegistry` status outputs |
| `<your-image-repository>` | Image repository name (e.g., `myapp/backend`) | Your container registry |
| `nyc1` | Target DigitalOcean region slug | [App Platform regions](https://docs.digitalocean.com/products/app-platform/) |

## Related Presets

- **01-git-source-web** -- Use instead when deploying directly from a GitHub repository with auto-build
