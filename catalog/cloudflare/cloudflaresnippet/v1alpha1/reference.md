# CloudflareSnippet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareSnippetSpec defines one Cloudflare Snippet: a small JavaScript
module deployed to the zone's edge, invoked on requests that match snippet
rules (managed separately by CloudflareSnippetRules). Snippets are the
lightweight sibling of Workers -- same runtime, no bindings, sized for
header rewrites, redirects, and request/response touch-ups.

The snippet NAME is the identity: Cloudflare's create call is an upsert, so
deploying a snippet whose name already exists in the zone silently adopts and
overwrites it. Pick names deliberately, and treat renaming as
delete-and-recreate (the provider replaces the resource when the name
changes). Snippets are a PAID-PLAN product: a free-plan zone refuses every
snippet upload with "snippets are not allowed" (measured live 2026-08-27 on
pending and active free zones alike); Pro and above include them at no
extra cost, with per-plan COUNT limits checked at create.

Code travels INLINE in the manifest as file content strings -- the same
convention as the Worker kind's script. Cloudflare caps snippet code size
(32 KB per snippet at the time of writing); oversized code fails at the API.

## Example

```yaml
# A complete, protovalidate-valid CloudflareSnippet example: a single-file
# snippet that adds a response header. Code travels inline; keep content
# byte-stable (Cloudflare re-serves the stored content on reads).
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSnippet
metadata:
  name: add-debug-header
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  snippet_name: add_debug_header
  files:
    - name: main.js
      content: "export default { async fetch(request) { const response = await fetch(request); const newResponse = new Response(response.body, response); newResponse.headers.set('x-served-by', 'snippet'); return newResponse; } };"
  main_module: main.js
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.snippetName` | `string` | yes |  |  |
| `spec.files` | `[]CloudflareSnippetFile` | yes |  |  |
| `spec.files[].name` | `string` | yes |  |  |
| `spec.files[].content` | `string` | yes |  |  |
| `spec.mainModule` | `string` | yes |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone the snippet is deployed to.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.snippetName

`string` · required

The snippet's name -- its IDENTITY within the zone. IMMUTABLE in practice:
changing it replaces the snippet (new identity), and any snippet rule
referencing the old name keeps pointing at whatever the old name holds.
Letters, numbers, and underscores only (Cloudflare's naming rule for
snippets; the API rejects other characters).

- rule: snippet_name may contain only letters, numbers, and underscores (e.g. redirect_legacy_urls)
- rule: {"required":true}

### spec.files

`[]CloudflareSnippetFile` · required

The snippet's source files. Most snippets are a single file; multi-file
snippets import siblings by name via ES module imports.

- rule: {"repeated":{"minItems":"1"}}

### spec.files[].name

`string` · required

The file name (e.g. main.js). Referenced by main_module and by ES module
imports from sibling files.

- rule: {"required":true}

### spec.files[].content

`string` · required

The file's JavaScript source, inline. Keep the content byte-stable
(consistent line endings, no trailing whitespace churn) -- Cloudflare
re-serves the stored content on reads, and gratuitous formatting differences
read back as configuration drift.

- rule: {"required":true}

### spec.mainModule

`string` · required

The entry module -- which file's default export handles the request. Must
name one of the files above.

- rule: {"required":true}

## Validation Rules

- `spec.main_module_in_files`: main_module must name one of the files -- the entry module has to be part of the snippet's own file list

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareSnippet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.snippet_name` | `string` | The snippet's name -- its identity within the zone, and what snippet rules reference. Exported as an output so consumers can wire references without repeating the literal. |
| `status.outputs.zone_id` | `string` | The zone the snippet is deployed to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareSnippetRules | `spec.rules[].snippetName` | `status.outputs.snippet_name` |

## See Also

- [Overview](../README.md)
