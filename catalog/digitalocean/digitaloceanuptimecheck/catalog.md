# DigitalOcean Uptime Check

Deploys an availability and latency probe on any external endpoint -- a site, an API, anything reachable -- run from DigitalOcean's global vantage regions, with alert rules for downtime, latency, and certificate expiry delivered by email or Slack. Alert rules are composed on the check as rows, one alert object per row with its own channels, because DigitalOcean's standalone alert resource leaves the parent relationship mutable -- a corruption class where re-pointing an alert orphans it on the old check. Destroying the check destroys its alerts with it.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Uptime check** -- the probe itself: target, protocol (`ping`, `http`, or `https`), and vantage regions, as one `digitalocean_uptime_check` resource
- **Uptime alerts** -- created only when `alerts` rows are set: one `digitalocean_uptime_alert` per row (`down`, `down_global`, `latency`, `ssl_expiry`), each carrying its own notification channels, with Slack webhook URLs accepted only as managed-secret references and encrypted in Pulumi stack state

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### DigitalOcean Account

- **Nothing for the probe itself** -- the probed target is external to the account.
- **Slack incoming webhook** (only for Slack delivery) -- the webhook URL is a credential; store it as a managed secret and reference it as `$secret/<name>` in the manifest.
- **Verified alert recipients** -- every email address in `notifications.emails` must belong to a verified member of the DigitalOcean team; the API rejects any other address at create time (`invalid email`). Invite and verify shared inboxes or pager bridges before naming them.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Uptime Check**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Website Availability Check** preset in the [Presets](#presets) tab to probe a public site from all four vantage regions.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanUptimeCheck
metadata:
  name: homepage-availability
  org: acme-corp
  env: prod
spec:
  checkName: homepage
  target: https://www.example.com
  regions:
    - us_east
    - us_west
    - eu_west
    - se_asia
  alerts:
    - alertName: homepage-down
      type: down_global
      period: 2m
      notifications:
        emails:
          - ops@acme-corp.com
```

```shell
planton apply -f do-uptime-check.yaml
```

This probes the site over https from all four vantage regions and mails ops (a verified member of the DigitalOcean team) only when every region agrees it is down. Probing starts immediately, and results appear in the control panel's Monitoring -> Uptime section. A Stack Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an uptime check. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**down versus down_global** -- a `down` alert fires when ANY vantage region cannot reach the target, which includes that region's own network weather; `down_global` fires only when ALL regions agree the target is unreachable. Page humans on `down_global`; route `down` to a low-urgency channel if you want early regional signals.

**Regions are always declared** -- `regions` is required here even though DigitalOcean can default it: the provider never reconciles a defaulted region set, so an omitted value would leave every subsequent plan trying to remove what the API chose. Pick the regions your users actually connect from (`us_east`, `us_west`, `eu_west`, `se_asia`); more vantage points make `down_global` sharper and per-region history richer.

**Each alert type owns a different slice of threshold / comparison / period** -- `period` is required on every row (the API rejects an alert without it). `latency` honors and requires both `threshold` (milliseconds) and `comparison`. `ssl_expiry` honors and requires `threshold` (DAYS before expiry -- give it your renewal pipeline's worst-case turnaround, 14+ days, because DigitalOcean accepts 0 and an alert at zero days is a post-mortem) and always evaluates `less_than`, so `comparison` is rejected there. `down` and `down_global` are fixed by DigitalOcean at `threshold: 1, comparison: less_than` whatever is sent, so both fields are rejected on those rows and the modules send the API's own pair. A value DigitalOcean would overwrite is never accepted as input: it would read back different from what you wrote and re-plan forever.

**Target and protocol pair up** -- a URL for `http`/`https` probes (`https://www.example.com`), a hostname or IP for `ping`; DigitalOcean enforces the pairing at request time. Only `https` probes can carry `ssl_expiry` rules -- the certificate-renewal safety net is a reason on its own to probe over https.

**Alert rules live and die with the check** -- deleting the check deletes every alert rule under it: nothing to clean up, and nothing survives to alert on a target you stopped probing. Renaming an alert row replaces that row on both provisioners (new id, fresh alert history) because the row's address is keyed by its name -- DigitalOcean itself would rename in place, the replacement is the module's doing. The check renames in place.

**Comparison spelling** -- this API spells `comparison` snake_case (`greater_than`, `less_than`); monitor alerts spell the same concept CamelCase. The two are different DigitalOcean APIs and are deliberately not unified -- copy each kind's own spelling.

**Slack webhooks are credentials** -- the `url` field is marked sensitive in the spec, so in manifests it must be a managed-secret reference (`$secret/<name>`), never a literal URL; the Pulumi module additionally encrypts it in stack state. Terraform state stores every value in plain text -- on that engine the protection is the state backend's own encryption.

**Pausing without deleting** -- `enabled: false` keeps the check defined but stops probing (and with it, alerting); unset defaults to enabled.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies -- the probed target is an external endpoint, declared as a literal URL, hostname, or IP.

### What This Component Provides

`status.outputs` carries `check_id`, the check's UUID -- its API identity and its import id -- and `alert_ids`, the composed rows' UUIDs keyed by `<row index>-<alert name>` (each row imports as `{check_id},{alert_id}`, and the map supplies the second half so adopting a whole check needs no lookup). No downstream Cloud Resource consumes an uptime check by reference, so there is no ValueFromRef story to teach.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Website availability paging** -- probe from all four vantage regions and page only on `down_global`, filtering out single-region network weather. The baseline monitor every public endpoint deserves. Start from the **Website Availability Check** preset.

**Latency and certificate safety net** -- two quality signals on one https probe: a latency rule that fires on sustained slowness, and an `ssl_expiry` rule that turns a certificate lapse into a calendar item. Start from the **API Latency and Certificate Expiry** preset.

**Split urgency across channels** -- each alert row carries its own notification channels, so one check can route `down` to a low-urgency Slack channel while `down_global` mails the on-call inbox -- early signal and human paging without two probes.

## Works With

- [**DigitalOcean Monitor Alert**](/cloud-catalog/digital-ocean-monitor-alert) -- the inside view: metric alerts on the droplets, balancers, and databases behind the endpoint this check probes from outside
- [**DigitalOcean Load Balancer**](/cloud-catalog/digital-ocean-load-balancer) -- the public entry point most checks end up probing
- [**DigitalOcean App Platform App**](/cloud-catalog/digital-ocean-app) -- App Platform's default ingress and custom domains are natural probe targets
