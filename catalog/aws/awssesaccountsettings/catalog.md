# AWS SES Account Settings

Manages the account-level SES settings for one AWS region: the account suppression list and the Virtual Deliverability Manager (VDM) posture. This is a settings singleton — AWS keeps exactly one SES account object per account and region, so deploy at most one instance per region; two instances targeting the same region fight over the same object. These settings live here rather than on configuration sets or identities precisely because multiple of those would contend for the one account object.

## What Gets Created

This component creates nothing new at AWS — it adopts the region's existing SES account object and configures its account-wide attributes:

- **Suppression List Posture** — which events (hard bounces, spam complaints) automatically add recipient addresses to the account-level suppression list, skipping them on every future send from the account.
- **VDM Posture** — whether the Virtual Deliverability Manager is enabled, with its engagement-metrics dashboard and Guardian delivery-optimization sub-toggles.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with SES account-level permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Nothing else — the settings target the region's SES account object itself. Sending email still requires verified identities and, usually, production (non-sandbox) SES access, but neither is needed to apply these settings.

## Deploy

### Console

Open the deployment store, find **AWS SES Account Settings**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the two settings arms. Start from the **Reputation Defaults** preset in the [Presets](#presets) tab for the posture AWS recommends: bounces and complaints auto-suppress account-wide.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSesAccountSettings
metadata:
  name: ses-account-us-east-1
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  suppression:
    reasons:
      - BOUNCE
      - COMPLAINT
```

```shell
planton apply -f ses-account-settings.yaml
```

This sets the region's account-wide suppression posture: hard-bounced and complaining addresses are automatically suppressed and skipped on every future send from the account. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring SES account settings. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One instance per region, period** — the region IS the resource identity; `metadata.name` never reaches AWS. A second instance targeting the same region does not error — it silently fights the first over the same account object, each apply overwriting the other. Give each region exactly one owner.

**Omitted and empty are different postures** — an omitted arm leaves the account's current setting untouched (omission is configuration); a `suppression` arm with an EMPTY `reasons` list explicitly turns account-level auto-suppression OFF. Both are real, and the spec requires at least one arm — an instance managing neither is dead configuration.

**Suppression outlives destroy** — SES has no delete for suppression attributes: destroying this component leaves the last-applied reasons in effect indefinitely. To actually stop suppressing, apply an empty `reasons` list BEFORE destroying. VDM behaves the opposite way — destroy resets it to disabled — and the asymmetry is upstream SES behavior, not a module choice.

**BOUNCE + COMPLAINT is the reputation default** — removing BOUNCE risks re-sending to hard-bounced addresses, which is how sending reputations die. Deviate only with a concrete reason, and remember configuration sets can still layer their own suppression on top.

**VDM is a billing decision, not a feature flag** — enabling it starts AWS charges for the account. The `engagementMetrics` and `optimizedSharedDelivery` sub-toggles only matter while `enabled` is true; drop `optimizedSharedDelivery` to keep the analytics without Guardian changing delivery behavior.

**Suppression is account-wide** — it applies to every identity and configuration set in the region, including ones other teams manage. Coordinate before turning it off; another team may be depending on it.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — it configures the region's own SES account object, which needs no reference to locate.

### What This Kind Provides

`status.outputs` contains a single value: `account_id`, the 12-digit AWS account ID the settings belong to (also the provider's import ID for the suppression singleton). It is an identity echo for auditing and imports, not a composition input — downstream components compose with identities and configuration sets, never with the account settings.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Reputation defaults everywhere** — BOUNCE + COMPLAINT suppression in every production SES region, applied before sending volume scales: reputation damage is far easier to prevent than repair. Start from the **Reputation Defaults** preset.

**Analytics for serious senders** — layer VDM on top of the suppression defaults when the sending program is large enough that deliverability analytics pay for themselves: the engagement dashboard for open/click tuning, Guardian for delivery-behavior protection. Start from the **VDM Analytics** preset.

**Deliberate teardown** — when decommissioning SES in a region, apply the suppression arm with an empty `reasons` list first, verify, then destroy. Skipping the first step leaves auto-suppression running forever with no component managing it.

## Works With

- [**AWS SES Email Identity**](/infra-catalog/aws-ses-email-identity) — the verified domains and addresses whose sends the account-wide suppression list protects
- [**AWS SES Configuration Set**](/infra-catalog/aws-ses-configuration-set) — per-stream sending configuration that can layer its own suppression overrides on top of the account posture
