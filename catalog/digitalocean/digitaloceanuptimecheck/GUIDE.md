# DigitalOcean Uptime Check -- Operational Guide

What experience with this component teaches that the field reference cannot.

## down vs down_global: page on global, watch regional

A `down` alert fires when ANY vantage region cannot reach the target -- which includes that region's own network weather. A `down_global` alert fires only when ALL regions agree the target is unreachable. Page humans on `down_global`; route `down` to a low-urgency channel if you want early regional signals. Probing from more regions makes the distinction sharper -- declare at least two.

## Always declare regions (the spec makes you)

DigitalOcean happily defaults the region set when omitted -- and the provider then reads the default back into state and tries to remove it on every subsequent plan, forever. That is why `regions` is required here even though the API can choose. Pick the regions your users actually connect from.

## Each alert type owns a different slice of threshold / comparison / period

DigitalOcean's API decides more of an alert row than its schema admits, and the spec encodes exactly what was measured against the live API so a manifest and DigitalOcean agree on every apply:

- **`period` is required on every row.** The API rejects any alert without it ("missing required field 'period'"), whatever the provider's schema says.
- **`latency`** honors both `threshold` (milliseconds) and `comparison`, so both are required -- a latency rule without a bar or a direction is ambiguous, and DigitalOcean would silently pick one.
- **`ssl_expiry`** honors `threshold` (DAYS before expiry; required, because the API accepts 0 -- an alert on the day the certificate expires is a post-mortem, not a warning; give it your renewal pipeline's worst-case turnaround, 14+ days) and always evaluates `less_than`, so `comparison` is rejected and the module sends it.
- **`down` / `down_global`** are fixed by the API at `threshold: 1, comparison: less_than` no matter what you send (an explicit `3 / greater_than` reads back as `1 / less_than`), so both fields are rejected and the module sends the API's own pair.

The rule of thumb: a value DigitalOcean is going to overwrite is never accepted as input, because it would read back different from what you wrote and re-plan forever.

## Alert email goes only to verified team members

DigitalOcean delivers alert email only to addresses that belong to verified members of the team that owns the check, and rejects any other address at create time (`invalid email`). A shared inbox, a pager bridge, or an external on-call address must be invited to the team and verified before it can appear in `notifications.emails`. The same rule applies to monitor alerts. Slack rows have no such restriction.

## ping targets are hosts, http(s) targets are URLs

`https://www.example.com` for http/https probes; `www.example.com` (or an IP) for ping. DigitalOcean enforces the pairing at request time -- the spec documents it rather than guessing your intent.

## Alert rules live and die with the check

Deleting the check deletes every alert rule under it -- there is nothing to clean up, and nothing survives to alert on a target you stopped probing. Renaming an alert row replaces that row on both provisioners (new id, fresh alert history): the row's address is keyed by its name, so a rename is a delete-and-create in the module's eyes, even though DigitalOcean itself would have renamed the alert in place. The check renames in place.

## Slack webhooks are credentials

The webhook URL lets anyone post to your channel. The spec marks it sensitive, so the platform accepts only a managed-secret reference (`$secret/<name>`) for it, never a literal URL, and the Pulumi module additionally encrypts it in stack state. Terraform state stores every value in plain text -- on that engine the protection is your state backend's own encryption, so treat state as you would the credential itself.

## A check deleted in the console breaks the next plan -- remove it from state by hand

DigitalOcean's Uptime API answers **403 "not authorized"**, not 404, for any check the account no longer owns -- deleted a moment ago or never created. The provider only treats 404 as "gone", so a check someone deleted in the control panel makes every later plan fail with `Error retrieving check: ... 403` instead of quietly dropping the resource from state. The remedy is a manual state removal of the check (and its alert rows) before the next apply recreates them. Every other DigitalOcean API in this catalog 404s normally; this one is the exception.

## What is deliberately NOT here

Metric alerts on droplets/load balancers/databases (that is the DigitalOceanMonitorAlert kind); authenticated probes, custom headers, and response-body assertions (DigitalOcean's Uptime API has none of them); and standalone alert objects pointing at existing checks -- the mutable parent id upstream is a corruption class, so rules are declared on their check, period.
