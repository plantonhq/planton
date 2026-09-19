# GcpTagValue Guide

The judgment this guide protects: a tag value is one word in a governed
vocabulary. Declare the words a policy will test for, one value each,
and never reuse a word for a different meaning.

## One value per allowed setting

`environment/prod`, `environment/staging`, `environment/dev` are three
`GcpTagValue` resources under one `GcpTagKey`. A policy tests a specific
value (`resource.matchTag('123456789012/environment', 'prod')`), so the
set of values IS the set of things a policy can distinguish. Declare the
values the guardrails need; leave out the ones nobody tests for.

## Reference the key

`tagKey` takes a reference to the `GcpTagKey` (its `name` output,
`tagKeys/{id}`), so a chart declares the key and its values together and
the platform orders them. The `tagKeys/{id}` literal works for a key
created elsewhere; the numeric id comes from the key's outputs or from
`gcloud resource-manager tags keys describe`.

## The name is immutable

The key and the short name cannot change; only the description can.
Renaming a value means creating a new one and re-binding every resource,
so pick short names as they will appear in conditions -- lowercase, no
slashes or quotes, stable. A deleted value's short name stays reserved
under its key for 30 days.

## Delete bindings first

Google refuses to delete a value while any binding uses it. A chart that
declares the bindings by reference destroys them before the value; a
hand-managed estate must do the same. `deletionPolicy: PREVENT` protects
a value every environment's policies test.

## Values under a dynamic key

If the key has an `allowedValuesRegex`, every declared value's short name
must match it -- and bindings may also carry undeclared values that match.
Declare the values you want to be able to enumerate anyway; the regex is
for the long tail.
