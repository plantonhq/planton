# GcpGcsBucketIamMember Guide

The judgment this guide protects: a bucket has two homes for grants, and picking the wrong one either breaks the chart or doubles a grant. The bucket's own `iam_members` is the default. This kind is for the one case `iam_members` cannot serve: a grantee that only exists because of the bucket.

## Which home a grant belongs in

Ask one question: does the grantee depend on the bucket?

- **No** — a workload's service account, a group, a service agent you can name up front. Put the grant in the bucket's `iam_members`. It travels with the bucket and needs no extra resource.
- **Yes** — a logging sink exporting into the bucket. The sink names the bucket as its destination, and Google mints the identity that must write there. A grant on the bucket that referenced that identity would make the bucket depend on the sink while the sink depends on the bucket. Use this kind.

Never declare the same (role, member) pair in both places. Both are additive, so they do not fight while both exist — but removing either one removes the grant, and the survivor silently stops working.

## The sink pattern, in order

1. The `GcpGcsBucket`.
2. The `GcpLoggingSink` whose destination is the bucket's `bucket_id`.
3. This grant: `bucket` from the bucket's `bucket_id`, `member` from the sink's `writer_identity`, role `roles/storage.objectCreator`.

Keep all three in one chart. Cloud Logging writes hourly batches, and a batch it cannot write is lost. `objectCreator` is write-only: the sink can add files and never read or overwrite what is there.

## Reference the identity, never copy it

A sink's writer identity is minted by Google, and recreating the sink can mint a new one. A `valueFrom` on `writer_identity` follows it; a pasted literal keeps granting the old identity while the new sink writes nothing.

## Conditions need uniform bucket-level access

Cloud Storage accepts conditional role bindings, but only on buckets with uniform bucket-level access. A condition is the way to scope a writer to one prefix (`resource.name.startsWith("projects/_/buckets/<bucket>/objects/<prefix>/")`) or to give it an expiry. The condition is part of the grant's identity, so the same role with and without a condition are two separate grants.

## Conventions and gotchas

- The bucket arrives by name, never a `gs://` URL. Bucket names are global, so there is no project field.
- Every field is immutable. Changing the role, member, or condition replaces the grant, and the member cannot write for the moment between delete and create.
- `allUsers` and `allAuthenticatedUsers` grant public access to every object the role covers. Buckets with public access prevention enforced refuse them, which is the right default.

## Pairs well with

- `GcpGcsBucket` — the destination; its `bucket_id` feeds `bucket`, and its own `iam_members` holds every grantee that does not depend on it.
- `GcpLoggingSink` — a log export into the bucket; its `writer_identity` feeds `member`.
- `GcpPubSubTopicIamMember` — the same pattern when the sink exports to a topic.
