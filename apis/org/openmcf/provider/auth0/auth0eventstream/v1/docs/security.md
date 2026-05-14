# Auth0 Event Stream - Security

## Platform Security Posture

Auth0 maintains the following certifications and security standards:

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
