# Messaging with Default Bucket

Firebase, Cloud Messaging, and the project's default Cloud Storage for
Firebase bucket in a multi-region location, protected from accidental
destroy.

## What it configures

- `defaultStorageLocation: US` — creates the default bucket the Firebase
  client SDKs use when no bucket is named, geo-redundant across the US
  multi-region. Created once per project; immutable.
- `deletionPolicy: PREVENT` — a destroy fails rather than deleting the
  bucket with its objects.

## Adjust before deploying

- **The project must be on the pay-as-you-go (Blaze) plan** — linked to a
  Cloud Billing account. The bucket create fails without it.
- **`defaultStorageLocation`** — `EU` for European residency, a region such
  as `us-central1` to sit beside Firestore and Functions (not
  geo-redundant).
- Object-level access is governed by Firebase Security Rules on the bucket,
  authored with your app.

## When to choose something else

A push-only project takes the **Enable Only** preset; a project that needs
enforcement over its backends adds the **App Check Enforced** shape.
