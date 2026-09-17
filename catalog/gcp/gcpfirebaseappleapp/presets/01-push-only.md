# Push Only

The registration alone: the shape a notifications-only iOS app needs. The
app gets its `GoogleService-Info.plist`, receives Cloud Messaging once the
APNs key is uploaded, and nothing else is configured.

## What it configures

- The app registration inside the referenced `GcpFirebaseProject`, with
  the display name, the immutable bundle id, and the Apple Developer Team
  ID (harmless now, required the day App Check is added).
- `deletionPolicy: PREVENT`, because a shipped app's plist is in users'
  hands and `DELETE` would remove the app immediately and permanently.
- No App Store id (not needed for push), no API key reference (Firebase
  provisions an unrestricted one), no App Check.

## Adjust before deploying

- **`bundleId`** -- the bundle id the shipped app carries. It is the app's
  permanent identity; a project accepts it once.
- **`teamId`** -- your team's 10-character id from the Apple Developer
  Membership page.
- **After deploying**, upload the APNs authentication key in the Firebase
  console under Cloud Messaging for this app -- Firebase has no API for it,
  and iOS push does not deliver without it.
- Add `apiKeyId` referencing a `GcpApiKey` restricted to this bundle id if
  the key that ships in the bundle must not be usable by another app.

## When to choose something else

Apps that must prove they are genuine before the backend answers take the
**App Attest with DeviceCheck** preset.
