# Pulumi Module: DigitalOcean Project

Provisions a DigitalOcean project -- the complete `digitalocean_project` resource surface, at 100% behavioral parity with the Terraform module (same arguments, same outputs).

## Layout

- `main.go` -- entrypoint (`package main`), loads the stack input and calls the module
- `module/main.go` -- orchestration: locals, provider, resource
- `module/project.go` -- the `Project` resource and output exports
- `module/locals.go` -- target handle (a project has no tag surface, so no label set applies)
- `module/outputs.go` -- output key constants (the `DigitalOceanProjectStackOutputs` contract)

## Behavior notes

- Optional strings are set only when non-empty so the provider's defaults apply (purpose defaults to "Web Application" upstream).
- Membership references are resolved to literal URNs before the module runs; an empty list stays unset (membership unmanaged).
- `owner_id` is exported as a string (the SDK surfaces an integer; the outputs contract is engine-identical).
- `resource_urns` is the SDK's read-back of membership, sorted before export so both engines emit the same list from the API's unordered set.
- The resource carries `CustomTimeouts{Delete: "10m"}` (twin of the Terraform module's `timeouts` block): destroy relocates members to the default project and the provider retries the delete through the API's 412 while the asynchronous moves settle, which were measured to outlast its 3-minute default once. A destroy that still fails on the 412 has moved the members already and succeeds when run again.
