# Kafka Producer with Topic ACLs

This preset creates a Kafka cluster user scoped to producing on a topic prefix, with produce-and-consume rights on a dead-letter topic. Kafka users authenticate with mutual TLS: DigitalOcean issues the certificate pair, exported as the `access_cert` / `access_key` secret outputs.

## When to Use

- Producer services that must not read application topics
- Least-privilege Kafka access instead of the cluster's admin default
- Wildcard topic families (`events-*`) owned by one producer

## Key Configuration Choices

- **`produce` on `events-*`** -- write-only on the topic family; the wildcard tracks new topics automatically.
- **`produceconsume` on `dead-letter`** -- producers typically re-read their own dead-letter queue.
- **ACLs are write-only upstream** -- DigitalOcean returns them only at create time, so this manifest is the source of truth for what the user may do; review changes here, not in the console.

## What You Get

A Kafka user whose topic rights match the manifest exactly, with the mTLS credential pair and password exported as secret outputs.
