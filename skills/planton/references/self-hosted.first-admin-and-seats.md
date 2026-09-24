---
title: First Admin, Seats, and the License — Who Gets In, and How Many
description: The first sign-in on a fresh self-hosted Planton (the one-time setup code, the local administrator it creates, why that account is the install's break-glass), the seat ceiling every account passes through, the Community edition's five seats, and how a license key is delivered and renewed. Read when a person asks how to get into a new install, sees "seats are all in use", asks where the license key goes, or wonders whether to keep the local admin once a directory is connected.
---

# First Admin, Seats, and the License

Read this on a fresh install ("how do I get in?"), at the seat door ("seats are all in use"), and whenever the license comes up. The people-facing steps live in `site/public/docs/self-hosting/index.md` ("Your first sign-in" and "Licensing"); this file is the agent's version, with the reasons.

## The doctrine in one sentence

The first administrator is claimed once with a code the operator minted, every later account passes the same seat door, and the license is a key the operator delivers and the control plane judges -- three facts an agent reads before it explains any refusal at sign-in.

## The first sign-in

A fresh install has no accounts. The operator mints a one-time setup code and stores it in a Secret named `<platform>-identity-setup-code` in the platform's namespace; the `planton` chart's install notes print the exact `kubectl get secret … | base64 -d` command, and the identity component's `Ready` sentence names the Secret. Opening the console for the first time shows the guided setup: paste the code, choose the administrator's email and password, sign in. The code is consumed by that one claim; the account it creates is a **local** account in the platform's own identity server (Keycloak), holding the organization's owner role.

Two things people often ask at this point:

- **Where is the admin's password kept?** Nowhere but the person's head. The operator wrote a bootstrap Secret for the identity server's own admin console, not for this account. "Forgot password?" appears on the sign-in page exactly when `spec.email` is declared (the identity server sends the reset mail through the platform's provider); without mail, a Keycloak admin resets it in the realm.
- **Is the local admin still needed after we connect our directory?** Yes. It is the install's break-glass: the one account that signs in without the directory. Keep it, keep its password in the adopter's own secret store, and reach it with `kc_idp_hint=local` when the directory is primary (`self-hosted.identity-primary-and-break-glass.md`). Never offboard it, never map it to a directory group.

## Seats

Every account creation -- the setup claim, an invitation accepted, a first federated sign-in, `planton` CLI's first login -- passes the same seat door. **Accounts that already exist are unaffected**; the door only judges new ones. When it refuses, the sentence is:

> This installation's seats are all in use (N of M). Ask your platform administrator to free a seat or install a license with more seats -- accounts that already exist are unaffected.

No half-account is created at a refusal: the person can try again once a seat is free, and nothing about them was recorded. The Community edition (no license) is the full core under a ceiling of **five** seats. Freeing a seat means removing a member from the organization (the console's Members page); a licensed key raises the ceiling to what it grants.

The seat door is also where a directory sign-in can be turned away: a federated account whose directory record carries no email is refused in words before it consumes a seat (`self-hosted.identity-offboarding.md` has the sentence). When a person reports "I signed in with our company account and got a message instead of the console", the message is the diagnosis -- relay it.

## Inviting people before mail is set up

`planton invite <email>` creates a personal invitation pinned to that email, seated with the role named (`Developer` by default), and prints the link to share; on an install without email delivery, the link IS the invitation -- hand it to the person over whatever channel the adopter trusts. The console's Members page does the same. The invitation is a mutation (it seats a person when accepted): one clear yes.

## The license

The operator only delivers the key; the control plane verifies it and resolves what it grants. Two delivery shapes on the platform declaration, at most one set:

```yaml
spec:
  license:
    key: <the compact signed token from the purchase email>    # convenient for a first install
    # or, preferred in GitOps trees so the key never lives in the CR:
    secretKeyRef:
      name: planton-license
      key: license.key
```

The platform's `LICENSE` column echoes the delivery mode (`Community`, `InlineKey`, `SecretRef`), not the verdict; the console's License page (Settings) shows what the control plane resolved -- the plan, the seats, the expiry. Renewal is a Secret edit: rotate the value and the key is re-delivered on the next control-plane restart, never a reinstall. Expiry never blocks running workloads.

Installing or rotating a license is a mutation (an edit to the declaration or the Secret): exact command, one clear yes, then read the License page to confirm the verdict.

## What never to do

- Never share or store the setup code once it has been claimed; it is consumed, and a fresh one only comes with a fresh realm.
- Never create a second local administrator "for the agent" or "for automation"; a service account with an API key (`craft.planton-cli.md`) is the shape for that, and it passes no seat door of its own.
- Never explain a seat refusal as an outage or a directory problem; it is the ceiling, and the sentence says how to raise it.
