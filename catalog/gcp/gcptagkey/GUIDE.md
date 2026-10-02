# GcpTagKey Guide

The judgment this guide protects: a tag is a governed fact about a
resource that policies trust, so its vocabulary is designed once, owned
centrally, and treated as immutable. Labels are for everything else.

## Tags are for policy; labels are for people

A label is free-form metadata anyone can set and nothing enforces on. A
tag is created centrally, permissioned (`tagUser` to bind it), immutable
in name, and evaluated by organization policies (`resource.matchTag`),
IAM conditions, and firewall policy rules. Use tags for the facts a
guardrail must trust -- environment, data classification, network
posture -- and labels for cost slicing and search.

## Own keys at the organization

An organization-owned key's values can be bound to any resource in the
estate; a project-owned key's values only inside that project. The
landing-zone vocabulary (`environment`, `data-classification`) belongs to
the organization. A project-owned key is for a tag one application
manages for itself -- and it is also the shape that needs no
organization-level grant, which is why it is the one that can be proved
in a single project. A folder cannot own a key; Google's rule.

## Design the vocabulary before creating the key

Owner, short name, purpose, and purpose data are immutable; Google
refuses to delete a key that still has values, and a value that still has
bindings. Renaming `env` to `environment` after fifty bindings means
fifty new bindings. Only the description and the allowed-values regex
change in place. Choose short names as they will appear in conditions:
`resource.matchTag('123456789012/environment', 'prod')`.

## Declared values or a dynamic pattern

A key's vocabulary is the set of `GcpTagValue`s declared under it -- a
closed list a policy can enumerate. Set `allowedValuesRegex` when the
values are high-cardinality (a ticket number, a team code): the key
becomes dynamic, bindings may carry undeclared values that match, and the
regex is the whole contract. Do not mix the two on one key without a
reason.

## Purposes change the key's nature

`GCE_FIREWALL` makes the key's values secure tags usable as sources and
targets in network firewall policy rules, scoped to the VPC named in
`purposeData.network`; `DATA_GOVERNANCE` marks values as data
classifications for Sensitive Data Protection and BigQuery. Both are
immutable once set -- a firewall key cannot later become an ordinary key.

## Delete in order

Bindings, then values, then the key. A chart that declares all three by
reference destroys them in that order automatically; `deletionPolicy:
PREVENT` protects a key the whole estate's policies key on.
