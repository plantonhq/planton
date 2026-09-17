# GCP Firebase Android App

Registers an Android app in a Firebase-enabled Google Cloud project and composes the app's anti-abuse attestation: App Check with Play Integrity, plus debug tokens for development builds and CI. The registration is what lets the Android build receive Firebase Cloud Messaging and use every other Firebase product — its `google-services.json` comes out as an output for the build to consume. The app lives inside a GcpFirebaseProject; declare that first and reference it from `projectId`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **The app registration** -- one `firebase_android_app` (`projects.androidApps`) identified by its immutable package name
- **Play Integrity attestation** -- the app's `firebase_app_check_play_integrity_config` when `appCheck.playIntegrity` is configured
- **Debug tokens** -- one `firebase_app_check_debug_token` per `appCheck.debugTokens` entry
- **API enablement** -- `firebaseappcheck.googleapis.com` exactly when App Check is composed (never disabled on destroy)
- **The configuration file** -- `google-services.json`, read after registration and exported base64-encoded

## Before You Deploy

### Identity Is Immutable, Delete Is Permanent — Read This First

- **The package name is the app's identity in Firebase.** It cannot change: a new package name is a new registration (a new app id, a new `google-services.json`). A project accepts each package name exactly once.
- **`deletionPolicy: DELETE` (the default) removes the app IMMEDIATELY and PERMANENTLY.** Firebase normally keeps a removed app recoverable for 30 days; this path skips that window. A shipped app whose users hold its `google-services.json` should carry `PREVENT`.
- **App Check configurations are never deleted on Google's side.** Removing `playIntegrity` from the spec forgets it; destroying the app removes it with the app.

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
kind: GcpFirebaseAndroidApp
metadata:
  name: mobile-android
spec:
  projectId:
    valueFrom:
      kind: GcpFirebaseProject
      name: app-firebase
      fieldPath: status.outputs.project_id
  displayName: Acme Mobile (Android)
  packageName: com.acme.mobile
  deletionPolicy: PREVENT
```

```shell
planton apply -f android-app.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `displayName` | `string` | The app's name as the Firebase console shows it. Updatable. |
| `packageName` | `string` | The Android package name (a Java package: `com.acme.mobile`). IMMUTABLE; accepted once per project. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | The Firebase-enabled project. The value is the GCP project id; the natural reference is a GcpFirebaseProject (its `project_id` output), which orders the app after the enablement. Immutable. |
| `sha1Hashes` | `list<string>` | `[]` | SHA-1 signing certificate fingerprints (40 hex, optionally colon-separated). Not needed for push; needed by Google Sign-In, Dynamic Links, Phone Auth. |
| `sha256Hashes` | `list<string>` | `[]` | SHA-256 signing certificate fingerprints (64 hex). Not needed for push; REQUIRED for Play Integrity attestation to work at runtime. |
| `apiKeyId` | `StringValueOrRef` | Firebase-provisioned | The API key the app presents, by UID -- a GcpApiKey's `uid` output. A client identifier that ships in the APK, not a secret. Omit to let Firebase associate or provision an unrestricted key. |
| `appCheck.playIntegrity` | `object` | — | Play Integrity attestation: `enabled` (default `true` when the block is present; `false` declares it off) and `tokenTtl` (`1800s`–`604800s`, default `3600s`). |
| `appCheck.debugTokens` | `list` | `[]` | Debug tokens: `displayName` (unique within the app) and `token` (a UUID4, SENSITIVE -- a managed secret reference). |
| `deletionPolicy` | `string` | `DELETE` | The app and its debug tokens: `DELETE` (immediate and permanent), `PREVENT` (refuse), or `ABANDON`. |

### Validation Rules

- **`packageName`**: a valid Java package name -- two or more dot-separated segments, each starting with a letter, made of letters, digits, underscores.
- **`sha1Hashes` / `sha256Hashes`**: 40 / 64 hex digits, optionally colon-separated in pairs; no duplicates.
- **`tokenTtl`**: a duration in seconds ending in `s` (e.g. `3600s`); Google accepts 30 minutes to 7 days.
- **Debug tokens**: `displayName` and `token` required; display names unique within the app.
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `app_id` | `string` | The Firebase-assigned app id (`mobilesdk_app_id` in `google-services.json`) |
| `name` | `string` | The app's full resource name, `projects/{project}/androidApps/{app_id}` |
| `api_key_id` | `string` | The UID of the API key associated with the app -- from the spec, or the one Firebase associated or provisioned |
| `config_filename` | `string` | `google-services.json` |
| `config_file_contents` | `string` | The configuration file, base64-encoded -- a build input that ships in the APK, not a secret |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Push needs only the package name and the config file.** Certificate fingerprints are for Play Integrity, Google Sign-In, Dynamic Links, and Phone Auth -- not for Cloud Messaging.
- **A key referenced by `apiKeyId` must be valid for this app**: unrestricted, or restricted to this package name and certificate with API restrictions that include the Firebase APIs the app uses (Firebase Installations, FCM Registration for push). The GcpApiKey `firebase-android-key` preset is that shape.
- **`google-services.json` is a build input.** Decode `config_file_contents` and write it to `app/google-services.json` for the Google Services Gradle plugin. It contains the API key and app id by design; access to the project's backends is governed by IAM, Security Rules, and App Check, not by hiding it.
- **The Terraform module uses the `google-beta` provider for the registration and its config lookup** -- Google publishes them only there. Both are recorded in the catalog's beta admission list; App Check and API enablement stay on the GA provider.
- **Both engines set `user_project_override`** -- the Firebase Management API attributes quota to the caller's project on user-credential calls.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpFirebaseProject](/docs/catalog/gcp/gcpfirebaseproject) — the Firebase enablement this app is registered in
- [GcpApiKey](/docs/catalog/gcp/gcpapikey) — the restricted key the app references
- [GcpFirebaseAppleApp](/docs/catalog/gcp/gcpfirebaseappleapp) — the same product's iOS registration
- [GcpFirebaseWebApp](/docs/catalog/gcp/gcpfirebasewebapp) — the same product's web registration

## Additional Resources

- [Firebase Management API: androidApps](https://firebase.google.com/docs/projects/api/reference/rest/v1beta1/projects.androidApps)
- [Firebase App Check for Android (Play Integrity)](https://firebase.google.com/docs/app-check/android/play-integrity-provider)
- [Firebase Cloud Messaging on Android](https://firebase.google.com/docs/cloud-messaging/android/client)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
