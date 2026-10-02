# GcpPscServiceAttachment Guide

The judgment this guide protects: a service attachment is a door from
other people's networks into yours. Who may walk through it, how many at a
time, and what happens when you close it are the decisions that matter;
the rest is plumbing.

## Producer and consumer are two kinds

Private Service Connect has two halves. The PRODUCER publishes an internal
load balancer as a service attachment -- this kind. Each CONSUMER creates a
PSC endpoint in its own VPC: a regional `GcpGlobalForwardingRule` with an
empty `loadBalancingScheme` whose `target` is the attachment's `self_link`
(or its PSC id when the consumer is in another organization). Neither side
sees the other's address space, nothing is peered, and ranges may overlap
freely. The attachment's region must match the producer load balancer's;
the consumer endpoint lives in the consumer's region.

## Manual acceptance is the default posture

`ACCEPT_AUTOMATIC` lets any project in Google Cloud connect, with only the
reject list standing in the way. `ACCEPT_MANUAL` admits only the projects,
networks, or endpoints in `consumerAcceptLists`, each with a
`connectionLimit`; everyone else's endpoint stays PENDING until they are
listed. Start manual. A consumer removed from the accept list keeps its
existing connection unless `reconcileConnections` is true -- decide that
consciously, because true means a list edit can cut a running consumer off.

## NAT subnets are capacity

Consumer traffic is translated into `natSubnets` -- subnets of purpose
`PRIVATE_SERVICE_CONNECT` in the producer VPC and region, used for nothing
else. Every connected endpoint consumes NAT addresses from them, so the
subnet size caps how many consumers you can serve; add subnets to grow
capacity (the list is a set, and order never diffs). Google reserves the
PSC subnet at creation time, so declare the purpose on the `GcpSubnetwork`
when it is created.

## PROXY protocol changes what the backends see

Behind an attachment, backends see the NAT address, not the consumer's.
`enableProxyProtocol: true` prepends the consumer's original address data
to each TCP connection -- but only backends that speak the PROXY protocol
can read it; others break. It is a producer-and-backend decision, made
once per attachment.

## Closing the door

`deletionPolicy` unset means DELETE: the attachment goes and every consumer
endpoint connected through it loses the service at once. `PREVENT` makes
destroy fail, the guard for a service other teams depend on; `ABANDON`
leaves the attachment serving but unmanaged. Removing a consumer from the
accept list (with reconciliation on) is the surgical alternative to
deleting the attachment.
