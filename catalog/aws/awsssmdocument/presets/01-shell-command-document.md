# Shell Command Document

This preset creates the workhorse Command document: a parameterized
shell script Systems Manager runs on managed EC2 nodes.

## When to Use

- Fleet operations expressed as scripts — install agents, rotate
  certs, collect diagnostics
- Anywhere you would SSH-and-run, done instead through SSM's audited,
  IAM-gated channel

## What You Get

- A schema-2.2 Command document with a declared, defaulted parameter
  callers can override per run
- A target-type restriction so the document only runs against EC2
  instances

## Customize

- Add `versionName` labels per release (immutable forever — treat them
  like git tags)
- Schedule it against tagged instances with
  [AWS SSM Association](/infra-catalog/aws-ssm-association) or run it
  in a window with
  [AWS SSM Maintenance Window](/infra-catalog/aws-ssm-maintenance-window)
- Share it to sibling accounts with `shareWithAccountIds`
