# GcpFirebaseAndroidApp Guide

The judgment this guide protects: an app registration is IDENTITY, not
configuration. Its package name is permanent, its removal is permanent,
and the file it produces ships inside every installed copy of the app.

## The package name is forever

`packageName` is the app's identity in Firebase and in
`google-services.json`. It cannot be changed -- a different package name is
a different registration with a new app id and a new configuration file --
and a project accepts each package name exactly once. Register the package
name the shipped APK actually carries, in the project you mean to keep.

## DELETE is immediate and permanent

Firebase normally keeps a removed app recoverable for 30 days.
`deletionPolicy: DELETE` (the default) removes the app with `immediate=true`
and skips that window: the app id is gone, the configuration file installed
in users' devices stops working, and the package name is registrable again
at once. That is the right behavior for a throwaway registration and the
wrong one for a shipped app -- set `PREVENT` the day the app reaches users.
`ABANDON` leaves the app registered and simply stops managing it.

## google-services.json is a build input, not a secret

`config_file_contents` is the base64 of the file the Google Services Gradle
plugin reads from `app/google-services.json`. It carries the app id, the
project number (the FCM sender id), and the API key -- client identifiers
that ship inside every APK by design. Access to the project's backends is
governed by IAM, Firebase Security Rules, and App Check, never by keeping
this file hidden. Decode it in the build; do not treat it as a credential.

## Which key the app presents

When `apiKeyId` is empty, Firebase associates an existing valid key or
provisions a new one -- unrestricted. Google's recommended practice is one
restricted key per app: reference a `GcpApiKey` whose Android restriction
names this package name and its SHA-1, and whose API restrictions include
the Firebase APIs the app uses (Firebase Installations and FCM Registration
for push). A key that excludes those APIs, or names another package, is
rejected when the app is created or updated -- the Firebase Management API
requires the key to be valid for the app.

## Push needs no certificate; Play Integrity needs the SHA-256

Cloud Messaging works with the package name and the configuration file
alone. Certificate fingerprints serve other products: SHA-1 for Google
Sign-In, Dynamic Links, and Phone Authentication; SHA-256 for App Check
with Play Integrity, which attests against the certificate the installed
APK is signed with -- for a Play release that is Play App Signing's
certificate, not the upload key's. Configuring `playIntegrity` without the
matching SHA-256 is accepted by the API and fails at attestation time, so
the module does not refuse it; this guide does.

## App Check configurations never delete

The Play Integrity configuration is a per-app singleton Google never
deletes. Removing `playIntegrity` from the spec (or declaring it
`enabled: false`) makes the module forget it; the configuration disappears
only when the app itself is removed. Debug tokens are different: each is a
real resource with a real delete, governed by `deletionPolicy`, and
removing an entry revokes that token.

## Debug tokens are secrets

A debug token lets any build that presents it pass App Check as this app --
that is the point, for emulators and CI, and the danger, for anyone else.
The `token` field is sensitive: on the platform it holds a managed secret
reference, never plaintext. Name each token for the machine or lane it
serves (the name is the key the modules manage it by), and revoke by
removing the entry.

## Enforcement lives on the project

This kind configures HOW the app attests. WHETHER a backend rejects
unattested requests is the `GcpFirebaseProject`'s `appCheck` -- start every
service at `UNENFORCED`, watch the metrics, then `ENFORCED` once every
shipped client attests, or you lock out your own users.

## Two blocks ride the beta provider

Google publishes the registration and its configuration lookup only in the
`google-beta` Terraform provider. The Terraform module attaches that
provider to exactly those two blocks under a recorded admission
(`pkg/providerparity/admissions/google-beta.yaml`); App Check and API
enablement stay on the GA provider. Pulumi's SDK is bridged from the beta
provider, so one provider instance serves everything there. Both engines
set `user_project_override` -- the Firebase Management API attributes quota
to the caller's project on user-credential calls.
