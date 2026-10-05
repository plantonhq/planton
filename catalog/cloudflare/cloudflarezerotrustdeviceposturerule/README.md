# Cloudflare Zero Trust Device Posture Rule

## Overview

`CloudflareZeroTrustDevicePostureRule` creates a device posture rule: a health check (disk encrypted? OS current? EDR agent healthy?) that WARP evaluates on enrolled devices and that Access and Gateway policies can then require before admitting a device. A plain CRUD object -- real create, update, delete.

## Key Features

- **23 check types** -- WARP-client checks (file, application, os_version, disk_encryption, firewall, domain_joined, serial_number, client certificates, antivirus, and kin) plus service-to-service integrations (CrowdStrike, Intune, Workspace ONE, SentinelOne, Tanium, Kolide, custom)
- **Platform targeting** -- run a rule on windows/mac/linux/android/ios/chromeos, per rule
- **Polling and freshness** -- `schedule` controls how often the client re-checks (min 1m); `expiration` bounds how long a result stays valid
- **The full input surface** -- every parameter of every check family, enum-walled to the provider's own value lists

## Use Cases

**Ideal for:**

- "Disk must be encrypted" / "OS at least version X" gates in front of internal apps
- Requiring a healthy EDR agent (via a posture integration) before granting network access

**Not ideal for:**

- The policies that CONSUME the checks -- those are `CloudflareZeroTrustAccessPolicy` and `CloudflareZeroTrustGatewayPolicy`
- The third-party integration credentials -- posture integrations are managed outside this catalog today (`connection_id` takes their UUID)

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `name` | string | Yes | The rule's name. Required here although the provider marks it optional -- an unnamed rule is indistinguishable everywhere rules are listed. |
| `type` | string | Yes | The check to run (23 provider values, enum-walled). |

### Key Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `match` | list | Platform rows (`windows`, `mac`, `linux`, `android`, `ios`, `chromeos`). |
| `input` | object | The check's parameters; which fields apply depends on `type` (each field's comment names its check families). |
| `schedule` | string | Polling frequency (e.g. `5m`; min `1m`). |
| `expiration` | string | How long a result stays valid (e.g. `1h`). |

### Outputs

| Field | Description |
|-------|-------------|
| `rule_id` | The Cloudflare-assigned UUID (what Access and Gateway policies reference) |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDevicePostureRule
metadata:
  name: macos-current
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: macos-current
  type: os_version
  schedule: 5m
  match:
    - platform: mac
  input:
    version: 14.4.1
    operator: ">="
```

## Destroy Semantics

Destroy is a real delete. Policies referencing the rule stop evaluating it -- review consumers before deleting.

## Related Resources

- **CloudflareZeroTrustAccessPolicy** / **CloudflareZeroTrustGatewayPolicy** -- the policies that require these checks
- **CloudflareZeroTrustDeviceDefaultProfile** -- the WARP client configuration the checks run under

## Further Reading

For operational judgment -- the type/input pairing, deletion-order discipline, integration-backed checks -- see GUIDE.md.

## References

- [Cloudflare device posture](https://developers.cloudflare.com/cloudflare-one/identity/devices/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
