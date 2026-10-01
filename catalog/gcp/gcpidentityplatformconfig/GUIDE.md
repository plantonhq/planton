# GcpIdentityPlatformConfig Guide

The judgment this guide protects: this resource is a ONE-WAY door. The
first apply permanently initializes Identity Platform on the project, and
every later mistake lands on a live authentication surface.

## Choose the project deliberately

Initialization cannot be undone — GCP has no de-initialize, billing is
required, and destroy abandons the config in place rather than deleting
anything. A config applied to the wrong project leaves that project
Identity-Platform-enabled forever. Treat the first deploy of this kind
with the same care as creating the project itself.

## Already-initialized projects need adoptExisting

Initialization is also ONCE-ONLY: GCP rejects a second initializeAuth
with "Identity Platform has already been enabled for this project"
(live-verified). Any project where Identity Platform was ever enabled —
a console click, a Firebase Auth setup, a previous deploy of this kind
(destroy abandons, so a re-deploy hits this too) — needs
`adoptExisting: true`, which imports the project's config singleton and
applies your spec as an update. The failure mode without it is loud and
immediate; the fix is one field. Fresh, never-initialized projects leave
it unset.

## The tenant gate lives here

`multiTenant.allowTenants: true` is the prerequisite for every
GcpIdentityPlatformTenant in the project — the tenant API rejects
creation otherwise. If multi-tenancy is anywhere on the roadmap, arm the
gate in the same manifest that initializes the project; flipping it later
is a safe update, but forgetting it is the most common first-tenant
failure.

## enabled=false must reach the API

Every sign-in arm's `enabled` is sent explicitly whenever the arm is
present in the spec. This is deliberate: these flags drive live sign-in
methods, and a manifest transition from true to false must actively
disable the method rather than being silently omitted (the
send-true-or-omit class of bug). Omit an arm entirely to leave it
unmanaged; set it with `enabled: false` to shut it off.

## SMS region policy is a money control

Phone sign-in and SMS MFA send real, billable SMS — toll fraud against
unrestricted projects is an established attack. `allowlistOnly` with the
regions you actually serve is the right default posture; the deny-list
arm (`allowByDefault`) exists for genuinely global products.

## Blocking functions sit in the hot path

Every trigger fires synchronously inside sign-up/sign-in, so the
function's latency is added to every user's authentication. Keep the
function fast and regional to your users, and forward only the tokens
(`forwardInboundCredentials`) its logic actually inspects.

## The api_key output is a live credential

It is what client apps bootstrap the Identity Platform / Firebase Auth
SDK with — public by design but abusable when unrestricted. Restrict it
by domain/app in the console before shipping; the engines already mark it
sensitive, and Pulumi encrypts it in state.
