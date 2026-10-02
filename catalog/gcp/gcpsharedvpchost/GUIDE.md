# GcpSharedVpcHost Guide

The judgment this guide protects: one project owns the networks, many
projects use them. The host is that one project, and enabling it is a
separate act from creating it -- declared once, referenced by every
attachment.

## Why a separate block

A project becomes a host by a single API call (`enableXpnHost`) that
can happen long after the project exists and must be undone before it
is deleted. Modeling it as a flag on `GcpProject` would tie the host
role to the project's lifecycle and put an organization-level
permission (`roles/compute.xpnAdmin`) on every project create. As its
own block, the host is declared once, and every
`GcpSharedVpcServiceProject` references it -- which is exactly the
order Google needs on create and on destroy.

## The permission is organization-level

`roles/compute.xpnAdmin` on the project is not enough; Google checks it
on the organization (or a folder above the project). The deploying
identity for this block is therefore the organization's network
administrator, not an application team's project deployer. Plan the
connection accordingly.

## Leave the project empty for the everyday case

An empty `projectId` enables the project the credentials are configured
for -- the same "ambient project" every GCP block honors -- so a chart
that deploys into the network project needs no configuration here.
Name a project (by literal or `GcpProject` reference) when the host is
somewhere else.

## Destroy in order

Google refuses to disable a host while any service project is
attached. A chart whose attachments reference the host through
`status.outputs.host_project_id` destroys them first automatically; a
hand-managed estate must detach by hand. `deletionPolicy: PREVENT` is
the right guard for a host the whole organization depends on.

## The attachment is only half the story

Enabling the host and attaching service projects gives the service
projects permission to SEE the host's networks. Placing a VM, a GKE node
pool, or a Cloud SQL instance in a host subnetwork also needs
`roles/compute.networkUser` for the deployer and for Google's service
agents (`service-{number}@container-engine-robot.iam.gserviceaccount.com`
and friends) on that subnetwork or on the host project. Declare those
grants with `GcpProjectIamMember` on the host project, or subnetwork
IAM, per team -- they have their own lifecycle and are not this block's.
