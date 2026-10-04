# CloudflareSnippetRules

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareSnippetRulesSpec manages the zone's snippet routing table: the
ordered list of expressions deciding which requests invoke which snippet.
The zone has exactly ONE such table -- this resource is a zone singleton, and
every apply replaces the whole list (Cloudflare's API is a full-replacement
PUT). Keep all of a zone's snippet rules in ONE manifest; a second manifest
against the same zone would silently overwrite the first's rules on every
apply.

Destroying this resource deletes ALL snippet rules in the zone -- including
any created outside this manifest (dashboard, API). The snippets themselves
survive; only the routing table empties.

Rules evaluate in list order against Cloudflare's Rules language
(https://developers.cloudflare.com/ruleset-engine/rules-language/) -- the same
wirefilter expressions rulesets use, e.g.
`starts_with(http.request.uri.path, "/legacy")`. Use the FUNCTION form for
prefix/suffix matches: the operator form (`... starts_with "/legacy"`) is a
paid-grammar extension the API rejects with error 20127 on most plans
(measured live 2026-08-27 -- rejected even on a Pro zone).

## Example

```yaml
# A complete, protovalidate-valid CloudflareSnippetRules example: the zone's
# whole snippet routing table (this list REPLACES all snippet rules in the
# zone on every apply).
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSnippetRules
metadata:
  name: zone-snippet-routing
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  rules:
    - expression: 'starts_with(http.request.uri.path, "/legacy")'
      snippet_name:
        value: redirect_legacy_urls
      description: "Send legacy URLs to the redirect snippet"
    - expression: 'http.host eq "beta.example.com"'
      snippet_name:
        value: add_debug_header
      enabled: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.rules` | `[]CloudflareSnippetRule` | yes |  |  |
| `spec.rules[].expression` | `string` | yes |  |  |
| `spec.rules[].snippetName` | `string \| valueFrom` | yes |  | CloudflareSnippet (`status.outputs.snippet_name`) |
| `spec.rules[].description` | `string` |  |  |  |
| `spec.rules[].enabled` | `bool` |  | `true` |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone whose snippet routing table is managed.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.rules

`[]CloudflareSnippetRule` · required

The zone's snippet rules, evaluated in order. This list is the WHOLE table:
every apply replaces all snippet rules in the zone with exactly this list.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].expression

`string` · required

The match expression in Cloudflare's Rules language (wirefilter), e.g.
`starts_with(http.request.uri.path, "/api/legacy")`. Validated by
Cloudflare at apply time. Prefer the always-legal FUNCTION form for
prefix/suffix matches -- the operator form (`... starts_with "..."`) is
rejected with error 20127 outside the paid grammar extension (measured
live 2026-08-27, rejected even on a Pro zone).

- rule: {"required":true}

### spec.rules[].snippetName

`string | valueFrom` · required

The snippet the rule invokes on matching requests.
When using value_from, defaults to CloudflareSnippet kind and
status.outputs.snippet_name field path.

- references: CloudflareSnippet (`status.outputs.snippet_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareSnippet, name: <that resource's name>, fieldPath: status.outputs.snippet_name}} -- a bare string does not parse

### spec.rules[].description

`string`

Optional description shown in the dashboard's rule list.

### spec.rules[].enabled

`bool` · optional (explicit presence)

Whether the rule is active. FOOTGUN: Cloudflare defaults this to FALSE -- a
rule that omits enabled is created DISABLED and matches nothing (this
flipped from true in older API generations, and migrated configurations
have been burned by it). This spec defaults it to true so a manifest that
declares a rule gets a rule that runs; set enabled: false explicitly to
stage a rule without activating it.

- default: `true`

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareSnippetRules, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_id` | `string` | The zone whose snippet routing table is managed. The table is a zone singleton -- the zone ID is its identity. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |
| `spec.rules[].snippetName` | CloudflareSnippet | `status.outputs.snippet_name` |

## See Also

- [Overview](../README.md)
