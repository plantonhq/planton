# Push Only

The registration alone: the shape a notifications-only Android app needs.
The app gets its `google-services.json`, receives Cloud Messaging, and
nothing else is configured.

## What it configures

- The app registration inside the referenced `GcpFirebaseProject`, with
  the display name and the immutable package name.
- `deletionPolicy: PREVENT`, because a shipped app's configuration file is
  in users' hands and `DELETE` would remove the app immediately and
  permanently.
- No certificate fingerprints (push needs none), no API key reference
  (Firebase provisions an unrestricted one), no App Check.

## Adjust before deploying

- **`packageName`** -- the package name the shipped APK carries. It is the
  app's permanent identity; a project accepts it once.
- **`projectId`** -- the `GcpFirebaseProject` that enabled Firebase on the
  target project.
- Add `apiKeyId` referencing a `GcpApiKey` restricted to this package name
  if the key that ships in the APK must not be usable by another app.

## When to choose something else

Apps that must prove they are genuine before the backend answers take the
**Play Integrity Attested** preset.
