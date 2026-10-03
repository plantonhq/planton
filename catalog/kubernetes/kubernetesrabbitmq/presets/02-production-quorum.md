# Production quorum preset

A production RabbitMQ: 3 nodes (the quorum posture — quorum queues
and the Raft-based metadata store survive one node loss), a 50Gi data
volume per node, 4Gi of memory with requests equal to limits (the
memory-high-watermark rule), required anti-affinity so a node loss
takes one broker instead of the quorum, and automatic feature-flag
enablement so upgrades never stall on manual flag management.

Know the two quorum facts that shape this manifest. First, the count
is ODD on purpose: 3, 5 or 7 nodes tolerate a minority loss, while a
2-node cluster loses availability when either node fails — even
counts buy cost without buying availability. Second, replicas are a
one-way door: the operator does not support scaling down, because
removed brokers strand their queue replicas — a cluster that turns
out oversized is migrated to a smaller one, not shrunk. Grow in odd
steps, deliberately. The 7-day termination grace period (the operator
default, kept here) is part of the same posture: a draining node gets
however long it needs to hand off cleanly rather than lose messages.

Change first: `disk_size`, from your actual retention — queues,
quorum-queue Raft logs and the metadata store all live on it, and
PVCs cannot shrink. Then memory, keeping requests and limits EQUAL —
RabbitMQ reads its memory high watermark from the limit, so a gap
triggers flow control at the wrong threshold. Add `tls` (a
KubernetesCertificate's secret output plugs in directly) before
anything crosses a trust boundary; applications read credentials from
the operator-generated `prod-rabbitmq-default-user` Secret exported
in the outputs — never from a manifest.

See [02-production-quorum.yaml](./02-production-quorum.yaml) for the
manifest.
