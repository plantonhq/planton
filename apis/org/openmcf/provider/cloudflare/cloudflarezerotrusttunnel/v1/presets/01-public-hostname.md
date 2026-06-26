# Preset: Publish a private app on a public hostname

Expose a single private web app at a public hostname through the tunnel — the most common
Cloudflare Tunnel setup.

## When to use

- You have an internal HTTP service and want it reachable at `app.example.com` without
  opening inbound ports.

## Key choices

- `ingress[].service`: the local address of your app (e.g. `http://localhost:8080`).
- The trailing `service: http_status:404` rule is the required catch-all.
- After applying, CNAME `app.example.com` to `status.outputs.tunnel_cname` with a
  `CloudflareDnsRecord`, and run the connector with `status.outputs.tunnel_token`.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
