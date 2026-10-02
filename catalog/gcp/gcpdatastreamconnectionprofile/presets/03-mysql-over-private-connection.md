# MySQL over a Private Connection

## Use Case

Reach a MySQL server that has no public address -- a VM or an on-premises database -- through a shared Datastream private connection, with the session encrypted by TLS.

## When to Use

- Self-managed MySQL inside a VPC or behind VPN or Interconnect
- Security policies that forbid database public IPs

## What This Creates

- A MySQL source profile in `us-central1` for `10.10.0.20`, over the `data-vpc` private connection, with TLS on

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `mysqlProfile.hostname` | `10.10.0.20` | The server's private address. |
| `mysqlProfile.sslConfig` | TLS, no verification | Add `caCertificate` to verify the server; add a client certificate and key for mutual TLS. |
| `privateConnection` | `GcpDatastreamPrivateConnection` reference | The link to the network the server lives in. |

The server needs binary logging in ROW format and a user with REPLICATION SLAVE, REPLICATION CLIENT, and SELECT.
