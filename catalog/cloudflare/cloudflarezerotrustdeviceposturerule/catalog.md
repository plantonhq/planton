# Cloudflare Zero Trust Device Posture Rule

Deploys a Cloudflare Zero Trust device posture rule: a health check — disk encrypted, OS current, EDR agent healthy — that WARP evaluates on enrolled devices and that Access and Gateway policies can require before admitting a device. The rule's `type` selects which of the 23 supported checks runs, from WARP-client checks (file, os_version, disk_encryption, firewall, client certificates) to service-to-service reads of posture integrations (CrowdStrike, Intune, SentinelOne, Tanium, Workspace ONE, Kolide, custom). A plain CRUD object: real create, update, and delete.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Device Posture Rule** — a `cloudflare_zero_trust_device_posture_rule` carrying the check `type`, its `input` parameters, platform `match` targeting, and the `schedule`/`expiration` timing pair

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with a Cloudflare API token carrying Account → Zero Trust → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** — the account must have completed Cloudflare Zero Trust onboarding (the team-name step).
- **A posture integration** (only for `*_s2s`, `intune`, and `workspace_one` types) — a configured integration holding the vendor's credentials; its UUID goes in `input.connectionId`. Posture integrations are managed outside this catalog today.
- **A Zero Trust list of device identifiers** (only for `serial_number` and `unique_client_id` types) — the list whose ID goes in `input.id`.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Device Posture Rule**, and click **Deploy**. The creation wizard walks you through the owning account, the check type, platform targeting, the check's input parameters, and the polling schedule. Start from the **OS version floor** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDevicePostureRule
metadata:
  name: macos-current
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: macos-current
  type: os_version
  match:
    - platform: mac
  input:
    version: 14.4.1
    operator: ">="
  schedule: 5m
  expiration: 1h
```

```shell
planton apply -f posture-rule.yaml
```

This creates a rule requiring macOS devices to run at least 14.4.1, re-checked every five minutes with results trusted for an hour. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a posture rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Which input fields a type reads is Cloudflare's contract** — the spec enum-walls every value list the provider walls (operators, compliance statuses, states, risk levels), but deliberately does not wall which input fields belong to which type: 23 types times ~36 fields is Cloudflare's pairing to evolve. A wrong pairing (say, `version` on a `file` rule) fails at the API with a clear message, not silently. Each input field's reference documentation names the check families that read it.

**Delete consumers before rules** — Access and Gateway policies reference rules by UUID. Deleting a rule a policy still requires makes that requirement unevaluable; depending on the policy's shape, that means nobody passes (a soft outage) or the check silently stops gating (a security hole). Retire the policy reference first, then the rule.

**Integration-backed checks inherit the integration's health** — the `*_s2s`, `intune`, and `workspace_one` types read a posture integration holding vendor credentials. If the integration is deleted or its credentials lapse, every rule pointing at it starts failing devices. Treat integration health as part of these rules' operational surface.

**`schedule` and `expiration` trade freshness against noise** — `schedule` is how often the client re-checks (default 5m, minimum 1m); `expiration` is how long a result may be trusted. An expiration shorter than the schedule makes devices oscillate to "unknown" between polls — keep expiration comfortably above schedule, or leave it empty to trust the latest result.

**Platform targeting narrows where the check runs** — `match` scopes the rule to platforms (windows, mac, linux, android, ios, chromeos); empty runs wherever the check type applies. OS version checks are platform-specific in practice: clone the rule per platform, and on Linux use `input.osDistroName`/`input.osDistroRevision` rather than the bare version.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. Posture integrations (`input.connectionId`), Zero Trust lists (`input.id`), and client certificates (`input.certificateId`) are referenced as literal UUID strings because those surfaces carry no typed reference here.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `rule_id` | The Cloudflare-assigned UUID of the rule | The device posture requirement in Access and Gateway policies |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**OS version floor** — the patch-hygiene shape: devices must run at least a named OS version, re-checked every five minutes with results trusted for an hour. Clone per platform and reference the rule from Access or Gateway policies. Start from the **OS version floor** preset.

**Disk encryption required** — the data-protection baseline: every disk on every desktop platform must be encrypted (`input.requireAll` checks all volumes; `input.checkDisks` names specific volumes instead). The single most common posture gate in front of internal applications. Start from the **Disk encryption required** preset.

**EDR presence** — a `sentinelone`, `carbonblack`, or `crowdstrike_s2s` rule proving the endpoint agent is installed and healthy before the device reaches production networks; the s2s variants also gate on the vendor's own threat and state signals.

## Works With

- [**Cloudflare Zero Trust Access Policy**](/infra-catalog/cloudflare-zero-trust-access-policy) — requires posture checks in front of applications.
- [**Cloudflare Zero Trust Gateway Policy**](/infra-catalog/cloudflare-zero-trust-gateway-policy) — requires posture checks for network egress.
- [**Cloudflare Zero Trust Device Default Profile**](/infra-catalog/cloudflare-zero-trust-device-default-profile) — the WARP client the checks run under.
- [**Cloudflare Zero Trust List**](/infra-catalog/cloudflare-zero-trust-list) — holds the device identifiers `serial_number` and `unique_client_id` checks read via `input.id`.
