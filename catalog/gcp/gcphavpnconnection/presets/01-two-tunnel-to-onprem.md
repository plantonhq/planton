# Two Tunnel To Onprem

Google's recommended 99.99% topology to an on-premises site with two
public addresses: one tunnel from each gateway interface to each address,
two BGP sessions with BFD, protected against destroy.

## What it configures

- `gateway`, `router`, `region` — three references to the same
  `GcpHaVpnGateway`, so the connection can never name a region or router
  its gateway is not in.
- `peer.externalGateway` — `TWO_IPS_REDUNDANCY` with the device's two
  addresses as interfaces 0 and 1.
- `tunnels` — `hq-tunnel-0` from gateway interface 0 to device interface 0,
  `hq-tunnel-1` from 1 to 1; each with its own link-local /30 and a BGP
  session to the device's ASN; BFD `ACTIVE` for ~1 s failure detection.
- `sharedSecret` — wired from a secrets manager, never written into the
  manifest.
- `deletionPolicy: PREVENT` — a production site link.

## Adjust before deploying

- **`gateway`/`router`/`region` `valueFrom.name`** — your gateway's manifest
  name.
- **The two `ipAddress`es**, **`peerAsn`**, and the two secrets — from the
  on-premises team; configure the device toward the gateway's
  `interface_0_ip_address` / `interface_1_ip_address` outputs with
  `169.254.10.2` and `169.254.11.2` as its BGP addresses and the gateway's
  `router_asn` as its neighbor ASN.

## When to choose something else

A device with four addresses uses **Four Tunnel To Onprem**; another
Google Cloud VPC uses **Gcp To Gcp**.
