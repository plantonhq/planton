# Environment Folder

A top-level folder directly under the organization, one per environment
(`production`, `nonprod`). The node every environment-wide guardrail and
grant attaches to, protected against accidental destroy twice over.

## What it configures

- `parent.organizationId` — the folder sits directly under the organization.
- `deletionProtection: true` — a destroy fails until this is set to false and
  applied first (Google's default, sent explicitly).
- `deletionPolicy: PREVENT` — the second guard: destroy fails outright.

## Adjust before deploying

- **`organizationId`** — your numeric organization ID (`gcloud organizations list`).
- **`displayName`** — unique among the organization's top-level folders.
- Drop `PREVENT` only for folders you expect to tear down.

## When to choose something else

A folder inside another folder takes the **Nested Team Folder** preset; a
folder that must carry a tag from creation takes **Tagged At Create**.
