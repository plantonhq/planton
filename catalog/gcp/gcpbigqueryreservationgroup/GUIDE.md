# GcpBigQueryReservationGroup Guide

The judgment this guide protects: idle capacity should stay with the people who paid for it before it helps anyone else. A reservation group draws that boundary.

## How sharing works

By default, a reservation's idle baseline slots can be borrowed by any other reservation in the same admin project (unless the lender sets `ignoreIdleSlots`). Reservations in a group lend and borrow idle slots among themselves first, and only then with the rest of the project. The group holds no slots and costs nothing; it is a boundary.

## Membership

A reservation joins by referencing the group's `name` output from its `reservationGroup` field -- the group itself lists nobody, so each team's reservation manifest carries its own membership. The group and its reservations live in the same admin project and location, and the reservations depend on the group, so they are destroyed before it.
