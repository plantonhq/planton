# Auth0 Event Stream Guide

## Security
## Platform Security Posture

The certifications below are Auth0's own published claims about their hosted platform (verify current status on Auth0's compliance page). They describe the vendor's service — never this catalog kind, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Auth0's published certifications and security standards:

- SOC 2 Type II (annual audit)
- ISO 27001, ISO 27018 (privacy controls)
- HIPAA BAA available on enterprise plans
- PCI DSS Level 1 Service Provider
- FedRAMP Authorized (moderate baseline)
- CSA STAR Level 2
- GDPR compliant with Data Processing Agreement

## Data Protection

- **Data residency**: US, EU, AU regions
- **Encryption in transit**: TLS 1.2+
- **Encryption at rest**: AES-256
- **Penetration testing**: Annual third-party assessments

## Event Stream-Specific Security Notes

### Webhook Credentials

Webhook-type event streams require authentication tokens or credentials to deliver events to your endpoint. These credentials are:

- Stored encrypted in Auth0's configuration store
- Sent with each webhook delivery (typically as an Authorization header or custom token)
- Must be rotated periodically to maintain security posture

### Event Data Content

Auth0 event streams deliver tenant log events that may contain user PII:

- User email addresses and names (in login events)
- IP addresses and geolocation data
- User agent strings
- Authentication method details

Ensure the destination system handles this data in accordance with your data protection obligations.

### Transport Security

- **Webhook endpoints**: Must use HTTPS (TLS 1.2+). Auth0 will not deliver events to HTTP endpoints.
- **Amazon EventBridge**: Uses AWS's built-in encryption for event delivery.
- **Third-party integrations** (Datadog, Splunk, etc.): Use each provider's secure ingestion endpoints with API key authentication.

### Delivery Reliability

Auth0 retries failed webhook deliveries with exponential backoff. Failed events are retried for up to 24 hours. If the destination remains unavailable, events are dropped. There is no dead-letter queue for failed deliveries.

### Access to Stream Configuration

Event stream configurations contain sensitive credentials (webhook tokens, API keys). Limit Management API access to `read:log_streams` to minimize exposure of these credentials.

## Permissions
## Management API Scopes

Auth0 event stream (log stream) resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:log_streams` | List and retrieve event stream configurations |
| Create | `create:log_streams` | Create new event streams (webhook, EventBridge, etc.) |
| Update | `update:log_streams` | Modify stream destination, filters, and credentials |
| Delete | `delete:log_streams` | Remove event streams from the tenant |

## Related Scopes

Event streams deliver tenant log data. These additional scopes relate to the underlying log data:

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read logs | `read:logs` | Query tenant logs directly via the Management API |
| Read log users | `read:logs_users` | Access user-specific log entries |

## Minimum Required Scopes

For basic lifecycle management (create, read, update, delete), the minimum required scopes are:

```
read:log_streams create:log_streams update:log_streams delete:log_streams
```

The `read:logs` and `read:logs_users` scopes are not required for managing event streams but may be useful for verifying that events are being generated correctly.

## Credential Sensitivity

Event stream configurations contain destination credentials (webhook tokens, API keys). The `read:log_streams` scope exposes these credentials in API responses. Grant this scope only to trusted automation principals.

## Compliance
## Regulatory Frameworks

The table below records Auth0's own published compliance posture for their hosted platform, as Auth0 states it (verify current status in Auth0's trust documentation). These are vendor facts about the service this component configures — not properties of the component or of your deployment, and nothing here transfers to your application without your own assessment.

Auth0's published framework posture:

| Framework | Status | Notes |
|-----------|--------|-------|
| SOC 2 Type II | Certified | Annual audit cycle |
| ISO 27001:2022 | Certified | Information security management |
| ISO 27018:2019 | Certified | PII protection in public clouds |
| HIPAA | BAA Available | Enterprise plans only |
| PCI DSS Level 1 | Certified | Service provider certification |
| FedRAMP Moderate | Authorized | US government workloads |
| CSA STAR Level 2 | Certified | Cloud security assurance |
| GDPR | Compliant | Data Processing Agreement available |
| CCPA | Compliant | California consumer privacy |
| Privacy Shield | Historical | Deprecated framework, replaced by SCCs |

## Event Stream-Specific Compliance Notes

### PII in Event Data

Auth0 log events contain user PII including email addresses, IP addresses, and user agent strings. When streaming events to external systems, ensure the destination complies with applicable data protection regulations (GDPR, CCPA). Data Processing Agreements may be required with third-party log aggregation providers.

### Cross-Border Data Transfer

Event streams may deliver data to destinations outside your Auth0 tenant's region. If your tenant is in the EU region and events stream to a US-based service, this constitutes a cross-border data transfer subject to GDPR Chapter V requirements. Use Standard Contractual Clauses (SCCs) or other approved transfer mechanisms.

### Log Retention and Right to Erasure

Auth0's built-in log retention is time-limited. Event streams export logs to external systems where retention policies are independently managed. For GDPR right-to-erasure compliance, ensure your external log storage can identify and delete user-specific log entries upon request.

### Audit Trail

All event stream CRUD operations are recorded in Auth0 tenant logs. Stream delivery status (success/failure) is also logged. These operational logs provide evidence of log pipeline health for compliance audits.

### Data Minimization

Auth0 streams all tenant log event types by default. Event streams support filtering by event type. For compliance-sensitive environments, configure filters to exclude event types containing unnecessary PII.

## Cost
## Pricing Model

Auth0 pricing is based on Monthly Active Users (MAUs), not on the number of resources created. Event streams (log streams) are free API objects with no per-resource cost.

## Free Tier

The Auth0 Free plan includes:

- 25,000 MAUs
- 1 tenant
- Log streams available on all plans

## Cost Impact

Creating, updating, or deleting Auth0 event stream resources has no direct billing impact from Auth0. There is no charge per log stream definition.

The only Auth0 cost driver is the number of monthly active users authenticating through your tenant.

## Downstream Cost Considerations

While Auth0 does not charge for event streams, the destination services may incur costs:

| Destination Type | Cost Consideration |
|-----------------|-------------------|
| Webhook (HTTP) | Cost depends on your endpoint's hosting |
| Amazon EventBridge | EventBridge charges per event published |
| Datadog | Log ingestion charges apply |
| Splunk | Log volume-based pricing applies |
| Sumo Logic | Ingestion-based pricing applies |
| Azure Event Hubs | Throughput unit pricing applies |

## Event Volume

Event volume scales with authentication activity, not with the number of log streams. Creating multiple streams to the same destination type multiplies the downstream ingestion cost but not the Auth0 cost.

## Log Retention

Auth0's built-in log retention is limited by plan tier (2 days free, 30 days enterprise). Event streams provide a mechanism to export logs to external systems for longer retention without affecting Auth0 costs.
