# Tagged At Create

A folder that carries a Resource Manager tag from its first second, so an
organization policy conditioned on that tag (`resource.matchTag(...)`)
governs it immediately -- the sandbox exemption pattern. The guard is off
because a sandbox folder is meant to be torn down.

## What it configures

- `tags` — `tagKeys/{id}` -> `tagValues/{id}`, the `name` outputs of a
  `GcpTagKey` and a `GcpTagValue`. Bound at creation.
- `deletionProtection: false` and `deletionPolicy: DELETE` — disposable.

## Adjust before deploying

- **`tags`** — the real key and value ids. These are IMMUTABLE: changing them
  recreates the folder, which Google refuses while it holds projects. For
  a tag on an existing folder use `GcpTagBinding` instead.

## When to choose something else

Every folder that will hold real workloads takes the **Environment Folder**
or **Nested Team Folder** preset and binds tags afterwards with
`GcpTagBinding`.
