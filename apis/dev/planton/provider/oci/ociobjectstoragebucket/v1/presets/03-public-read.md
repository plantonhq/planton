# Public Read

This preset creates a public-read Object Storage bucket for serving static assets directly to end users. Individual objects can be accessed via URL without authentication, but the bucket contents cannot be listed, preventing directory enumeration. Object events are enabled for download tracking.

## When to Use

- Static website assets (CSS, JavaScript, images, fonts) served via CDN or direct URL
- Public file downloads (documentation PDFs, software releases, open datasets)
- Shared assets across applications where authenticated access adds unnecessary complexity
- Any content intended for anonymous public consumption

## Key Configuration Choices

- **Public read without list** (`accessType: object_read_without_list`) -- anonymous users can read individual objects by exact URL but cannot list bucket contents. This prevents directory browsing while allowing direct object access. Use `object_read` instead if listing is acceptable.
- **Standard storage tier** -- public assets are typically accessed frequently, making Standard tier the right choice for low-latency retrieval.
- **Versioning disabled** -- public asset buckets typically use a deploy-and-replace workflow. Versioning is unnecessary when old assets are simply overwritten with new versions.
- **Object events enabled** (`objectEventsEnabled: true`) -- emits events for object uploads, downloads, and deletes via OCI Events. Useful for tracking asset usage and triggering CDN cache invalidation.
- **No KMS encryption** -- public assets do not need customer-managed encryption. Oracle-managed encryption is used by default (data is still encrypted at rest).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the bucket | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<object-storage-namespace>` | Object Storage namespace for your tenancy | `oci os ns get` CLI command, or OCI Console > Object Storage |
| `<bucket-name>` | Globally unique bucket name within the namespace | Choose a name (e.g., `myapp-static-assets`) |

## Related Presets

- **01-private-versioned** -- Use instead for application data that should not be publicly accessible
- **02-archive-storage** -- Use instead for long-term retention data that is rarely accessed
