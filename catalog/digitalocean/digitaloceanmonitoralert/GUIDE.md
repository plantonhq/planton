# DigitalOcean Monitor Alert -- Operational Guide

What experience with this component teaches that the field reference cannot.

## Prefer tags over id lists for droplet fleets

An id-targeted policy watches exactly the droplets listed -- replacements and autoscaled additions are NOT covered until the manifest changes. A tag-targeted policy tracks membership automatically: every droplet carrying the tag is watched the moment it exists. Use id references for singular pets, tags for fleets.

The tag does not have to exist when the policy is created. DigitalOcean stores it as a selector and neither checks nor creates it -- a policy naming a tag no droplet carries is accepted and simply watches nothing until one does, so you can declare the alert before the fleet. That is the opposite of firewalls, which reject a tag no resource carries; do not carry the firewall rule over.

## Alert email goes only to verified team members

DigitalOcean delivers alert email only to addresses that belong to verified members of the team that owns the policy, and rejects any other address at create time (`email is not verified`). A shared inbox, a pager bridge, or an external on-call address must be invited to the team and verified before it can appear in `alerts.emails`. The same rule applies to uptime alerts. Slack rows have no such restriction.

## The metric name IS the contract -- copy it exactly

Metric names are DigitalOcean's raw API paths with their inconsistencies intact: droplet CPU is `v1/insights/droplet/cpu` (no `_utilization_percent` suffix, unlike memory and disk), and database metrics live under `v1/dbaas/alerts/` with `_alerts` suffixes. Validation carries the exact 28-value list, so a typo fails at validation -- but read the error's list rather than guessing the spelling.

## Thresholds are float32 upstream

DigitalOcean stores `value` as a 32-bit float. More than 7 significant digits silently truncate -- 99.999999 becomes 100 by the time it evaluates. Round thresholds are also easier for humans at 3 AM.

## One policy per symptom, not per droplet

Policies accept many targets; a CPU policy covering the whole web fleet beats ten identical per-droplet policies. Split policies when the THRESHOLD differs (databases at 80%, batch workers at 95%), not per target.

## Slack webhooks are credentials

The webhook URL lets anyone post to your channel. The spec marks it sensitive, so the platform accepts only a managed-secret reference (`$secret/<name>`) for it, never a literal URL, and the Pulumi module additionally encrypts it in stack state. Terraform state stores every value in plain text -- on that engine the protection is your state backend's own encryption, so treat state as you would the credential itself.

## Disabling beats deleting

`enabled: false` keeps the policy defined but silent -- ideal for maintenance windows or pre-staging alerts before a service carries traffic. Deleting the policy loses nothing but its UUID, though: recreating it is cheap and the manifest is the source of truth.

## What is deliberately NOT here

Uptime probing of external endpoints (that is the DigitalOceanUptimeCheck kind); Kubernetes and App Platform alerting (DigitalOcean exposes no monitor-alert metrics for them at the pinned provider); and PagerDuty/webhook delivery beyond email and Slack (the API supports only these two).
