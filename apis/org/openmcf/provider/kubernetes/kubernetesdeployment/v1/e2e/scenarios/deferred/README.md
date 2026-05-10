# Deferred Scenarios

Scenarios in this directory are **valid test cases** that are temporarily
excluded from the E2E test suite due to known limitations.

The E2E scenario scanner only picks up `*.yaml` files directly in the
`scenarios/` directory. Files in subdirectories like `deferred/` are
automatically skipped -- no code changes or skip annotations needed.

Each deferred scenario file includes a comment block explaining:
- What the limitation is
- When it was discovered
- What needs to change to re-enable it

To re-enable a scenario, move it back to the parent `scenarios/` directory.
