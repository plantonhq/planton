# GcpSharedVpcServiceProject Guide

The judgment this guide protects: attaching a project to a Shared VPC
host is a declaration of membership, not of permission. It says "this
project may place workloads in the host's networks"; who may actually do
so is IAM, declared separately.

## Reference the host kind, not the project

`hostProjectId` points at a `GcpSharedVpcHost` (its `host_project_id`
output), not at the host's `GcpProject`. The difference is ordering: a
chart that references the host kind enables the project as a host
before any attachment is tried, and detaches every service project
before the host is disabled -- both of which Google requires. A literal
host project ID works when the host was enabled elsewhere.

## One attachment per project, one host per project

A project attaches to at most one host, and a host cannot itself be a
service project. Declare one `GcpSharedVpcServiceProject` per
application project; a chart for a new environment usually carries the
project, its attachment, and its `networkUser` grants together.

## The grants that make it useful

Attaching gives the service project visibility of the host's networks
and nothing more. Before a VM, node pool, or Cloud SQL instance can land
in a host subnetwork, `roles/compute.networkUser` is needed on that
subnetwork (or on the whole host project) for:

- the identity that deploys into the service project;
- Google's service agents for the products in play --
  `service-{number}@container-engine-robot.iam.gserviceaccount.com`
  for GKE, `service-{number}@compute-system.iam.gserviceaccount.com`
  for Compute, the Cloud SQL and Serverless VPC Access agents likewise.

Declare those with `GcpProjectIamMember` on the host project (all
subnetworks) or subnetwork-level IAM (one subnetwork). They are per
team and per subnetwork, with their own lifecycle, which is why they
are not folded into this block.

## Detaching has a precondition

Google refuses to detach a service project while any of its resources
still uses a host subnetwork. A chart that destroys the workloads first
(they reference the subnetworks) detaches cleanly. For a project whose
workloads must survive the chart -- a platform project handed to another
team -- set `deletionPolicy: ABANDON`; it is the only alternative this
resource accepts, and it leaves the attachment live.

## The permission is organization-level

Like the host, attaching needs `roles/compute.xpnAdmin` on the
organization or a folder above BOTH projects. The connection that
deploys this block is the network administrator's, not the application
team's.
