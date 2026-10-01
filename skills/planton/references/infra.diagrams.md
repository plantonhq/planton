---
title: How the Platform Draws What You Author
description: The picture every Planton surface draws from a chart or a project -- the account rooms, the rooms and what nests in them, the lines, and the metadata.group trays -- which of it the platform decides on its own and the three authoring choices that genuinely change it (a reference or a literal, runs_on or another relationship, a dedicated component or a buried flag), plus how to use metadata.group well. Read when writing or reviewing manifests and you want the architecture view to read like a reference diagram, when someone asks why a resource draws where it does, when choosing between valueFrom and a literal or between relationship types, or before adding metadata.group.
---

# How the Platform Draws What You Author

Every chart, project, pipeline, and service is drawn as an architecture
diagram: on the studio canvas while you write, on a project's page, on a
pipeline run, and on the organization's Infrastructure Map. Almost all of
that picture is decided by the platform from facts it already has. Author
for correctness first; the picture follows. This file says what the
platform draws on its own, the few choices that change the picture, and how
to use `metadata.group`.

## What the platform draws on its own

- **Account rooms.** The outermost box is the cloud account each resource
  lands in, resolved exactly as a deploy resolves it: the connection a
  deployed resource landed through, else the `planton.dev/connection`
  annotation on the manifest, else the environment's default connection for
  the provider, else the organization's. A chart spanning an AWS account, a
  Google Cloud project, and an Azure subscription draws three rooms side by
  side. When nothing names the connection yet (a chart template with no
  annotation), the room says the provider ("AWS") and the account is decided
  at deploy. A Kubernetes resource lands in the cloud account behind its
  cluster. Never restate the account anywhere else.
- **Rooms and what nests in them.** A kind the catalog marks as a place is
  drawn as a room when something sits in it. The places are the things
  others are created INSIDE: networks and subnets, clusters, namespaces,
  DNS zones, Google Cloud projects and Azure resource groups, key rings,
  and servers that hold databases, topics, or file systems. Compute never
  is: a node group, a node pool, or a VM is a card, and what runs on it
  draws wherever it sits. A resource sits in the room its
  reference PLACES it in: a subnet's `vpcId` puts it in the VPC; a node
  group's subnets put it in their shared VPC. A reference that only grants
  access (a function allowed into a subnet, a controller writing to a zone)
  draws a line and never nests -- the catalog marks those fields, you do
  not.
- **Namespaces a component creates.** A Kubernetes component whose
  `namespace` field holds a literal draws inside a namespace room even when
  no namespace resource exists in the chart; the platform draws the room and
  explains it on hover.
- **Kubernetes resources inside their cluster.** A resource deployed
  through a Kubernetes connection the platform published for a cluster
  draws inside that cluster, with no relationship authored (see
  `kubernetes-on-cluster.md` for the wiring itself).
- **Lines.** Every `valueFrom` draws a line and orders the deploy. So does
  every `metadata.relationships` entry. A literal value in a reference field
  that names a resource of the field's kind in the same set also draws a
  line. A small attached kind -- an address, a certificate, a DNS record,
  a key pair, an IAM binding -- hangs as a plate on the one thing it names
  in its room instead of standing as a card. Only resources in the set
  count: a binding wired to a service account and to a role in the same
  set names two things and stays a card with two lines; the same binding
  with the role as a literal names one and hangs on the service account.
- **External things.** A reference to a resource another chart deployed
  draws as an external card: the chart is not the owner, and the picture
  says so.

## The three choices that change the picture

1. **A reference or a literal.** `valueFrom` carries the value, the deploy
   order, and the line. A hardcoded id or ARN carries none of that: the
   picture shows two unrelated cards, and the deploy may run them in the
   wrong order. Wire it (`dependencies.md`).
2. **`runs_on` or another relationship.** Among relationships only
   `runs_on` can PLACE: a resource that `runs_on` a room (a cluster, say)
   draws inside it. `depends_on`, `uses`, and `managed_by` order the deploy
   and draw a line, never a nesting. Say `runs_on` when the resource truly
   runs on the target; never use it to push a card into a box. A
   `runs_on` to machines (a node group or node pool) places the resource
   wherever those machines sit -- their cluster. A Kubernetes workload
   already draws inside its cluster through its connection, so its
   `runs_on` to the cluster's node group agrees with the connection and
   adds the deploy order.
3. **A dedicated component or a buried flag.** A resource authored as its
   own component is a card someone can see, click, and deploy on its own; a
   capability buried as a flag or an inline block inside another resource
   draws nothing. When the picture should show it, author it as its own
   component -- when the catalog offers one.

## `metadata.group` -- trays, used well

A group is the author's statement of CONCERN, the one thing the platform
cannot infer. It never changes the deploy order.

- **Shape.** A slash path, `layer/concern`: `infrastructure/networking`,
  `infrastructure/iam`, `platform/certificates`, `platform/dns`,
  `apps/storefront`. One vocabulary per chart, used on every resource that
  shares the concern.
- **Where it draws.** As a tray INSIDE the room its members live in -- a
  member keeps its room and joins the tray there. Inside a cluster, the
  certificates grouped `platform/certificates` draw in a "Certificates" tray
  within the cluster.
- **What it will not draw.** A tray is skipped where a tray of the same
  top-level family already holds the room (a cluster grouped
  `infrastructure/cluster` inside a VPC grouped `infrastructure/networking`
  stands straight in the VPC), and a tray that would hold fewer than two
  things is not drawn at all: a group of one is not a group.
- **When it earns its place.** When a room (an account, a VPC, a cluster)
  holds six or more cards of mixed concerns -- identities beside storage
  beside network pieces -- grouping them turns a pile into a reading order.
  In a small chart, leave it out.
- **Never** use a group to name an account, an environment, or a network:
  the platform draws those truthfully, and a group that restates them will
  drift from them.

## Checking the picture

No CLI command prints a diagram's structure. On a Planton surface the studio
draws the chart live from the folder as you write, and a project's or
chart's page shows its architecture view after a build. Read it the way a
reviewer would: each resource in the account and room it should be in,
every line meaning a real reference, no card standing alone that should be
wired. When a resource draws somewhere surprising, the cause is almost
always a literal where a reference belongs, a relationship type that does
not mean what was intended, or a missing connection annotation.
