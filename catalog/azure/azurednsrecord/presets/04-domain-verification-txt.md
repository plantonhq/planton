# Domain Verification TXT

The ownership-proof TXT record that custom-domain flows require before they bind a hostname: Container Apps checks `asuid.{host}`, Front Door checks `_dnsauth.{host}`, and most SaaS domain verifications follow the same pattern.

## When to Use

- Binding a custom domain to a Container App: publish the app's `custom_domain_verification_id` output at `asuid.{host}` BEFORE deploying the AzureContainerAppCustomDomain binding
- Validating a Front Door custom domain: publish the domain's `validation_token` output at `_dnsauth.{host}`
- Any provider-issued domain-ownership token (Google, Microsoft 365, certificate authorities)

## Key Configuration Choices

- `name` -- the service defines it; underscore-led and dotted names are fully supported
- `ttlSeconds: 60` -- validation services poll public DNS; a short TTL makes the proof visible fast
- The token can be referenced instead of pasted -- e.g. `valueFrom` an `AzureContainerApp`'s `custom_domain_verification_id` output

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-resource-group>` | The zone's resource group | `AzureResourceGroup.status.outputs.resource_group_name` |
| `example.com` | Replace with the zone name | `AzureDnsZone.status.outputs.zone_name` |
| `asuid.app` | Replace `app` with the hostname being verified, relative to the zone | The custom-domain binding you are creating |
| `<your-custom-domain-resource-name>` | The AzureFrontDoorCustomDomain whose `validation_token` output the TXT record carries (for another service, replace the reference with its token: `AzureContainerApp.status.outputs.custom_domain_verification_id`, or a literal `value` from the service's setup screen) | Your Front Door composition |

## Related Presets

- `03-mail-mx-records` -- pair with SPF/DKIM/DMARC TXT records for mail
