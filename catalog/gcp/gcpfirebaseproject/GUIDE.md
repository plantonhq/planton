# GcpFirebaseProject Guide

The judgment this guide protects: this resource is a ONE-WAY door on a
shared project. Enabling Firebase cannot be undone, and every Firebase app
in the project lives inside the single enablement this kind declares.

## One Firebase project per GCP project, forever

Google offers no way to remove Firebase from a project: destroy detaches
this resource and the project stays enabled. A project that already has
Firebase (enabled from the console, or by an earlier deploy) is ADOPTED on
apply -- the provider reads first and never issues a second enablement --
so bringing an existing project under management is safe. What is not
reversible is the choice of project: enable Firebase on the project you
mean to keep.

Because the enablement is per project, every environment that shares a GCP
project shares one Firebase project. Environments separate by IDENTITY --
which control plane's service account holds `roles/firebasemessaging.admin`
-- not by Firebase project. Firebase rejects a second app registration with
the same package name or bundle id in one project, so declare each app
registration exactly once, in the layer that owns the shared project.

## Cloud Messaging is declared, not assumed

`fcm.googleapis.com` is enabled explicitly alongside `firebase.googleapis.com`.
Enablement's side effects are not a contract; the push channel this kind
exists for is written down. Sending push then needs two things this kind
does not declare: the sender's identity (a `GcpServiceAccount`) and its
grant (`GcpProjectIamMember` with `roles/firebasemessaging.admin`). Bind
that identity to the control plane through Workload Identity and the sender
runs keyless -- no service-account key file anywhere.

## The default bucket is a one-time, billing-changing act

`defaultStorageLocation` creates the project's default Cloud Storage for
Firebase bucket. It can be created once per project, its location is
immutable, and it requires the pay-as-you-go plan: the project must be
linked to a Cloud Billing account. Set it when app content needs Firebase
Storage; leave it empty on a push-only project or one without billing. A
multi-region location (`US`, `EU`) is geo-redundant; a region is not.

## App Check: UNENFORCED first, ENFORCED second

App Check makes a Firebase backend reject requests that do not prove they
came from a genuine registered app. This kind decides WHICH backends
enforce (Firestore, Storage, Realtime Database, Authentication); the app
registration kinds configure HOW each app attests (Play Integrity, App
Attest, DeviceCheck, reCAPTCHA). Always start a service at `UNENFORCED`:
it collects metrics on what would be rejected while rejecting nothing.
Switch to `ENFORCED` only when every shipped client attests, or you lock
out your own users.

## deletionPolicy protects the composed resources, never the enablement

`DELETE` (the default) removes the default bucket WITH ITS OBJECTS and
switches App Check OFF for every listed service. `PREVENT` fails the
destroy -- the right value once the bucket holds user content or a
production backend runs `ENFORCED`. `ABANDON` leaves everything in place,
unmanaged. The enablement itself is unaffected by all three: it is always
detached.

## Outputs arrive after enablement

`project_number` is the FCM sender id (`messagingSenderId`). The three
Admin SDK values -- `database_url`, `storage_bucket`, `location_id` -- are
read after the enablement and are each present only when the project has
the corresponding resource (a Realtime Database, a default bucket, a
finalized default location); each degrades to empty otherwise. Offline
plans stay credential-free because the read is deferred to apply.

## Two resources ride the beta provider

Google publishes the enablement and the default bucket only in the
`google-beta` Terraform provider. The Terraform module attaches that
provider to exactly those two resources under a recorded admission
(`pkg/providerparity/admissions/google-beta.yaml`); App Check and API
enablement stay on the GA provider. Pulumi's SDK is bridged from the beta
provider, so one provider instance serves everything there -- provider
packaging, not a behavioral difference.

## Quota project attribution

Both engines set `user_project_override`: the Firebase Management API
attributes quota to the caller's project on user-credential calls, and a
deploy under plain ADC fails with "requires a quota project" without it --
Google's own Firebase provider guidance, and the same posture the Identity
Platform kinds take.
