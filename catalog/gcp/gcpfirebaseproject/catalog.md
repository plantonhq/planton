# GCP Firebase Project

Enables Firebase on an existing Google Cloud project and configures the project-level Firebase surface every app on it shares: the default Cloud Storage for Firebase bucket and App Check enforcement. Firebase Cloud Messaging is enabled with it, so a control plane bound to the messaging role can send push the moment this resource is live. This is the room the Firebase app registrations live in -- declare one per GCP project, then the Android, Apple, and Web apps inside it.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Firebase enablement** -- the `firebase_project` PROJECT SINGLETON (`projects.addFirebase`); an already-enabled project is adopted, not re-enabled
- **API enablement** -- `firebase.googleapis.com` and `fcm.googleapis.com` in the target project (never disabled on destroy), plus the Storage and App Check APIs exactly when the spec composes their resources
- **Default storage bucket** -- the project's one `firebase_storage_default_bucket` when `defaultStorageLocation` is set
- **App Check enforcement** -- one `firebase_app_check_service_config` per `appCheck.serviceConfigs` entry and one `firebase_app_check_resource_policy` per `appCheck.resourcePolicies` entry

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Project

- **Enabling Firebase is ONE-WAY.** Google offers no way to remove Firebase from a project. Destroying this resource detaches it from management and leaves the project Firebase-enabled; the composed bucket and App Check settings follow `deletionPolicy`. Choose the project deliberately -- one Firebase project serves every environment that shares the GCP project.
- **Already-enabled projects are adopted.** Applying against a project where Firebase was enabled from the console brings it under management without error.
- **The default bucket needs the pay-as-you-go (Blaze) plan** -- a Cloud Billing account linked to the project. Leave `defaultStorageLocation` empty on a project without billing.
- **IAM**: the deploying identity needs `roles/firebase.admin` or broader (plus `roles/firebaseappcheck.admin` when App Check is configured).

## Deploy

### Console

Open the deployment store, find **GCP Firebase Project**, and click **Deploy**. The creation wizard walks you through the target project, the optional default bucket location, and App Check enforcement. Start from the **Enable Only** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirebaseProject
metadata:
  name: app-firebase
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-app-prod
```

```shell
planton apply -f firebase.yaml
```

This enables Firebase and Cloud Messaging on the project; the `project_number` output is the FCM sender id every client registers with. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the enablement references its project via ValueFromRef, and the app registrations reference the enablement:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: app-project
      fieldPath: status.outputs.project_id
  appCheck:
    serviceConfigs:
      - serviceId: firestore.googleapis.com
        enforcementMode: UNENFORCED
```

The InfraPipeline deploys the project first, then the enablement, then every Firebase app whose `projectId` references it.

## Key Configuration

These are the most important decisions when configuring a Firebase project. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The project** -- one Firebase project per GCP project, permanently. Environments that share a GCP project share the Firebase project and separate by identity (which control plane may send push), not by Firebase project.

**Default storage location** -- creates the default Cloud Storage for Firebase bucket the client SDKs use when no bucket is named. Created at most once per project, immutable, and it moves the project onto the pay-as-you-go plan. A multi-region (US, EU) is geo-redundant; a region is not.

**App Check enforcement** -- `serviceConfigs` decides which Firebase backends (Firestore, Storage, Realtime Database, Authentication) reject requests that do not prove they came from a genuine registered app. `UNENFORCED` first (metrics only), `ENFORCED` once every shipped client attests. Per-app attestation providers are configured on the app registration kinds.

**Deletion policy** -- governs ONLY the composed resources: `DELETE` removes the default bucket (with its objects) and switches App Check OFF; `PREVENT` fails the destroy; `ABANDON` leaves them unmanaged. The enablement itself is always detached.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `project_id` | The GCP project Firebase is enabled on | The `projectId` reference on GcpFirebaseAndroidApp / AppleApp / WebApp |
| `project_number` | The project number -- the FCM sender id | `messagingSenderId` in client configuration |
| `display_name` | The project's Firebase display name | Console and tooling |
| `database_url` | Default Realtime Database URL (when one exists) | Client and Admin SDK initialisation |
| `storage_bucket` | Default Cloud Storage for Firebase bucket (when one exists) | Client and Admin SDK initialisation |
| `location_id` | Default GCP resource location (once finalized) | Placing Firestore / Functions alongside |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Enable only** -- Firebase plus Cloud Messaging on a project, nothing else; the shape a push-only backend needs. Start from the **Enable Only** preset.

**Messaging with storage** -- enablement plus the default bucket for app content. Start from the **Messaging with Default Bucket** preset.

**App Check enforced** -- every Firebase backend rejects unattested clients. Start from the **App Check Enforced** preset.

## Works With

- [**GCP API Key**](/cloud-catalog/gcp-api-key) -- the restricted key each app registration references
- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the project Firebase is enabled on
- [**GCP Project IAM Member**](/cloud-catalog/gcp-project-iam-member) -- grants `roles/firebasemessaging.admin` to the identity that sends push
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- the keyless sender identity a control plane runs as
