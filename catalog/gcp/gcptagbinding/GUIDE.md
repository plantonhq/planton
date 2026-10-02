# GcpTagBinding Guide

The judgment this guide protects: a binding is the moment a tag starts
governing a resource. It is immutable, it inherits downward, and Google
insists on the exact identity of the tagged resource -- so bind by
reference and let the platform get the identity right.

## Bind the hierarchy, inherit everywhere

A tag bound to a folder applies to every project and folder beneath it;
bound to a project, to every resource in it. Bind `environment/prod` once
on the production folder and every organization policy conditioned on it
governs everything inside. Bind individual resources only for facts that
differ per resource (a data classification on one bucket).

## No parent means "this project"

A binding with an empty `parent` tags the project the deploying
credentials are configured for -- the everyday case of "tag the project I
am working in". Name a parent when the tag belongs somewhere else.

## Projects are addressed by number, and the module handles it

Google's Tag Bindings API wants a project's NUMBER in the parent, not its
ID. A `GcpProject` reference resolves to the number automatically; a
numeric literal is used as is; a project ID literal (or the empty parent)
is looked up once at apply time. Both engines gate that lookup the same
way, so a plan on a reference or a number performs no live read.

## Regional and zonal resources need a location

Organizations, folders, and projects are global. A Compute instance, a
Cloud SQL instance, or a GKE cluster is served from a location, and its
binding must be too: set `parent.resourceName` to the resource's full
resource name and `location` to its zone or region. The module switches
to the location-scoped binding resource; leaving `location` empty for
such a resource fails at apply time.

## A binding is replaced, never edited

Every field is immutable, so a change destroys the old binding and
creates the new one -- and Google allows one value per key per resource,
so moving a project from `environment/staging` to `environment/prod`
means destroying the staging binding before the prod one can exist. A
chart does this when the old binding is removed and the new one added in
one apply; watch the order when doing it by hand.

## Delete order and protection

Bindings go before values, values before keys. Deleting a binding is
immediate, with no soft-delete window -- the tag simply stops applying.
For a binding an organization policy depends on (the production folder's
`environment/prod`), `deletionPolicy: PREVENT` turns an accidental destroy
into a failed plan instead of a lifted guardrail.
