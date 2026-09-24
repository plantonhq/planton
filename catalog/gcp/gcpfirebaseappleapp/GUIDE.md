# GcpFirebaseAppleApp Guide

The judgment this guide protects: an app registration is IDENTITY, not
configuration. Its bundle id is permanent, its removal is permanent, the
file it produces ships inside every installed copy of the app -- and on
Apple platforms, push needs one key this resource cannot declare.

## The bundle id is forever

`bundleId` is the app's identity in Firebase and in
`GoogleService-Info.plist`. It cannot be changed -- a different bundle id is
a different registration with a new app id and a new plist -- and a project
accepts each bundle id exactly once. Register the bundle id the shipped app
actually carries, in the project you mean to keep.

## DELETE is immediate and permanent

Firebase normally keeps a removed app recoverable for 30 days.
`deletionPolicy: DELETE` (the default) removes the app with `immediate=true`
and skips that window: the app id is gone, the plist installed in users'
devices stops working, and the bundle id is registrable again at once.
Right for a throwaway registration, wrong for a shipped app -- set
`PREVENT` the day the app reaches users. `ABANDON` leaves the app
registered and simply stops managing it.

## The APNs key is the one console step

Sending push to an Apple device goes through Apple Push Notification
service, and Firebase needs the APNs authentication key Apple mints (a
`.p8` with its Key ID and your Team ID) to do it. Firebase exposes no API
to upload that key -- neither Terraform nor Pulumi can declare it -- so it
is uploaded once, in the Firebase console under Cloud Messaging, for the
app this resource registers. Everything else about the app is declared
here; that one step is recorded where the estate keeps its vendor
procedures, not pretended away.

## The DeviceCheck key is not the APNs key

Apple issues two different `.p8` keys that are easy to confuse.
`deviceCheck.privateKey` is the DeviceCheck key (Certificates, Identifiers &
Profiles, Keys, with the DeviceCheck service enabled), used by Google to
verify DeviceCheck tokens server-side. The APNs key is a separate key with
the Apple Push Notifications service enabled, and it never enters this
resource. Both are secrets; only the DeviceCheck key is declared here, as
a managed secret reference Google never returns (the module reports only
whether it is set).

## GoogleService-Info.plist is a build input, not a secret

`config_file_contents` is the base64 of the plist the Firebase Apple SDK
reads from the app bundle. It carries the app id, the project number (the
FCM sender id), and the API key -- client identifiers that ship inside
every installed app by design. Access to the project's backends is governed
by IAM, Firebase Security Rules, and App Check, never by keeping this file
hidden. Decode it in the build; do not treat it as a credential.

## Which key the app presents

When `apiKeyId` is empty, Firebase associates an existing valid key or
provisions a new one -- unrestricted. Google's recommended practice is one
restricted key per app: reference a `GcpApiKey` whose iOS restriction names
this bundle id, and whose API restrictions include the Firebase APIs the
app uses (Firebase Installations and FCM Registration for push). A key that
excludes those APIs, or names another bundle id, is rejected when the app
is created or updated -- the Firebase Management API requires the key to
be valid for the app.

## App Attest first, DeviceCheck as the fallback, team id for both

App Attest (iOS 14+) is Google's recommended attestation for Apple apps;
DeviceCheck works on every iOS 11+ device and is the fallback for the rest.
Google's guidance is to configure both. Either one requires `teamId`: Apple
validates attestations against the Developer Team that signed the app, and
Google refuses the configuration without it -- this kind refuses at
validation rather than at apply. A `teamId` on a push-only app is harmless
and future-proof.

## App Check configurations never delete

The App Attest and DeviceCheck configurations are per-app singletons Google
never deletes. Removing one from the spec (or declaring App Attest
`enabled: false`) makes the module forget it; the configuration disappears
only when the app itself is removed. Debug tokens are different: each is a
real resource with a real delete, governed by `deletionPolicy`, and
removing an entry revokes that token.

## Debug tokens are secrets

A debug token lets any build that presents it pass App Check as this app --
the point, for simulators and CI, and the danger, for anyone else. The
`token` field is sensitive: on the platform it holds a managed secret
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
