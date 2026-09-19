# GcpHaVpnConnection Guide

The judgment this guide protects: a connection is one site's worth of
tunnels and BGP sessions on a gateway that outlives it. Declare one per
site, build it for failover from the first day, and never rotate a key
in place.

## One connection per peer

A hub gateway with five branch offices is one `GcpHaVpnGateway` and
five `GcpHaVpnConnection`s. Each connection owns its tunnels, its BGP
sessions on the gateway's router, and -- for an external device -- the
external VPN gateway resource that holds the device's addresses. Adding
a site touches nothing but its own connection; removing one deletes
only its tunnels.

## Two tunnels, two interfaces, two addresses

Google's 99.99% SLA needs a tunnel from EACH gateway interface to a
redundant peer address. With an external device that means
`redundancyType: TWO_IPS_REDUNDANCY` and two tunnels:
`vpnGatewayInterface: 0` to device interface 0, `vpnGatewayInterface: 1`
to device interface 1. Four addresses (`FOUR_IPS_REDUNDANCY`) take four
tunnels, two per gateway interface. A single-address device
(`SINGLE_IP_INTERNALLY_REDUNDANT`) still gets two tunnels from the two
gateway interfaces to the one address -- 99.9%. Both sessions advertise
the same routes, so a failed tunnel loses traffic only until BGP notices
(three keepalives, 60 s by default) or BFD notices (about a second).

## Every tunnel needs its own /30

Each BGP session runs on a link-local /30 from 169.254.0.0/16:
`interfaceIpRange: 169.254.10.1/30` makes Google's end `169.254.10.1`
and the device's end `169.254.10.2`. Every session on a router needs a
different /30, and it must not overlap the router's `identifierRange`.
Pick a scheme (`169.254.{site}.{tunnel*4+1}/30`) and keep it in the
chart; the device is configured with the mirror addresses.

## Keys rotate by add-then-remove

The pre-shared key is immutable, like nearly everything on a tunnel: a
new key recreates the tunnel and drops its traffic for the duration.
Rotate by adding a third tunnel with the new key, waiting for its BGP
session to establish, then removing the old one -- the other tunnel
carries traffic throughout. The same motion applies to a cipher change
or an interface re-pairing. Wire keys from a secrets manager
(`${secrets-group.<name>.<key>}`); they are `sensitive` and never appear
in outputs or logs.

## Google to Google

Two VPCs connect over HA VPN by each declaring a gateway and a
connection with `peer.gcpGateway` pointing at the other's gateway.
Google pairs interfaces itself (interface 0 to interface 0), so the
tunnels leave `peerExternalGatewayInterface` empty; the same pre-shared
key is declared on both sides for each tunnel pair; each side's
`peerAsn` is the other side's router ASN. Neither side needs the other
to exist first, which is why the gateway is its own block.

## Advertisement, priority, and policies

The session inherits the router's advertisement (`DEFAULT` = all
subnets). Override per peer with `advertiseMode: CUSTOM` and the ranges
this site should learn. `advertisedRoutePriority` is the MED the peer
sees: give the two tunnels different values to make the device prefer
one (active/passive), or leave them equal for active/active. Route
policies (`importPolicies` / `exportPolicies`) are Cloud Router
objects created outside this block and named here.

## BFD and MD5

`bfd.sessionInitializationMode: ACTIVE` with the defaults (1000 ms,
multiplier 5) detects a dead tunnel in about five seconds instead of
sixty; the device must support BFD. `md5AuthenticationKey` guards the
BGP session against a misconfigured peer (IPsec already guards it
against an attacker); each session declares its own key because Google
requires a key to be used by exactly one peer.

## What it costs

Each tunnel bills per hour from the moment it exists, up or down, plus
the traffic that leaves Google through it at internet egress rates. Two
tunnels are two tunnel-hours per hour; four are four. Idle tunnels are
not free.
