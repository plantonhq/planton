# DigitalOcean Certificate -- Operational Guide

Judgment calls that matter when you run SSL certificates on DigitalOcean.

## Reference certificates by name, never by UUID

DigitalOcean auto-renews Let's Encrypt certificates by issuing a NEW certificate object with a new UUID; only the name survives. Anything holding the UUID — a script, a hand-wired API call — silently goes stale after the first renewal (typically within 90 days). The platform's wiring already does the right thing: the `certificate_id` output carries the name, and load balancer forwarding rules consume it. Keep any out-of-band automation on the same rule.

## Let's Encrypt is a package deal with DigitalOcean DNS

DigitalOcean validates domain ownership through zones it hosts in the SAME account. A domain managed at Cloudflare, Route 53, or another DigitalOcean account fails issuance — no challenge records get planted. Before requesting a certificate, make sure the `DigitalOceanDnsZone` for every listed domain exists in this account and the registrar delegates to DigitalOcean's nameservers. Wildcards (`*.example.com`) work under the same condition.

## Custom certificates rotate by re-apply, and the timing is yours

There is no renewal machinery for uploaded certificates: DigitalOcean serves exactly what you gave it until it expires. Rotation is editing the manifest with the new PEM material and re-applying — every field is create-only, so this replaces the certificate, and the replacement is created BEFORE the old one is destroyed (consumers referencing the name never observe a gap). Watch the `expiry_rfc3339` output; alert well before it, because an expired certificate fails loudly in every client.

## The private key goes in, never comes out

The DigitalOcean API never returns certificate material, and provisioner state stores only hashes of what was written. Two consequences worth internalizing: an imported certificate always shows empty PEM fields (that is fidelity, not data loss), and there is no "download my key from DigitalOcean" recovery path — the manifest (or the secret store feeding it) is the only home your key has. Treat the manifest's `privateKey` value with secret discipline end to end.

## Adopting an existing custom certificate replaces it once

Because the PEM fields come back empty on import and every one of them is create-only, the first apply after importing a custom certificate sees the manifest supplying material the state does not have and replaces the certificate — created before the old one is destroyed, so consumers referencing the name see no gap. That single replacement is correct behavior, not drift: rotating certificate material is supposed to replace the certificate, and the modules deliberately do not suppress it. Expect it, apply it in a window you chose, and the manifest and DigitalOcean agree from then on. Let's Encrypt certificates carry no PEM in the manifest and adopt with no change at all.

## Deletion can stall behind a load balancer

DigitalOcean refuses to delete a certificate while a load balancer still references it, and the provisioner retries deletion for several minutes waiting for the reference to clear. When restructuring HTTPS termination, update or destroy the load balancer's forwarding rule first, then the certificate — the reverse order looks hung.
