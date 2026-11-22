# A Network Endpoint Group's VM Endpoint Names Its Address, Linked to the Instance

**Date**: November 22, 2026
**Type**: Fix (breaking change on a `v1alpha1` field)
**Components**: `GcpNetworkEndpointGroup`

## Summary

**A VM endpoint states its IP address, and the address can come from the instance.** Google fills a missing VM endpoint address with the instance's primary internal IP. An endpoint declared without one then planned a change on every apply, and detaching it sent the empty address back, which Google could not match to the endpoint it held, so the endpoint was never removed. `endpoints[].ipAddress` is now a value-or-reference whose default is a `GcpComputeInstance`'s `internal_ip` output, and every endpoint of a `GCE_VM_IP` or `GCE_VM_IP_PORT` group must name it.

## Breaking change

### `GcpNetworkEndpointGroup.spec.endpoints[].ipAddress` takes a value or a link

```yaml
# Before
endpoints:
  - instance:
      valueFrom: {kind: GcpComputeInstance, name: web-1, fieldPath: status.outputs.instance_name}
    ipAddress: 10.0.1.5
# After: a link to the instance's primary internal IP
endpoints:
  - instance:
      valueFrom: {kind: GcpComputeInstance, name: web-1, fieldPath: status.outputs.instance_name}
    ipAddress:
      valueFrom: {kind: GcpComputeInstance, name: web-1, fieldPath: status.outputs.internal_ip}
# After: a literal (an alias IP, or a hybrid or internet address)
endpoints:
  - ipAddress:
      value: 192.168.10.11
```

A VM endpoint that left `ipAddress` unset is now refused before deploy (`vm_endpoints_name_an_ip_address`). The rules that read the address (`endpoint_address_or_fqdn`, `address_endpoints_name_an_ip`) read the reference's content, and a literal must still be an IPv4 or IPv6 address. The Terraform variable stays a string: the CLI flattens the reference before the module runs, as it does for `instance`.

## Verification

- The kind's spec tests, with new cases for the required VM address, a referenced address, and a referenced address beside an FQDN.
- The Pulumi module builds; both modules read the resolved string.
- `planton validate-manifest` on the three presets, the reference manifest and the scenarios; `planton validate-refs --check` resolves `status.outputs.internal_ip` on `GcpComputeInstance`.
- `make generate-reference` regenerated the reference pages, the foreign-key graph and the documentation index; `variables.tf` regenerated and the drift gate is green.
