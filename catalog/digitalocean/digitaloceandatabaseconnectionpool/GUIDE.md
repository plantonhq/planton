# DigitalOcean Database Connection Pool -- Operational Guide

What experience with this kind teaches that the field reference cannot.

## Every change is an outage, so decide sizes early

The provider has no update path for pools: size, mode, database, user -- all of it is create-only, and any edit REPLACES the pool, dropping its live connections. (DigitalOcean's raw API can update pools in place; the provider does not expose it -- mirrored here deliberately, revisited if the provider grows an update path.) Practical consequence: size generously up front, and schedule pool edits like brief connection outages with client retry in place.

## Transaction vs session vs statement

`transaction` is the right default for web apps: a server connection per transaction, maximum reuse -- but session state dies (LISTEN/NOTIFY, session prepared statements, advisory locks held across transactions). `session` preserves all of that at the cost of one server connection per connected client -- size for concurrent CLIENTS. `statement` is for autocommit-style workloads only; multi-statement transactions fail on it.

## Sizing against the cluster's connection budget

`size` is backend connections HELD OPEN, and the cluster's connection limit scales with its node size (DigitalOcean reserves a few for itself). Pools on the same cluster share that budget with direct connections. If the pool needs to grow past the budget, the real fix is a bigger cluster size slug.

## The inbound-user pattern

Omit `user` and the pool proxies each client's own credentials -- per-client identity, grants, and rotation stay intact, which is why it is the safer default for shared pools. The `password` output is legitimately empty for this shape. Name a user only when exactly one service owns the pool.

## Connect to the pool's port, not the cluster's

The pool listens on its own port beside the cluster's (both on the same hosts). Clients connect to `pool_name` as if it were a database name, at the pool's port -- the `uri`/`private_uri` outputs have this right already; hand-built connection strings routinely get it wrong.

## What is deliberately NOT here

In-place pool updates (no provider surface -- see above), non-PostgreSQL engines (DigitalOcean pools are PostgreSQL-only), and PgBouncer tunables beyond mode/size (not exposed by the API).
