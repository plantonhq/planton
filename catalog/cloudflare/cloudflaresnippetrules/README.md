# Cloudflare Snippet Rules

## Overview

`CloudflareSnippetRules` manages the zone's snippet routing table: the ordered list of expressions deciding which requests invoke which snippet. The zone has exactly ONE such table -- this resource is a zone singleton, and every apply replaces the whole list (Cloudflare's API is a full-replacement PUT). Keep all of a zone's snippet rules in ONE manifest; a second manifest against the same zone would silently overwrite the first's rules on every apply.

Destroying this resource deletes ALL snippet rules in the zone -- including any created outside this manifest (dashboard, API). The snippets themselves survive; only the routing table empties.

## Key Features

- **Zone singleton** -- one table per zone; every apply is a full-replacement PUT
- **One manifest owns the WHOLE table** -- a second manifest against the same zone silently overwrites the first
- **`enabled` defaults TRUE here** -- Cloudflare's own default is FALSE (a rule that omits `enabled` is created disabled and matches nothing). This spec flips that so a declared rule runs
- **`snippet_name` is a foreign key** -- defaults to `CloudflareSnippet` / `status.outputs.snippet_name`

## Use Cases

**Ideal for:**

- Invoking a redirect snippet on a path prefix
- An ordered list of snippet routes you want in one reviewable manifest

**Not ideal for:**

- The snippet code itself -- that is `CloudflareSnippet`
- Expression-based WAF or cache rules -- that is `CloudflareRuleset`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone whose snippet routing table is managed. Can reference a `CloudflareDnsZone` via `value_from` (defaults to `status.outputs.zone_id`). |
| `rules` | list | Yes | The WHOLE table. At least one rule. Each rule needs `expression` and `snippet_name`. |

### Rule Fields

| Field | Type | Description |
|-------|------|-------------|
| `expression` | string | Cloudflare Rules language (wirefilter), e.g. `starts_with(http.request.uri.path, "/legacy")`. Use the function form for prefix/suffix matches -- the operator form (`... starts_with "..."`) is rejected with error 20127 outside the paid grammar extension (measured even on Pro). |
| `snippet_name` | StringValueOrRef | The snippet to invoke. Can reference a `CloudflareSnippet` via `value_from` (defaults to `status.outputs.snippet_name`). |
| `description` | string | Shown in the dashboard. |
| `enabled` | bool | Defaults to **true** in this spec. Cloudflare's provider default is false -- omit it here and the rule still runs. Set `false` explicitly to stage a rule. |

### Outputs

| Field | Description |
|-------|-------------|
| `zone_id` | The zone whose snippet routing table is managed (the singleton's identity) |

## Example Manifests

One rule that invokes a snippet on a path prefix:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSnippetRules
metadata:
  name: legacy-routes
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  rules:
    - expression: 'starts_with(http.request.uri.path, "/legacy")'
      snippet_name:
        value: redirect_legacy
      description: Redirect legacy paths
```

## Destroy Semantics

Destroy wipes ALL snippet rules in the zone, including ones created outside this manifest. The snippets themselves survive. There is no "delete only my rows" -- the API replaces the table with empty. If the dashboard has rules you did not declare, they disappear on destroy (and on every apply that omitted them).

## Related Resources

- **CloudflareSnippet** -- the code this table invokes by `snippet_name`
- **CloudflareRuleset** -- expression-based WAF and cache rules, a different table
- **CloudflareDnsZone** -- `zone_id` foreign key

## Further Reading

For operational judgment -- full-replacement PUT, destroy-wipes-everything, and the enabled-defaults-true footgun -- see GUIDE.md.

## References

- [Cloudflare Snippet rules](https://developers.cloudflare.com/rules/snippets/create-snippet/#snippet-rules)
- [Cloudflare Rules language](https://developers.cloudflare.com/ruleset-engine/rules-language/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
