# Dev single node preset

The fastest declarable RabbitMQ: one node on ephemeral storage
(emptyDir — no PVC, no StorageClass involved), 1Gi of memory with
requests equal to limits, and a 60-second termination grace period in
place of the operator's 7-day production default. For developers who
need a real AMQP endpoint — wiring a task queue, testing a consumer,
prototyping an RPC flow — without any production ceremony.

Know what ephemeral means here: every queue, message and user
vanishes with each pod restart. That disposability is the point — the
cluster tears down and comes back in seconds — but nothing real can
depend on it. Credentials still work the production way: the operator
generates them into the `dev-rabbitmq-default-user` Secret (keys
username, password, host, port, connection_string, ...), exported in
the outputs. In-cluster clients connect at the exported
`amqp_endpoint` (port 5672); the management UI is a
`port_forward_command` away on 15672.

Change first: drop `ephemeral` and declare `disk_size` the moment
anything should survive a restart — the two are mutually exclusive,
so it is a swap, not an addition. Then grow `replicas` to an odd
count (3, 5, 7) when availability starts to matter, keeping in mind
the one-way door: the operator does not support scaling down.

See [01-dev-single-node.yaml](./01-dev-single-node.yaml) for the
manifest.
