# Push Web Client

The registration with a referrer-restricted key: the shape a
notifications-only site needs. The page gets its `firebaseConfig`,
receives Cloud Messaging, and the key it presents works only from the
site's origins.

## What it configures

- The app registration inside the referenced `GcpFirebaseProject`, named
  for the client it serves.
- `apiKeyId` referencing a `GcpApiKey` restricted by HTTP referrer (the key
  kind's **Browser Key** preset), so the key that ships in the page cannot
  be used from another origin.
- `deletionPolicy: PREVENT`, because a shipped site's pages carry this
  app's `firebaseConfig` and `DELETE` would remove the app immediately and
  permanently.
- No App Check.

## Adjust before deploying

- **`displayName`** -- a web app's only identity; name it for the client.
- **`projectId`** -- the `GcpFirebaseProject` that enabled Firebase on the
  target project.
- **The key** -- the referenced `GcpApiKey`'s referrers must list the
  site's origins, and its API targets must include the Firebase APIs the
  page uses.

## When to choose something else

Sites that must prove a real browser produced each request take the
**reCAPTCHA Enterprise Attested** preset.
