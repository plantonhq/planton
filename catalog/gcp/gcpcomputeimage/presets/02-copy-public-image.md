# Copy a Public Image

## Use Case

Pin a copy of a public OS image in your own project, so VMs keep booting a version you tested even after the public family moves on.

## When to Use

- Controlling when a new OS release reaches your fleet
- Keeping images in a chosen storage location

## What This Creates

- The Compute Engine API on the project
- One copy of Debian 12's newest public image in the `debian-12-pinned` family

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `sourceImage` | Debian 12 family | Another public image or family. |
| `storageLocations` | `us` | Another multi-region or region. |
