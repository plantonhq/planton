# Pulumi Module: DigitalOcean Database Firewall

Provisions the inbound trusted-sources rule set of a DigitalOcean managed database cluster -- the complete `digitalocean_database_firewall` resource surface, at 100% behavioral parity with the Terraform module (same fan-out, same outputs).

## Layout

- `main.go` -- entrypoint (`package main`), loads the IaC input and calls the module
- `module/main.go` -- orchestration: locals, provider, resource
- `module/database_firewall.go` -- the typed-list fan-out and the `DatabaseFirewall` resource
- `module/locals.go` -- target handle (the firewall's tag rules TARGET tags, they don't apply them -- no label set)
- `module/outputs.go` -- output key constants (the `DigitalOceanDatabaseFirewallOutputs` contract)

## Behavior notes

- Rule type tokens are the PROVIDER's values (`ip_addr`, `droplet`, `k8s`, `app`, `tag`) -- the bridged SDK's doc comment renders "ipAddr", but values pass through untranslated and the camelCase spelling would be rejected.
- Destroy PUTs an EMPTY rule set (the cluster then accepts connections from anywhere); there is no object to 404 on.
