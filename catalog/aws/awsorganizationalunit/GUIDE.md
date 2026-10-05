# AwsOrganizationalUnit — Kind Guide

Authored operational judgment for the organizational-unit kind:
the design decisions behind the spec's shape, and what to know before
operating an OU tree in production.

## Design decisions

- **The name is an explicit spec field.** AWS allows OU names with
  spaces and arbitrary characters ("Core Services") that
  `metadata.name` cannot carry — the explicit-name-field convention,
  same as the account and policy kinds in this family.
- **The parent is a value-or-reference with the root as the default
  wiring.** First-level OUs are the common case, so the reference
  default resolves an AwsOrganization's `root_id`; nested OUs
  reference a parent OU's `ou_id` explicitly, and literals carry
  pre-existing trees. A literal-arm CEL enforces the provider's own
  `r-...`/`ou-...` pattern.
- **The organization is a registry prerequisite** — an OU cannot exist
  outside one, and the parent reference is schema-required.

## Operating an OU tree in production

- **Parents are immutable — design the tree before populating it.**
  AWS has no MoveOrganizationalUnit: re-parenting means recreating the
  OU, and a populated OU cannot be deleted. Accounts move freely
  between OUs; OUs do not.
- **Deleting an OU requires it empty** — move member accounts out and
  detach/delete child OUs and policies first. Destroy failures here
  are AWS's ordering contract, not module defects.
- **Nesting runs five levels deep** (the root plus five OU levels, per
  AWS's quota) — deep trees complicate policy inheritance reasoning
  long before the quota bites.
- **Policy inheritance flows down the tree**: an SCP attached to an OU
  governs every account and OU beneath it. Prefer attaching guardrails
  high (root or first-level OUs) and exceptions low.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
