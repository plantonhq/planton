# GcpFirebaseAndroidApp

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpFirebaseAndroidAppSpec registers an Android app in a Firebase-enabled
Google Cloud project and composes the app's anti-abuse attestation (App
Check with Play Integrity, plus debug tokens for development builds).
The registration is what lets the Android build receive Firebase Cloud
Messaging and use every other Firebase product: its google-services.json
configuration file is produced as an output for the build to consume.

The app lives INSIDE a GcpFirebaseProject (the project's Firebase
enablement); declare that first and reference it from project_id.

Two facts to know before acting. The package name is the app's permanent
identity in Firebase: it cannot change, and a project accepts each
package name exactly once. And destroying this resource with
deletion_policy DELETE removes the app IMMEDIATELY and PERMANENTLY --
Firebase's 30-day recoverable window is skipped -- so a shipped app's
registration should carry PREVENT.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirebaseAndroidApp
metadata:
  name: my-android-app
spec:
  # The Firebase-enabled project the app is registered in. The value is the
  # GCP project id; the natural reference is the GcpFirebaseProject that
  # enabled Firebase on it -- referencing it orders this app after the
  # enablement and places it inside the Firebase project on diagrams. Omit
  # to use the provider's default project. Immutable.
  projectId:
    valueFrom:
      kind: GcpFirebaseProject
      name: app-firebase
      fieldPath: status.outputs.project_id

  # As the Firebase console shows the app. Updatable.
  displayName: My Android App

  # The app's IDENTITY in Firebase and in google-services.json. IMMUTABLE
  # (a change replaces the registration) and accepted once per project.
  packageName: com.example.app

  # Signing certificate fingerprints. Cloud Messaging needs none; Play
  # Integrity needs the SHA-256 of the certificate the shipped APK is
  # signed with (Play App Signing's for a Play release); Google Sign-In,
  # Dynamic Links, and Phone Auth need the SHA-1.
  sha1Hashes:
    - DA:39:A3:EE:5E:6B:4B:0D:32:55:BF:EF:95:60:18:90:AF:D8:07:09
  sha256Hashes:
    - E3:B0:C4:42:98:FC:1C:14:9A:FB:F4:C8:99:6F:B9:24:27:AE:41:E4:64:9B:93:4C:A4:95:99:1B:78:52:B8:55

  # The restricted API key the app presents, by the key's UID -- a
  # reference to a GcpApiKey (its uid output). A client identifier that
  # ships in the APK, not a secret. Omit to let Firebase associate or
  # provision an unrestricted key.
  apiKeyId:
    valueFrom:
      kind: GcpApiKey
      name: firebase-android-key
      fieldPath: status.outputs.uid

  # App Check: prove requests come from the genuine app on a genuine
  # device. Whether backends ENFORCE it is configured on the
  # GcpFirebaseProject.
  appCheck:
    # Play Integrity: Google's attestation for apps distributed through
    # Google Play. Requires the SHA-256 fingerprint above. `enabled`
    # defaults to true when the block is present.
    playIntegrity:
      enabled: true
      tokenTtl: 3600s
    # Debug tokens let development builds and CI emulators pass App Check.
    # Each token is a SECRET -- a managed secret reference on the platform.
    debugTokens:
      - displayName: ci-emulator
        token: $secret/firebase-app-check-ci-emulator-token

  # What destroy does to the app and its debug tokens. DELETE (default) is
  # IMMEDIATE and PERMANENT -- Firebase's 30-day recoverable window is
  # skipped. A shipped app should carry PREVENT.
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpFirebaseProject (`status.outputs.project_id`) |
| `spec.displayName` | `string` | yes |  |  |
| `spec.packageName` | `string` | yes |  |  |
| `spec.sha1Hashes` | `[]string` |  |  |  |
| `spec.sha256Hashes` | `[]string` |  |  |  |
| `spec.apiKeyId` | `string \| valueFrom` |  |  | GcpApiKey (`status.outputs.uid`) |
| `spec.appCheck` | `GcpFirebaseAndroidAppAppCheck` |  |  |  |
| `spec.appCheck.playIntegrity` | `GcpFirebaseAndroidAppAppCheckPlayIntegrity` |  |  |  |
| `spec.appCheck.playIntegrity.enabled` | `bool` |  | `true` |  |
| `spec.appCheck.playIntegrity.tokenTtl` | `string` |  |  |  |
| `spec.appCheck.debugTokens` | `[]GcpFirebaseAndroidAppAppCheckDebugToken` |  |  |  |
| `spec.appCheck.debugTokens[].displayName` | `string` | yes |  |  |
| `spec.appCheck.debugTokens[].token` | `string` (sensitive) | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the app is registered in -- the Firebase-enabled
project. Its VALUE is the GCP project id (the same string every GCP
kind's project_id carries), so a literal works; the default REFERENCE
is a GcpFirebaseProject resource, because an app registration exists
only inside a project's Firebase enablement: referencing the enablement
orders this app after it in a chart and places it inside the Firebase
project on diagrams. If omitted, the provider's default project is
used. Immutable.

- references: GcpFirebaseProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFirebaseProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.displayName

`string` · required

The user-assigned display name of the app, as the Firebase console
shows it. Updatable.

- rule: {"required":true}

### spec.packageName

`string` · required

The canonical package name of the Android app, as it appears in the
Play Console (e.g. com.example.app). This is the app's IDENTITY in
Firebase and in google-services.json: it is immutable (a change
replaces the registration -- a new app id, a new config file), and a
project accepts each package name exactly once. Must be a valid Java
package name: two or more dot-separated segments, each starting with a
letter and made of letters, digits, and underscores.

- rule: package_name must be a valid Java package name -- two or more dot-separated segments, each starting with a letter and made of letters, digits, and underscores (e.g. com.example.app)
- rule: {"required":true}

### spec.sha1Hashes

`[]string`

SHA-1 fingerprints of the app's signing certificates (debug, release,
Play App Signing), 40 hex digits, with or without colon separators.
Cloud Messaging needs NONE of these; they are required by Google
Sign-In, Dynamic Links, and Phone Authentication. Updatable.

- rule: {"repeated":{"unique":true,"items":{"cel":[{"id":"valid_sha1_hash","message":"each sha1_hashes entry must be a SHA-1 fingerprint: 40 hex digits, optionally colon-separated in pairs","expression":"this.matches('^([0-9A-Fa-f]{40}|([0-9A-Fa-f]{2}:){19}[0-9A-Fa-f]{2})$')"}]}}}

### spec.sha256Hashes

`[]string`

SHA-256 fingerprints of the app's signing certificates, 64 hex digits,
with or without colon separators. Cloud Messaging needs none; App Check
with Play Integrity REQUIRES the fingerprint of the certificate the
shipped APK is signed with (Play App Signing's certificate for a Play
release) -- without it attestation fails at runtime. Updatable.

- rule: {"repeated":{"unique":true,"items":{"cel":[{"id":"valid_sha256_hash","message":"each sha256_hashes entry must be a SHA-256 fingerprint: 64 hex digits, optionally colon-separated in pairs","expression":"this.matches('^([0-9A-Fa-f]{64}|([0-9A-Fa-f]{2}:){31}[0-9A-Fa-f]{2})$')"}]}}}

### spec.apiKeyId

`string | valueFrom`

The API key the app presents to Firebase, by the key's Google-assigned
UID: a literal, or a reference to a GcpApiKey resource (its uid
output). The key is a CLIENT IDENTIFIER that ships inside the APK in
google-services.json -- not a secret -- and it must be valid for this
app: unrestricted, or restricted to this package name and certificate
with API restrictions that include the Firebase APIs the app uses
(Firebase Installations and FCM Registration for push). Google's
recommended practice is one restricted key per app. When omitted,
Firebase associates an existing valid key or provisions a new,
unrestricted one; the api_key_id output reports which. Updatable.

- references: GcpApiKey (`status.outputs.uid`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpApiKey, name: <that resource's name>, fieldPath: status.outputs.uid}} -- a bare string does not parse

### spec.appCheck

`GcpFirebaseAndroidAppAppCheck`

Firebase App Check for this app: the attestation provider that proves
requests come from the genuine app on a genuine device (Play
Integrity), and debug tokens that let development builds and CI pass
App Check without a real device. Whether the project's backends
ENFORCE App Check is configured on the GcpFirebaseProject.

- rule: app_check.debug_tokens must use each display_name at most once

### spec.appCheck.playIntegrity

`GcpFirebaseAndroidAppAppCheckPlayIntegrity`

Attest with the Play Integrity API -- Google's attestation provider for
Android apps distributed through Google Play. Requires the app's
SHA-256 signing fingerprint in sha256_hashes. Omit the block to leave
the provider unconfigured.

### spec.appCheck.playIntegrity.enabled

`bool` · optional (explicit presence)

Whether Play Integrity attestation is configured for the app. Defaults
to true when the block is present, so `playIntegrity: {}` means
"configured, with Google's default token lifetime"; set false to
declare the block and leave the provider unconfigured. The wire has no
switch of its own -- the configuration's existence is the switch -- so
this field is how the manifest states the intent explicitly.

- default: `true`

### spec.appCheck.playIntegrity.tokenTtl

`string`

How long an App Check token exchanged from a Play Integrity verdict
stays valid, as a duration in seconds with an "s" suffix (e.g. "3600s",
"1800s", "604800s"). Google accepts 30 minutes to 7 days inclusive and
assumes 1 hour when unset. Shorter lifetimes re-attest more often;
longer ones spare the device.

- rule: token_ttl must be a duration in seconds ending in 's' (e.g. 3600s), between 1800s and 604800s

### spec.appCheck.debugTokens

`[]GcpFirebaseAndroidAppAppCheckDebugToken`

Debug tokens: pre-shared secrets a development build or a CI emulator
presents in place of a device attestation. Each token is a SECRET --
anyone holding it passes App Check as this app -- so it is a managed
secret reference on the platform, never plaintext. Revoke by removing
the entry.

### spec.appCheck.debugTokens[].displayName

`string` · required

A name for the token as the Firebase console shows it -- the
developer's machine or the CI lane it is for. Unique within the app;
updatable.

- rule: {"required":true}

### spec.appCheck.debugTokens[].token

`string` · required · sensitive

The token value itself: a UUID (version 4), case-insensitive, that the
development build registers with the App Check debug provider. Google
never returns it after creation, and it cannot be changed -- a new
value is a new token. It is a SECRET: the value must be a managed
secret reference on the platform.

- rule: {"required":true}

### spec.deletionPolicy

`string`

What destroying this resource does to the app registration (and to
the composed debug tokens):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the app is removed IMMEDIATELY and PERMANENTLY: Firebase
               normally keeps a removed app recoverable for 30 days;
               this path skips that window. The package name becomes
               registrable again at once.
  "PREVENT" -- destroy FAILS; use for a shipped app whose users hold
               its google-services.json
  "ABANDON" -- the app is left registered and removed from management
The Play Integrity configuration is a per-app singleton Google never
deletes: it is removed from management on destroy and disappears only
with the app itself.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpFirebaseAndroidApp, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.app_id` | `string` | The Firebase-assigned app id (mobilesdk_app_id in google-services.json), e.g. 1:123456789012:android:0123456789abcdef. Globally unique and immutable. |
| `status.outputs.name` | `string` | The app's full resource name, projects/{project}/androidApps/{app_id}. The handle the Firebase Management API addresses the app by. |
| `status.outputs.api_key_id` | `string` | The Google-assigned UID of the API key associated with the app -- the one from the spec, or the key Firebase associated or provisioned when none was given. |
| `status.outputs.config_filename` | `string` | The configuration file's name: google-services.json. Place it at the app module's root (app/google-services.json) for the Google Services Gradle plugin. |
| `status.outputs.config_file_contents` | `string` | The configuration file's contents, base64-encoded. Decode and write to config_filename in the Android build. A build input that ships in the APK, not a secret. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpFirebaseProject | `status.outputs.project_id` |
| `spec.apiKeyId` | GcpApiKey | `status.outputs.uid` |

## See Also

- [Overview](../README.md)
