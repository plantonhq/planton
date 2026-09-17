# App Attest with DeviceCheck

An app that proves it is the genuine signed app on a genuine Apple device
before the project's backends answer it -- Google's recommended pairing of
App Attest (iOS 14+) with DeviceCheck as the fallback -- plus a restricted
API key and a simulator debug token.

## What it configures

- The registration with its team id (required for both attestations) and
  App Store id.
- `apiKeyId` referencing a `GcpApiKey` restricted to this bundle id (the
  key kind's **Firebase iOS Key** preset), so the key that ships in the
  bundle cannot be lifted into another app.
- App Attest at Google's default one-hour token lifetime; DeviceCheck with
  the DeviceCheck key's Key ID and its `.p8` contents held as a managed
  secret reference; one debug token for the simulator, also a secret
  reference.
- `deletionPolicy: PREVENT`.

## Adjust before deploying

- **`teamId`** and **`deviceCheck.keyId`** -- from the Apple Developer
  account; the DeviceCheck key is NOT the APNs key.
- **The secret references** -- point them at the DeviceCheck `.p8` text and
  a UUID4 you generated for the simulator.
- **After deploying**, upload the APNs authentication key in the Firebase
  console; push does not deliver without it.
- Enforcement is the project's decision: list the backends on the
  `GcpFirebaseProject`'s `appCheck.serviceConfigs`, `UNENFORCED` first.

## When to choose something else

A notifications-only app with no backend to protect takes the **Push Only**
preset.
