# Publish to Any Consumer (Automatic Acceptance)

## Use Case

Publish an internal platform service every project in the organization may reach -- a metrics gateway, a package mirror, an internal DNS forwarder -- without maintaining an accept list. `ACCEPT_AUTOMATIC` admits any consumer that creates an endpoint; a reject list blocks the exceptions. The PROXY protocol gives the backends the consumer's original address for logging and per-consumer policy, and a Cloud DNS domain registers connected endpoints under one internal name.

## When to Use

- Organization-wide platform services behind an internal load balancer
- Services whose consumers are many and change often, where a manual accept list would not keep up
- Only when admitting any Google Cloud project (minus the reject list) is acceptable -- otherwise start from the manual-acceptance preset

## What This Creates

- A service attachment in `us-central1` in front of `metrics-gateway-ilb`
- Two PSC NAT subnets for a large consumer fleet
- `ACCEPT_AUTOMATIC` with one rejected project
- PROXY protocol on, the `metrics.internal.example.` domain, `deletionPolicy: DELETE`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `natSubnets` | two subnets | Size for the expected endpoint count; add subnets to grow. |
| `consumerRejectLists` | one project | The projects that must never connect. |
| `enableProxyProtocol` | `true` | `false` when the backends do not speak the PROXY protocol -- they would fail to parse connections. |
| `domainNames` | `metrics.internal.example.` | Your internal domain, with the trailing dot; at most one. |
| `deletionPolicy` | `DELETE` | `PREVENT` once consumers depend on it. |

Under `ACCEPT_AUTOMATIC` there is no `consumerAcceptLists`; the spec rejects one.
