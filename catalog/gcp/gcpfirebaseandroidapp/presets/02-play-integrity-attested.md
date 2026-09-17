# Play Integrity Attested

An app that proves it is the genuine APK on a genuine Google Play device
before the project's backends answer it, with a restricted API key and a
debug token so CI can still pass App Check.

## What it configures

- The registration with the release certificate's SHA-256 fingerprint --
  Play Integrity attests against the certificate the installed APK is
  signed with (Play App Signing's for a Play release).
- A SHA-1 fingerprint for Google Sign-In, Dynamic Links, and Phone Auth.
- `apiKeyId` referencing a `GcpApiKey` restricted to this package name and
  certificate (the key kind's **Firebase Android Key** preset), so the key
  that ships in the APK cannot be lifted into another app.
- App Check with Play Integrity at Google's default one-hour token
  lifetime, and one debug token for the CI emulator held as a managed
  secret reference.
- `deletionPolicy: PREVENT`.

## Adjust before deploying

- **The fingerprints** -- replace both with the shipped APK's; the SHA-256
  must be the signing certificate's or attestation fails at runtime.
- **The debug token** -- point the secret reference at a UUID4 you
  generated; remove the entry when the lane it serves is gone.
- Enforcement is the project's decision: list the backends on the
  `GcpFirebaseProject`'s `appCheck.serviceConfigs`, `UNENFORCED` first.

## When to choose something else

A notifications-only app with no backend to protect takes the **Push Only**
preset.
