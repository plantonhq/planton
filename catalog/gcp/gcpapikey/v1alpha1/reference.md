# GcpApiKey

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpApiKeySpec defines a Google Cloud API key: a project-scoped credential
that identifies a CLIENT application to Google APIs. API keys are the
identifier Firebase app registrations carry (a Firebase Android/Apple/Web
app references a key by its `uid` output), the credential Maps and other
client-facing Google APIs read from the calling application, and -- when
bound to a service account -- a shortcut for server-to-server calls that
do not need OAuth.

What an API key IS and is not: it ships inside the application that uses
it (a mobile binary, a web page), so it is not a secret in the way a
password is -- anyone with the app has the key. It IS still a credential
Google bills and rate-limits against your project, so the whole design of
this kind is RESTRICTION: which platform may present it (one Android app
with a signing certificate, one set of iOS bundle ids, one set of web
referrers, or one set of server IPs) and which Google APIs it may call
(`api_targets`). An unrestricted key is what Firebase auto-provisions when
an app registration names none; declaring a restricted key here and
referencing it from the app is Google's recommended posture.

The key STRING (the value the client presents) is an output of this kind
and is treated as sensitive by both engines; the key's `uid` is the
referenceable identity other resources point at.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpApiKey
metadata:
  name: my-sample-api-key
spec:
  # GCP project the key belongs to. Omit to use the provider's default
  # project. Immutable: a key cannot move between projects.
  projectId:
    value: my-gcp-project-123

  # The key's resource id (projects/{project}/locations/global/keys/{keyId}):
  # lowercase letters, digits, hyphens; starts with a letter; 1-63
  # characters. Immutable -- changing it recreates the key and rotates the
  # key string. A deleted key's id stays reserved for 30 days.
  keyId: firebase-android-key

  # Console label. Freely updatable.
  displayName: Firebase Android key (production signing)

  # Optional: bind the key to a service account, making requests that carry
  # it authenticate AS that account (for APIs that accept API-key auth in
  # place of OAuth). Immutable. Leave out for a client-identifying key --
  # the Firebase case.
  # serviceAccountEmail:
  #   value: sender@my-gcp-project-123.iam.gserviceaccount.com

  # What may present the key and what it may call. At most ONE client arm
  # (android / ios / browser / server); api_targets any number. Omitting
  # the whole block makes an unrestricted key -- accepted, discouraged.
  restrictions:
    # Android: package name AND signing-certificate SHA-1, together. List the
    # package once per certificate it ships under (debug, release, Play App
    # Signing).
    androidKeyRestrictions:
      allowedApplications:
        - packageName: ai.planton.mobile
          sha1Fingerprint: DA:39:A3:EE:5E:6B:4B:0D:32:55:BF:EF:95:60:18:90:AF:D8:07:09
    # The other client arms, each exclusive with the one above:
    # iosKeyRestrictions:
    #   allowedBundleIds: [ai.planton.mobile]
    # browserKeyRestrictions:
    #   allowedReferrers: ["https://app.example.com/*"]
    # serverKeyRestrictions:
    #   allowedIps: ["203.0.113.7", "2001:db8::/32"]

    # The Google APIs the key may call (empty = every API enabled on the
    # project). For a Firebase Cloud Messaging client: installations and
    # FCM registrations, plus whatever else the app's SDKs call.
    apiTargets:
      - service: firebaseinstallations.googleapis.com
      - service: fcmregistrations.googleapis.com
      - service: translate.googleapis.com
        methods:
          - google.cloud.translate.v2.*

  # What destroy does to the key: DELETE (default; soft-deleted, recoverable
  # 30 days), PREVENT (destroy fails -- for a key baked into a shipped
  # binary), ABANDON (unmanaged, stays live).
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.keyId` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.serviceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.restrictions` | `GcpApiKeyRestrictions` |  |  |  |
| `spec.restrictions.androidKeyRestrictions` | `GcpApiKeyAndroidKeyRestrictions` |  |  |  |
| `spec.restrictions.androidKeyRestrictions.allowedApplications` | `[]GcpApiKeyAndroidApplication` | yes |  |  |
| `spec.restrictions.androidKeyRestrictions.allowedApplications[].packageName` | `string` | yes |  |  |
| `spec.restrictions.androidKeyRestrictions.allowedApplications[].sha1Fingerprint` | `string` | yes |  |  |
| `spec.restrictions.iosKeyRestrictions` | `GcpApiKeyIosKeyRestrictions` |  |  |  |
| `spec.restrictions.iosKeyRestrictions.allowedBundleIds` | `[]string` | yes |  |  |
| `spec.restrictions.browserKeyRestrictions` | `GcpApiKeyBrowserKeyRestrictions` |  |  |  |
| `spec.restrictions.browserKeyRestrictions.allowedReferrers` | `[]string` | yes |  |  |
| `spec.restrictions.serverKeyRestrictions` | `GcpApiKeyServerKeyRestrictions` |  |  |  |
| `spec.restrictions.serverKeyRestrictions.allowedIps` | `[]string` | yes |  |  |
| `spec.restrictions.apiTargets` | `[]GcpApiKeyApiTarget` |  |  |  |
| `spec.restrictions.apiTargets[].service` | `string` | yes |  |  |
| `spec.restrictions.apiTargets[].methods` | `[]string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the key belongs to. Can be a literal project ID or a
reference to a GcpProject resource. If omitted, the provider's default
project is used. Immutable: a key cannot move between projects.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.keyId

`string` · required

The key's resource id -- the last segment of
projects/{project}/locations/global/keys/{key_id}. Unique within the
project; lowercase letters, digits, and hyphens, starting with a letter,
1-63 characters (RFC 1034 label). Immutable: changing it destroys and
recreates the key, which changes the key string every client holds.

A deleted key's id stays reserved for 30 days (Google keeps deleted keys
recoverable via undelete); a fresh key cannot reuse the id in that
window, so an ephemeral key needs a fresh id per lifetime.

- rule: key_id must start with a lowercase letter and contain only lowercase letters, digits, and hyphens, ending alphanumeric (1-63 characters) -- e.g. firebase-android-key
- rule: {"required":true,"string":{"maxLen":"63"}}

### spec.displayName

`string`

Human-readable name shown in the Cloud console's Credentials page.
Freely updatable.

### spec.serviceAccountEmail

`string | valueFrom`

Bind the key to a service account, making it a service-account-bound
key: requests carrying it are authenticated AS that service account
(for APIs that accept API-key authentication in place of OAuth). Can be
a literal email or a reference to a GcpServiceAccount resource.
Immutable: binding cannot be added, changed, or removed after creation.
Leave empty for an ordinary client-identifying key -- the Firebase case.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.restrictions

`GcpApiKeyRestrictions`

What may present the key and what it may call. Omit for an
unrestricted key (accepted by Google, discouraged: any caller anywhere
can spend your quota). Every arm is freely updatable in place -- the
key string does not change when restrictions do.

- rule: restrictions may carry at most one of android_key_restrictions, ios_key_restrictions, browser_key_restrictions, server_key_restrictions -- a key identifies one kind of caller

### spec.restrictions.androidKeyRestrictions

`GcpApiKeyAndroidKeyRestrictions`

Android apps allowed to present the key, each identified by package
name AND the SHA-1 of its signing certificate. The pairing is what
makes the restriction real: a package name alone is trivially spoofed.
For a Firebase Android app, list the app's package name with every
signing certificate it ships under (debug, release, Play App Signing).

### spec.restrictions.androidKeyRestrictions.allowedApplications

`[]GcpApiKeyAndroidApplication` · required

At least one application; each pairs a package name with a signing
certificate SHA-1 fingerprint.

- rule: {"repeated":{"minItems":"1"}}

### spec.restrictions.androidKeyRestrictions.allowedApplications[].packageName

`string` · required

The app's package name, e.g. ai.planton.mobile.

- rule: package_name must be a dotted Java package name (segments of letters, digits, and underscores starting with a letter) -- e.g. ai.planton.mobile
- rule: {"required":true}

### spec.restrictions.androidKeyRestrictions.allowedApplications[].sha1Fingerprint

`string` · required

The SHA-1 fingerprint of the certificate the app is signed with, as 40
hex characters with or without colon separators
(DA:39:A3:EE:... or DA39A3EE...). Get it with
`keytool -list -v -keystore <keystore>` or from the Play Console's App
signing page. Google stores and returns the colon-free form.

- rule: sha1_fingerprint must be 40 hex characters, optionally colon-separated in pairs -- e.g. DA:39:A3:EE:5E:6B:4B:0D:32:55:BF:EF:95:60:18:90:AF:D8:07:09
- rule: {"required":true}

### spec.restrictions.iosKeyRestrictions

`GcpApiKeyIosKeyRestrictions`

iOS apps allowed to present the key, by bundle id (the same bundle id a
GcpFirebaseAppleApp registers).

### spec.restrictions.iosKeyRestrictions.allowedBundleIds

`[]string` · required

At least one bundle id, e.g. ai.planton.mobile.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.restrictions.browserKeyRestrictions

`GcpApiKeyBrowserKeyRestrictions`

Websites allowed to present the key, by HTTP referrer pattern (the key
a GcpFirebaseWebApp's firebaseConfig carries).

### spec.restrictions.browserKeyRestrictions.allowedReferrers

`[]string` · required

At least one referrer pattern. Google's referrer grammar: a URL with
optional wildcards, e.g. `https://app.example.com/*`,
`*.example.com/*`. Requests with no Referer header (native apps,
curl) are rejected by a browser-restricted key.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.restrictions.serverKeyRestrictions

`GcpApiKeyServerKeyRestrictions`

Servers allowed to present the key, by caller IP address or CIDR.

### spec.restrictions.serverKeyRestrictions.allowedIps

`[]string` · required

At least one caller address: an IPv4/IPv6 address or CIDR block, e.g.
203.0.113.7 or 2001:db8::/32.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.restrictions.apiTargets

`[]GcpApiKeyApiTarget`

The Google APIs (and optionally methods) the key may call. Empty means
every API enabled on the project -- restrict to the APIs the client
actually uses (for Firebase Cloud Messaging on a mobile client:
`firebaseinstallations.googleapis.com`, `fcmregistrations.googleapis.com`,
plus whatever else the app's SDKs call).

### spec.restrictions.apiTargets[].service

`string` · required

The service's canonical name, e.g. translate.googleapis.com,
fcmregistrations.googleapis.com.

- rule: service must be a canonical Google API service name ending in .googleapis.com -- e.g. translate.googleapis.com
- rule: {"required":true}

### spec.restrictions.apiTargets[].methods

`[]string`

Methods the key may call on the service. Empty means every method. A
trailing wildcard is allowed, e.g. `google.cloud.translate.v2.*` or
`TranslateText`.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.deletionPolicy

`string`

What destroying this resource does to the key in GCP:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the key is deleted; Google keeps it recoverable for 30
               days (console/undelete), during which the key string
               stops working and the key_id stays reserved
  "PREVENT" -- destroy FAILS; use for a key baked into a shipped
               mobile binary you cannot rotate on demand
  "ABANDON" -- the key is removed from management but stays live in
               GCP, still presentable by every client that holds it

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpApiKey, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The key's full resource name: projects/{project}/locations/global/keys/{key_id}. What the API Keys API and IAM conditions address the key by. |
| `status.outputs.uid` | `string` | The key's unique id (a UUID). This is the value a Firebase app registration's api_key_id takes -- reference it from GcpFirebaseAndroidApp / GcpFirebaseAppleApp / GcpFirebaseWebApp rather than passing the key string. |
| `status.outputs.key_string` | `string` | The key string -- the value a client presents to Google APIs. A credential Google bills and rate-limits against the project, so both engines mark it sensitive in their state and outputs; it reaches the client build through the platform's secret handling, never a log or a plan. (It ships inside the client binary by design; the restrictions on the key, not the secrecy of the string, are what bound its blast radius.) |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |

## See Also

- [Overview](../README.md)
