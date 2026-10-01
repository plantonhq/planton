# CloudflareWebAnalyticsSite guide

The judgment this guide protects you from: the API token needs a READ permission its write half does not imply, this component's rules are write-only from the provider's point of view, and several site fields never come back from the API at all.

## The token needs Account Settings READ, not just Write

Web Analytics is one of the product families where Cloudflare splits read from write: Account Settings Write creates, updates, and deletes sites, but every read-back -- including the refresh Terraform performs on every plan and the read-after-create the modules perform -- answers 403 "Authentication error" without the separate Account Settings Read permission (measured live). The failure is nasty because the create SUCCEEDS first: the deploy breaks midway, leaving a site the same token can still delete but not see. Grant both permissions before the first deploy.

## Manage rules here or in the dashboard -- never both

The provider's read for `cloudflare_web_analytics_rule` is an empty function: it makes the API call for nothing and never compares what came back. So rules are effectively write-only. A rule someone adds or edits in the Cloudflare dashboard is invisible to IaC -- no drift shows in any plan -- and it stays invisible until the next apply re-asserts your declared rows and quietly overwrites the difference. Pick one owner for the rule set. If it is this manifest, then dashboard edits are temporary by definition.

One sharp edge on adoption (measured live): the rule create REJECTS duplicates -- creating a rule identical to one already on the ruleset answers 409 "duplicateFound", it does not converge. So when adopting an existing site that already carries rules, delete the existing rules first (in the dashboard, or via the API's rule DELETE), then apply -- the apply re-creates exactly your declared rows. Skipping the delete step fails the apply against every surviving duplicate.

## Four site fields also never come back

`enabled`, `host`, `lite`, and `zone_tag` are marked no-refresh at the provider: they are sent on write and never populated from a read. State carries whatever you configured, so drift on them is undetectable by design. On import the provider works around this for one field only -- it reaches into the computed ruleset to backfill `zone_tag` -- which is worth knowing when an imported site looks like it lost its other settings.

The asymmetry runs the other way too: Cloudflare's CREATE response omits the site's `ruleset` object entirely (measured live -- it exists only on the GET), so the modules perform a read-after-create to wire the folded rules and the `ruleset_id`/`site_token`/`snippet` outputs. That is why the Read permission above is not optional even for a deploy-only workflow.

## The site token is public by nature

The beacon runs in visitors' browsers, so the token inevitably ships inside your pages -- it is not a secret in the way an API key is. The `site_token` and `snippet` outputs are therefore plain values, not secrets. Nothing breaks if the token is seen.

## Rules exist only on zone-linked sites

Cloudflare creates the ruleset -- the container every rule attaches to -- only when the site is identified by `zone_tag`. A host-identified site has no ruleset in any API response, ever (measured live), so it cannot carry rules and its `ruleset_id` output is empty. Validation rejects `host` + `rules` up front, because deployed it would fail midway with a null-reference error rather than a clear message. If you need path exclusions on a site that is not a Cloudflare zone, the beacon-side answer is to embed the snippet only on the pages you want measured.

## host and zone_tag are different products of the same feature

With `host` you measure any site and embed the snippet yourself. With `zone_tag` you measure a Cloudflare zone, and `auto_install` can put the beacon on every proxied page with no code change. `auto_install` needs the zone to actually be proxied (orange-clouded) -- on a DNS-only zone Cloudflare has no response to inject into, so the setting is accepted and does nothing.

## Deleting the site deletes the history

A real delete here is not just "stop measuring": the site's collected analytics stop being reachable. If the data matters, export what you need before retiring the site -- and note there is no provider-side deletion guard to catch a careless destroy.

## Rules are ordered, and order is positional

The module manages one provider object per declared row, keyed by position. Reordering rows therefore rewrites those objects rather than reordering them in place. That is harmless (the end state is what the list says) but it makes plans noisier than expected -- expect churn if you insert a row at the top of a long list.

## Pairs well with

- [CloudflareDnsZone](../cloudflarednszone/README.md) -- the zone a zone-measured site references.
- [CloudflareLogpushJob](../cloudflarelogpushjob/README.md) -- the server-side view of the same traffic, including non-browser requests.
- [CloudflareNotificationPolicy](../cloudflarenotificationpolicy/README.md) -- the `web_analytics_metrics_update` alert family.
