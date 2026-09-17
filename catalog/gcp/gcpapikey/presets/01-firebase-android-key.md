# Firebase Android Key

The key a Firebase Android app registration references: presentable only by
the named package signed with the named certificate, and only against the
Firebase installation and Cloud Messaging registration APIs.

## What it configures

- `androidKeyRestrictions.allowedApplications` — one package name paired with
  its signing-certificate SHA-1. The pairing is the restriction; a package
  name alone is trivially spoofed.
- `apiTargets` — `firebaseinstallations.googleapis.com` and
  `fcmregistrations.googleapis.com`, the two APIs an FCM client calls to
  register for push.

## Adjust before deploying

- **`packageName`** — your app's application id (e.g. `ai.planton.mobile`).
- **`sha1Fingerprint`** — the SHA-1 of the certificate the app is signed
  with; add one entry per certificate (debug, upload, Play App Signing).
  `keytool -list -v -keystore <keystore>` prints it.
- **`apiTargets`** — add the services your app's other Firebase SDKs call
  (Firestore, Auth, Remote Config) or drop the list to allow every enabled
  API on the project.
- Reference this key's `uid` output from the `GcpFirebaseAndroidApp`'s
  `apiKeyId` so Firebase uses it instead of auto-provisioning an
  unrestricted one.

## When to choose something else

An iOS app takes the **Firebase iOS Key** preset (bundle id restriction);
a web app the **Browser Key** preset (referrer restriction). A backend that
calls Google APIs takes the **Server Key** preset, or better, no key at all
and Workload Identity.
