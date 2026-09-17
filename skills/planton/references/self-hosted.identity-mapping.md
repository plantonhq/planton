---
title: Mapping Directory Groups to Roles — Joining a Group IS Getting the Access
description: How access is granted from the company directory on a self-hosted Planton: a mapping rule joins a directory group (by its stable id) to a role at a scope, a mirrored team materializes for it, and membership follows the directory at sign-in and on the sync. The preview that answers a group's blast radius before the rule exists, the three interfaces (the Directory page, planton directory, the MCP tools), what sync health records, and why the mirrored team is read-only. Read when a person asks to give a directory group access, asks why someone is or is not in a team, or wants to know how many people a rule reaches.
---

# Mapping Directory Groups to Roles

Read this when an adopter has connected their directory and asks "how do our engineers get access?". The command's own summary says it: *map your company directory's groups to roles -- joining a group IS getting the access*. This file carries how a rule works, how to preview it, the three doors, and the reading of sync health.

## The doctrine in one sentence

A mapping rule binds one directory group to one role at one scope; Planton mirrors the group as a team it owns, joins people to it from their directory groups at sign-in and on the sync, and the access lives and dies with the group.

## How a rule works

- **The group is named by its stable id**, not its display name. Renaming a group in the directory keeps the rule and rewrites the mirrored team's name and description; the rule never re-targets by accident.
- **The role and scope are Planton's**: a role such as `developer` at the organization, or an environment-scoped role (`release-approver`) at one environment -- the same vocabulary `planton grant <role>` uses, and `planton grant --help` lists the ambient-scope rule.
- **A mirrored team materializes** for the rule, born empty. Members arrive only two ways: a person's sign-in (their groups arrive in the token on the OIDC arm, or from the federation on the LDAP arm) and the directory sync pass. Nobody can add or remove a member of a mirrored team by hand; the team's origin is the directory and the console says so.
- **Nested groups count.** A person who holds the mapped group through a nested group (platform-eng inside engineering) is joined, and stays joined on later sync passes -- the sync consults the person's own recursive memberships before removing anyone, the same view the sign-in used to join them.
- **The access dies with the rule.** Deleting a mapping removes the role binding the mirrored team carried; the people keep their accounts.

## Preview first

Mapping is a mutation with a blast radius, and every door previews it:

```bash
planton directory groups [search]        # browse the groups as the identity server mirrors them
planton directory preview <group>        # how many people the rule reaches, before it exists
planton directory map <group> --role developer [--reason "..."]    # confirms interactively; --yes non-interactively
planton directory mappings               # this scope's rules with their recorded sync health
planton directory unmap <mapping-id>     # the access it granted dies with it
```

The scope is the ambient one: the organization from the CLI context (`planton context get`), or the environment when the context carries one or `--env` names one. `preview` is read-only and free to run; `map` and `unmap` follow the protocol in `cloud.exploration.md` -- the preview is the blast-radius sentence, the confirmation is the yes.

The preview has three honest answers. On the LDAP arm, a count of the group's direct members and how many already hold accounts (members without one gain the access at their first sign-in). On the brokered arm: *"This group isn't in the identity server's directory mirror, so member counts can't be known ahead of sign-ins. That is the normal state for a cloud-brokered directory; the mapping takes effect as members sign in."* And for a parent group, its direct members -- nested holders are counted from their own sign-in, not the preview.

The MCP tools are the same verbs under the same law: `find_directory_groups`, `preview_directory_group`, `create_directory_group_mapping` (its text says to preview first), `list_directory_group_mappings`, `get_directory_group_mapping`, `update_directory_group_mapping`, `delete_directory_group_mapping` (requires `confirm: true` as a deliberate act; without it the tool refuses: *"retracting a mapping is irreversible; pass 'confirm: true' to proceed"*).

Creating a mapping needs the install's plan to carry the SSO capability; the server names the doorway when it is missing. Reads and deletes are never gated on it, so a lapsed license still lets an administrator see and retract rules -- offboarding never rides the entitlement.

The console's Settings → Directory page is the third door: the connection panel, the mapping list with sync health, a create-mapping sheet with the blast radius shown before the confirm, and a delete dialog that names what the rule granted.

## Reading sync health

Each rule records what its last sync pass did, in a sentence: how many direct members the group holds in the directory, how many were joined or left, and when. The Directory page shows it as a chip -- `Awaiting first sync` (the periodic pass converges it next), `Synced`, `Grants at sign-in` (the brokered arm's normal state: nothing to mirror, membership arrives with each token), or `Sync failed` (the sentence says what could not be read; nobody was removed). Read it before answering "why is X not in the team":

- **Joined at sign-in, present in the team** -- the ordinary case.
- **In the directory group but not in the team** -- the person has not signed in since the rule was created (accounts materialize at first sign-in; the sync joins only people who already have an account), or on the OIDC arm the token carried no `groups` claim for them (`self-hosted.identity-connecting.md`, the app registration's groups claim).
- **Was in the team, now gone** -- the person left the group in the directory (or a nested group that held it), and the sync or their next sign-in removed them (`self-hosted.identity-offboarding.md` has the timing per arm).
- **Renders as an account id rather than a name** -- a display index lag; a directory-born member's name and email appear within moments on current releases (older ones waited for a periodic index pass, up to four hours).

## What was verified

On running installs, both arms: mapping through all three doors; the blast-radius preview; nested-group membership joining at sign-in and surviving the sync; a directory rename rewriting the mirrored team's name and description; the delete cascade removing the binding and keeping the accounts.

## What never to do

- Never grant a directory user a role directly (`planton grant`) as a substitute for a mapping when the adopter's intent is "this group gets access"; the direct grant does not follow the directory, and the two paths reading differently is the confusion you were asked to prevent. Direct grants are for individuals and service accounts.
- Never try to add or remove a member of a mirrored team by hand; the door is closed by design, and the change belongs in the directory.
- Never map a group without running the preview, and never present the preview's count as the number of accounts that will appear today -- it is the number of people the rule reaches when they sign in.
- Never map the local administrator's group or the break-glass account; that account stays local (`self-hosted.identity-primary-and-break-glass.md`).
