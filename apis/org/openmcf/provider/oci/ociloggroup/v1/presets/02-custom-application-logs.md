# Custom Application Logs

This preset creates a log group with a custom log that accepts application-level log entries pushed via the OCI Logging Ingestion API. Unlike service logs that are auto-collected from OCI infrastructure, custom logs are populated by your application code, sidecar agents, or log shipping pipelines (e.g., Fluentd, Fluent Bit, or the OCI Logging agent). This is the standard pattern for centralizing application telemetry in OCI Logging.

## When to Use

- Centralizing application logs (HTTP access logs, business events, error traces) in OCI Logging
- Shipping container or VM application logs via Fluent Bit, Fluentd, or the OCI Logging agent
- Building a unified observability pipeline where application logs feed into OCI Log Analytics or Object Storage
- Any custom telemetry that does not originate from an OCI managed service

## Key Configuration Choices

- **Custom log type** (`logType: custom`) -- the log accepts entries pushed via the Logging Ingestion API. No source configuration is needed because the application controls what is sent.
- **30-day retention** (`retentionDuration: 30`) -- the minimum retention period and OCI default. Suitable for operational troubleshooting. Increase to 60 or 90 days if logs are needed for compliance or post-incident analysis beyond the immediate operational window.
- **Enabled on creation** (`isEnabled: true`) -- the log accepts ingestion immediately. Set to `false` to create the log endpoint without activating ingestion (useful when the application is not yet deployed).
- **No configuration block** -- custom logs do not require the `configuration` field. The service/resource/category fields are only relevant for service logs that auto-collect from OCI resources.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the log group will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |

## Related Presets

- **01-vcn-flow-logs** -- use instead for auto-collecting network traffic logs from VCN subnets
