# Auth0CustomDomain — Terraform Module

Terraform/OpenTofu module that creates a custom domain for an Auth0 tenant.

## What It Creates

- `auth0_custom_domain` — the custom domain, in the tenant the provider's credential belongs to. The optional settings (client-IP header, TLS policy, relying party, metadata) are sent only when the spec sets them. Changing the domain or its type replaces it; destroy deletes it.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `create:custom_domains`, `read:custom_domains`, `update:custom_domains` and `delete:custom_domains`.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `domain` and `type` (required); `custom_client_ip_header`, `tls_policy`, `domain_metadata`, `relying_party_identifier` (optional) |

## Outputs

| Name | Description |
|---|---|
| `id` | The custom domain's identifier in Auth0 (`cd_...`) |
| `domain` | The custom domain's name |
| `status` | Where the domain is in its life (`pending_verification`, `ready`, ...) |
| `origin_domain_name` | The tenant host the domain serves from |
| `dns_record_name` | The name of the DNS record that proves control of the domain |
| `dns_record_type` | Its type (`CNAME` or `TXT`) |
| `dns_record_value` | Its value |

`locals.tf` picks the record from Auth0's verification methods: the CNAME when offered, otherwise the first method -- the same rule as the Pulumi module.
