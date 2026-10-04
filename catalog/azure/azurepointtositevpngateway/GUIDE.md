# Azure Point-to-Site VPN Gateway -- Operational Guide

Judgment that does not fit in field references.

## This is the expensive, slow object -- the policy next to it is free

The gateway bills from creation (per scale unit, whether anyone
connects) and creates in 30-45 minutes; the VPN Server Configuration
it points at is free and edits in seconds. Design accordingly: iterate
on the CONFIGURATION (auth methods, certificates, policy groups) and
touch the gateway only for capacity and pools. Both of its references
(`virtualHubId`, `vpnServerConfigurationId`) replace the gateway when
changed -- pick them like you mean them.

## Size the pool before the first client connects, not after

Azure hands each connected device an address from the pool, and a /24
serves roughly 250 concurrent connections BEFORE scale-unit math
matters. Growing the pool updates in place, but the pool must never
overlap the hub, its spokes, or anything reachable through them --
carve client space out of a range your network plan reserves for it
(nothing routable uses this kind's example `172.16.201.0/24`
by convention).

## Split tunnel vs forced tunnel is a per-pool product decision

`internetSecurityEnabled: false` (the default) is split tunneling:
clients reach hub-connected networks over the tunnel and the internet
locally -- cheap, fast, and invisible to users. `true` advertises
0.0.0.0/0 into the tunnel: every packet rides into the hub, which only
makes sense WITH an inspection point there (a hub firewall via routing
intent) and with the egress bandwidth bill accepted.

## One gateway per hub -- capacity is scale units, not gateway count

ARM allows one point-to-site gateway per hub (the slot is separate
from the hub's site-to-site VPN gateway). More users means more
`scaleUnit` (updates in place, 500 connections each) or more hubs --
never a second gateway in the same hub. Regional expansion is a new
hub with its own gateway sharing the same server configuration.

## Deleting the gateway strands nobody gracefully

Every connected client drops immediately, and the 30-45 minute
recreate is your recovery window if the delete was a mistake. Drain
expectations accordingly: point-to-site outages are user-visible the
second the gateway goes, unlike most hub-side changes.
