# GcpApiKey Guide

The judgment this guide protects: an API key is an identifier that ships
inside the client, so its safety comes from what it is allowed to do, not
from hiding it. Every field on this kind is about drawing that boundary.

## Restrict first, then hand it to the app

An unrestricted key is accepted by Google and is what Firebase
auto-provisions when an app registration names none. It is also a
credential any caller anywhere can present, billed to your project. Declare
the client arm that matches the platform (Android package plus signing
certificate, iOS bundle ids, browser referrers, server IPs) and list the
APIs the client actually calls in `apiTargets`. Then reference the key's
`uid` from the Firebase app registration instead of letting Firebase mint
its own.

## One key per client platform

Google allows exactly one client arm per key, and the spec rejects a
second. A product with an Android app, an iOS app, and a web app declares
three keys, each restricted to its platform and referenced by its own app
registration. Sharing one unrestricted key across platforms is the shape
this kind exists to replace.

## What makes a key valid for a Firebase app

The Firebase Management API refuses an app registration whose `api_key_id`
names a key that is not valid for the app -- and it checks. A valid key is
unrestricted, or restricted ONLY in ways that fit this app: an Android arm
naming this package name (and its certificate), an iOS arm naming this
bundle id, browser referrers covering the site's origins. Its API
restrictions, when set, must include every Firebase API the app uses --
for push that is Firebase Installations (`firebaseinstallations.googleapis.com`)
and FCM Registration (`fcmregistrations.googleapis.com`); add Identity
Toolkit for Authentication, Firestore for Firestore, and so on. A key
restricted to the wrong platform or missing one of those APIs fails the
app's create or update, not the key's own apply.

## Android needs the certificate, not just the package

The Android arm pairs a package name with the SHA-1 of the signing
certificate, because a package name alone is trivially spoofed. List the
package once per certificate the app ships under: the debug certificate
developers build with, the upload certificate, and the Play App Signing
certificate. Get fingerprints with `keytool -list -v` or from the Play
Console's App signing page; both colon and colon-free forms are accepted.

## Identity is immutable; restrictions are not

`keyId`, `projectId`, and `serviceAccountEmail` are immutable: changing any
of them destroys and recreates the key, which rotates the key string every
shipped client holds. Restrictions and the display name update in place,
so tightening a key that is already in the field is safe and cheap. Pick
the id deliberately and treat it like the package name.

## Deletion is soft, and ids are reserved

`DELETE` soft-deletes the key: the string stops working immediately, Google
keeps the key recoverable for 30 days, and the id cannot be reused in that
window. For a key baked into a shipped mobile binary you cannot rotate on
demand, set `PREVENT` so a destroy fails instead of breaking installed
apps. Ephemeral keys (test fixtures) need a fresh id per lifetime.

## The service-account arm is for callers that cannot use OAuth

`serviceAccountEmail` makes requests carrying the key authenticate as that
service account, for the APIs that accept API-key authentication in place
of OAuth. It is still a static bearer credential; pair it with a server IP
restriction and prefer Workload Identity with no key at all wherever the
caller runs on Google Cloud.

## Quota project attribution

Both engines set `user_project_override`: the API Keys API attributes
quota to the caller's project on user-credential calls, and a deploy under
plain ADC fails with "requires a quota project" without it. The override
attributes quota to the key's own project under every credential mode,
the same posture the Identity Platform kinds take.
