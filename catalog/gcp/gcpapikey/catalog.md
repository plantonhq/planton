# GCP API Key

Creates a Google Cloud API key: the project-scoped credential a client application presents to Google APIs -- the key a Firebase mobile or web app is registered with, the key a Maps or Translation client calls with, or a server key bound to a service account. The kind is built around RESTRICTION: which platform may present the key (one Android app and its signing certificate, iOS bundle ids, web referrers, or server IPs) and which APIs it may call.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API key** -- the `apikeys_key` in the project's `global` location, named by `keyId`, with its display name and optional service-account binding
- **Restrictions** -- at most one client arm (`androidKeyRestrictions`, `iosKeyRestrictions`, `browserKeyRestrictions`, `serverKeyRestrictions`) plus any number of `apiTargets`, updated in place whenever the spec changes

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Project

- **The APIs the key unlocks must be enabled on the project** -- a key does not enable anything; `apiTargets` names services already enabled (the Firebase kinds enable the Firebase APIs they need).
- **`keyId` is immutable and reserved for 30 days after deletion** -- a deleted key is soft-deleted and recoverable; a new key cannot reuse the id in that window. Ephemeral keys need a fresh id per lifetime.
- **IAM**: the deploying identity needs `roles/serviceusage.apiKeysAdmin` or broader (it includes `apikeys.keys.getKeyString`, which the modules need to export the key string).

## Deploy

### Console

Open the deployment store, find **GCP API Key**, and click **Deploy**. The creation wizard walks you through the target project, the key id, the client restriction arm, and the API targets. Start from the **Firebase Android Key** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpApiKey
metadata:
  name: firebase-android-key
  org: acme-corp
  env: prod
spec:
  keyId: firebase-android-key
  displayName: Firebase Android key
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

This creates a key only the signed Android app can present, and only against the Firebase installation and FCM registration APIs. The `uid` output is what a Firebase app registration references; the `key_string` output is what ships in the app. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the key references its project and (for a bound key) its service account via ValueFromRef, and the Firebase app registrations reference the key's `uid`:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: app-project
      fieldPath: status.outputs.project_id
  keyId: firebase-android-key
  restrictions:
    androidKeyRestrictions:
      allowedApplications:
        - packageName: com.acme.app
          sha1Fingerprint: DA39A3EE5E6B4B0D3255BFEF95601890AFD80709
```

The InfraPipeline deploys the project first, then the key, then any Firebase app that points its `apiKeyId` at `status.outputs.uid`.

## Key Configuration

These are the most important decisions when configuring an API key. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One client arm** -- Google allows exactly one kind of caller per key: Android (package name AND signing-certificate SHA-1, together), iOS (bundle ids), browser (referrer patterns), or server (IP addresses/CIDRs). The spec rejects two arms on one key; make one key per client platform.

**API targets** -- the services the key may call. Empty means every API enabled on the project; list the ones the client actually uses. Restrictions are updated in place -- tightening them never rotates the key string.

**Service-account binding** -- `serviceAccountEmail` makes requests carrying the key authenticate AS that account (for APIs that accept API-key auth in place of OAuth). Immutable, and still a bearer credential: pair it with a server IP restriction.

**Deletion policy** -- `DELETE` (default) soft-deletes the key for 30 days; `PREVENT` fails the destroy -- the guard for a key baked into a shipped mobile binary; `ABANDON` leaves the key live and unmanaged.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** (optional) | `serviceAccountEmail` | `status.outputs.email` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | `projects/{project}/locations/global/keys/{keyId}` | Addressing the key in IAM conditions and tooling |
| `uid` | The key's unique id | A Firebase app registration's `apiKeyId` |
| `key_string` | The key string clients present (sensitive) | The client build's configuration, through the platform's secret handling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Firebase Android key** -- package name plus signing certificate, FCM and installations API targets. Start from the **Firebase Android Key** preset.

**Firebase iOS key** -- bundle id, the same API targets. Start from the **Firebase iOS Key** preset.

**Browser key** -- referrer patterns for a web app (Maps JavaScript, a Firebase web app). Start from the **Browser Key** preset.

**Server key** -- caller IPs plus tightly scoped API targets for a backend that cannot use OAuth. Start from the **Server Key** preset.

## Works With

- [**GCP Firebase Project**](/cloud-catalog/gcp-firebase-project) -- Firebase enabled on the same project; the app registrations that reference this key live inside it
- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the project the key belongs to
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- the identity a service-account-bound key authenticates as
