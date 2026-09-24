---
title: Offboarding and Sync — What Leaving the Directory Does, and How Fast
description: The offboarding guarantee on each arm of a connected directory (brokered: access converges at every sign-in from the provider's fresh token and a removed person cannot get another; LDAP: the sync pass revokes mapped access within the manifest's sync period plus Planton's fifteen-minute pass, with no sign-in needed), the laws the sync obeys (removals only on the directory's own answer, accounts never created, leaving never gated), the email-less refusal, invited directory users, and the alerts on sync health transitions. Read when a person asks how quickly a departed employee loses access, why someone still shows in a team, what happens when the directory is unreachable, or whether the alert email arrived.
---

# Offboarding and Sync

Read this when an adopter asks the security question every directory integration exists to answer: "when someone leaves, how fast are they out?". This file carries the guarantee per arm, the laws behind it, and the edge cases that come up in the first month.

## The doctrine in one sentence

Membership follows the directory -- joined at sign-in, converged by a fifteen-minute sync pass that removes people only on the directory's own answer -- so a person removed from a mapped group, or from the directory entirely, loses the mapped access without anyone in Planton doing anything.

## The guarantee, per arm

**Brokered (OIDC: Entra ID, Okta, …).** Planton never mirrors the provider's groups; each sign-in carries the person's current groups in a fresh token, and the mirrored teams converge to it then. A person removed from a mapped group is out of the team at their next sign-in. A person removed from the directory entirely cannot sign in again, and is out when their current session ends. For an immediate cut on this arm, remove the person from the organization on the Members page -- the one offboarding step that is an administrator's, and only when the session's end is too slow.

**LDAP (Active Directory).** The identity server imports users and groups on the manifest's `syncPeriodMinutes` (default 60), and Planton's own pass runs every fifteen minutes against that record. A person removed from a mapped group -- directly or through a nested group -- or deleted from the directory loses every mapped grant within **`syncPeriodMinutes` plus fifteen minutes, with no sign-in needed**. Tighten `syncPeriodMinutes` to tighten the window. (Proven live: a person deleted from the directory lost every mapped grant, direct and nested, fifty-nine minutes later with no sign-in and no admin action.)

Say the window in the adopter's own numbers when they ask; "within an hour and a quarter by default, and you set the hour" is the honest sentence for the LDAP arm.

## The laws the sync obeys

- **Removals only on the directory's own answer.** A failed read of the mirror or of a group's members is *unknown*: the pass records the failure and removes nobody. An unreachable directory never offboards anyone by accident.
- **Nested membership is asked about as a person.** The directory's member list names direct members only; a person absent from it is checked against their own recursive memberships before removal, because a mapped parent group admits people through nested groups. That is why access never flaps between what a sign-in granted and what the sweep takes.
- **Accounts are never created by the sync.** A directory member with no Planton account is not materialized; accounts arrive at first sign-in, through the seat door. The sync only converges people who already exist.
- **Joining is a licensed capability; leaving is not.** An install whose license lost the SSO capability gains no new members through mappings, but every leave still applies. Offboarding never rides the entitlement.
- **The local administrator is outside all of this.** It is not a directory account and no pass touches it (`self-hosted.identity-primary-and-break-glass.md`).

## Sentences people will meet

**The email-less refusal.** A directory account with no email address cannot become a Planton account; the person is told at the door, before a seat is consumed:

> Your account has no email address, and Planton needs one to identify you. Ask your directory administrator to add an email address to your account, then sign in again.

**Invited directory users.** An organization can still invite a person by email. When that email belongs to the directory, the invitation page offers the directory's sign-in (with the manifest's `signInButtonLabel`), and accepting is signing in through it; the account that results is the directory's, correlated on the stable id, not a second local one.

**Sync health.** Each mapping records its last pass (`planton directory mappings`, the Directory page): direct members counted, joined, left, and any failure in words. Alerts go to the organization's manage-access holders on **transitions only** -- a standing failure is announced once, and recovery once -- by push and by email through the platform's `spec.email` provider. Alert failures never fail the pass; the recorded status is the durable truth.

## Deleting the manifest

Deleting the `PlantonIdentityProvider` removes the broker or the federation from the identity server: directory sign-in stops at once, the local administrator still signs in, and the mapping rules remain recorded. Re-applying the manifest restores provisioning and directory users sign in again. It is a mutation with an organization-wide blast radius (every directory user loses sign-in until it is back): say so in one sentence, get one clear yes.

## What was verified, and what to confirm on your install

Verified on running installs: the LDAP window end to end (removal from a group, from a nested group, and deletion from the directory); the brokered arm converging at sign-in against a live Entra tenant; the email-less refusal; an invited directory user accepting through the directory; alerts recorded on transitions and rendered in the console's Access History; the delete-and-re-apply of the manifest. **Not verified: the alert email arriving in a mailbox.** The lab install had no `spec.email` provider declared, so the email leg of the alert was never observed. On the adopter's install, declare `spec.email`, then force one transition (a mapping whose group is deleted in the directory) and confirm a manage-access holder receives the message. If the push alert is recorded and the email never arrives, treat it as a platform gap: file it on the open-source repository with the platform's `EMAIL` column, the Email settings page's verdict, and the alert's timestamp from Access History (`craft.filing-platform-gaps.md`).

## What never to do

- Never promise a window tighter than the arm supports: on the brokered arm a group change lands at the next sign-in and a removal at the session's end; on the LDAP arm it is the sync period plus fifteen minutes.
- Never remove a departed person from a mirrored team by hand to "speed it up"; the door is closed by design, and the directory is the place to act. Removing them from the organization (the Members page) is the immediate lever when one is truly needed.
- Never switch a live install between arms as a routine change. A person who first arrived through the brokered arm is a local realm user; after a switch to LDAP the directory's namesake cannot import until an administrator removes that twin. Plan an arm switch as its own journey, with the adopter's approval.
- Never read a `Failed` sync entry as "people were removed"; the law is the opposite -- a failure removes nobody, and the entry says what could not be read.
