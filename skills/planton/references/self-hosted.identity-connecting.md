---
title: Connecting a Company Directory — One Manifest, Two Arms, Verdicts in Words
description: How a self-hosted Planton is joined to a company directory through one PlantonIdentityProvider manifest -- the OIDC arm for Entra ID, Okta, and any OIDC provider, the Active Directory arm for LDAP over TLS -- what each field means, the verification checks the operator writes back and how to read a failed one, the dry-run that catches shape mistakes before anything reaches the cluster, and the sign-in that proves it. Read when a person asks to connect Entra ID, Active Directory, Okta, or "our SSO" to their Planton, when a Directory page shows a failed check, or when a first company sign-in does not land.
---

# Connecting a Company Directory

Read this when an adopter wants their people to sign in with the company account. The definition itself is documented field by field in `operator/api/v1/plantonidentityprovider_types.go` in the open-source repository, and the console's Settings → Directory page carries a guided setup that composes the manifest from a few answers (arm, issuer or servers, private CA, primary) and then checks the connection once the admin says it is applied. This file carries the reading order, the exact fields and defaults, the verdict sentences, and what to do at each one -- so you can compose the manifest by hand when the console is not in front of the person, and read the results either way.

## The doctrine in one sentence

A directory is connected by one `PlantonIdentityProvider` manifest with exactly one arm, the operator provisions the identity server from it and writes each verification verdict back onto the manifest in words, and a first sign-in is the only proof that matters.

## Which arm

- **`spec.oidc` -- brokering.** For Microsoft Entra ID, Okta, Google Workspace, Keycloak, any OIDC issuer. The company's own sign-in page authenticates; Planton receives a token with a stable subject and (when the provider sends them) the groups. Nothing is synced on a schedule; each sign-in is the truth.
- **`spec.activeDirectory` -- LDAP federation.** For on-premises Active Directory (or any LDAP directory) reachable from the cluster over LDAPS. The identity server authenticates against the directory and imports users and groups on a schedule.

An organization with Entra ID picks the OIDC arm even if it also runs on-premises AD: the company's own sign-in page does the authenticating, nothing runs on a schedule, and a removed person cannot get another token (`self-hosted.identity-offboarding.md` has both arms' windows). Exactly one arm per manifest, one manifest per platform.

## The manifest

Both arms share the head:

```yaml
apiVersion: planton.ai/v1
kind: PlantonIdentityProvider
metadata:
  name: corporate                   # any name; the Directory page shows it
  namespace: <the platform's namespace>
spec:
  platformRef: {name: <platform>}   # optional when exactly one platform lives in the namespace
  signInButtonLabel: "Sign in with Acme"
```

**OIDC arm (Entra ID shown)**:

```yaml
  oidc:
    issuerUrl: https://login.microsoftonline.com/<tenant-id>/v2.0
    clientId: <application (client) id>
    clientSecretRef: {name: entra-client, key: client-secret}
    scopes: [openid, profile, email]        # the default
    groupsClaim: groups                      # the default; where memberships arrive in the token
    subjectClaim: oid                        # Entra: oid, not the default sub
    primary: false                           # true sends every sign-in straight to the directory
```

Field truths worth saying out loud:

- `issuerUrl` must be the **single-tenant** form. The multi-tenant endpoints (`/common`, `/organizations`, `/consumers`) cannot satisfy strict issuer validation; the operator refuses them by name in the `issuerTenancy` check before touching the realm.
- `subjectClaim: oid` on Entra. Planton correlates accounts, sync, and offboarding on this claim; `sub` differs per app registration on Entra, `oid` does not.
- The app registration must send a `groups` claim in the ID token (Entra: Token configuration → Add groups claim, security groups, as group ID) for mapping to work; without it, sign-in works and every group mapping stays empty.
- "Who may sign in at all" is the tenant's decision, not a Planton feature: Entra's **Assignment required** on the enterprise application turns the app registration into the sign-in gate, and a person outside the assignment is refused by Microsoft before Planton sees a token. Planton's mappings decide what a signed-in person may do; the tenant decides who gets through.
- The redirect URI registered upstream is `<origin>/idp/realms/planton/broker/corp-directory/endpoint` -- the broker's alias is fixed, whatever the manifest is named; the Directory page's guided setup prints it.

**Active Directory arm**:

```yaml
  activeDirectory:
    servers: [ldaps://dc1.corp.acme.com:636]
    caBundleSecretRef: {name: ad-ca, key: ca.crt}     # when the directory's CA is private
    bindDn: CN=svc-planton,OU=Service Accounts,DC=corp,DC=acme,DC=com
    bindCredentialSecretRef: {name: ad-bind, key: password}
    usersDn: OU=Users,DC=corp,DC=acme,DC=com
    groupsDn: OU=Groups,DC=corp,DC=acme,DC=com
    syncPeriodMinutes: 60                              # the default; this plus Planton's 15-minute pass is the offboarding window
```

Defaults that fit Active Directory are already set (`userObjectClasses`, `usernameAttribute: sAMAccountName`, `emailAttribute: mail`, `groupNameAttribute: cn`, `groupMemberAttribute: member`, `nestedGroups: true`, `editMode: READ_ONLY`); set them only when the directory differs. `startTls: true` is for directories that upgrade a plain connection rather than listen on LDAPS. The bind account needs read on both DNs and nothing more.

## Applying it, and the dry-run first

The manifest's shape is validated by the cluster itself before anything runs. Compile it against the definition without creating it:

```bash
kubectl apply --dry-run=server -f identity.yaml
```

A field that does not exist, two arms at once, `startTls: true` on an `ldaps://` server, or a value outside its rule comes back as the API server's sentence (the definition carries these rules and they fire at admission); fix the manifest and try again. When the dry run passes, `kubectl apply -f identity.yaml` is a mutation under the protocol in `cloud.exploration.md`: what it changes (the identity server's realm gains a broker or a federation), one clear yes, then the read below.

The Secrets the manifest names must exist in the same namespace before the apply; a missing one is a `Refused` verdict, not a crash.

## Reading the verdicts

```bash
kubectl -n <ns> get plantonidentityprovider <name> -o yaml
```

`status.conditions` carries `Bound` (the manifest resolved its platform) and `Provisioned` (the identity server accepted the configuration). `status.verification.checks` carries named checks, each `Passed`, `Failed`, or `Unknown`, each with a sentence:

- OIDC: `issuerTenancy`, `issuer` ("issuer discovered; sign-in will authorize at …"), `clientAuthentication` (always `Unknown`: "the client id and secret are verifiable only at a real sign-in … the first sign-in attempt is the proof"), and `seededAdminCollision` when a declared administrator's email is also a local user.
- LDAP: `connection` ("the directory answered at …"), `bind` ("authenticated as …"), `usersSearch` ("N directory users are visible to the identity server (a added, b updated this sync)"), `groupsSearch` ("N directory groups mirrored under … (a added, b updated this sync)").

A `Failed` sentence names the field to look at (`searching the users DN failed -- check usersDn and the service account's read permissions: …`). An `Unknown` on the LDAP arm right after the apply ("the federation component has not been provisioned yet; the next reconcile pass verifies it") clears within a pass. The console's Settings → Directory page renders the same checks in its connection panel with a refresh control; when the person is on it, read that first.

The operator verifies the directory's certificate chain itself: a private CA the manifest names is trusted after verification, and when the directory re-issues its certificate the operator re-verifies and rolls the identity server to pick up the new trust -- a `PKIX path validation failed` in a check right after a directory certificate change resolves on its own within a pass or two.

## The proof: one sign-in

Open the console in a fresh browser context. On the OIDC arm with `primary: false`, the sign-in page shows the local form and a button carrying `signInButtonLabel`; on the LDAP arm the local form itself accepts directory credentials. Sign in as a real directory user and expect to land in the console with the person's name and email. That sign-in creates the person's Planton account (through the seat door, `self-hosted.first-admin-and-seats.md`) and, when groups arrive, joins them to any mapped teams (`self-hosted.identity-mapping.md`).

Verified on running installs: both arms end to end -- Entra ID as a live tenant and Samba Active Directory over LDAPS -- through the console, the CLI (`planton login`, both of its paths), and the MCP tools.

## Tenant-side facts not exercised

Three behaviors belong to the adopter's Entra tenant and could not be proven from a lab tenant without them:

- **Group-claim filtering ("groups assigned to the application")** -- a P1 feature; the lab sent all security groups.
- **Group overage** -- a user in more than roughly two hundred groups gets a `_claim_names` pointer instead of a `groups` array; Planton reads only the claim.
- **Conditional Access and tenant MFA policies** during a brokered sign-in.

On the adopter's install, confirm each with one real user: sign in, then read the person's teams on the Members page against their directory groups. If a mapped group does not join, or a sign-in the tenant allows does not land, treat it as a platform gap: file it on the open-source repository with the manifest (secrets removed), the tenant feature in use, and what the sign-in showed (`craft.filing-platform-gaps.md`).

## What never to do

- Never put the client secret or the bind password in the manifest; both arms take a `SecretKeyRef`, and a GitOps tree should never carry the value.
- Never "fix" a failed check in the identity server's admin console; the operator owns the broker and the federation component and converges them back.
- Never declare both arms, or a second manifest for the same platform, to "try the other one"; change the arm on the one manifest.
- Never move a live install to `primary: true` in the same apply that first connects the directory; prove the sign-in with the button first, then flip (`self-hosted.identity-primary-and-break-glass.md`).
