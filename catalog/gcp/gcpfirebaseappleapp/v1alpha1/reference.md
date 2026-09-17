# GcpFirebaseAppleApp

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpFirebaseAppleAppSpec registers an iOS / macOS app in a Firebase-enabled
Google Cloud project and composes the app's anti-abuse attestation (App
Check with App Attest and DeviceCheck, plus debug tokens for development
builds). The registration is what lets the Apple build receive Firebase
Cloud Messaging and use every other Firebase product: its
GoogleService-Info.plist configuration file is produced as an output for
the build to consume.

The app lives INSIDE a GcpFirebaseProject (the project's Firebase
enablement); declare that first and reference it from project_id.

Three facts to know before acting. The bundle id is the app's permanent
identity in Firebase: it cannot change, and a project accepts each bundle
id exactly once. Destroying this resource with deletion_policy DELETE
removes the app IMMEDIATELY and PERMANENTLY -- Firebase's 30-day
recoverable window is skipped -- so a shipped app's registration should
carry PREVENT. And push on Apple platforms additionally needs the APNs
authentication key Apple mints, uploaded in the Firebase console: Firebase
exposes no API for it, so that one step stays manual and is not part of
this resource.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirebaseAppleApp
metadata:
  name: my-ios-app
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
  displayName: My iOS App

  # The app's IDENTITY in Firebase and in GoogleService-Info.plist.
  # IMMUTABLE (a change replaces the registration) and accepted once per
  # project.
  bundleId: com.example.app

  # The numeric Apple ID from App Store Connect (Dynamic Links, App Store
  # redirects). Leave empty until the app has an App Store record.
  appStoreId: "1234567890"

  # The Apple Developer Team ID that signs the app. REQUIRED when App Attest
  # or DeviceCheck is configured below.
  teamId: ABCDE12345

  # The restricted API key the app presents, by the key's UID -- a
  # reference to a GcpApiKey (its uid output). A client identifier that
  # ships in the app bundle, not a secret. Omit to let Firebase associate
  # or provision an unrestricted key.
  apiKeyId:
    valueFrom:
      kind: GcpApiKey
      name: firebase-ios-key
      fieldPath: status.outputs.uid

  # App Check: prove requests come from the genuine app on a genuine
  # device. Whether backends ENFORCE it is configured on the
  # GcpFirebaseProject.
  appCheck:
    # App Attest: Apple's attestation for iOS 14+, Google's recommended
    # primary provider. `enabled` defaults to true when the block is present.
    appAttest:
      enabled: true
      tokenTtl: 3600s
    # DeviceCheck: the fallback for devices that cannot use App Attest. The
    # private key is the DeviceCheck .p8 from the Apple Developer account --
    # NOT the APNs key -- and is a managed secret reference on the platform.
    deviceCheck:
      keyId: ABC123DEFG
      privateKey: $secret/firebase-devicecheck-private-key
      tokenTtl: 3600s
    # Debug tokens let simulators and CI pass App Check. Each token is a
    # SECRET -- a managed secret reference on the platform.
    debugTokens:
      - displayName: simulator
        token: $secret/firebase-app-check-simulator-token

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
| `spec.bundleId` | `string` | yes |  |  |
| `spec.appStoreId` | `string` |  |  |  |
| `spec.teamId` | `string` |  |  |  |
| `spec.apiKeyId` | `string \| valueFrom` |  |  | GcpApiKey (`status.outputs.uid`) |
| `spec.appCheck` | `GcpFirebaseAppleAppAppCheck` |  |  |  |
| `spec.appCheck.appAttest` | `GcpFirebaseAppleAppAppCheckAppAttest` |  |  |  |
| `spec.appCheck.appAttest.enabled` | `bool` |  | `true` |  |
| `spec.appCheck.appAttest.tokenTtl` | `string` |  |  |  |
| `spec.appCheck.deviceCheck` | `GcpFirebaseAppleAppAppCheckDeviceCheck` |  |  |  |
| `spec.appCheck.deviceCheck.keyId` | `string` | yes |  |  |
| `spec.appCheck.deviceCheck.privateKey` | `string` (sensitive) | yes |  |  |
| `spec.appCheck.deviceCheck.tokenTtl` | `string` |  |  |  |
| `spec.appCheck.debugTokens` | `[]GcpFirebaseAppleAppAppCheckDebugToken` |  |  |  |
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

### spec.bundleId

`string` · required

The app's bundle identifier, as registered with Apple (e.g.
com.example.app). This is the app's IDENTITY in Firebase and in
GoogleService-Info.plist: it is immutable (a change replaces the
registration -- a new app id, a new plist), and a project accepts each
bundle id exactly once. Apple allows letters, digits, hyphens, and
periods; the reverse-DNS shape is Apple's convention, not a rule.

- rule: bundle_id must use only letters, digits, hyphens, and periods (e.g. com.example.app)
- rule: {"required":true}

### spec.appStoreId

`string`

The app's Apple ID in App Store Connect -- the numeric identifier Apple
assigns when the app record is created (the number in the App Store
URL). Used by Dynamic Links and App Store redirects; not needed for
push. Updatable; leave empty until the app has an App Store record.

- rule: app_store_id must be the numeric Apple ID from App Store Connect (digits only)

### spec.teamId

`string`

The Apple Developer Team ID that signs the app: ten uppercase letters
and digits, from the Membership page of the Apple Developer account.
REQUIRED when app_check configures App Attest or DeviceCheck (Apple's
attestations validate against the signing team); otherwise optional.
Updatable.

- rule: team_id must be the 10-character Apple Developer Team ID (uppercase letters and digits)

### spec.apiKeyId

`string | valueFrom`

The API key the app presents to Firebase, by the key's Google-assigned
UID: a literal, or a reference to a GcpApiKey resource (its uid
output). The key is a CLIENT IDENTIFIER that ships inside the app
bundle in GoogleService-Info.plist -- not a secret -- and it must be
valid for this app: unrestricted, or restricted to this bundle id with
API restrictions that include the Firebase APIs the app uses (Firebase
Installations and FCM Registration for push). Google's recommended
practice is one restricted key per app. When omitted, Firebase
associates an existing valid key or provisions a new, unrestricted
one; the api_key_id output reports which. Updatable.

- references: GcpApiKey (`status.outputs.uid`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpApiKey, name: <that resource's name>, fieldPath: status.outputs.uid}} -- a bare string does not parse

### spec.appCheck

`GcpFirebaseAppleAppAppCheck`

Firebase App Check for this app: the attestation providers that prove
requests come from the genuine app on a genuine device (App Attest,
with DeviceCheck as the fallback for devices that cannot use it), and
debug tokens that let development builds and CI pass App Check without
a real device. Whether the project's backends ENFORCE App Check is
configured on the GcpFirebaseProject.

- rule: app_check.debug_tokens must use each display_name at most once

### spec.appCheck.appAttest

`GcpFirebaseAppleAppAppCheckAppAttest`

Attest with App Attest -- Apple's attestation for iOS 14+ that proves
the app is genuine and unmodified. Google's recommended primary
provider for Apple apps; configure DeviceCheck beside it as the
fallback for devices that cannot use App Attest. Omit the block to
leave the provider unconfigured.

### spec.appCheck.appAttest.enabled

`bool` · optional (explicit presence)

Whether App Attest attestation is configured for the app. Defaults to
true when the block is present, so `appAttest: {}` means "configured,
with Google's default token lifetime"; set false to declare the block
and leave the provider unconfigured. The wire has no switch of its own
-- the configuration's existence is the switch -- so this field is how
the manifest states the intent explicitly.

- default: `true`

### spec.appCheck.appAttest.tokenTtl

`string`

How long an App Check token exchanged from an App Attest artifact
stays valid, as a duration in seconds with an "s" suffix (e.g. "3600s",
"1800s", "604800s"). Google accepts 30 minutes to 7 days inclusive and
assumes 1 hour when unset.

- rule: token_ttl must be a duration in seconds ending in 's' (e.g. 3600s), between 1800s and 604800s

### spec.appCheck.deviceCheck

`GcpFirebaseAppleAppAppCheckDeviceCheck`

Attest with DeviceCheck -- Apple's attestation available on every iOS
11+ device, verified server-side with a DeviceCheck private key you
generate in the Apple Developer account. Omit the block to leave the
provider unconfigured.

### spec.appCheck.deviceCheck.keyId

`string` · required

The Key ID of the DeviceCheck private key, as shown in the Apple
Developer account (Certificates, Identifiers & Profiles > Keys): ten
uppercase letters and digits. This identifies the key; it is not
secret. Updatable (rotate by generating a new key and setting both
fields).

- rule: device_check.key_id must be the 10-character Key ID of the DeviceCheck key (uppercase letters and digits)
- rule: {"required":true}

### spec.appCheck.deviceCheck.privateKey

`string` · required · sensitive

The contents of the DeviceCheck private key file (.p8) Apple issued for
key_id -- the PEM text, including its BEGIN and END lines. This is a
SECRET Google never returns after it is set (the module reports only
whether it is set), so the value must be a managed secret reference on
the platform. NOTE: this is the DeviceCheck key, NOT the APNs
authentication key -- Apple issues them separately, and the APNs key is
uploaded in the Firebase console, not declared here.

- rule: {"required":true}

### spec.appCheck.deviceCheck.tokenTtl

`string`

How long an App Check token exchanged from a DeviceCheck token stays
valid, as a duration in seconds with an "s" suffix (e.g. "3600s").
Google accepts 30 minutes to 7 days inclusive and assumes 1 hour when
unset.

- rule: token_ttl must be a duration in seconds ending in 's' (e.g. 3600s), between 1800s and 604800s

### spec.appCheck.debugTokens

`[]GcpFirebaseAppleAppAppCheckDebugToken`

Debug tokens: pre-shared secrets a development build, a simulator, or a
CI lane presents in place of a device attestation. Each token is a
SECRET -- anyone holding it passes App Check as this app -- so it is a
managed secret reference on the platform, never plaintext. Revoke by
removing the entry.

### spec.appCheck.debugTokens[].displayName

`string` · required

A name for the token as the Firebase console shows it -- the
developer's machine, the simulator, or the CI lane it is for. Unique
within the app; updatable.

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
               this path skips that window. The bundle id becomes
               registrable again at once.
  "PREVENT" -- destroy FAILS; use for a shipped app whose users hold
               its GoogleService-Info.plist
  "ABANDON" -- the app is left registered and removed from management
The App Attest and DeviceCheck configurations are per-app singletons
Google never deletes: they are removed from management on destroy and
disappear only with the app itself.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `team_id_required_for_apple_attestation`: team_id is required when app_check.app_attest or app_check.device_check is configured -- Apple attestation validates against the Developer Team that signed the app

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpFirebaseAppleApp, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.app_id` | `string` | The Firebase-assigned app id (GOOGLE_APP_ID in GoogleService-Info.plist), e.g. 1:123456789012:ios:0123456789abcdef. Globally unique and immutable. |
| `status.outputs.name` | `string` | The app's full resource name, projects/{project}/iosApps/{app_id} -- the Firebase Management API's path for Apple apps keeps the historical iosApps segment. The handle the API addresses the app by. |
| `status.outputs.api_key_id` | `string` | The Google-assigned UID of the API key associated with the app -- the one from the spec, or the key Firebase associated or provisioned when none was given. |
| `status.outputs.config_filename` | `string` | The configuration file's name: GoogleService-Info.plist. Add it to the Xcode project's main target so it ships in the app bundle. |
| `status.outputs.config_file_contents` | `string` | The configuration file's contents, base64-encoded. Decode and write to config_filename in the Apple build. A build input that ships in the bundle, not a secret. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpFirebaseProject | `status.outputs.project_id` |
| `spec.apiKeyId` | GcpApiKey | `status.outputs.uid` |

## See Also

- [Overview](../README.md)
