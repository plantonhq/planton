# GcpHaVpnGateway Guide

The judgment this guide protects: the gateway is the thing the other
side is configured against, so it is declared once and changed rarely.
Sites come and go as connections; the gateway stays.

## Why the gateway and the connection are two blocks

An HA VPN is a gateway with two public addresses plus, per site, a set
of tunnels and BGP sessions. Bundling them would have made
Google-to-Google VPN impossible to declare: each side's tunnels must
name the other side's gateway, so two bundled blocks would each need
the other to exist first. Split, each VPC declares its gateway, then a
`GcpHaVpnConnection` pointing at the other's gateway -- no cycle. The
same split lets a hub add its fifth branch office without touching the
tunnels of the first four.

## The router rides with the gateway

Every tunnel needs a BGP session, every session needs a Cloud Router,
and Google allows many routers per network and region -- so the gateway
brings its own, exactly as `GcpRouterNat` brings its own for NAT. The
router's ASN is the one number the on-premises side configures as its
neighbor; pick a private ASN (64512-65534, or 4200000000 and up) that
differs from every peer's, because it is fixed for the router's life.
The default advertisement (`DEFAULT` = all subnets, or `CUSTOM` with
listed ranges) is what every session inherits; a connection's session
can override it per peer.

## Almost nothing changes in place

Network, region, IP version, stack type, interface pinning, the
router's ASN: all recreate the resource. A recreated gateway has NEW
public IPs, which means every device on every site must be updated.
Treat the gateway spec as write-once; labels and the router's
description and advertisement are the mutable surface.

## Two interfaces, two tunnels, 99.99%

The gateway always has two interfaces. Google's 99.99% SLA applies when
each interface carries a tunnel to a redundant peer address
(`TWO_IPS_REDUNDANCY` on the connection, or a Google peer, which always
has two). One tunnel from one interface is 99.9% at best and a single
point of failure on Google's side. The connection's presets show both
shapes.

## Over the internet or over Interconnect

Leave `vpnInterfaces` empty and the interfaces get public IPs. Pin
them to encrypted Cloud Interconnect VLAN attachments (and set
`router.encryptedInterconnectRouter: true`) and the same gateway
encrypts traffic that would otherwise cross a Dedicated or Partner
Interconnect in the clear. The two are different products with the
same block; the spec's rule refuses the mix.

## Destroy order

Google refuses to delete a gateway while tunnels reference it. A chart
whose connections reference the gateway destroys them first; a
hand-managed estate must. `deletionPolicy: PREVENT` guards the gateway
every site's device is configured against.

## What it costs

The gateway is free. Google bills each tunnel per hour from the moment
it exists and bills tunnel traffic as internet egress -- both live on
the connection, which is where the tunnels are declared.
