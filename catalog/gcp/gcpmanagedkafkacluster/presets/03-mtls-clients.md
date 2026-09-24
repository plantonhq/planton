# Mutual TLS Clients

## Use Case

Clients that authenticate with certificates instead of Google IAM tokens -- partner systems or workloads outside Google Cloud -- using client certificates issued by a Certificate Authority Service pool, with their names mapped to the short principals your Kafka ACLs use.

## When to Use

- Clients that cannot obtain Google IAM tokens
- Certificate-based access policies
- Kafka ACLs keyed on certificate identities

## What This Creates

- A 3-vCPU cluster in `europe-west1` trusting one CA pool for mTLS on port 9192, with a principal mapping rule

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `tlsConfig.caPools` | `kafka-clients` | Up to ten pools, any project or location. |
| `tlsConfig.sslPrincipalMappingRules` | CN only | Kafka's `ssl.principal.mapping.rules` syntax; changing it restarts brokers one by one. |
