# Private NPM Repository

This preset creates a private Artifact Registry repository for NPM packages. Use this for hosting internal JavaScript/TypeScript libraries that should not be published to the public NPM registry.

## When to Use

- Internal shared libraries for JavaScript or TypeScript projects
- Monorepo setups where packages are published to a private registry
- Organizations that want to control package distribution within their teams

## Key Configuration Choices

- **NPM format** (`repoFormat: NPM`) -- hosts NPM packages with `npm publish` / `npm install` support
- **Private** (`enablePublicAccess: false`) -- requires IAM authentication; configure `.npmrc` with Artifact Registry credentials
- **Regional** -- place near your CI/CD runners for faster publish/install

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your preferred region |

## Related Presets

- **01-docker-private** -- Use for container image storage
