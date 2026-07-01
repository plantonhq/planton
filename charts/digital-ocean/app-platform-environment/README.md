# DigitalOcean App Platform Environment InfraChart

The **DigitalOcean App Platform Environment** InfraChart provisions all
cloud resources required to run a containerised service on
DigitalOcean App Platform—optionally securing traffic with Let’s Encrypt
TLS on your own domain.

Like the ACS / GKE charts, it uses **Jinja‑templated manifests** so you
can turn HTTPS support, autoscaling and other features on or off with a
single flag in`values.yaml`.

---

## What gets created <a id="included-resources"></a>

| Resource                                     | Always created | Controlled by flag |
|----------------------------------------------|----------------|--------------------|
| **DigitalOcean Container Registry**          | ✅              | —                  |
| **DigitalOcean DNS Zone**                    | ✅              | —                  |
| **DigitalOcean App Platform Service**        | ✅              | —                  |
| **DigitalOcean Certificate (Let’s Encrypt)** | ❌              | `httpsEnabled`     |

> **Why a separate `DigitalOceanCertificate`?**  
> App Platform will happily auto‑issue Let’s Encrypt certs when you map a
> verified domain. Modelling it explicitly in Planton lets you:
> * audit expiry,
> * reference the cert ARN/ID elsewhere,
> * and control renewal through the usual StackJob lifecycle.

---

## Quick start

```bash
# 1. Fork or pull the chart
git clone https://github.com/plantonhq/planton.git
cd planton/charts/digital-ocean/app-platform-environment

# 2. Copy values and tweak for your environment
cp values.yaml values.dev.yaml
$EDITOR values.dev.yaml

# 3. Render → apply
planton chart build --values values.dev.yaml \
                         --org acme \
                         --env dev

