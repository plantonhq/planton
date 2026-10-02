# GcpProject Guide

The judgment this guide protects: the project is the blast-radius
boundary for everything — billing, IAM, quotas, APIs. It is the one
resource whose accidental destruction takes everything else with it,
which is why its destroy default differs from the whole catalog.

## The destroy default here is the provider's PREVENT-shaped exception

Everywhere else in this catalog, an unset `deletionPolicy` means DELETE.
The provider makes google_project the exception (its own default is
PREVENT); this kind's contract keeps the catalog-wide DELETE default —
the module sends DELETE explicitly when the spec is silent — so destroy
semantics stay uniform across kinds. The consequence: for any real
project, SET `deletionPolicy: PREVENT` explicitly. A deleted project
enters a 30-day pending-deletion window (restorable, but APIs and
serving stop immediately); `ABANDON` is the right hand-off when
ownership moves to another management plane.

## Tags recreate the project — with PREVENT, they deadlock instead

`tags` (resource-manager tags) are create-time only: changing them
plans a project REPLACE. Under `deletionPolicy: PREVENT` that plan
cannot apply — the destroy half fails. This is the intended safety
interlock, not a bug: to change tags on a live project, decide
consciously (flip the policy, or apply tags via the tag-binding tooling
outside this kind).

## enabled_apis is authoritative, and disabling cascades

The API list is reconciled: removing an entry disables the service AND
(module-fixed) its dependent services — the honest interpretation of
"this list says what the project offers". Add APIs freely; remove them
with the same care as removing a permission. Destroying the project
resource itself never disables APIs one-by-one (disable_on_destroy is
false) — the project teardown handles everything.

## auto_create_network false still needs network quota

Disabling the default network is implemented by GCP as
create-then-delete, so project creation still momentarily consumes one
network slot. In quota-constrained organizations, a project create can
fail on network quota even with `autoCreateNetwork: false`.

## Parent and billing moves are org-level events

`parentType`/`parentId` (or `folderId`) moves re-scope every inherited IAM policy and
org constraint; `billingAccountId` changes who pays retroactively for
nothing and prospectively for everything. Both are one-line edits here
and multi-team conversations in real life — treat the spec change as
the LAST step of the move, not the first.
