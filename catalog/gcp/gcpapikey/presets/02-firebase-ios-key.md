# Firebase iOS Key

The key a Firebase Apple app registration references: presentable only by
the named bundle id, and only against the Firebase installation and Cloud
Messaging registration APIs.

## What it configures

- `iosKeyRestrictions.allowedBundleIds` — the app's bundle id. iOS keys
  are restricted by bundle id alone (there is no certificate pairing on
  Apple's side; App Attest and DeviceCheck through App Check are the
  attestation layer).
- `apiTargets` — the two APIs an FCM client calls to register for push.

## Adjust before deploying

- **`allowedBundleIds`** — your app's bundle id (e.g. `ai.planton.mobile`);
  list a second one for a separate development build if it uses another id.
- **`apiTargets`** — add the services your app's other Firebase SDKs call,
  or drop the list to allow every enabled API on the project.
- Reference this key's `uid` output from the `GcpFirebaseAppleApp`'s
  `apiKeyId`. Push on iOS additionally needs the APNs authentication key
  uploaded in the Firebase console — Google exposes no API for it.

## When to choose something else

Android takes the **Firebase Android Key** preset; a web app the **Browser
Key** preset.
