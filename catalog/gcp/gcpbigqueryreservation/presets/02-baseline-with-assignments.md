# Baseline with Assignments

## Use Case

Production analytics capacity: 200 always-on Enterprise slots with 400 more on demand, serving every project in a folder for queries and an ETL project for load and export jobs, protected from destroy.

## When to Use

- Steady production load with bursts
- Capacity shared by a folder of projects
- Pairing with a capacity commitment

## What This Creates

- An Enterprise reservation with two assignments and `PREVENT` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `slotCapacity` | `200` | Size to steady load; commit it with a `GcpBigQueryCapacityCommitment` to lower its rate. |
| `assignments` | folder + ETL project | Every assignee's jobs of that type run here. |
