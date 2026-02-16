---
title: "Project User Member Assignment"
description: "This preset assigns the `member` role to a user on a project. The member role is the standard privilege level for users who need to create and manage cloud resources (instances, volumes, networks)..."
type: "preset"
rank: "01"
presetSlug: "01-project-user-member"
componentSlug: "role-assignment"
componentTitle: "Role Assignment"
provider: "openstack"
icon: "package"
order: 1
---

# Project User Member Assignment

This preset assigns the `member` role to a user on a project. The member role is the standard privilege level for users who need to create and manage cloud resources (instances, volumes, networks) within a project. This is the most common role assignment in OpenStack.

## When to Use

- Granting a user standard access to a project for day-to-day cloud operations
- Onboarding new team members to an existing project
- Automating user provisioning as part of project setup

## Key Configuration Choices

- **Member role** -- standard operational access (not admin, not reader)
- **Project scope** (`projectId`) -- the role applies only within the specified project
- **User principal** (`userId`) -- role is assigned to an individual user (use `groupId` instead for group-based assignment)
- **ForceNew** -- all fields are immutable; changing any field recreates the assignment

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<member-role-id>` | UUID of the `member` role | `openstack role list` or Keystone API |
| `<project-id>` | ID of the project to grant access to | OpenStack console or `OpenStackProject` status outputs |
| `<user-id>` | UUID of the user receiving the role | `openstack user list` or Keystone API |
