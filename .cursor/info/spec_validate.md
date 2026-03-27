# Spec Validate Authoring Guide

Purpose: add validation rules to an existing `spec.proto` without changing its schema.

## Scope
- Do not rename/add/remove fields or messages.
- Only add validations and brief comments.

## Validation Message Philosophy

**Every CEL `message` string you write will be read by a human who is trying to configure a cloud component. That human may be a senior platform engineer or a developer who has never used this cloud provider before. Your message is the only guidance they receive when their configuration is invalid.**

**You are not a compiler reporting an error. You are a knowledgeable colleague reviewing their configuration and helping them succeed.**

### The Three Questions

Every validation message MUST answer at least one of these questions, ideally all three:

1. **What's wrong?** -- State the issue clearly in plain English
2. **Why it matters?** -- Briefly explain what would break or what the field controls (when non-obvious)
3. **How to fix it?** -- Provide an example of a valid value or describe the expected format

### Message Quality Standards

**DO:**
- Write complete sentences that a non-expert can act on
- Explain character constraints in plain English ("lowercase letters, numbers, and hyphens")
- Include examples inline when the constraint is non-obvious (e.g., regex patterns)
- Explain conditional requirements with context ("hostname is required when ingress is enabled")
- Use the field's semantic purpose, not its technical path ("Image repository" not "this.repo")

**DON'T:**
- Expose raw regex patterns as the primary message (put regex in a code comment, explain the pattern in the message)
- Write messages shorter than 5 words ("required" alone is not enough -- "Image repository is required" is)
- Use proto field names verbatim ("min_replicas must be >= 1" -- say "Minimum replica count must be at least 1")
- Write messages that assume knowledge of the component's internals
- Use jargon without explanation

### Good vs. Bad Examples

```proto
// BAD: Terse, unexplained
message: "required"
message: "invalid format"
message: "must match ^[a-z0-9-]+$"

// GOOD: Clear, actionable, empathetic
message: "Image repository is required -- this is the registry URL where your container image is hosted (e.g., gcr.io/my-project/my-app)"
message: "Only lowercase letters, numbers, and hyphens are allowed"
message: "Must not end with a hyphen"
message: "Hostname is required when ingress is enabled -- this is the DNS name users will use to reach your service"
message: "Certificate ARN must be set when custom aliases are configured -- the ALB needs a TLS certificate to serve HTTPS for your domain"
```

### Regex Constraints

When a field has a regex pattern validation:
1. Put the regex in a code comment (with a regex101 link if complex)
2. Write the message in plain English describing what characters/format are allowed
3. If the pattern is complex, include an example of a valid value in the message

```proto
// GOOD: Human message + regex in comment
(buf.validate.field).cel = {
  id: "spec.version.chars"
  message: "Only lowercase letters, numbers, and hyphens are allowed"
  // https://regex101.com/r/NKTohE/1
  expression: "this.matches('^[a-z0-9-]+$')"
}

// BAD: Regex in the message
(buf.validate.field).cel = {
  id: "spec.version.chars"
  message: "Must match pattern ^[a-z0-9-]+$"
  expression: "this.matches('^[a-z0-9-]+$')"
}
```

### Conditional Constraints

When a field is only required under certain conditions, the message MUST explain both the condition and the requirement:

```proto
// GOOD: Explains the condition and the requirement
message: "Hostname is required when ingress is enabled -- without it, the load balancer cannot route external traffic to your service"

// BAD: Missing context
message: "hostname required"
```

### Self-Test Before Committing

Before finalizing validation messages, read each `message` string aloud and ask:
- Would someone who has never used this cloud provider understand what to do?
- Does the message explain WHY this constraint exists, not just WHAT it is?
- If there's a regex or format constraint, is there a human-readable description?
- Is the message specific to this field (not generic like "invalid value")?

## Import
- Add once near the top:
  - `import "buf/validate/validate.proto";`

## Field-level validations (80/20)
- Strings:
  - Names/IDs: `(buf.validate.field).string.min_len = 1`
  - ARNs/URIs: add `min_len` and, if stable, a simple `pattern` (avoid brittle regexes)
  - Domains/emails: prefer minimal `pattern` or `min_len`
- Enums:
  - `(buf.validate.field).enum.defined_only = true`
  - If field must be set and enum has `*_UNSPECIFIED = 0`, add CEL to forbid 0
- Numbers:
  - Use `gt/gte/lt/lte` as appropriate
- Booleans:
  - Typically no direct validation; enforce via CEL when tied to other fields
- Repeated:
  - `(buf.validate.field).repeated.min_items = 1` when at least one is required
  - Consider `(buf.validate.field).repeated.unique = true` for sets like aliases
- Bytes/Maps:
  - Apply min/max sizes if applicable

## Message-level CEL validations
- Require B when A is set:
  - Example: if `aliases` non-empty, `certificate_arn` must be non-empty
  - `this.aliases.size() == 0 || this.certificate_arn != ""`
- Mutually exclusive fields:
  - `this.x == "" || this.y == ""`
- Enum-dependent constraints (e.g., DynamoDB billing mode):
  - If `billing_mode == PROVISIONED`, then `read_capacity_units > 0 && write_capacity_units > 0`, else both 0/unset
- Ordered ranges:
  - `(this.min_ttl == 0 || this.default_ttl >= this.min_ttl) && (this.max_ttl == 0 || this.default_ttl <= this.max_ttl)`

## Example (adapt to your schema)
```proto
syntax = "proto3";
package org.openmcf.provider.aws.awscloudfront.v1;

import "buf/validate/validate.proto";

message AwsCloudFrontSpec {
  repeated string aliases = 1 [(buf.validate.field).repeated = {min_items: 1, unique: true}];

  string certificate_arn = 2 [(buf.validate.field).string.min_len = 1];

  enum PriceClass {
    PRICE_CLASS_UNSPECIFIED = 0;
    PRICE_CLASS_100 = 1;
    PRICE_CLASS_200 = 2;
    PRICE_CLASS_ALL = 3;
  }
  PriceClass price_class = 3 [(buf.validate.field).enum.defined_only = true];

  option (buf.validate.message).cel = {
    id: "aliases_require_cert",
    message: "certificate_arn must be set when aliases are provided",
    expression: "this.aliases.size() == 0 || this.certificate_arn != \"\""
  };
}
```

## Default Field Options (Separate from Validation)

Default values use a **separate** field option from buf.validate rules:

### Import

```proto
import "org/openmcf/shared/options/options.proto";
```

### Syntax

When a field should have a default value:
1. Mark as `optional`
2. Add `(org.openmcf.shared.options.default)` option

```proto
// Default: ghcr.io/actions/actions-runner
optional string repository = 1 [(org.openmcf.shared.options.default) = "ghcr.io/actions/actions-runner"];

// Default: 443
optional int32 port = 2 [(org.openmcf.shared.options.default) = "443"];
```

### Combining with Validation

Default options can be combined with buf.validate rules:

```proto
optional string namespace = 1 [
  (org.openmcf.shared.options.default) = "external-dns",
  (buf.validate.field).string.min_len = 1
];
```

### Build Enforcement

The `DEFAULT_REQUIRES_OPTIONAL` linter fails builds if `(org.openmcf.shared.options.default)` is set without `optional` keyword.

### When to Use Defaults vs Validation

- **Default option**: When OpenMCF should automatically provide a sensible value
- **Validation**: When there are constraints on what values are acceptable

Example:
```proto
// Has default AND validation
optional string image_tag = 1 [
  (org.openmcf.shared.options.default) = "2.331.0",
  (buf.validate.field).string.min_len = 1
];
```

## Notes
- Prefer pragmatic rules with low false positives.
- Ensure compatibility with protovalidate-go (uses `buf/validate/validate.proto`).
- **Every CEL message is a user-facing string.** It will be rendered in UIs, CLIs, and API error responses. Write it like a colleague would say it, not like a compiler would emit it. See the "Validation Message Philosophy" section above -- this is non-negotiable.
