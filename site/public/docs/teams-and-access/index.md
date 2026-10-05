---
title: "Teams and Access"
description: "Manage organization membership, team structure, and role-based access control across your Planton organization"
icon: team
order: 80
tags:
  - Teams
  - Access Control
  - IAM
  - Permissions
---

# Teams and Access

Every Planton organization needs to answer three questions: who can access the platform, what can they do, and how do you manage that at scale? Teams and Access is where those answers live.

## The Problem This Solves

As organizations grow, managing individual permissions becomes unsustainable. Adding a new engineer means manually granting access to each resource. Offboarding someone means tracking down every permission they were given. Auditing who has access to production means checking one person at a time.

Planton provides organization-level membership with role-based access, team grouping for shared permissions, and a fine-grained authorization system built on [OpenFGA](https://openfga.dev/). You invite members, assign roles, organize them into teams, and the platform enforces access consistently across every resource, environment, and operation.

## Members

Members are the individual users in your organization. Every person who logs into Planton and accesses your organization is a member with an identity account.

### Inviting Members

New members join through an invitation workflow. You invite someone by email, optionally assigning one or more IAM roles at the time of invitation. The invitee receives an email with a unique invitation link.

The invitation follows a simple lifecycle:

- **Pending** — Invitation sent, waiting for the recipient to accept
- **Accepted** — Recipient clicked the link, created an account (if new to Planton), and joined the organization
- **Removed** — Invitation was revoked before acceptance

If the invitee already has a Planton account, they join the organization immediately upon accepting. If they are new to the platform, they create an account first and are then added to the organization with the roles specified in the invitation.

<!-- SCREENSHOT: Invite member modal
  Page: /orgs/{org}/settings/members
  Action: Click "Invite Member" button to show the invitation modal
  Focus: The modal showing email input and role checkboxes
  Alt: Invite Member modal with email field and role selection checkboxes
-->

### Managing Members

The Members page in Organization Settings shows all current members and pending invitations. From here you can:

- View all organization members and their roles
- Switch between the **Members** list and the **Invitations** list
- Copy an invitation link to share directly (useful if the email was not received)
- Remove a pending invitation before it is accepted

<!-- SCREENSHOT: Members list
  Page: /orgs/{org}/settings/members
  Action: Show the members list with at least 3 members visible
  Focus: The members table showing names, emails, and roles
  Alt: Organization members list showing team members with their assigned roles
-->

## Teams

Teams group members who share the same access needs. Instead of granting permissions to individuals one at a time, you create a team, add members, and grant permissions to the team. When someone joins or leaves a team, their access updates automatically.

### Team Structure

A team has a name, description, and a list of members. Members can be individual identity accounts or other teams — this nesting allows you to build hierarchies. For example, a "Platform Engineering" team could include the "Infrastructure" team and the "SRE" team as members, and any permissions granted to "Platform Engineering" would flow to members of both sub-teams.

### Creating and Managing Teams

Navigate to **Settings > Teams** in the web console to create and manage teams. Each team shows its members, description, and associated permissions.

<!-- SCREENSHOT: Teams list
  Page: /orgs/{org}/settings/teams
  Action: Show the teams list with at least 2 teams
  Focus: The teams table showing team names, descriptions, and member counts
  Alt: Organization teams list showing team names, descriptions, and number of members
-->

You can also list teams using the CLI:

```bash
# List all teams in the organization
planton get team
```

## Roles and Permissions

Planton uses a role-based access control system backed by [OpenFGA](https://openfga.dev/), a fine-grained authorization engine. Roles define what actions a principal (user or team) can perform on a specific type of resource.

### How Roles Work

Each IAM role specifies:

- **What kind of resource** it applies to (organization, environment, Infra Component, service, team, etc.)
- **What actions** it grants (such as creating resources, updating configurations, managing IAM policies, or viewing details)
- **What kind of principal** it is assigned to (user or organization)

Roles are not generic "admin" or "viewer" labels — they are scoped to specific resource types. A role granting full access to services does not automatically grant access to Infra Components.

### Assigning Roles

Roles are assigned through IAM policies that bind a principal (identity account or team) to a resource with a specific relation. For example, you can grant a team the "operator" relation on a specific environment, which gives team members operational access to all resources within that environment.

Roles can be assigned:

- At invitation time, when you specify which roles a new member receives
- After the fact, by managing IAM policies through the CLI or web console
- Indirectly through team membership — if a team has a role on a resource, all team members inherit that access

### Managing IAM Policies with the CLI

```bash
# Add an IAM policy — grant a role on a resource to a principal
planton iam iam-policy add \
  --resource-kind organization \
  --resource-id org-acme \
  --principal-id ia-usr-alice \
  --role operator

# View IAM policies for a resource
planton iam iam-policy get \
  --resource-kind environment \
  --resource-id env-production

# View policies grouped by role
planton iam iam-policy get \
  --resource-kind environment \
  --resource-id env-production \
  --group-by-role

# Remove an IAM policy
planton iam iam-policy remove \
  --resource-kind organization \
  --resource-id org-acme \
  --principal-id ia-usr-alice \
  --role operator

# List all available IAM roles
planton iam role list
```

## API Keys

For automation and CI/CD integration, Planton supports API keys that authenticate non-interactive access. API keys are scoped to the user who created them and carry that user's permissions.

```bash
# Create a new API key
planton api-key new --name "ci-pipeline"

# List existing API keys
planton api-key list
```

API keys display their fingerprint, creation date, last-used date, and expiration status. Keys can be set to never expire or to expire on a specific date.

## CLI Reference

| Command | Description |
|---------|-------------|
| `planton iam invite <email>` | Invite a member to the organization by email |
| `planton iam lookup-invitations` | Look up invitations by email |
| `planton iam remove-invitation` | Remove a pending invitation |
| `planton iam iam-policy add` | Grant a role on a resource to a principal |
| `planton iam iam-policy get` | View IAM policies for a resource |
| `planton iam iam-policy remove` | Remove a role binding |
| `planton iam role list` | List all available IAM roles |
| `planton api-key new` | Create a new API key |
| `planton api-key list` | List existing API keys |
| `planton get team` | List all teams in the organization |

## Related Documentation

- [Billing](/docs/teams-and-access/billing) — Subscription plans and billing management
- [Authentication and Authorization](/docs/security/authentication-and-authorization) — How the permission model works under the hood
- [Security Overview](/docs/security) — Platform security architecture
- [Connections](/docs/connections) — Credential and integration management
- [Runner Security Model](/docs/runner/security-model) — How credentials are isolated in customer infrastructure
- [Platform Overview](/docs/platform) — Organization structure and core concepts
