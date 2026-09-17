# GCP API Key

Creates a Google Cloud API key — the project-scoped credential a client application presents to Google APIs: the key a Firebase mobile or web app is registered with, the key a Maps or Translation client calls with, or a server key bound to a service account. The kind is built around RESTRICTION: which platform may present the key (one Android app and its signing certificate, iOS bundle ids, web referrers, or server IPs) and which APIs it may call.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API key** -- the `apikeys_key` in the project's `global` location, named by `keyId`, with its display name and optional service-account binding
- **Restrictions** -- at most one client arm (`androidKeyRestrictions`, `iosKeyRestrictions`, `browserKeyRestrictions`, `serverKeyRestrictions`) plus any number of `apiTargets`, updated in place whenever the spec changes

## Before You Deploy

### A Key Is an Identifier That Ships in the Client — Read This First

- **The key string is not a secret in the password sense.** It ships inside the app that uses it (a mobile binary, a web page), so anyone with the app has the key. It IS a credential Google bills and rate-limits against your project. Restriction — not secrecy — is the control: declare the client arm and the API targets. Both engines still mark the `key_string` output sensitive so it never prints in plans or logs.
- **Every identity field is immutable.** Changing `keyId`, `projectId`, or `serviceAccountEmail` destroys and recreates the key, which rotates the key string every shipped client holds. Restrictions and the display name update in place.
- **A deleted key's id is reserved for 30 days.** Deletion is a soft delete (recoverable from the console or the API); a new key cannot reuse the id in that window.

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **The APIs the key unlocks must be enabled on the project.** A key enables nothing; `apiTargets` names services that are already enabled (the Firebase kinds enable the Firebase APIs they need).
- **IAM**: the deploying identity needs `roles/serviceusage.apiKeysAdmin` or broader — it includes `apikeys.keys.getKeyString`, which both modules need to export the key string.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpApiKey
metadata:
  name: firebase-android-key
spec:
  keyId: firebase-android-key
  restrictions:
    androidKeyRestrictions:
      allowedApplications:
        - packageName: com.acme.app
          sha1Fingerprint: DA:39:A3:EE:5E:6B:4B:0D:32:55:BF:EF:95:60:18:90:AF:D8:07:09
    apiTargets:
      - service: firebaseinstallations.googleapis.com
      - service: fcmregistrations.googleapis.com
```

```shell
planton apply -f api-key.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `keyId` | `string` | The key's resource id — lowercase letters, digits, hyphens; starts with a letter; 1-63 characters. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project the key belongs to. Can reference a GcpProject resource. Immutable. |
| `displayName` | `string` | — | Console label. Freely updatable. |
| `serviceAccountEmail` | `StringValueOrRef` | — | Bind the key to a service account (requests carrying it authenticate as that account). Can reference a GcpServiceAccount. Immutable. |
| `restrictions.androidKeyRestrictions` | `message` | — | `allowedApplications[]` of `packageName` + `sha1Fingerprint` pairs (at least one). |
| `restrictions.iosKeyRestrictions` | `message` | — | `allowedBundleIds[]` (at least one). |
| `restrictions.browserKeyRestrictions` | `message` | — | `allowedReferrers[]` — referrer URL patterns with wildcards (at least one). |
| `restrictions.serverKeyRestrictions` | `message` | — | `allowedIps[]` — IPv4/IPv6 addresses or CIDRs (at least one). |
| `restrictions.apiTargets` | `list` | `[]` | `service` (a `*.googleapis.com` name) plus optional `methods[]`. Empty means every enabled API. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (soft-delete, recoverable 30 days), `PREVENT` (destroy fails), or `ABANDON` (unmanaged, stays live). |

### Validation Rules

- **One client arm**: at most one of the four `*KeyRestrictions` arms on a key (Google's rule — a key identifies one kind of caller).
- **No hollow arms**: every arm's list needs at least one entry.
- **`keyId`**: RFC 1034 label, `^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$`.
- **Android**: `packageName` is a dotted Java package name; `sha1Fingerprint` is 40 hex characters, optionally colon-separated in pairs.
- **API targets**: `service` ends in `.googleapis.com`.
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/global/keys/{keyId}` — the key's full resource name |
| `uid` | `string` | The key's unique id — what a Firebase app registration's `apiKeyId` references |
| `key_string` | `string` | The key string clients present. Sensitive in both engines. |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Unrestricted keys are accepted and discouraged** — omitting `restrictions` makes a key any caller anywhere can spend your quota with. It is also what Firebase auto-provisions when an app registration names no key; declaring a restricted key here and referencing its `uid` from the app is the recommended posture.
- **Restrictions never rotate the key string** — tightening `apiTargets` or adding a certificate fingerprint is an in-place update; only the immutable identity fields recreate the key.
- **Both engines set `user_project_override`** — the API Keys API attributes quota to the caller's project on user-credential calls; without the override a deploy under plain ADC fails with "requires a quota project" (the Identity Toolkit precedent).
- **Deletion is soft** — `DELETE` keeps the key recoverable for 30 days; recovery is an operator act (console or `undelete`), not something the modules do.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpFirebaseProject](/docs/catalog/gcp/gcpfirebaseproject) — Firebase on the same project; the app registrations that reference this key live inside it
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the project the key belongs to
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the identity a service-account-bound key authenticates as

## Additional Resources

- [API Keys Overview](https://cloud.google.com/api-keys/docs/overview)
- [Applying API key restrictions](https://cloud.google.com/api-keys/docs/add-restrictions-api-keys)
- [API Keys API Reference](https://cloud.google.com/api-keys/docs/reference/rest)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
