# KubernetesNeo4j Guide

The judgment this guide carries: the two things a composer gets wrong
with Neo4j are the credential (declaring or escrowing a password the
module would have minted) and the size (shrinking below the chart's own
floor, where it refuses to install at all).

## Leave the credential to the module

Declare no `auth` and the module generates the `neo4j` admin password,
materializes it as the `<name>-auth` Secret before the Helm release
(the chart looks that Secret up at template time and fails without
it), and reports it two ways: `auth_secret_name` for the chart's own
`NEO4J_AUTH: neo4j/<password>` pair and `password_secret` for the bare
password a workload's driver wants. A manifest therefore carries no
secret and no placeholder, nobody escrows a value, and the credential is
stable across re-applies. Declare `auth.password` only to bring your
own — and then as a managed-secret reference, never a literal in a
manifest that lives in git. Declare `auth.existing_secret` when a Secret
you own already carries the chart's pair; the module then creates
nothing and leaves `password_secret` unset, because that Secret's layout
is yours, not its. What breaks if you choose wrong: a literal password
in a manifest is a leaked credential the first time the file is shared,
and a hand-minted one is a value somebody has to keep somewhere forever.

## Do not shrink below the chart's floor

The official chart rejects any install under 500m CPU and 2Gi memory;
the smallest Neo4j is the chart's smallest, not this catalog's. Size the
container above the page cache plus heap you tune in `memory`, and keep
the page cache large relative to heap — traversals live or die on graph
data staying in memory.

## Community is one server; enterprise is the cluster

`cluster_name` is refused on the community edition by the spec itself
(single-instance by license). Enterprise members sharing a
`cluster_name` form one cluster and are each their own release; the
bolt endpoint every driver connects to is per member.

## Diagram

A Neo4j server renders as one node in its namespace with its Secret
beside it; the Secret is the module's, so a consumer that reads
`password_secret` draws its edge to the Neo4j node, never to a Secret
somebody declared by hand.
