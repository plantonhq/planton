# App Check Enforced

Every Firebase backend rejects requests that do not prove they came from
one of the project's genuine registered apps; Authentication collects
metrics first.

## What it configures

- `appCheck.serviceConfigs` — Firestore, Cloud Storage for Firebase, and
  Realtime Database at `ENFORCED`; Authentication at `UNENFORCED` (metrics
  only) because locking sign-in out of unattested clients is the change
  with the widest blast radius.
- `deletionPolicy: PREVENT` — a destroy fails rather than silently switching
  enforcement OFF.

## Adjust before deploying

- **Start everything at `UNENFORCED`** on a project with shipped clients,
  read the App Check metrics in the Firebase console, and move services to
  `ENFORCED` one at a time once every client attests.
- **Attestation providers live on the app registrations** — Play Integrity
  on `GcpFirebaseAndroidApp`, App Attest / DeviceCheck on
  `GcpFirebaseAppleApp`, reCAPTCHA on `GcpFirebaseWebApp`. Enforcement here
  without attestation there rejects your own apps.
- Add `resourcePolicies` to override enforcement per iOS OAuth client under
  Google Identity for iOS.

## When to choose something else

A project with no App Check clients yet takes the **Enable Only** preset
and adds enforcement later — enforcement is an in-place update.
