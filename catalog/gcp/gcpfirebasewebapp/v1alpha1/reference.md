# GcpFirebaseWebApp

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpFirebaseWebAppSpec registers a web app in a Firebase-enabled Google
Cloud project and composes the app's anti-abuse attestation (App Check
with reCAPTCHA v3 or reCAPTCHA Enterprise, plus debug tokens for
development). The registration is what lets a browser client receive
Firebase Cloud Messaging and use every other Firebase product: the
firebaseConfig object the web SDK is initialised with is produced as
outputs for the front-end build to consume.

The app lives INSIDE a GcpFirebaseProject (the project's Firebase
enablement); declare that first and reference it from project_id.

Two facts to know before acting. A web app has no identity beyond its
display name -- several web apps may coexist in one project, so name
them for the client each serves. And destroying this resource with
deletion_policy DELETE removes the app IMMEDIATELY and PERMANENTLY --
Firebase's 30-day recoverable window is skipped -- so a shipped app's
registration should carry PREVENT.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirebaseWebApp
metadata:
  name: my-web-app
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

  # As the Firebase console shows the app -- the only identity a web app
  # has, so name it for the client it serves. Updatable.
  displayName: My Web App

  # The referrer-restricted API key the app presents, by the key's UID -- a
  # reference to a GcpApiKey (its uid output). A client identifier that
  # ships in the page as firebaseConfig.apiKey, not a secret. Omit to let
  # Firebase associate or provision an unrestricted key.
  apiKeyId:
    valueFrom:
      kind: GcpApiKey
      name: browser-key
      fieldPath: status.outputs.uid

  # App Check: prove requests come from your genuine site. Whether backends
  # ENFORCE it is configured on the GcpFirebaseProject. A client initialises
  # with ONE provider; both may be configured during a v3-to-Enterprise
  # migration.
  appCheck:
    # reCAPTCHA v3: Google verifies the client's token with the site SECRET
    # -- a managed secret reference on the platform.
    recaptchaV3:
      siteSecret: $secret/recaptcha-v3-site-secret
      tokenTtl: 3600s
    # reCAPTCHA Enterprise: Google verifies against the PUBLIC site key --
    # the same value the page embeds.
    recaptchaEnterprise:
      siteKey: 6LcExampleSiteKey_ReplaceWithYours
      tokenTtl: 3600s
    # Debug tokens let local development pass App Check. Each token is a
    # SECRET -- a managed secret reference on the platform.
    debugTokens:
      - displayName: localhost
        token: $secret/firebase-app-check-localhost-token

  # What destroy does to the app and its debug tokens. DELETE (default) is
  # IMMEDIATE and PERMANENT -- Firebase's 30-day recoverable window is
  # skipped. A shipped site should carry PREVENT.
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpFirebaseProject (`status.outputs.project_id`) |
| `spec.displayName` | `string` | yes |  |  |
| `spec.apiKeyId` | `string \| valueFrom` |  |  | GcpApiKey (`status.outputs.uid`) |
| `spec.appCheck` | `GcpFirebaseWebAppAppCheck` |  |  |  |
| `spec.appCheck.recaptchaV3` | `GcpFirebaseWebAppAppCheckRecaptchaV3` |  |  |  |
| `spec.appCheck.recaptchaV3.siteSecret` | `string` (sensitive) | yes |  |  |
| `spec.appCheck.recaptchaV3.tokenTtl` | `string` |  |  |  |
| `spec.appCheck.recaptchaEnterprise` | `GcpFirebaseWebAppAppCheckRecaptchaEnterprise` |  |  |  |
| `spec.appCheck.recaptchaEnterprise.siteKey` | `string` | yes |  |  |
| `spec.appCheck.recaptchaEnterprise.tokenTtl` | `string` |  |  |  |
| `spec.appCheck.debugTokens` | `[]GcpFirebaseWebAppAppCheckDebugToken` |  |  |  |
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
shows it. The only identity a web app has; updatable.

- rule: {"required":true}

### spec.apiKeyId

`string | valueFrom`

The API key the app presents to Firebase, by the key's Google-assigned
UID: a literal, or a reference to a GcpApiKey resource (its uid
output). The key is a CLIENT IDENTIFIER that ships in the page as
firebaseConfig.apiKey -- not a secret -- and it must be valid for this
app: unrestricted, or restricted to the site's HTTP referrers with API
restrictions that include the Firebase APIs the app uses (Firebase
Installations and FCM Registration for push). Google's recommended
practice is one referrer-restricted key per web app. When omitted,
Firebase associates an existing valid key or provisions a new,
unrestricted one; the api_key_id and api_key outputs report which.
Updatable.

- references: GcpApiKey (`status.outputs.uid`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpApiKey, name: <that resource's name>, fieldPath: status.outputs.uid}} -- a bare string does not parse

### spec.appCheck

`GcpFirebaseWebAppAppCheck`

Firebase App Check for this app: the attestation provider that proves
requests come from your genuine site (reCAPTCHA v3 or reCAPTCHA
Enterprise), and debug tokens that let local development pass App
Check. Whether the project's backends ENFORCE App Check is configured
on the GcpFirebaseProject.

- rule: app_check.debug_tokens must use each display_name at most once

### spec.appCheck.recaptchaV3

`GcpFirebaseWebAppAppCheckRecaptchaV3`

Attest with reCAPTCHA v3 -- the score-based, no-interaction reCAPTCHA
registered at google.com/recaptcha (site type "reCAPTCHA v3"). Google
verifies the client's reCAPTCHA token with the site SECRET configured
here. Omit the block to leave the provider unconfigured.

### spec.appCheck.recaptchaV3.siteSecret

`string` · required · sensitive

The reCAPTCHA v3 SITE SECRET from the reCAPTCHA admin console -- the
server-side half of the key pair (the site key goes in the page). A
SECRET Google never returns after it is set (the module reports only
whether it is set), so the value must be a managed secret reference on
the platform. Updatable (rotate by setting a new secret).

- rule: {"required":true}

### spec.appCheck.recaptchaV3.tokenTtl

`string`

How long an App Check token exchanged from a reCAPTCHA v3 token stays
valid, as a duration in seconds with an "s" suffix (e.g. "3600s",
"1800s", "604800s"). Google accepts 30 minutes to 7 days inclusive and
assumes 1 hour when unset.

- rule: token_ttl must be a duration in seconds ending in 's' (e.g. 3600s), between 1800s and 604800s

### spec.appCheck.recaptchaEnterprise

`GcpFirebaseWebAppAppCheckRecaptchaEnterprise`

Attest with reCAPTCHA Enterprise -- the Google Cloud reCAPTCHA product
with its own assessments, billing, and console. Google verifies the
client's token against the site KEY (a public key) configured here.
Omit the block to leave the provider unconfigured.

### spec.appCheck.recaptchaEnterprise.siteKey

`string` · required

The reCAPTCHA Enterprise SITE KEY, created in the Google Cloud console
for the site's domains (a score-based key). This is the PUBLIC half of
the key -- the same value the page embeds -- so it is an identifier,
not a secret; reCAPTCHA Enterprise has no site secret. Updatable.

- rule: {"required":true}

### spec.appCheck.recaptchaEnterprise.tokenTtl

`string`

How long an App Check token exchanged from a reCAPTCHA Enterprise
assessment stays valid, as a duration in seconds with an "s" suffix
(e.g. "3600s"). Google accepts 30 minutes to 7 days inclusive and
assumes 1 hour when unset.

- rule: token_ttl must be a duration in seconds ending in 's' (e.g. 3600s), between 1800s and 604800s

### spec.appCheck.debugTokens

`[]GcpFirebaseWebAppAppCheckDebugToken`

Debug tokens: pre-shared secrets a local development build presents in
place of a reCAPTCHA attestation. Each token is a SECRET -- anyone
holding it passes App Check as this app -- so it is a managed secret
reference on the platform, never plaintext. Revoke by removing the
entry.

### spec.appCheck.debugTokens[].displayName

`string` · required

A name for the token as the Firebase console shows it -- the
developer's machine or the CI lane it is for. Unique within the app;
updatable.

- rule: {"required":true}

### spec.appCheck.debugTokens[].token

`string` · required · sensitive

The token value itself: a UUID (version 4), case-insensitive, that the
development build registers with the App Check debug provider
(self.FIREBASE_APPCHECK_DEBUG_TOKEN). Google never returns it after
creation, and it cannot be changed -- a new value is a new token. It
is a SECRET: the value must be a managed secret reference on the
platform.

- rule: {"required":true}

### spec.deletionPolicy

`string`

What destroying this resource does to the app registration (and to
the composed debug tokens):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the app is removed IMMEDIATELY and PERMANENTLY: Firebase
               normally keeps a removed app recoverable for 30 days;
               this path skips that window.
  "PREVENT" -- destroy FAILS; use for a shipped site whose pages carry
               this app's firebaseConfig
  "ABANDON" -- the app is left registered and removed from management
The reCAPTCHA configurations are per-app singletons Google never
deletes: they are removed from management on destroy and disappear
only with the app itself.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpFirebaseWebApp, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.app_id` | `string` | The Firebase-assigned app id (firebaseConfig.appId), e.g. 1:123456789012:web:0123456789abcdef. Globally unique and immutable. |
| `status.outputs.name` | `string` | The app's full resource name, projects/{project}/webApps/{app_id}. The handle the Firebase Management API addresses the app by. |
| `status.outputs.api_key_id` | `string` | The Google-assigned UID of the API key associated with the app -- the one from the spec, or the key Firebase associated or provisioned when none was given. |
| `status.outputs.app_urls` | `[]string` | The URLs where the web app is hosted, as Firebase records them (set by Firebase Hosting or in the console). Empty when none are recorded. |
| `status.outputs.api_key` | `string` | firebaseConfig.apiKey -- the API key STRING the page presents. A client identifier that ships in the page, restricted (when the key is a GcpApiKey) by HTTP referrer and API targets. |
| `status.outputs.auth_domain` | `string` | firebaseConfig.authDomain -- the domain Firebase Authentication redirects through (e.g. my-project.firebaseapp.com). |
| `status.outputs.database_url` | `string` | firebaseConfig.databaseURL -- the default Realtime Database URL, when the project has one (empty otherwise). |
| `status.outputs.storage_bucket` | `string` | firebaseConfig.storageBucket -- the default Cloud Storage for Firebase bucket, when the project has one (empty otherwise). |
| `status.outputs.location_id` | `string` | firebaseConfig.locationId -- the project's default GCP resource location, once finalized (empty until then). |
| `status.outputs.messaging_sender_id` | `string` | firebaseConfig.messagingSenderId -- the project number a browser client registers with to receive push. |
| `status.outputs.measurement_id` | `string` | firebaseConfig.measurementId -- the Google Analytics stream id, when the app is linked to a Google Analytics 4 property (empty otherwise). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpFirebaseProject | `status.outputs.project_id` |
| `spec.apiKeyId` | GcpApiKey | `status.outputs.uid` |

## See Also

- [Overview](../README.md)
