# GCP Firebase Web App

Registers a web app in a Firebase-enabled Google Cloud project and composes the app's anti-abuse attestation: App Check with reCAPTCHA v3 or reCAPTCHA Enterprise, plus debug tokens for local development. The registration is what lets a browser client receive Firebase Cloud Messaging and use every other Firebase product — the `firebaseConfig` object the web SDK is initialised with comes out as outputs for the front-end build to consume. The app lives inside a GcpFirebaseProject; declare that first and reference it from `projectId`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **The app registration** -- one `firebase_web_app` (`projects.webApps`), identified only by its display name
- **reCAPTCHA v3 attestation** -- the app's `firebase_app_check_recaptcha_v3_config` when `appCheck.recaptchaV3` is configured
- **reCAPTCHA Enterprise attestation** -- the app's `firebase_app_check_recaptcha_enterprise_config` when `appCheck.recaptchaEnterprise` is configured
- **Debug tokens** -- one `firebase_app_check_debug_token` per `appCheck.debugTokens` entry
- **API enablement** -- `firebaseappcheck.googleapis.com` exactly when App Check is composed (never disabled on destroy)
- **firebaseConfig** -- the seven client initialisation values, read after registration

## Before You Deploy

### Delete Is Permanent — Read This First

- **`deletionPolicy: DELETE` (the default) removes the app IMMEDIATELY and PERMANENTLY.** Firebase normally keeps a removed app recoverable for 30 days; this path skips that window. A shipped site whose pages carry this app's `firebaseConfig` should carry `PREVENT`.
- **A web app has no identity beyond its display name.** Several may coexist in one project; name each for the client it serves.
- **App Check configurations are never deleted on Google's side.** Removing a reCAPTCHA block from the spec forgets it; destroying the app removes it with the app.

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **Firebase must be enabled on the project** -- a GcpFirebaseProject deployed first. The registration exists only inside the enablement.
- **IAM**: the deploying identity needs `roles/firebase.admin` or broader; `roles/firebaseappcheck.admin` when `appCheck` is configured.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirebaseWebApp
metadata:
  name: console-web
spec:
  projectId:
    valueFrom:
      kind: GcpFirebaseProject
      name: app-firebase
      fieldPath: status.outputs.project_id
  displayName: Acme Console
  deletionPolicy: PREVENT
```

```shell
planton apply -f web-app.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `displayName` | `string` | The app's name as the Firebase console shows it -- its only identity. Updatable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | The Firebase-enabled project. The value is the GCP project id; the natural reference is a GcpFirebaseProject (its `project_id` output), which orders the app after the enablement. Immutable. |
| `apiKeyId` | `StringValueOrRef` | Firebase-provisioned | The API key the app presents, by UID -- a GcpApiKey's `uid` output. A client identifier that ships in the page, not a secret. Omit to let Firebase associate or provision an unrestricted key. |
| `appCheck.recaptchaV3` | `object` | — | reCAPTCHA v3 attestation: `siteSecret` (SENSITIVE -- a managed secret reference) and `tokenTtl` (`1800s`–`604800s`, default `3600s`). |
| `appCheck.recaptchaEnterprise` | `object` | — | reCAPTCHA Enterprise attestation: `siteKey` (the PUBLIC site key the page embeds) and `tokenTtl`. |
| `appCheck.debugTokens` | `list` | `[]` | Debug tokens: `displayName` (unique within the app) and `token` (a UUID4, SENSITIVE -- a managed secret reference). |
| `deletionPolicy` | `string` | `DELETE` | The app and its debug tokens: `DELETE` (immediate and permanent), `PREVENT` (refuse), or `ABANDON`. |

### Validation Rules

- **`recaptchaV3.siteSecret`** and **`recaptchaEnterprise.siteKey`** are required inside their blocks; both blocks may be set at once (the v3-to-Enterprise migration shape).
- **`tokenTtl`**: a duration in seconds ending in `s` (e.g. `3600s`); Google accepts 30 minutes to 7 days.
- **Debug tokens**: `displayName` and `token` required; display names unique within the app.
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `app_id` | `string` | The Firebase-assigned app id (`firebaseConfig.appId`) |
| `name` | `string` | The app's full resource name, `projects/{project}/webApps/{app_id}` |
| `api_key_id` | `string` | The UID of the API key associated with the app |
| `app_urls` | `list<string>` | The URLs Firebase records the app as hosted at (empty when none) |
| `api_key` | `string` | `firebaseConfig.apiKey` -- the key string the page presents; a client identifier |
| `auth_domain` | `string` | `firebaseConfig.authDomain` |
| `database_url` | `string` | `firebaseConfig.databaseURL` (empty without a Realtime Database) |
| `storage_bucket` | `string` | `firebaseConfig.storageBucket` (empty without a default bucket) |
| `location_id` | `string` | `firebaseConfig.locationId` (empty until the default location is finalized) |
| `messaging_sender_id` | `string` | `firebaseConfig.messagingSenderId` -- the project number a browser client registers with for push |
| `measurement_id` | `string` | `firebaseConfig.measurementId` (empty without a linked Google Analytics property) |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Every `firebaseConfig` value ships in the page by design**, the API key included. They are client identifiers; access to the project's backends is governed by IAM, Security Rules, and App Check, not by hiding them.
- **A key referenced by `apiKeyId` must be valid for this app**: unrestricted, or restricted to the site's HTTP referrers with API restrictions that include the Firebase APIs the app uses (Firebase Installations, FCM Registration for push). The GcpApiKey `browser-key` preset is that shape.
- **A web client initialises App Check with ONE provider** (`ReCaptchaV3Provider` or `ReCaptchaEnterpriseProvider`); both configurations may exist on the backend so a site can migrate from v3 to Enterprise without a gap.
- **The Terraform module uses the `google-beta` provider for the registration and its config lookup** -- Google publishes them only there, under the catalog's beta admission list; App Check and API enablement stay on the GA provider.
- **Both engines set `user_project_override`** -- the Firebase Management API attributes quota to the caller's project on user-credential calls.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpFirebaseProject](/docs/catalog/gcp/gcpfirebaseproject) — the Firebase enablement this app is registered in
- [GcpApiKey](/docs/catalog/gcp/gcpapikey) — the referrer-restricted key the app references
- [GcpFirebaseAndroidApp](/docs/catalog/gcp/gcpfirebaseandroidapp) — the same product's Android registration
- [GcpFirebaseAppleApp](/docs/catalog/gcp/gcpfirebaseappleapp) — the same product's iOS registration

## Additional Resources

- [Firebase Management API: webApps](https://firebase.google.com/docs/projects/api/reference/rest/v1beta1/projects.webApps)
- [Firebase App Check for web (reCAPTCHA)](https://firebase.google.com/docs/app-check/web/recaptcha-provider)
- [Firebase Cloud Messaging on the web](https://firebase.google.com/docs/cloud-messaging/js/client)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
