# SigNoz for development

The smallest honest SigNoz: the kind's defaults against a composed
`KubernetesClickHouse` named `telemetry` in the same namespace — the
whole platform (UI, API, alerting, the ingestion collector) wired
through three references, no password anywhere in the manifest.

Deploy order and the two contracts to honor on the ClickHouse side:

1. A `KubernetesAltinityOperator`, then a `KubernetesClickHouse` named
   `telemetry` in the `telemetry` namespace at SigNoz's tested version
   (`version: "25.12.5"` at chart 0.133.0 — older servers fail the
   schema migrations), declaring a `signoz` user
   (no grants = unrestricted config-user access, which covers SigNoz's
   schema migrations) with explicit `networks` (e.g. `0.0.0.0/0` +
   `::/0` — a networks-less user is fenced to the ClickHouse pods and
   localhost, and SigNoz's pods get what reads as a password failure),
   and `coordination.type: managed_keeper` — SigNoz migrates
   `ON CLUSTER`, and a bare single-replica topology defaults to no
   coordination.
2. This SigNoz into the SAME namespace: the password travels by
   secretKeyRef, which cannot cross namespaces.

**When to use:** local clusters, spikes, previews — anywhere you want
traces flowing in minutes.

**When to move on:** production traffic deserves the production preset
(TLS to ClickHouse, alert email, external URL) and a sized ClickHouse.
