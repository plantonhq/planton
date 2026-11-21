# Parked GPU

## Use Case

A GPU runtime that exists, fully configured, but stays stopped until a researcher needs it -- auto-upgraded to Google's latest image whenever it starts.

## When to Use

- Expensive GPU runtimes used a few days a month
- Keeping a runtime's disk (installed packages, local data) between sessions

## What This Creates

- A runtime from the `private-gpu-t4` template, assigned to one researcher, created `STOPPED` (a user can still start it in Colab Enterprise until idle shutdown stops it), auto-upgrading during deploys that change it, with `PREVENT` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `desiredState` | `STOPPED` | `RUNNING` for the days it is in use; clearing it also starts the runtime. |
| `autoUpgrade` | `true` | `false` to hold an image version. |
