# GCP Firebase Web App

Registers a web app in a Firebase-enabled Google Cloud project and composes the app's anti-abuse attestation: App Check with reCAPTCHA v3 or reCAPTCHA Enterprise, plus debug tokens for local development. The registration is what lets a browser client receive Firebase Cloud Messaging and use every other Firebase product -- the `firebaseConfig` object the web SDK is initialised with comes out as outputs for the front-end build to consume. The app lives inside a GcpFirebaseProject; declare that first and reference it from `projectId`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **The app registration** -- one `firebase_web_app`, identified only by its display name
- **reCAPTCHA attestation** -- the app's `firebase_app_check_recaptcha_v3_config` and `firebase_app_check_recaptcha_enterprise_config` when configured
- **Debug tokens** -- one `firebase_app_check_debug_token` per `appCheck.debugTokens` entry
- **firebaseConfig** -- the seven client initialisation values, read after registration

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Project

- **Firebase must be enabled on the project** -- deploy a GcpFirebaseProject first; the registration exists only inside it.
- **`deletionPolicy: DELETE` is immediate and permanent** -- Firebase's 30-day recoverable window is skipped. A shipped site should carry `PREVENT`.
- **IAM**: the deploying identity needs `roles/firebase.admin` or broader (plus `roles/firebaseappcheck.admin` when App Check is configured).

## Deploy

### Console

Open the deployment store, find **GCP Firebase Web App**, and click **Deploy**. The creation wizard walks you through the Firebase project, the app's display name and API key, and App Check. Start from the **Push Web Client** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirebaseWebApp
metadata:
  name: console-web
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-app-prod
  displayName: Acme Console
  deletionPolicy: PREVENT
```

```shell
planton apply -f web-app.yaml
```

This registers the app and produces its `firebaseConfig`; the seven outputs are what the front-end build writes into `initializeApp({...})`. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the app references its Firebase project and its API key via ValueFromRef:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpFirebaseProject
      name: app-firebase
      fieldPath: status.outputs.project_id
  displayName: Acme Console
  apiKeyId:
    valueFrom:
      kind: GcpApiKey
      name: browser-key
      fieldPath: status.outputs.uid
  appCheck:
    recaptchaEnterprise:
      siteKey: 6LcExampleSiteKey
```

The InfraPipeline deploys the enablement and the key first, then the registration.

## Key Configuration

These are the most important decisions when configuring a web app registration. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The display name** -- a web app's only identity. Several web apps may coexist in one project; name each for the client it serves.

**The API key** -- reference a GcpApiKey restricted to the site's HTTP referrers so the key that ships in the page cannot be used from another origin. Omit it and Firebase provisions an unrestricted one.

**App Check** -- `recaptchaV3` (the free, score-based reCAPTCHA, verified with a site SECRET) or `recaptchaEnterprise` (Google Cloud's reCAPTCHA, verified against the PUBLIC site key, billed per assessment beyond the free tier); a client initialises with one, both may be configured during a migration. `debugTokens` let local development pass. Whether backends ENFORCE App Check is configured on the GcpFirebaseProject.

**Deletion policy** -- `DELETE` removes the app immediately and permanently; `PREVENT` fails the destroy; `ABANDON` leaves it registered, unmanaged. A shipped site carries `PREVENT`.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpFirebaseProject** | `projectId` | `status.outputs.project_id` |
| **GcpApiKey** (optional) | `apiKeyId` | `status.outputs.uid` |

### What This Component Provides

After provisioning, `status.outputs` contains the `firebaseConfig` values the front-end build consumes:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `app_id` | The Firebase-assigned app id | `firebaseConfig.appId` |
| `name` | The app's full resource name | The Firebase Management API handle |
| `api_key_id` | The UID of the associated API key | Auditing which key the app presents |
| `app_urls` | Hosting URLs Firebase records | Firebase Hosting integration |
| `api_key` | `firebaseConfig.apiKey` | The key string the page presents |
| `auth_domain` | `firebaseConfig.authDomain` | Firebase Authentication redirects |
| `database_url` | `firebaseConfig.databaseURL` | Realtime Database client (when one exists) |
| `storage_bucket` | `firebaseConfig.storageBucket` | Cloud Storage for Firebase client (when a default bucket exists) |
| `location_id` | `firebaseConfig.locationId` | Placing Functions / Firestore alongside |
| `messaging_sender_id` | `firebaseConfig.messagingSenderId` | The project number a browser registers with for push |
| `measurement_id` | `firebaseConfig.measurementId` | Google Analytics (when linked) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Push web client** -- the registration with a referrer-restricted key; the shape a notifications-only site needs. Start from the **Push Web Client** preset.

**reCAPTCHA Enterprise attested** -- Google Cloud's reCAPTCHA with a localhost debug token. Start from the **reCAPTCHA Enterprise Attested** preset.

## Works With

- [**GCP Firebase Project**](/cloud-catalog/gcp-firebase-project) -- the Firebase enablement this app is registered in
- [**GCP API Key**](/cloud-catalog/gcp-api-key) -- the referrer-restricted key the app references
- [**GCP Firebase Android App**](/cloud-catalog/gcp-firebase-android-app) -- the same product's Android registration
- [**GCP Firebase Apple App**](/cloud-catalog/gcp-firebase-apple-app) -- the same product's iOS registration
