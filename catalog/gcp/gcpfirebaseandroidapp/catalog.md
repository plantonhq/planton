# GCP Firebase Android App

Registers an Android app in a Firebase-enabled Google Cloud project and composes the app's anti-abuse attestation: App Check with Play Integrity, plus debug tokens for development builds and CI. The registration is what lets the Android build receive Firebase Cloud Messaging and use every other Firebase product -- its `google-services.json` comes out as an output for the build to consume. The app lives inside a GcpFirebaseProject; declare that first and reference it from `projectId`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **The app registration** -- one `firebase_android_app` identified by its immutable package name
- **Play Integrity attestation** -- the app's `firebase_app_check_play_integrity_config` when `appCheck.playIntegrity` is configured
- **Debug tokens** -- one `firebase_app_check_debug_token` per `appCheck.debugTokens` entry
- **The configuration file** -- `google-services.json`, read after registration and exported base64-encoded

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Project

- **Firebase must be enabled on the project** -- deploy a GcpFirebaseProject first; the registration exists only inside it.
- **The package name is permanent.** It is the app's identity in Firebase and in `google-services.json`; a project accepts each package name once.
- **`deletionPolicy: DELETE` is immediate and permanent** -- Firebase's 30-day recoverable window is skipped. A shipped app should carry `PREVENT`.
- **IAM**: the deploying identity needs `roles/firebase.admin` or broader (plus `roles/firebaseappcheck.admin` when App Check is configured).

## Deploy

### Console

Open the deployment store, find **GCP Firebase Android App**, and click **Deploy**. The creation wizard walks you through the Firebase project, the app's identity (display name and package name), its certificate fingerprints and API key, and App Check. Start from the **Push Only** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirebaseAndroidApp
metadata:
  name: mobile-android
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-app-prod
  displayName: Acme Mobile (Android)
  packageName: com.acme.mobile
  deletionPolicy: PREVENT
```

```shell
planton apply -f android-app.yaml
```

This registers the app and produces its `google-services.json`; the `config_file_contents` output is what the Android build decodes into `app/google-services.json`. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the app references its Firebase project and its API key via ValueFromRef:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpFirebaseProject
      name: app-firebase
      fieldPath: status.outputs.project_id
  displayName: Acme Mobile (Android)
  packageName: com.acme.mobile
  sha256Hashes:
    - E3:B0:C4:42:98:FC:1C:14:9A:FB:F4:C8:99:6F:B9:24:27:AE:41:E4:64:9B:93:4C:A4:95:99:1B:78:52:B8:55
  apiKeyId:
    valueFrom:
      kind: GcpApiKey
      name: firebase-android-key
      fieldPath: status.outputs.uid
  appCheck:
    playIntegrity: {}
```

The InfraPipeline deploys the enablement and the key first, then the registration.

## Key Configuration

These are the most important decisions when configuring an Android app registration. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The package name** -- the app's identity, permanent and unique per project. Register the package name the shipped APK actually carries.

**Certificate fingerprints** -- push needs none. Add the SHA-256 of the signing certificate the shipped APK carries (Play App Signing's for a Play release) when you configure Play Integrity; add SHA-1s for Google Sign-In, Dynamic Links, and Phone Auth.

**The API key** -- reference a GcpApiKey restricted to this package name and certificate so the key that ships in the APK cannot be lifted into another app. Omit it and Firebase provisions an unrestricted one.

**App Check** -- `playIntegrity` makes the app prove it is genuine and unmodified on a Google Play device; `debugTokens` let development builds and CI pass without one. Whether backends ENFORCE App Check is configured on the GcpFirebaseProject.

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
| `app_id` | The Firebase-assigned app id | `mobilesdk_app_id`; App Check and Analytics tooling |
| `name` | The app's full resource name | The Firebase Management API handle |
| `api_key_id` | The UID of the associated API key | Auditing which key the app presents |
| `config_filename` | `google-services.json` | The file the build writes |
| `config_file_contents` | The file, base64-encoded | Decoded into `app/google-services.json` by the Android build |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Push only** -- the registration alone; the shape a notifications-only app needs. Start from the **Push Only** preset.

**Play Integrity attested** -- the release certificate's SHA-256 plus Play Integrity and a CI debug token. Start from the **Play Integrity Attested** preset.

## Works With

- [**GCP Firebase Project**](/cloud-catalog/gcp-firebase-project) -- the Firebase enablement this app is registered in
- [**GCP API Key**](/cloud-catalog/gcp-api-key) -- the restricted key the app references
- [**GCP Firebase Apple App**](/cloud-catalog/gcp-firebase-apple-app) -- the same product's iOS registration
- [**GCP Firebase Web App**](/cloud-catalog/gcp-firebase-web-app) -- the same product's web registration
