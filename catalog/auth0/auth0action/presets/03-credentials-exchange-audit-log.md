# Credentials Exchange Audit Log

## Pattern

Log every machine-to-machine (M2M) token exchange to an external audit endpoint for compliance and security monitoring.

## What It Does

- Captures metadata from each client_credentials exchange: client identity, audience, requested scopes, timestamp, and IP address.
- POSTs the audit event to a configurable HTTPS endpoint with bearer token authentication.
- Fails gracefully — audit failures are logged but do not block token issuance.

## When to Use

- Compliance requirements mandate logging of all API-to-API authentication events.
- Security teams need visibility into which M2M clients are requesting tokens and for which audiences.
- You want to detect unusual M2M activity (new clients, unexpected scopes, off-hours access).

## Customization

- Store your SIEM or logging service's endpoint and bearer token as organization secrets (`planton secret set audit-endpoint --string`, `planton secret set audit-token --string`) and point `AUDIT_ENDPOINT` and `AUDIT_TOKEN` at them as `$secret/audit-endpoint` and `$secret/audit-token`.
- Add additional fields from the `event` object (e.g., `event.client.metadata`).
- To block suspicious exchanges, add `api.access.deny()` based on your criteria.
- Adjust the `timeout` value based on your audit endpoint's SLA.
