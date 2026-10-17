# Router with TLS

The router serves HTTPS only, from a certificate that a cert-manager `Certificate` (the catalog's **Certificate** kind) writes into `agent-sandbox-system`, with TLS 1.3 as the floor and bearer tokens required. The `router_endpoint` output becomes the `https://` address on port 8443.

Pick this when traffic to sandboxes must be encrypted inside the cluster, or when a Gateway in front of the router re-encrypts to it.

A **Namespace** resource named `agent-sandbox-system` owns the namespace here, not this resource. That is what lets the order work: the namespace first, then the Certificate issued into it, then this controller, whose router mounts the Certificate's Secret as it starts. It also means destroying the controller never deletes the certificate.

Change first:
- Point `certificateSecret` at your Certificate. Its Secret must live in `agent-sandbox-system` and carry `tls.crt` and `tls.key`.
- Add `mtlsMode` with `clientCa` to require client certificates.
- Keep `httpEnabled` true if some callers still use plain HTTP while you move them.

See [04-router-tls.yaml](./04-router-tls.yaml) for the manifest.
