# GCP Firebase Apple App

Registers an iOS / macOS app in a Firebase-enabled Google Cloud project and composes the app's anti-abuse attestation: App Check with App Attest and DeviceCheck, plus debug tokens for simulators and CI. The registration is what lets the Apple build receive Firebase Cloud Messaging and use every other Firebase product -- its `GoogleService-Info.plist` comes out as an output for the build to consume. The app lives inside a GcpFirebaseProject; declare that first and reference it from `projectId`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **The app registration** -- one `firebase_apple_app` identified by its immutable bundle id
- **App Attest and DeviceCheck attestation** -- the app's `firebase_app_check_app_attest_config` and `firebase_app_check_device_check_config` when configured
- **Debug tokens** -- one `firebase_app_check_debug_token` per `appCheck.debugTokens` entry
- **The configuration file** -- `GoogleService-Info.plist`, read after registration and exported base64-encoded

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Project

- **Firebase must be enabled on the project** -- deploy a GcpFirebaseProject first; the registration exists only inside it.
- **The bundle id is permanent.** It is the app's identity in Firebase and in the plist; a project accepts each bundle id once.
- **`deletionPolicy: DELETE` is immediate and permanent** -- Firebase's 30-day recoverable window is skipped. A shipped app should carry `PREVENT`.
- **Push needs the APNs key uploaded in the Firebase console** -- Apple mints it, Firebase has no API for it; this is the one manual step.
- **IAM**: the deploying identity needs `roles/firebase.admin` or broader (plus `roles/firebaseappcheck.admin` when App Check is configured).

## Deploy

### Console

Open the deployment store, find **GCP Firebase Apple App**, and click **Deploy**. The creation wizard walks you through the Firebase project, the app's identity (display name, bundle id, team id, App Store id), its API key, and App Check. Start from the **Push Only** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirebaseAppleApp
metadata:
  name: mobile-ios
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-app-prod
  displayName: Acme Mobile (iOS)
  bundleId: com.acme.mobile
  teamId: ABCDE12345
  deletionPolicy: PREVENT
```

```shell
planton apply -f ios-app.yaml
```

This registers the app and produces its `GoogleService-Info.plist`; the `config_file_contents` output is what the Xcode build decodes into the target. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the app references its Firebase project and its API key via ValueFromRef:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpFirebaseProject
      name: app-firebase
      fieldPath: status.outputs.project_id
  displayName: Acme Mobile (iOS)
  bundleId: com.acme.mobile
  teamId: ABCDE12345
  apiKeyId:
    valueFrom:
      kind: GcpApiKey
      name: firebase-ios-key
      fieldPath: status.outputs.uid
  appCheck:
    appAttest: {}
```

The InfraPipeline deploys the enablement and the key first, then the registration.

## Key Configuration

These are the most important decisions when configuring an Apple app registration. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The bundle id** -- the app's identity, permanent and unique per project. Register the bundle id the shipped app actually carries.

**The team id** -- the Apple Developer Team that signs the app. Required the moment App Attest or DeviceCheck is configured, because Apple's attestations validate against it.

**The API key** -- reference a GcpApiKey restricted to this bundle id so the key that ships in the bundle cannot be lifted into another app. Omit it and Firebase provisions an unrestricted one.

**App Check** -- `appAttest` is Google's recommended primary provider (iOS 14+); `deviceCheck` is the fallback for devices that cannot use it and needs a DeviceCheck private key from the Apple Developer account; `debugTokens` let simulators and CI pass without either. Whether backends ENFORCE App Check is configured on the GcpFirebaseProject.

**Deletion policy** -- `DELETE` removes the app immediately and permanently; `PREVENT` fails the destroy; `ABANDON` leaves it registered, unmanaged. A shipped app carries `PREVENT`.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpFirebaseProject** | `projectId` | `status.outputs.project_id` |
| **GcpApiKey** (optional) | `apiKeyId` | `status.outputs.uid` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources and the mobile build can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `app_id` | The Firebase-assigned app id | `GOOGLE_APP_ID`; App Check and Analytics tooling |
| `name` | The app's full resource name | The Firebase Management API handle |
| `api_key_id` | The UID of the associated API key | Auditing which key the app presents |
| `config_filename` | `GoogleService-Info.plist` | The file the build adds to the target |
| `config_file_contents` | The file, base64-encoded | Decoded into the Xcode target by the build |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Push only** -- the registration with its team id; the shape a notifications-only app needs. Start from the **Push Only** preset.

**App Attest with DeviceCheck fallback** -- Google's recommended attestation pairing plus a simulator debug token. Start from the **App Attest with DeviceCheck** preset.

## Works With

- [**GCP Firebase Project**](/cloud-catalog/gcp-firebase-project) -- the Firebase enablement this app is registered in
- [**GCP API Key**](/cloud-catalog/gcp-api-key) -- the restricted key the app references
- [**GCP Firebase Android App**](/cloud-catalog/gcp-firebase-android-app) -- the same product's Android registration
- [**GCP Firebase Web App**](/cloud-catalog/gcp-firebase-web-app) -- the same product's web registration
