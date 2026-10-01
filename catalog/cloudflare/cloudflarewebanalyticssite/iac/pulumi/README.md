# CloudflareWebAnalyticsSite Pulumi Module

Pulumi (Go) IaC module for Web Analytics (RUM) sites.

## Architecture

```
main.go                         — stack-input loading + module entry
module/main.go                  — provider setup + resource orchestration
module/locals.go                — metadata/credential references
module/web_analytics_site.go    — WebAnalyticsSite + one WebAnalyticsRule per rules[] row
module/outputs.go               — site_tag, site_token, snippet, ruleset_id
```

## Behavior

A plain CRUD resource: real create/update/delete for the site and every rule. The site is identified by `host` OR `zone_tag` (spec validation enforces exactly one).

Rules fan out one SDK resource per declared row, named by position and parented to the site, with the ruleset id taken from the created site. The provider never reads rules back (its Read is an empty stub), so each apply re-asserts exactly the declared rows -- dashboard edits to the rule set are invisible and will be overwritten. Four site attributes (`enabled`, `host`, `lite`, `zone_tag`) are also never refreshed by the provider.

`site_token` and `snippet` are plain outputs: the token ships inside public pages once deployed.

Import the site as `{account_id}/{site_tag}` (the site tag, not the id copy); rules ship no importer upstream.

## Outputs

| Name | Description |
|------|-------------|
| `site_tag` | The Cloudflare-assigned site tag |
| `site_token` | The beacon's measurement token |
| `snippet` | The ready-to-embed script tag |
| `ruleset_id` | The parent object the measurement rules live under |

## SDK Version

Uses `pulumi-cloudflare` v6.19.0.
