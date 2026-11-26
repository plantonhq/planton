# A Shared VPC Service Project Reads Both Projects Through the Organization

**Date**: November 26, 2026
**Type**: Fix (permission data; no schema change)
**Components**: `GcpSharedVpcServiceProject`, the GCP permission reach rule (`pkg/iac/permissions`, `pkg/iac/actioninventory`), `PermissionReach`'s documentation

## Summary

**Attaching a service project to a Shared VPC host now works through a connection.**

Before attaching, Google reads both projects through Resource Manager, reads the host's attachments with `compute.projects.get`, and polls the host's operation. The kind declared those reads at `reach: target`, which a connection grants on its own project only. An attachment always joins two projects, so one of them was always out of reach: Google refused the deploy with "Required 'resourcemanager.projects.get' permission for 'projects/<the other project>'", while a check of the connection's own project found nothing missing.

The two groups, `ReadSharedVpcAttachment` and `ResolveAndPollHostProject`, now sit at `reach: organization`. That is the one grant that reaches both projects, and the kind already needs it for its `compute.organizations.*Xpn*` permissions.

## The reach rule's third case

`TestGcpPermissionReach` holds a group's reach to the narrowest scope Google's inventory gives each of its permissions. It allowed a higher reach in two cases:
- the shape lives above the project;
- Google authorizes the permission above a project only.

A third, machine-checked case joins them. `permissions.GcpKindsJoiningTwoProjects` lists the kinds whose one resource always joins two projects. Any group of such a kind may sit at `organization`.

It is a curated list, holding only kinds that join two projects by their nature. It does not hold a kind that merely may point at another project, such as a VPC peering, a DNS zone's peering network or an HA VPN peer. Each entry says why. Its one entry is `GcpSharedVpcServiceProject`.

`PermissionReach`'s comment names the three cases (comment-only proto change).

## Proven live

The kind ran on both engines in the test organization as an identity holding exactly its manifest:
- every group was granted on the organization and nothing on either project;
- each engine deployed and destroyed the attachment;
- the sweep read zero after.

The scenario no longer names both projects for the identity (`planton.dev/e2e-identity-project`), because the organization's grant reaches both.

## What changes for a connection

A GCP connection with Organization Access now carries `compute.projects.get` and `resourcemanager.projects.get` in its organization role. Both are read-only. Without Organization Access, the kind was already outside the grant, because its attachment permissions are organization-level.
