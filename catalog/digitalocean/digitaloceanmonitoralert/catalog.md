# DigitalOcean Monitor Alert

Deploys an alert policy on DigitalOcean's built-in metrics -- droplet CPU, memory, disk, and bandwidth, load-balancer health and error rates, or managed-database utilization -- with email and Slack delivery. Targets are wired as typed references to droplets, load balancers, or database clusters, or as droplet tags, and validation rejects any target outside the chosen metric's family. Every field updates in place, so thresholds, windows, and channels tune without replacing the policy.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Monitor alert policy** -- one `digitalocean_monitor_alert` resource carrying the metric, comparison, threshold, sampling window, targets, and notification channels. The typed reference lists (`dropletIds`, `loadBalancerIds`, `databaseClusterIds`) merge back into the provider's single entities argument; a tag-targeted policy sends only `tags`, and DigitalOcean resolves membership from the tag. Slack webhook URLs are accepted only as managed-secret references and encrypted in Pulumi stack state.

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Alert targets** -- the DigitalOceanDroplet, DigitalOceanLoadBalancer, or DigitalOceanDatabaseCluster resources the policy watches, referenced by name -- or a droplet tag, which needs no resource reference at all.

### DigitalOcean Account

- **Slack incoming webhook** (only for Slack delivery) -- the webhook URL is a credential; store it as a managed secret and reference it as `$secret/<name>` in the manifest.
- **Verified alert recipients** -- every email address in `alerts.emails` must belong to a verified member of the DigitalOcean team; the API rejects any other address at create time (`email is not verified`). Invite and verify shared inboxes or pager bridges before naming them.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Monitor Alert**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Tag-Targeted Fleet CPU Alert** preset in the [Presets](#presets) tab to watch a droplet fleet by tag.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanMonitorAlert
metadata:
  name: web-fleet-cpu
  org: acme-corp
  env: prod
spec:
  description: CPU of the web fleet above 90 percent for 10 minutes
  metricType: v1/insights/droplet/cpu
  compare: GreaterThan
  value: 90
  window: 10m
  tags:
    - web
  alerts:
    emails:
      - ops@acme-corp.com
```

```shell
planton apply -f do-monitor-alert.yaml
```

This creates a policy watching CPU across every droplet tagged `web`, mailing ops when the 10-minute average crosses 90 percent -- droplets gaining or losing the tag join and leave the alert with no manifest change. A Stack Job tracks the provisioning in real time.

### InfraChart

When the policy watches resources deployed in the same InfraPipeline, wire the targets with ValueFromRef:

```yaml
spec:
  description: Load balancer 5xx error rate above 5 percent
  metricType: v1/insights/lbaas/increase_in_http_error_rate_percentage_5xx
  compare: GreaterThan
  value: 5
  window: 5m
  loadBalancerIds:
    - valueFrom:
        kind: DigitalOceanLoadBalancer
        name: web-lb
        fieldPath: status.outputs.load_balancer_id
  alerts:
    slack:
      - channel: "#incidents"
        url: $secret/incidents-slack-webhook
```

The InfraPipeline resolves the dependency graph, deploys the load balancer first, then provisions the policy with the resolved UUID. The Slack `url` stays a managed-secret reference -- the field is sensitive and never carries a literal webhook.

## Key Configuration

These are the most important decisions when configuring a monitor alert. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Metric family fixes the target list** -- `metricType` is namespaced per resource family: droplet metrics under `v1/insights/droplet/`, load-balancer metrics under `v1/insights/lbaas/`, managed-database metrics under `v1/dbaas/alerts/`. Validation rejects any cross-family pairing before a provisioner runs: droplet metrics accept `dropletIds` and/or `tags`, while load-balancer and database metrics accept only their own id lists -- no tags.

**Metric names are DigitalOcean's raw API paths** -- their inconsistencies are deliberate facts of that API, never "corrected" here: droplet CPU is bare `v1/insights/droplet/cpu` (no `_utilization_percent` suffix, unlike memory and disk), and the database family carries `_alerts` suffixes. Validation holds the exact 28-value list, so a typo fails at validation -- read the error's list rather than guessing the spelling.

**Tags versus id lists** -- an id-targeted policy watches exactly the droplets listed; replacements and autoscaled additions are not covered until the manifest changes. A tag-targeted policy tracks membership automatically: every droplet carrying the tag is watched the moment it exists. Use id references for singular pets, tags for fleets. The tag need not exist when the policy is created -- DigitalOcean stores it as a selector and neither checks nor creates it, so the alert can be declared before the fleet (firewalls behave the opposite way and reject a tag no resource carries).

**One policy per symptom, not per target** -- policies accept many targets, so a CPU policy covering the whole web fleet beats ten identical per-droplet policies. Split policies when the threshold differs -- databases at 80 percent, batch workers at 95 -- not per target.

**Threshold precision and units** -- `value`'s units follow the metric: percent for utilization metrics, load units for `load_*`, bytes per second for bandwidth and disk I/O. DigitalOcean stores the threshold as a 32-bit float, so more than 7 significant digits silently truncate -- 99.999999 becomes 100 by the time it evaluates. The `window` (5m to 1h) sets how long the metric aggregates before comparison: 10m ignores boot spikes; tighten to 5m for latency-sensitive services.

**Slack webhooks are credentials** -- the `url` field is marked sensitive in the spec, so in manifests it must be a managed-secret reference (`$secret/<name>`), never a literal URL; the Pulumi module additionally encrypts it in stack state. Terraform state stores every value in plain text -- on that engine the protection is the state backend's own encryption.

**Disabling beats deleting** -- `enabled: false` keeps the policy defined but silent, ideal for maintenance windows or pre-staging alerts before a service carries traffic. Unset defaults to enabled, and the policy starts evaluating the moment it provisions. Deleting loses nothing but the policy's UUID -- the manifest is the source of truth and recreating is cheap.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDroplet** (droplet metrics) | `dropletIds[]` | `status.outputs.droplet_id` |
| **DigitalOceanLoadBalancer** (load-balancer metrics) | `loadBalancerIds[]` | `status.outputs.load_balancer_id` |
| **DigitalOceanDatabaseCluster** (database metrics) | `databaseClusterIds[]` | `status.outputs.cluster_id` |

### What This Component Provides

`status.outputs` carries a single value: `alert_id`, the policy's UUID -- its API identity and its import id. No downstream Cloud Resource consumes an alert policy by reference, so there is no ValueFromRef story to teach; the manifest itself is the source of truth for recreating the policy.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Fleet CPU alert by tag** -- the standard "something is running hot" signal for droplet fleets behind load balancers or autoscale pools, where members come and go. Start from the **Tag-Targeted Fleet CPU Alert** preset.

**Load-balancer 5xx paging** -- the 5xx error rate is the closest built-in metric to user pain; wire the balancer by reference and deliver to an incidents channel. Pair it with a droplet CPU policy so cause (hot backends) and effect (errors) page together. Start from the **Load Balancer 5xx Error-Rate Alert** preset.

**Database utilization guardrails** -- `v1/dbaas/alerts/` metrics on referenced database clusters, with thresholds split from the compute fleets' policies so each symptom pages at its own bar.

## Works With

- [**DigitalOcean Droplet**](/cloud-catalog/digital-ocean-droplet) -- CPU, memory, disk, and bandwidth targets, wired by reference or by tag
- [**DigitalOcean Load Balancer**](/cloud-catalog/digital-ocean-load-balancer) -- health and HTTP error-rate targets
- [**DigitalOcean Database Cluster**](/cloud-catalog/digital-ocean-database-cluster) -- managed-database utilization targets
- [**DigitalOcean Uptime Check**](/cloud-catalog/digital-ocean-uptime-check) -- the outside view: external endpoint probing that complements these inside-view metrics
