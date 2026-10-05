# CI standard preset

The full Tekton control plane — Pipelines, Triggers, Dashboard and
Chains — with the one piece of configuration no production cluster
should skip: a pruner. Completed runs keep their pods around until
something cleans them up; this preset prunes daily, keeping the newest
100 of each kind (their durable history belongs in your logging and
attestation stores, not in etcd).

Requires a KubernetesTektonOperator on the cluster first — this
resource is the declaration that operator reconciles, and exactly one
per cluster is allowed (an upstream contract).

Know the dashboard's posture before exposing it: it has NO built-in
authentication, and in its default writable mode anyone who reaches it
can run and delete pipelines. Set `dashboard.readonly: true` (or keep
it unexposed — the port-forward command lands in the outputs).

Change first: `pipeline.cloud_events_sink_url` to stream every run's
lifecycle events to your CI orchestration — it is ONE cluster-global
URL by Tekton's design, so multi-tenant clusters point it at a fan-out
service. Disable the internet-reaching resolvers
(`pipeline.resolvers`) on restricted networks.

See [01-ci-standard.yaml](./01-ci-standard.yaml) for the manifest.
