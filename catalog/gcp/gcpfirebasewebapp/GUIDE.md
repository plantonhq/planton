# GcpFirebaseWebApp Guide

The judgment this guide protects: everything this resource produces ships
in the page source, its removal is permanent, and a web client attests
with one provider even when two are configured.

## A web app has no identity beyond its name

Unlike an Android package name or an Apple bundle id, nothing on a web app
is unique or immutable: Firebase identifies it by the app id it assigns,
and several web apps may coexist in one project. Name each registration
for the client it serves (the console, the marketing site, the internal
tool) so the outputs land in the right build.

## DELETE is immediate and permanent

Firebase normally keeps a removed app recoverable for 30 days.
`deletionPolicy: DELETE` (the default) removes the app with `immediate=true`
and skips that window: the app id is gone and every page carrying its
`firebaseConfig` stops working. Right for a throwaway registration, wrong
for a shipped site -- set `PREVENT` the day the site reaches users.
`ABANDON` leaves the app registered and simply stops managing it.

## firebaseConfig ships in the page, API key included

The seven outputs -- `api_key`, `auth_domain`, `database_url`,
`storage_bucket`, `location_id`, `messaging_sender_id`, `measurement_id` --
are exactly what the web SDK's `initializeApp` takes, and they are
delivered to every browser that loads the page. That is how Firebase works
on the web: the values are client identifiers, and access to the project's
backends is governed by IAM, Firebase Security Rules, and App Check, never
by hiding them. The one lever that limits what the API key can be used FOR
is the key's own restrictions.

## Which key the app presents

When `apiKeyId` is empty, Firebase associates an existing valid key or
provisions a new one -- unrestricted. Google's recommended practice for a
web app is one key restricted to the site's HTTP referrers, with API
restrictions that include the Firebase APIs the app uses (Firebase
Installations and FCM Registration for push). Reference a `GcpApiKey` of
that shape; a key that excludes those APIs is rejected when the app is
created or updated -- the Firebase Management API requires the key to be
valid for the app.

## One client provider, two possible configurations

A browser initialises App Check with either `ReCaptchaV3Provider` or
`ReCaptchaEnterpriseProvider`, never both. Google nonetheless keeps the
two backend configurations independent, and that independence is the
migration path: configure `recaptchaEnterprise`, ship the client change,
then drop `recaptchaV3`. The v3 configuration takes the site SECRET (the
server-side half of a google.com/recaptcha key pair; a secret, held as a
managed reference Google never returns); the Enterprise configuration
takes the site KEY (the public value the page embeds; reCAPTCHA Enterprise
has no secret). Enterprise bills per assessment beyond its free tier on its
own service; v3 is free.

## App Check configurations never delete

The reCAPTCHA configurations are per-app singletons Google never deletes.
Removing one from the spec makes the module forget it; the configuration
disappears only when the app itself is removed. Debug tokens are
different: each is a real resource with a real delete, governed by
`deletionPolicy`, and removing an entry revokes that token.

## Debug tokens are secrets

A debug token lets any page that presents it
(`self.FIREBASE_APPCHECK_DEBUG_TOKEN`) pass App Check as this app -- the
point, for localhost, and the danger, for anyone else. The `token` field is
sensitive: on the platform it holds a managed secret reference, never
plaintext. Name each token for the machine or lane it serves, and revoke
by removing the entry.

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
