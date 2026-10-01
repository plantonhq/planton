# GcpPubSubTopicIamMember Guide

The judgment this guide protects: a Pub/Sub export is two resources, not one. The sink (or the Security Command Center notification config) names the topic, and the topic must let that sender publish. Leave the second half out and nothing fails loudly — the sender simply drops what it cannot deliver.

## Why the grant is its own resource

The identity that needs publish rights is created by the sender, and the sender names the topic. If the topic carried the grant and referenced that identity, the topic would depend on the sink while the sink depends on the topic. No chart can deploy that. A standalone grant depends on both, so the order is always:

1. The `GcpPubSubTopic`.
2. The `GcpLoggingSink` (or `GcpSccNotificationConfig`) whose destination is the topic's `topic_id`.
3. This grant: `topic` from the topic's `topic_id`, `member` from the sink's `writer_identity` (or the notification config's `service_account_member`), role `roles/pubsub.publisher`.

Keep all three in one chart. The grant then lands in the same deploy as the sink, and the export works from its first entry.

## Grant on the topic, not the project

A project-level `roles/pubsub.publisher` lets the sender publish to every topic in the project. The sink needs exactly one. The topic grant is the least-privilege shape, and it is the one that shows up as an edge in the graph instead of hiding in project IAM.

## Reference the identity, never copy it

A sink's writer identity is minted by Google. Recreating the sink (a rename, a scope change) can mint a new one. A `valueFrom` on `writer_identity` follows the new identity automatically; a pasted literal keeps granting the old one while the new sink drops entries. Use literals only for identities you own, like a group or a known service agent.

## No conditions on topics

Pub/Sub is not among the services Google lists as accepting conditional role bindings, so the grant has no `condition` field (the provider's schema carries one; the parity manifest records why it is excluded). Scope access by choosing the topic and the role instead, or put a conditioned grant on the project when time-boxed access is truly required.

## Conventions and gotchas

- The topic arrives by full name (`projects/<project>/topics/<topic>`); the project is read from it, so a topic in another project is granted the same way.
- `roles/pubsub.subscriber` on a topic lets the member attach subscriptions to it, which is how a consumer in another project reads a stream it does not own.
- Every field is immutable. Changing the role or member replaces the grant, and the sender cannot publish for the moment between delete and create.
- Never grant `allUsers` or `allAuthenticatedUsers` on a topic. The member-format rule admits them so the choice is always explicit and reviewable.

## Pairs well with

- `GcpPubSubTopic` — the destination; its `topic_id` feeds `topic`.
- `GcpLoggingSink` — a log export to the topic; its `writer_identity` feeds `member`.
- `GcpSccNotificationConfig` — findings notifications; its `service_account_member` feeds `member`.
- `GcpGcsBucketIamMember` — the same pattern when the sink exports to a bucket.
