# GcpFolder Guide

The judgment this guide protects: a folder is where governance attaches.
Everything placed beneath it -- projects, sub-folders, their resources --
inherits its IAM grants and its organization policies, so the shape of the
folder tree IS the shape of your guardrails. Build the tree for the
boundaries you want to enforce, not for the org chart.

## Fold by boundary, not by team

Google evaluates IAM and organization policies down the hierarchy. A
folder per environment (`production`, `nonprod`) lets one policy say
"only these regions in production" and one grant say "the on-call group
may act in production" once, for every project that will ever live there.
Team folders nest inside environment folders when a team needs its own
grants; the reverse (environments inside teams) forces every guardrail to
be repeated per team. Ten levels are allowed; three is usually plenty.

## Reference the parent, do not paste its id

`parent.folderId` takes a reference to another `GcpFolder`, so a chart
declares `environments`, then `production` inside it, then a `GcpProject`
inside that with `folderId` -- and the platform deploys them in that order
and destroys them in reverse. Pasting a numeric id works, but breaks the
moment the chart is applied to a fresh organization.

## The destroy guard is on, and it should stay on

`deletionProtection` defaults to true: a destroy fails until you set it
to false and apply first. That two-step is the point -- deleting a folder
is a hierarchy-wide act, and a just-emptied folder is one accidental
destroy away from a 30-day recovery exercise. For the folders a landing
zone is built on, add `deletionPolicy: PREVENT` as well; `ABANDON` is the
lever for handing a folder to another owner without touching it.

## A move is an in-place update with policy consequences

Changing the parent moves the folder (Google's `folders.move`). Nothing is
recreated and everything inside travels along, but every grant and policy
inherited from the old parent stops applying at once and the new parent's
start. Treat a move as the last step of a governance change, not the
first: declare the new parent's grants and policies before the move.

## Tags at create time are the exception, not the rule

`tags` bind at creation and are immutable -- changing them recreates the
folder, which Google refuses for a non-empty one. Use them only when an
organization policy conditioned on the tag must govern the folder from
its first second. Otherwise bind tags after creation with `GcpTagBinding`,
which attaches and detaches without touching the folder.

## Names are reserved after deletion

A deleted folder keeps its display name reserved under the same parent
for the 30-day soft-delete window, and siblings can never share a name.
Ephemeral folders (test fixtures) need a fresh name per lifetime.
