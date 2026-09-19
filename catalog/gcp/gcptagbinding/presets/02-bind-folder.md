# Bind Folder

Tags the production folder with `environment/prod`. Every project and
folder beneath inherits the tag, so one binding makes every
tag-conditioned guardrail govern the whole environment. Protected against
destroy for exactly that reason.

## What it configures

- `parent.folderId` — a reference to the `GcpFolder`'s `folder_id`.
- `deletionPolicy: PREVENT` — deleting this binding would lift every policy
  conditioned on the tag for the whole environment.

## Adjust before deploying

- **The two references** — your value and folder manifest names.

## When to choose something else

A single project takes the **Bind Project** preset.
