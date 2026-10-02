# GcpBigQueryReservation Guide

The judgment this guide protects: a reservation turns BigQuery from a per-query bill into capacity you plan. Plan it as capacity -- a baseline you always pay for, headroom you pay for only when used -- and decide deliberately which workloads run on it.

## Capacity

`slotCapacity` is the baseline: slots allocated and billed around the clock. `autoscaleMaxSlots` is headroom: slots added only while queries need them and billed only while in use. A zero baseline with autoscaling is pay-while-used capacity; a baseline is for steady load, and pairs with a `GcpBigQueryCapacityCommitment` in the same admin project and location, which lowers its price. Idle baseline slots are lent to other reservations in the admin project unless `ignoreIdleSlots` is set; a `GcpBigQueryReservationGroup` makes a set of reservations share idle slots with each other first. `concurrency` is a soft cap on queries running at once (0 sizes it automatically).

## Edition

`STANDARD` offers autoscaling without commitments and a smaller feature set; `ENTERPRISE` adds commitments, BigQuery ML, BI Engine, and more; `ENTERPRISE_PLUS` adds managed disaster recovery and compliance controls. The edition is fixed at creation and keys the slot price.

## Assignments

A reservation serves only the jobs assigned to it. Each assignment names exactly one assignee -- a project, a folder, or an organization, by reference or literal -- and a job type: `QUERY` for queries and scripts, `PIPELINE` for load, export, and copy jobs, `CONTINUOUS` for continuous queries. A `principal` narrows it to one user, service account, or workload identity; everyone else in the assignee falls back to the next assignment up the hierarchy, then to on-demand. Assignments are immutable (a change replaces one) and need `bigquery.reservationAssignments.create` on both the admin project and the assignee. Forcing a project onto on-demand with BigQuery's built-in `none` reservation is not expressible here.

## Disaster recovery

`secondaryLocation` (Enterprise Plus) makes a failover reservation replicated to a second region; failing over is an operational action outside the spec. Setting or clearing it later converts the reservation.
