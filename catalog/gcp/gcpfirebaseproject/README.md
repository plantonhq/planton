# GCP Firebase Project

Enables Firebase on an existing Google Cloud project and configures the project-level Firebase surface every app on it shares: the default Cloud Storage for Firebase bucket and App Check enforcement. Firebase Cloud Messaging is enabled with it, so a control plane bound to the messaging role can send push the moment this resource is live. This is the room the Firebase app registrations live in — declare one per GCP project, then the Android, Apple, and Web apps inside it.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Firebase enablement** -- the `firebase_project` PROJECT SINGLETON (`projects.addFirebase`); an already-enabled project is adopted, not re-enabled
- **API enablement** -- `firebase.googleapis.com` and `fcm.googleapis.com` in the target project (never disabled on destroy), plus `firebasestorage.googleapis.com` / `firebaseappcheck.googleapis.com` exactly when the spec composes their resources
- **Default storage bucket** -- the project's one `firebase_storage_default_bucket` when `defaultStorageLocation` is set
- **App Check service configs** -- one `firebase_app_check_service_config` per `appCheck.serviceConfigs` entry
- **App Check resource policies** -- one `firebase_app_check_resource_policy` per `appCheck.resourcePolicies` entry

## Before You Deploy

### One-Way Enablement — Read This First

- **Enabling Firebase on a project cannot be undone.** Google offers no way to remove Firebase from a project. Destroying this resource DETACHES it from management and leaves the project Firebase-enabled; the composed bucket and App Check settings follow `deletionPolicy`. Choose the project deliberately.
- **Already-enabled projects are adopted.** The provider reads the project first; a project where Firebase was enabled from the console (or by a previous deploy) is brought under management with no error and no second enablement.
- **The enablement carries no deletion policy** — it is undeletable by construction. `spec.deletionPolicy` governs ONLY the composed resources (the default bucket and the App Check configurations).

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **The default bucket needs the pay-as-you-go (Blaze) plan** — a Cloud Billing account linked to the project. On a project without billing, leave `defaultStorageLocation` empty; the enablement and App Check need no billing.
- **IAM**: the deploying identity needs `roles/firebase.admin` or broader; `roles/firebaseappcheck.admin` when `appCheck` is configured.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirebaseProject
metadata:
  name: app-firebase
spec:
  projectId:
    value: acme-app-prod
```

```shell
planton apply -f firebase.yaml
```

## Configuration Reference

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project to enable Firebase on. Can reference a GcpProject resource. Immutable. |
| `defaultStorageLocation` | `string` | — | Create the default Cloud Storage for Firebase bucket in this Cloud Storage location (`US`, `EU`, `NAM4`, `us-central1`, ...). Once per project; immutable; needs the pay-as-you-go plan. |
| `appCheck.serviceConfigs` | `list` | `[]` | Per-service enforcement: `serviceId` (one of `firestore`, `firebasestorage`, `firebasedatabase`, `identitytoolkit` `.googleapis.com`) and `enforcementMode` (`""` OFF, `UNENFORCED`, `ENFORCED`). Each service at most once. |
| `appCheck.resourcePolicies` | `list` | `[]` | Per-resource overrides: `serviceId` (`oauth2.googleapis.com`), `targetResource` (`//oauth2.googleapis.com/projects/{number}/oauthClients/{id}`), `enforcementMode`. |
| `deletionPolicy` | `string` | `DELETE` | Governs ONLY the composed resources: `DELETE`, `PREVENT` (refuse), or `ABANDON`. The enablement itself is always detached. |

### Validation Rules

- **`defaultStorageLocation`**: a Cloud Storage location id (letters, digits, hyphens).
- **App Check services**: `serviceId` in the four supported services; each listed at most once; `enforcementMode` empty, `UNENFORCED`, or `ENFORCED`.
- **Resource policies**: `serviceId` must be `oauth2.googleapis.com`; `targetResource` starts with `//oauth2.googleapis.com/projects/`.
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `project_id` | `string` | The GCP project Firebase is enabled on — what the app registration kinds reference |
| `project_number` | `string` | The project number — the Firebase Cloud Messaging sender id (`messagingSenderId`) |
| `display_name` | `string` | The project's display name as Firebase shows it |
| `database_url` | `string` | The default Realtime Database URL, when the project has one (empty otherwise) |
| `storage_bucket` | `string` | The default Cloud Storage for Firebase bucket name, when the project has one (empty otherwise) |
| `location_id` | `string` | The project's default GCP resource location, once finalized (empty until then) |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **One Firebase project per GCP project, permanently.** Environments that share a GCP project share the Firebase project; they separate by identity (which control plane holds `roles/firebasemessaging.admin`), not by Firebase project. Firebase rejects a second app registration with the same package name or bundle id in one project, so declare each app exactly once.
- **Cloud Messaging is enabled explicitly.** `fcm.googleapis.com` is declared, never assumed from the enablement's side effects; sending push additionally needs `roles/firebasemessaging.admin` on the sender's identity (a GcpProjectIamMember) — no key file.
- **The Admin SDK config outputs are read after enablement.** `database_url`, `storage_bucket`, and `location_id` are each present only when the project has the corresponding resource, and degrade to empty otherwise; offline plans stay credential-free because the read is deferred to apply.
- **Both engines set `user_project_override`** — the Firebase Management API attributes quota to the caller's project on user-credential calls; without the override a deploy under plain ADC fails with "requires a quota project".
- **The Terraform module uses the `google-beta` provider for two resources** — Google publishes the enablement and the default bucket only there. Both are recorded in the catalog's beta admission list; everything else stays on the GA provider.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpApiKey](/docs/catalog/gcp/gcpapikey) — the restricted key each app registration references
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project Firebase is enabled on
- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — grants `roles/firebasemessaging.admin` to the identity that sends push
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the keyless sender identity a control plane runs as

## Additional Resources

- [Firebase Management API](https://firebase.google.com/docs/projects/api/reference/rest)
- [Add Firebase to an existing Google Cloud project](https://firebase.google.com/docs/projects/learn-more#firebase-cloud-relationship)
- [Firebase App Check](https://firebase.google.com/docs/app-check)
- [Cloud Storage for Firebase default bucket](https://firebase.google.com/docs/storage)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
