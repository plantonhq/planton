# GCP Firebase Apple App

Registers an iOS / macOS app in a Firebase-enabled Google Cloud project and composes the app's anti-abuse attestation: App Check with App Attest and DeviceCheck, plus debug tokens for simulators and CI. The registration is what lets the Apple build receive Firebase Cloud Messaging and use every other Firebase product — its `GoogleService-Info.plist` comes out as an output for the build to consume. The app lives inside a GcpFirebaseProject; declare that first and reference it from `projectId`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **The app registration** -- one `firebase_apple_app` (`projects.iosApps`) identified by its immutable bundle id
- **App Attest attestation** -- the app's `firebase_app_check_app_attest_config` when `appCheck.appAttest` is configured
- **DeviceCheck attestation** -- the app's `firebase_app_check_device_check_config` when `appCheck.deviceCheck` is configured
- **Debug tokens** -- one `firebase_app_check_debug_token` per `appCheck.debugTokens` entry
- **API enablement** -- `firebaseappcheck.googleapis.com` exactly when App Check is composed (never disabled on destroy)
- **The configuration file** -- `GoogleService-Info.plist`, read after registration and exported base64-encoded

## Before You Deploy

### Identity Is Immutable, Delete Is Permanent — Read This First

- **The bundle id is the app's identity in Firebase.** It cannot change: a new bundle id is a new registration (a new app id, a new plist). A project accepts each bundle id exactly once.
- **`deletionPolicy: DELETE` (the default) removes the app IMMEDIATELY and PERMANENTLY.** Firebase normally keeps a removed app recoverable for 30 days; this path skips that window. A shipped app whose users hold its plist should carry `PREVENT`.
- **Push on Apple platforms needs one step this resource cannot do:** the APNs authentication key Apple mints must be uploaded in the Firebase console. Firebase exposes no API for it.
- **App Check configurations are never deleted on Google's side.** Removing `appAttest` or `deviceCheck` from the spec forgets them; destroying the app removes them with the app.

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
kind: GcpFirebaseAppleApp
metadata:
  name: mobile-ios
spec:
  projectId:
    valueFrom:
      kind: GcpFirebaseProject
      name: app-firebase
      fieldPath: status.outputs.project_id
  displayName: Acme Mobile (iOS)
  bundleId: com.acme.mobile
  teamId: ABCDE12345
  deletionPolicy: PREVENT
```

```shell
planton apply -f ios-app.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `displayName` | `string` | The app's name as the Firebase console shows it. Updatable. |
| `bundleId` | `string` | The bundle identifier registered with Apple (`com.acme.mobile`). IMMUTABLE; accepted once per project. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | The Firebase-enabled project. The value is the GCP project id; the natural reference is a GcpFirebaseProject (its `project_id` output), which orders the app after the enablement. Immutable. |
| `appStoreId` | `string` | — | The numeric Apple ID from App Store Connect (Dynamic Links, App Store redirects). Not needed for push. |
| `teamId` | `string` | — | The 10-character Apple Developer Team ID. REQUIRED when `appAttest` or `deviceCheck` is configured. |
| `apiKeyId` | `StringValueOrRef` | Firebase-provisioned | The API key the app presents, by UID -- a GcpApiKey's `uid` output. A client identifier that ships in the bundle, not a secret. Omit to let Firebase associate or provision an unrestricted key. |
| `appCheck.appAttest` | `object` | — | App Attest attestation: `enabled` (default `true` when present; `false` declares it off) and `tokenTtl` (`1800s`–`604800s`, default `3600s`). |
| `appCheck.deviceCheck` | `object` | — | DeviceCheck attestation: `keyId` (the DeviceCheck key's 10-character Key ID), `privateKey` (the `.p8` contents, SENSITIVE -- a managed secret reference; NOT the APNs key), `tokenTtl`. |
| `appCheck.debugTokens` | `list` | `[]` | Debug tokens: `displayName` (unique within the app) and `token` (a UUID4, SENSITIVE -- a managed secret reference). |
| `deletionPolicy` | `string` | `DELETE` | The app and its debug tokens: `DELETE` (immediate and permanent), `PREVENT` (refuse), or `ABANDON`. |

### Validation Rules

- **`bundleId`**: letters, digits, hyphens, and periods (Apple's character set); reverse-DNS is Apple's convention, not a rule.
- **`appStoreId`**: digits only. **`teamId`**: 10 uppercase letters and digits.
- **`teamId` is required** when `appCheck.appAttest` (not declared off) or `appCheck.deviceCheck` is set -- Apple's attestations validate against the signing team.
- **`deviceCheck.keyId`**: 10 uppercase letters and digits; `privateKey` required.
- **`tokenTtl`**: a duration in seconds ending in `s` (e.g. `3600s`); Google accepts 30 minutes to 7 days.
- **Debug tokens**: `displayName` and `token` required; display names unique within the app.
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `app_id` | `string` | The Firebase-assigned app id (`GOOGLE_APP_ID` in the plist) |
| `name` | `string` | The app's full resource name, `projects/{project}/iosApps/{app_id}` |
| `api_key_id` | `string` | The UID of the API key associated with the app -- from the spec, or the one Firebase associated or provisioned |
| `config_filename` | `string` | `GoogleService-Info.plist` |
| `config_file_contents` | `string` | The configuration file, base64-encoded -- a build input that ships in the app bundle, not a secret |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The APNs key is the one console step.** Apple mints the APNs authentication key (`.p8`, Key ID, Team ID); Firebase has no API to accept it, so it is uploaded once in the Firebase console under Cloud Messaging. Without it, iOS push does not deliver.
- **The DeviceCheck key is NOT the APNs key.** Apple issues them separately; `deviceCheck.privateKey` is the DeviceCheck `.p8`.
- **A key referenced by `apiKeyId` must be valid for this app**: unrestricted, or restricted to this bundle id with API restrictions that include the Firebase APIs the app uses (Firebase Installations, FCM Registration for push). The GcpApiKey `firebase-ios-key` preset is that shape.
- **`GoogleService-Info.plist` is a build input.** Decode `config_file_contents` and add it to the Xcode target. It contains the API key and app id by design; access to the project's backends is governed by IAM, Security Rules, and App Check, not by hiding it.
- **The Terraform module uses the `google-beta` provider for the registration and its config lookup** -- Google publishes them only there, under the catalog's beta admission list; App Check and API enablement stay on the GA provider.
- **Both engines set `user_project_override`** -- the Firebase Management API attributes quota to the caller's project on user-credential calls.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpFirebaseProject](/docs/catalog/gcp/gcpfirebaseproject) — the Firebase enablement this app is registered in
- [GcpApiKey](/docs/catalog/gcp/gcpapikey) — the restricted key the app references
- [GcpFirebaseAndroidApp](/docs/catalog/gcp/gcpfirebaseandroidapp) — the same product's Android registration
- [GcpFirebaseWebApp](/docs/catalog/gcp/gcpfirebasewebapp) — the same product's web registration

## Additional Resources

- [Firebase Management API: iosApps](https://firebase.google.com/docs/projects/api/reference/rest/v1beta1/projects.iosApps)
- [Firebase App Check for Apple platforms (App Attest, DeviceCheck)](https://firebase.google.com/docs/app-check/ios/app-attest-provider)
- [Firebase Cloud Messaging on Apple platforms (APNs setup)](https://firebase.google.com/docs/cloud-messaging/ios/client)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
