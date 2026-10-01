# Golden Image from a Disk

## Use Case

Capture a configured boot disk as a versioned golden image in a family, so new VMs boot the newest build through the family name.

## When to Use

- A hardened OS image with your agents pre-installed
- Rolling builds: each build is a new image in the same family

## What This Creates

- The Compute Engine API on the project
- One image in the `web-base` family

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `metadata.name` / `imageName` | `web-base-20261001` | One name per build. |
| `family` | `web-base` | The stable name consumers boot from. |
| `sourceDisk` | a `GcpComputeDisk` reference | The disk you configured (stop its VM first). |
