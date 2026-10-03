# Kubernetes NATS

## When NOT to Use This

**One resource is ONE NATS system** — the lightweight, high-speed
messaging server (pub/sub, request/reply, queue groups) with JetStream
persistence (streams, consumers, key-value and object stores) from the
official `nats` Helm chart.

Not the right kind when:

- **You want to declare the streams it carries** — streams, consumers
  and KV buckets are data-plane objects with their own lifecycle:
  create them from applications (any NATS SDK), the `nats` CLI, or the
  bundled nats-box pod. This kind deploys the SERVER, never your
  streams.
- **You need the Kafka ecosystem** — partitioned logs with consumer
  groups, Connect pipelines and schema registries are
  `KubernetesKafka` territory. JetStream persists and replays
  messages, but it is not a drop-in Kafka.
- **AMQP routing topologies** — exchanges, bindings and per-queue
  dead-lettering are `KubernetesRabbitMq` territory.

## The auth model and where passwords live

With `auth` unset the server accepts unauthenticated connections —
fine inside a trusted cluster network, never for anything reachable
from outside. Declare flat `users` or multi-tenant `accounts` (never
both — enforced on the spec) and the module GENERATES every password
into the `<name>-auth` Secret, one key per username. The server reads
each password from a Secret-backed environment variable, so no
credential ever appears in the rendered config or Helm values; point
workloads' secretKeyRefs at the same Secret.

## JetStream is on by default

Each server gets a persistent file-store volume, so published messages
survive pod restarts — this kind's posture is persistent messaging
(the chart's raw default is off; disable explicitly for a pure
fire-and-forget broker). A single server is a complete JetStream
deployment for dev; REPLICATED (R3) streams need `cluster` enabled
with at least 3 servers — replicas can never exceed the server count.

## Exposure

The client Service stays ClusterIP by default; in-cluster clients
connect through the exported `client_endpoint`. For external clients,
set `service` to LoadBalancer with your cloud's annotations, or enable
`websocket` behind first-class exposure kinds. Nothing in this kind
does ingress.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
