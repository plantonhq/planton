# Scaled Tier

## Use Case

Run RAG Engine's managed vector database at production grade in one location: large corpora, latency-sensitive retrieval in a serving path, and autoscaling under load. `PREVENT` guards the location from an accidental destroy, which would unprovision it and delete every corpus's data.

## When to Use

- A RAG Engine corpus behind a production application
- Corpora too large or too hot for the Basic tier
- Any location whose data must survive a mistaken `destroy`

## What This Creates

- The RAG Engine configuration for `europe-west4` set to `SCALED`
- `deletionPolicy: PREVENT`, so a destroy fails until the policy is changed deliberately

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `europe-west4` | The location your corpora live in. |
| `tier` | `SCALED` | Drop to `BASIC` when the workload is small again; the change applies in place. |
| `deletionPolicy` | `PREVENT` | `ABANDON` to stop managing the tier while keeping the data; `DELETE` only for disposable projects. |

Raising `BASIC` to `SCALED` is a live upgrade; moving any tier to `UNPROVISIONED` deletes the managed database's data and cannot be undone for that data.
