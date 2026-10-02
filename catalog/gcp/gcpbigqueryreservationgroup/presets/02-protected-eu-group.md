# Protected EU Group

## Use Case

A production tier in the EU multi-region whose reservations share idle slots among themselves, protected so a destroy cannot remove the group its production reservations depend on.

## When to Use

- EU data-residency workloads with their own capacity tier
- Groups that production reservations reference

## What This Creates

- The `eu-production-tier` group in `EU` with `PREVENT` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `EU` | Match the reservations that join it. Fixed at creation. |
| `deletionPolicy` | `PREVENT` | `DELETE` for groups in short-lived environments. |
