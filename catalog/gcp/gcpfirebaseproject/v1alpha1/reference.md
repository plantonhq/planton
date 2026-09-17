# GcpFirebaseProject

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpFirebaseProjectSpec enables Firebase on an existing Google Cloud
project and configures the project-level Firebase surface that every app
registered on it shares: the default Cloud Storage for Firebase bucket and
App Check enforcement. The project's Firebase Cloud Messaging API is
enabled with it, so a control plane bound to the right role can send push
notifications the moment this resource is live.

This is a PROJECT SINGLETON with ONE-WAY creation. Enabling Firebase on a
project cannot be undone: Google offers no way to remove Firebase from a
project, so destroying this resource DETACHES it from management and
leaves the project Firebase-enabled (the composed default bucket and App
Check settings follow the deletion_policy). Applying it against a project
that already has Firebase enabled ADOPTS the existing enablement -- the
operation is idempotent, so a project enabled from the console can be
brought under management by declaring this resource.

This kind is the ROOM the Firebase app registrations live in:
GcpFirebaseAndroidApp, GcpFirebaseAppleApp, and GcpFirebaseWebApp each
reference a GcpFirebaseProject and are placed inside it. Declare one
GcpFirebaseProject per GCP project, then as many app registrations as the
product has client platforms.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirebaseProject
metadata:
  name: my-sample-firebase
spec:
  # GCP project to enable Firebase on. Omit to use the provider's default
  # project. Immutable. NOTE: enabling Firebase is ONE-WAY -- Google offers
  # no way to remove it; destroy detaches this resource and leaves the
  # project enabled. Applying against an already-enabled project ADOPTS it.
  projectId:
    value: my-gcp-project-123

  # Create the project's default Cloud Storage for Firebase bucket in this
  # location (US, EU, NAM4, us-central1, ...). Once per project; immutable;
  # needs the pay-as-you-go (Blaze) plan. Omit to skip.
  defaultStorageLocation: US

  # Project-level App Check: which Firebase backends enforce genuine-app
  # attestation. Per-app attestation providers are configured on the app
  # registrations (GcpFirebaseAndroidApp / AppleApp / WebApp).
  appCheck:
    # One entry per backend service. UNENFORCED collects metrics without
    # rejecting -- the safe first step; ENFORCED rejects requests without a
    # valid App Check token. Supported services: firestore, firebasestorage,
    # firebasedatabase, identitytoolkit (.googleapis.com).
    serviceConfigs:
      - serviceId: firestore.googleapis.com
        enforcementMode: UNENFORCED
      - serviceId: firebasestorage.googleapis.com
        enforcementMode: ENFORCED
    # Per-resource overrides, for the services that support them (today:
    # OAuth clients under Google Identity for iOS).
    resourcePolicies:
      - serviceId: oauth2.googleapis.com
        targetResource: //oauth2.googleapis.com/projects/123456789012/oauthClients/123456789012-abc.apps.googleusercontent.com
        enforcementMode: ENFORCED

  # What destroy does to the COMPOSED settings (default bucket, App Check
  # configs) -- the enablement itself is always detached. DELETE (default),
  # PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.defaultStorageLocation` | `string` |  |  |  |
| `spec.appCheck` | `GcpFirebaseProjectAppCheck` |  |  |  |
| `spec.appCheck.serviceConfigs` | `[]GcpFirebaseProjectAppCheckServiceConfig` |  |  |  |
| `spec.appCheck.serviceConfigs[].serviceId` | `string` | yes |  |  |
| `spec.appCheck.serviceConfigs[].enforcementMode` | `string` |  |  |  |
| `spec.appCheck.resourcePolicies` | `[]GcpFirebaseProjectAppCheckResourcePolicy` |  |  |  |
| `spec.appCheck.resourcePolicies[].serviceId` | `string` | yes |  |  |
| `spec.appCheck.resourcePolicies[].targetResource` | `string` | yes |  |  |
| `spec.appCheck.resourcePolicies[].enforcementMode` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project to enable Firebase on. Can be a literal project ID or a
reference to a GcpProject resource. If omitted, the provider's default
project is used. Immutable: Firebase enablement is a property of the
project itself.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.defaultStorageLocation

`string`

Create the project's DEFAULT Cloud Storage for Firebase bucket in this
location (a Cloud Storage location: a multi-region such as US or EU, a
dual-region such as NAM4, or a region such as us-central1). The default
bucket is the one the Firebase client SDKs use when no bucket is named
and is created at most once per project; leave empty to skip it.
Immutable: the bucket's location cannot change after creation.

Requires the project to be on Firebase's pay-as-you-go (Blaze) plan --
that is, linked to a Cloud Billing account. On a project without
billing the create fails.

- rule: default_storage_location must be a Cloud Storage location id (letters, digits, and hyphens) -- e.g. US, EU, NAM4, us-central1

### spec.appCheck

`GcpFirebaseProjectAppCheck`

Firebase App Check: verify that requests to your Firebase backends come
from your genuine apps. Per-app attestation providers (Play Integrity,
App Attest, DeviceCheck, reCAPTCHA) are configured on each app
registration; the PROJECT-level part -- which backend services enforce
App Check, and per-resource overrides -- lives here.

- rule: app_check.service_configs must name each service at most once

### spec.appCheck.serviceConfigs

`[]GcpFirebaseProjectAppCheckServiceConfig`

Enforcement per Firebase backend service. A service not listed here is
in Google's default OFF state (no enforcement, no metrics). Listing a
service with enforcement_mode UNENFORCED starts collecting metrics
without rejecting anything -- the recommended first step before
switching to ENFORCED.

### spec.appCheck.serviceConfigs[].serviceId

`string` · required

The service to configure. Google supports exactly these:
  firebasestorage.googleapis.com  -- Cloud Storage for Firebase
  firebasedatabase.googleapis.com -- Firebase Realtime Database
  firestore.googleapis.com        -- Cloud Firestore
  identitytoolkit.googleapis.com  -- Firebase Authentication
Immutable: to move enforcement to another service, add a new entry.

- rule: service_id must be one of: firebasestorage.googleapis.com, firebasedatabase.googleapis.com, firestore.googleapis.com, identitytoolkit.googleapis.com
- rule: {"required":true}

### spec.appCheck.serviceConfigs[].enforcementMode

`string`

How the service treats requests without a valid App Check token:
  ""           -- OFF: not enforced, no metrics (Google's default for
                  an unlisted service; listing a service with this
                  value records the intent explicitly)
  "UNENFORCED" -- not enforced, but metrics show what WOULD be
                  rejected -- the safe first step
  "ENFORCED"   -- requests without a valid token are rejected

- rule: enforcement_mode must be one of: UNENFORCED, ENFORCED (or empty for OFF)

### spec.appCheck.resourcePolicies

`[]GcpFirebaseProjectAppCheckResourcePolicy`

Per-resource overrides of a service's enforcement, for the services
that support them (today: individual OAuth clients under Google
Identity for iOS).

### spec.appCheck.resourcePolicies[].serviceId

`string` · required

The service the resource belongs to. Google supports exactly one today:
oauth2.googleapis.com (Google Identity for iOS). Immutable.

- rule: resource_policies.service_id must be oauth2.googleapis.com (the only service with per-resource App Check policies)
- rule: {"required":true}

### spec.appCheck.resourcePolicies[].targetResource

`string` · required

The resource the override applies to, in the service's own format --
for an iOS OAuth client:
//oauth2.googleapis.com/projects/{project_number}/oauthClients/{client_id}

- rule: target_resource must be a full resource name starting with //oauth2.googleapis.com/projects/ -- e.g. //oauth2.googleapis.com/projects/123456789/oauthClients/123456789-abc.apps.googleusercontent.com
- rule: {"required":true}

### spec.appCheck.resourcePolicies[].enforcementMode

`string`

The enforcement for this resource, overriding the service's setting:
"" (OFF), "UNENFORCED", or "ENFORCED" -- same semantics as the
service-level enforcement_mode.

- rule: resource_policies.enforcement_mode must be one of: UNENFORCED, ENFORCED (or empty for OFF)

### spec.deletionPolicy

`string`

What destroying this resource does to the COMPOSED settings (the
default storage bucket and the App Check configurations). Firebase
enablement itself is not deletable and is always detached:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the default bucket is deleted (and its objects with it)
               and App Check enforcement is switched OFF for every
               configured service
  "PREVENT" -- destroy FAILS; use for a production project whose
               default bucket holds user data
  "ABANDON" -- the bucket and the App Check settings are left in place
               and removed from management

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpFirebaseProject, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.project_id` | `string` | The GCP project the enablement lives on (projects/{project}). |
| `status.outputs.project_number` | `string` | The project's number. This is the Firebase Cloud Messaging SENDER ID (messagingSenderId in a web firebaseConfig; project_number in google-services.json) a client registers with to receive push. |
| `status.outputs.display_name` | `string` | The project's display name as Firebase shows it. |
| `status.outputs.database_url` | `string` | The default Firebase Realtime Database URL, when the project has one (empty otherwise). From the Admin SDK configuration. |
| `status.outputs.storage_bucket` | `string` | The default Cloud Storage for Firebase bucket name (e.g. my-project.firebasestorage.app), when the project has one -- created by default_storage_location or in the console. Empty otherwise. |
| `status.outputs.location_id` | `string` | The project's default GCP resource location id (e.g. us-central), once finalized in the Firebase console or by a Firestore/RTDB/default-bucket create that pins it. Empty until then. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpFirebaseAndroidApp | `spec.projectId` | `status.outputs.project_id` |
| GcpFirebaseAppleApp | `spec.projectId` | `status.outputs.project_id` |
| GcpFirebaseWebApp | `spec.projectId` | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
