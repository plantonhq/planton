# Interconnect Backed Gateway

HA VPN over Cloud Interconnect: the gateway's two interfaces are pinned to
two encrypted VLAN attachments on an encrypted-interconnect router, so
traffic that would cross a Dedicated or Partner Interconnect in the clear
is IPsec-encrypted end to end.

## What it configures

- `vpnInterfaces` — interface 0 and 1 bound to two Interconnect attachments
  in the gateway's region (encrypted attachments, `encryption: IPSEC`).
  The interfaces get no public IPs.
- `router.encryptedInterconnectRouter: true` — required for pinned
  interfaces; the router cannot be converted later.
- `asn: 64520` — the router's private ASN.

## Adjust before deploying

- **Both `interconnectAttachment` paths** — your attachments' resource
  paths, same region as the gateway.
- **`network.valueFrom.name`**, **`region`**, **`asn`** — yours.

## When to choose something else

Without Cloud Interconnect, use **Internet Facing Gateway**: the spec
refuses pinned interfaces on a non-encrypted router and vice versa, so the
two shapes cannot be mixed by accident.
