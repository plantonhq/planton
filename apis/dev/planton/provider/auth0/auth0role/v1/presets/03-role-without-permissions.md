# Preset: Role Without Permissions

## Pattern

A bare role with only a name and description. Use this when you want the role to exist as a stable, assignable identity but intend to manage its permissions elsewhere.

## What It Does

- Creates a role with no permissions attached.
- Leaves the role's permission set unmanaged by this component, so permissions added out-of-band (or by another process) are not reconciled away.

## When to Use

- The role is assigned to users now, but its permissions are still being designed.
- Permissions are managed by a separate workflow or team and should not be overwritten on every apply.
- You are scaffolding a set of roles before the backing resource servers and scopes exist.

## Customization

- Set `spec.name` and `spec.description` to describe the role's intent.
- When you are ready to manage permissions here, add a `permissions` list (see the other presets). Note that once you manage permissions through this component, the set becomes authoritative.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `metadata.org` | Your Planton organization |
| `spec.name` | The display name of the role |
| `spec.description` | What the role is for |
