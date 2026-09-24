# GcpManagedKafkaAcl Guide

The judgment this guide protects: Kafka access is granted per resource pattern, and each pattern has exactly one ACL on a cluster. Give every pattern one owner -- usually the team that owns the topic or the consuming service -- and keep the grants next to the thing they protect.

## Two layers of access

Google IAM decides who may connect: a client needs `roles/managedkafka.client` on the project. Kafka ACLs then decide what a connected client may do. Both are needed; an ACL alone does not let an identity connect, and IAM alone does not let it read a topic once ACLs are in use.

## Patterns

`aclId` names the pattern and becomes the ACL's id: `cluster` for cluster-wide operations, `topic/{name}`, `consumerGroup/{name}`, or `transactionalId/{name}` for one resource, and `topicPrefixed/{prefix}` (and the group and transactional-id forms) for every resource whose name starts with the prefix. `{name}` may be `*`. Because a cluster holds one ACL per pattern, two manifests declaring the same `aclId` would overwrite each other -- split grants by pattern, not by team.

## Entries

Each entry is a principal, an operation, and ALLOW (the default) or DENY; a DENY wins over any ALLOW for the same principal and operation. Principals are `User:` followed by a Google account (`User:orders-api@project.iam.gserviceaccount.com`) or, with mTLS, the short name the cluster's principal mapping rules produce; `User:*` is everyone. The common grants: a producer needs WRITE and DESCRIBE on its topic, a consumer READ on the topic and READ on its consumer group, a transactional producer WRITE on its transactional id. An ACL holds at most 100 entries.
