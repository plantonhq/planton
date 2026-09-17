---
title: Primary Sign-In and the Break-Glass — Sending Everyone to the Directory Without Locking Anyone Out
description: What spec.oidc.primary changes (every sign-in goes straight to the company directory, the identity server's page is skipped), the one law that makes it safe (the local form stays reachable at a hint the console, the CLI, and device sign-in all carry), how to flip it without a lockout, and what to do when the directory itself is down. Read when a person asks to make Entra ID (or any OIDC provider) the only sign-in, when a sign-in loops to Microsoft with no way to the local admin, or when the break-glass is being tested.
---

# Primary Sign-In and the Break-Glass

Read this when an adopter wants the company directory to be the sign-in, not a button beside a password form -- and whenever anyone asks "what if Entra is down?". The field is documented at `spec.oidc.primary` in `operator/api/v1/plantonidentityprovider_types.go` in the open-source repository; this file carries the law and the choreography.

## The doctrine in one sentence

`primary: true` sends every sign-in straight to the upstream provider, and it is safe only because the local username-and-password form stays reachable at one hint -- `kc_idp_hint=local` -- that every surface carries, so the local administrator can always get in.

## What primary changes

With `primary: false` (the default) the identity server's sign-in page shows the local form and a button carrying `signInButtonLabel`. With `primary: true` the identity server's page is skipped: opening the console lands on the company directory's own sign-in (Microsoft's, Okta's). The person never sees a Planton password form. Nothing else changes -- the same broker, the same claims, the same mapping and offboarding.

The operator implements it with the identity server's own redirector on the browser flow, pointed at the broker; it owns that redirector and converges it, so it is never set by hand in the admin console.

## The break-glass, on every surface

The hint is the same everywhere; each surface has a door for it:

| Surface | Door |
|---|---|
| Web console | `https://<origin>/login?local=1` |
| `planton` CLI | `planton login --local` |
| Device sign-in (the CLI's broker path and the desktop app) | the same `local` on the console's `/device/auth` picker, which `planton login --local` sets for you |

All three land on the identity server's local form instead of the directory. The local administrator from the first sign-in (`self-hosted.first-admin-and-seats.md`) signs in there. **Bookmark the console URL and keep the local admin's password in the adopter's own secret store before flipping primary** -- that is the whole lockout prevention.

The CLI flag's own words: *"sign in with the instance's own username and password even when its directory is the primary sign-in (break-glass)"*.

## Flipping it without a lockout

1. Connect the directory with `primary: false` and prove a real directory sign-in lands (`self-hosted.identity-connecting.md`).
2. Prove the break-glass **before** it is needed: in a fresh browser context open `/login?local=1` and sign in as the local admin; run `planton login --local` once.
3. Set `primary: true` on the manifest and apply (a mutation: one clear yes). The operator converges within a pass; existing sessions are untouched.
4. Prove both again: a fresh context at the console lands on the directory; a fresh context at `/login?local=1` lands on the local form.

Verified on a running install against a live Entra ID tenant: the redirect to Microsoft, the web break-glass in a fresh context, `planton login --local`, and the device path.

## When the directory is down

Sign-in through the directory fails at the directory; Planton is fine. The local administrator gets in through the break-glass, and the platform keeps running for every existing session (tokens are the identity server's, not the directory's, so nothing already signed in is affected). Do not flip `primary` back to "fix" an upstream outage; the break-glass is the design for it.

## Two things people ask

- **"The break-glass page still sends me to Microsoft."** Almost always a lingering session: the browser carries an identity-server cookie from an earlier directory sign-in, and a signed-in session does not need a form. Test in a fresh browser context (a private window). On older console builds a cold-load race could also drop the hint; a current release carries it deterministically. If a fresh context still lands on the directory, that is a platform gap -- file it with the console version and the request trace (`craft.filing-platform-gaps.md`).
- **"The local form said my login attempt timed out."** The identity server's sign-in page has a login window of a few minutes; a form left open past it expires, and the page says so. Reload `/login?local=1` and sign in again -- the hint is carried afresh.
- **"Does primary bypass our tenant's MFA or Conditional Access?"** No; those run inside the directory's own sign-in, which primary sends people to. The lab tenant's MFA prompt was met during verification; an adopter's specific Conditional Access policies were not exercised (`self-hosted.identity-connecting.md`, "Tenant-side facts not exercised").

## What never to do

- Never set `primary: true` on a manifest whose directory sign-in has not yet been proven with a real user; the first proof and the flip are two applies.
- Never map the local administrator to a directory group or offboard it; it is the break-glass, and the break-glass must be a local account.
- Never edit the browser flow or the redirector in the identity server's admin console; the operator owns them and converges them back.
- Never disable the local form "for security"; the design has no other door when the directory is unreachable.
