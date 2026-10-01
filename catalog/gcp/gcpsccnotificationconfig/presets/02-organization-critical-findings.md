# Organization Critical Findings

## Use Case

Send every active critical finding in the organization to the security operations center's topic, wherever it appears.

## When to Use

- A central SOC with a SIEM or SOAR subscribed to one topic
- Organization-level Security Command Center activation

## What This Creates

- An organization notification config streaming active critical findings, with destroy blocked

Grant the config's `service_account` output `roles/pubsub.publisher` on the topic.

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `scope.organizationId` | `123456789012` | Your organization's numeric ID. |
| `filter` | active CRITICAL | Widen to HIGH, or restrict to categories. |
| `deletionPolicy` | `PREVENT` | `DELETE` only when retiring the stream on purpose. |
