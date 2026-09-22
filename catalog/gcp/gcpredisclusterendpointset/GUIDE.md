# GcpRedisClusterEndpointSet Guide

The judgment this guide protects: this block is a registration, not a
network resource. The forwarding rules and addresses are built by their
own blocks in the consumer's VPC; this set tells the cluster they exist.
Get the count right, name everything by reference, and treat the list as
the whole truth.

## Why this is its own block

Google models user-created connections as a resource that names the
cluster's forwarding rules -- and every one of those rules targets a
service attachment the cluster publishes only once it exists. Folding the
registration into the cluster would make a block depend on its own
output. So the chain is three steps: the cluster (without `pscConfigs`),
then per consumer VPC one `GcpAddress` and one regional
`GcpGlobalForwardingRule` per attachment, then this set naming every
rule. Deploy in that order; destroy in reverse.

## One connection per attachment, per network

A cluster publishes a discovery attachment and a primary attachment, and
a reader attachment when it has replicas. Each `endpoints[]` entry is one
consumer VPC's group, and Google requires a connection in that group for
every attachment -- a group with only the discovery connection is
rejected at apply. The spec cannot count the cluster's attachments, so
the rule is Google's to enforce; the cluster's three `*_service_attachment`
outputs tell you which exist (the reader handle is empty without
replicas).

## Everything is a reference

A connection carries five identifiers and every one is an output of
another block: the rule's `self_link` and its `psc_connection_id` (the
same `GcpGlobalForwardingRule`, referenced twice because a reference
names one output path), the `GcpAddress`'s `address`, the
`GcpVpcNetwork`'s `network_id`, and the cluster's attachment handle. A
manifest that pastes a PSC connection id by hand will drift the first
time the rule is recreated; a manifest built from references never does.

## The list is the set

Google replaces the cluster's whole user-created endpoint list with this
manifest on every apply. Declare exactly one set per cluster -- two sets
overwrite each other -- and list every consumer network in it. Removing a
connection deregisters it and its forwarding rule stops working; delete
the rule in the same change. `deletionPolicy: ABANDON` leaves the
registration on the cluster when the block is removed from management.
