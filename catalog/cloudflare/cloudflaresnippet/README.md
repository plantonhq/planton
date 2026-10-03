# Cloudflare Snippet

## Overview

`CloudflareSnippet` is a small JavaScript module deployed to the zone's edge, invoked on requests that match snippet rules (managed separately by `CloudflareSnippetRules`). Snippets are the lightweight sibling of Workers -- same runtime, no bindings, sized for header rewrites, redirects, and request/response touch-ups.

The snippet NAME is the identity. Cloudflare's create call is an upsert, so deploying a snippet whose name already exists in the zone silently adopts and overwrites it. Pick names deliberately, and treat renaming as delete-and-recreate.

## Key Features

- **Name is identity** -- create is an upsert; a name collision silently overwrites the existing snippet
- **Inline files** -- source travels in the manifest as file content strings; `main_module` must name one of those files
- **Byte-stable content** -- Cloudflare re-serves stored content on reads (rebuilt from multipart); formatting churn reads back as drift
- **Real delete** -- destroy removes the snippet; rules that referenced the name keep pointing at whatever the old name now holds (nothing)

## Use Cases

**Ideal for:**

- A 302 from a legacy path
- A header rewrite or request/response touch-up too small for a Worker
- Multi-file snippets that import siblings by name

**Not ideal for:**

- Bindings, cron, or custom domains -- that is `CloudflareWorker`
- The routing table that invokes the snippet -- that is `CloudflareSnippetRules`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone the snippet is deployed to. Can reference a `CloudflareDnsZone` via `value_from` (defaults to `status.outputs.zone_id`). |
| `snippet_name` | string | Yes | Identity within the zone. Letters, numbers, and underscores only. Changing it replaces the snippet. |
| `files` | list of `{name, content}` | Yes | At least one source file. `content` must be byte-stable. |
| `main_module` | string | Yes | The entry file -- must name one of `files`. The provider argument is `metadata`. |

### Outputs

| Field | Description |
|-------|-------------|
| `snippet_name` | The snippet's name -- what snippet rules reference |
| `zone_id` | The zone the snippet is deployed to |

## Example Manifests

A 302 redirect snippet:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSnippet
metadata:
  name: redirect-legacy
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  snippet_name: redirect_legacy
  files:
    - name: main.js
      content: "export default { async fetch(request) { return Response.redirect(\"https://www.example.com\", 302); } };"
  main_module: main.js
```

## Destroy Semantics

Destroy is a real delete. The snippet is removed from the zone. Snippet rules that referenced `snippet_name` keep pointing at that name -- they start invoking nothing until you retarget or delete them. Renaming is also a replace (new identity); the old name is left behind if something else still holds it.

## Related Resources

- **CloudflareSnippetRules** -- the zone's routing table that invokes this snippet by name
- **CloudflareWorker** -- the full Worker when you need bindings, cron, or a custom domain
- **CloudflareDnsZone** -- `zone_id` foreign key

## Further Reading

For operational judgment -- name-as-identity upsert, byte-stable content, and the `main_module` / `metadata` seam -- see GUIDE.md.

## References

- [Cloudflare Snippets](https://developers.cloudflare.com/rules/snippets/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
