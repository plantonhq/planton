# Browser Key

A key for a web application: presentable only from pages whose referrer
matches the listed patterns, and only against the listed APIs. The shape a
Firebase web app's `firebaseConfig` carries, or a Maps JavaScript key.

## What it configures

- `browserKeyRestrictions.allowedReferrers` — referrer URL patterns with
  wildcards. Requests with no Referer header (native apps, curl) are
  rejected by a browser-restricted key, which is the point.
- `apiTargets` — Firebase installations, FCM registrations, and Identity
  Toolkit (Firebase Authentication) for a web client that signs users in
  and receives push.

## Adjust before deploying

- **`allowedReferrers`** — your production origins. Add `http://localhost:*/*`
  in a separate development key rather than here.
- **`apiTargets`** — swap for `maps-backend.googleapis.com` and friends for a
  Maps key, or add Firestore / Storage as your web SDKs need.
- Reference this key's `uid` output from the `GcpFirebaseWebApp`'s
  `apiKeyId`.

## When to choose something else

Mobile apps take the **Firebase Android Key** or **Firebase iOS Key**
preset; backends the **Server Key** preset.
