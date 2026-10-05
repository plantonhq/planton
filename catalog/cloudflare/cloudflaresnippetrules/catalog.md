# Cloudflare Snippet Rules

Deploys the zone's snippet routing table: the ordered list of expressions deciding which requests invoke which Cloudflare Snippet. A zone has exactly one such table, and Cloudflare's API is a full-replacement PUT — every apply replaces the entire table with exactly the `rules` list in this manifest, so one manifest must own all of a zone's snippet rules. Destroying the resource deletes every snippet rule in the zone, including rules created outside this manifest; the snippets themselves survive, only the routing table empties.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Snippet Rules** -- one `cloudflare_snippet_rules` on the zone, whose `rules` list is the whole routing table, evaluated in order against Cloudflare's Rules language (the same wirefilter expressions rulesets use)

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Snippets Edit on the target zone. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A Cloudflare zone** -- provide the zone ID directly or wire `zoneId` to a Cloudflare DNS Zone resource by reference.
- **The snippets the rules invoke** -- each rule's `snippetName` must exist in the zone; a missing name fails at apply, not at validation. Create the Cloudflare Snippet resources first.
- **Sole ownership of the zone's snippet-rule table** -- a second manifest, a dashboard edit, or another team's rules against the same zone will be silently overwritten by this manifest's next apply.

## Deploy

### Console

Open the deployment store, find **Cloudflare Snippet Rules**, and click **Deploy**. The creation wizard walks you through the zone and an ordered rule builder — expression, snippet, description, and enablement per rule. Start from the **Path-prefix route** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSnippetRules
metadata:
  name: legacy-routes
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  rules:
    - expression: 'starts_with(http.request.uri.path, "/legacy")'
      snippetName:
        value: redirect_legacy
      description: Redirect legacy paths
```

```shell
planton apply -f snippet-rules.yaml
```

This installs a one-row routing table that invokes the `redirect_legacy` snippet on every request under `/legacy` — the rule runs immediately because this spec defaults `enabled` to true. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire both the zone and the invoked snippet to resources managed in the same InfraPipeline:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  rules:
    - expression: 'starts_with(http.request.uri.path, "/legacy")'
      snippetName:
        valueFrom:
          kind: CloudflareSnippet
          name: redirect-legacy
          fieldPath: status.outputs.snippet_name
```

The InfraPipeline resolves the dependency graph, deploys the zone and the snippet first, then installs the routing table that binds them.

## Key Configuration

These are the most important decisions when configuring snippet rules. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One manifest owns the whole table.** Every apply sends exactly the declared `rules` list and the live table becomes that list — rules added in the dashboard, by another team, or by a second manifest disappear on the next apply, with no merge and no warning. Keep every snippet rule for a zone in this one resource; split ownership and the manifests fight each other one apply at a time.

**`enabled` defaults true here — Cloudflare's default is false.** This is the footgun: a rule that omits the field at the API is created disabled and matches nothing, and migrated configurations have been burned by the flip. Both engines coalesce an unset flag to true so a declared rule runs; set `enabled: false` explicitly to stage a rule without activating it.

**Order is evaluation order.** Rules run top to bottom in list order. Put narrower expressions above broader ones when both could match — reordering the list is a plain in-place update.

**`snippetName` is a foreign key with a literal escape hatch.** Wire it to a Cloudflare Snippet resource by reference, or pass a literal name to invoke a snippet that already exists in the zone. Either way Cloudflare is the wall: an unknown name fails at apply.

**Destroy is not selective.** It empties the zone's entire snippet routing table, outsiders included. To stop invoking snippets reversibly, set each rule's `enabled: false` and apply — that preserves table ownership and is a one-line rollback.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |
| **CloudflareSnippet** | `rules[].snippetName` | `status.outputs.snippet_name` |

### What This Kind Provides

This kind has no consumable outputs: `status.outputs` only echoes the zone ID back, because the routing table is a zone singleton whose identity is the zone itself — there is no separate table ID for downstream resources to reference.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Path-prefix route** -- One rule invoking a redirect snippet when the path starts with `/legacy`; pair it with the Cloudflare Snippet **Redirect snippet** preset, which creates `redirect_legacy`. Start from the **Path-prefix route** preset.

**One table, many snippets** -- The steady state for a zone using snippets seriously: a single manifest listing every route in evaluation order, each row wired to its snippet by reference so renames and redeploys cannot drift the table.

**Staged rollout** -- Add the new rule with `enabled: false`, apply, verify the expression against live traffic expectations, then flip the flag. The table never leaves manifest control, and rollback is the same one-line change.

## Works With

- [**Cloudflare Snippet**](/infra-catalog/cloudflare-snippet) -- the code this table invokes by name; create the snippet first.
- [**Cloudflare Ruleset**](/infra-catalog/cloudflare-ruleset) -- the other expression table on the zone (WAF, cache, redirects), same Rules language, different engine.
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the zone scope; wire `zoneId` via ValueFromRef.
