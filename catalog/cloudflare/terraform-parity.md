---
title: "Terraform Parity"
description: "Measured parity of the Cloudflare catalog against the pinned Terraform provider"
icon: "check-circle"
order: 90
---

<!-- GENERATED FILE -- DO NOT EDIT.
     Rendered from the committed provider schemas, the kind registry, the
     Terraform modules, the per-kind provider-parity manifests, the
     dispositions ledger, and the E2E profiles.
     parameters: provider=cloudflare ga-schema=cloudflare
     Regenerate: make generate-provider-parity-report -->

# Cloudflare Terraform Parity

This catalog is **built for 100% Terraform parity**: every configurable
argument of the pinned Terraform provider is representable through a kind,
and every provider resource carries exactly one recorded disposition --
omission is a decision, never an accident. This page is the measurement,
generated from the same accounting that gates the repository's CI. It makes
no achieved-parity claim: a kind counts as PROVEN only when live end-to-end
runs pass on both IaC engines, and the tables below show exactly how far
that has progressed.

## Measurement baseline

| | |
|---|---|
| Provider schema (parity baseline) | `cloudflare@5.23.0` |
| Supporting schema (pinned by this catalog's modules) | `aws@6.58.0` |
| Kinds in the catalog | 66 |
| Distinct provider resources consumed | 113 |
| Spec fields authored across all kinds | 2092 |
| Module pins on `aws` | `~> 5.0` × 1 |
| Module pins on `cloudflare` | `~> 5.23` × 66 |
| Module pins on `tls` | `~> 4.0` × 1 |

The GA provider is the parity baseline. Capability that exists only in a
secondary channel (for Google, the `google-beta` provider) enters per kind
through the admission list (`pkg/providerparity/admissions/`), never
wholesale: one entry per resource per kind, with the reason and where its
promotion to the baseline is tracked. The accounting reads the list -- an
admitted resource is measured against its channel's schema, an unadmitted
secondary-channel resource is a finding, and an admitted resource the
baseline serves at the pin is a stale admission.

## The provider block

Parity covers the provider's own configuration block -- credentials,
role-assumption chains, default tags, endpoint overrides, retry tuning --
under the same total-accounting rule as resources: every configurable,
non-deprecated provider-block argument is matched to a provider-config
field, mapped by recorded judgment, owned by the modules by recorded
judgment, or excluded with a recorded reason -- and arguments set inside
catalog modules' own provider blocks must carry that judgment too.

| Provider-block args | Matched | Mapped | Module-owned | Excluded | Open gaps | Accounted |
|---|---|---|---|---|---|---|
| 6 | 3 | 0 | 0 | 3 | 0 | ✅ |

## Depth: per-kind accounting

Every configurable, non-deprecated provider argument of a kind's consumed
resources must be matched to a spec field, mapped by recorded judgment, or
excluded with a recorded reason -- and every spec field must reach provider
surface. **Accounted** means both directions hold with zero unexplained
gaps. **Proven** means live end-to-end runs passed on both IaC engines.

**66 of 66 kinds are at total accounting; 58 proven live.**

| Kind | Provider args | Matched | Mapped | Excluded | Open gaps | Accounted | Proven |
|---|---|---|---|---|---|---|---|
| CloudflareAccountApiToken | 7 | 5 | 2 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareAiGateway | 27 | 15 | 12 | 0 | 0 | ✅ | — |
| CloudflareAuthenticatedOriginPulls | 4 | 2 | 2 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareAuthenticatedOriginPullsCertificate | 6 | 6 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareBotManagement | 16 | 16 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareCacheSettings | 12 | 6 | 6 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareCertificatePack | 7 | 7 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareCustomHostname | 6 | 5 | 1 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareCustomHostnameFallbackOrigin | 2 | 2 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareCustomSslCertificate | 9 | 8 | 1 | 0 | 0 | ✅ | — |
| CloudflareD1Database | 5 | 2 | 3 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareDnsRecord | 12 | 10 | 2 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareDnsZone | 38 | 21 | 12 | 5 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareEmailRoutingAddress | 3 | 3 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareEmailRoutingRule | 8 | 4 | 2 | 2 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareEmailRoutingZone | 10 | 3 | 4 | 3 | 0 | ✅ | — |
| CloudflareHealthcheck | 14 | 11 | 2 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareHyperdriveConfig | 6 | 3 | 3 | 0 | 0 | ✅ | — |
| CloudflareIpAccessRule | 5 | 4 | 1 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareKvNamespace | 2 | 1 | 1 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareList | 5 | 4 | 0 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareListItem | 7 | 5 | 2 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareLoadBalancer | 20 | 11 | 9 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareLoadBalancerMonitor | 17 | 16 | 1 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareLoadBalancerPool | 15 | 9 | 4 | 2 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareLogpushJob | 16 | 12 | 4 | 0 | 0 | ✅ | — |
| CloudflareMtlsCertificate | 5 | 5 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareNotificationPolicy | 8 | 6 | 2 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareNotificationWebhook | 4 | 4 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareOriginCaCertificate | 4 | 4 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflarePagesProject | 9 | 4 | 5 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareQueue | 9 | 5 | 2 | 2 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareR2Bucket | 34 | 22 | 12 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareRuleset | 7 | 5 | 2 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareSecretsStore | 2 | 2 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareSecretsStoreSecret | 6 | 6 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareSnippet | 4 | 2 | 2 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareSnippetRules | 2 | 1 | 1 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareTurnstileWidget | 9 | 9 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareWaitingRoom | 25 | 21 | 3 | 1 | 0 | ✅ | — |
| CloudflareWaitingRoomEvent | 17 | 17 | 0 | 0 | 0 | ✅ | — |
| CloudflareWebAnalyticsSite | 12 | 6 | 5 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareWorker | 40 | 10 | 20 | 10 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareWorkersKvPair | 5 | 5 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareWorkflow | 7 | 4 | 3 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustAccessApplication | 39 | 29 | 10 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustAccessGroup | 7 | 4 | 3 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustAccessIdentityProvider | 8 | 6 | 2 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustAccessInfrastructureTarget | 3 | 2 | 1 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustAccessPolicy | 14 | 8 | 6 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustAccessServiceToken | 6 | 6 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustDeviceCustomProfile | 29 | 21 | 6 | 2 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustDeviceDefaultProfile | 25 | 16 | 8 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustDevicePostureRule | 8 | 6 | 2 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustDnsLocation | 8 | 5 | 3 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustGatewayPolicy | 13 | 9 | 4 | 0 | 0 | ✅ | — |
| CloudflareZeroTrustGatewaySettings | 10 | 3 | 7 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustList | 5 | 4 | 1 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustMcpPortal | 7 | 6 | 1 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustMcpServer | 11 | 9 | 2 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustOrganization | 21 | 16 | 4 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustTunnel | 8 | 5 | 2 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustTunnelRoute | 5 | 5 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZeroTrustTunnelVirtualNetwork | 4 | 4 | 0 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZoneSettings | 16 | 5 | 9 | 2 | 0 | ✅ | ✅ pulumi, terraform |
| CloudflareZoneTlsSettings | 16 | 6 | 9 | 1 | 0 | ✅ | ✅ pulumi, terraform |

## Breadth: every GA resource, one disposition

All resources of `cloudflare@5.23.0` land in exactly one class:

| Disposition | Resources | Meaning |
|---|---|---|
| Modeled | 111 | consumed by a kind's Terraform module today |
| IAM-covered | 0 | per-resource IAM member/binding/policy triplets, covered by the owning kinds' additive `iam_members` fields |
| Composed | 0 | capability covered through an existing kind's surface rather than a kind of its own |
| Planned | 94 | judged to be covered by a planned kind or planned composition, not built yet |
| Deferred | 45 | deliberately not offered, each with the recorded reason |
| Excluded as deprecated | 7 | deprecated or superseded provider surface |
| **Total** | **257** | |

## The enumerated record

The full per-resource record, so the accounting above is verifiable
rather than trusted.

### Modeled (111)

| Resource | Consuming kinds |
|---|---|
| `cloudflare_access_rule` | consumed by CloudflareIpAccessRule |
| `cloudflare_account_token` | consumed by CloudflareAccountApiToken |
| `cloudflare_ai_gateway` | consumed by CloudflareAiGateway |
| `cloudflare_ai_gateway_dynamic_routing` | consumed by CloudflareAiGateway |
| `cloudflare_argo_smart_routing` | consumed by CloudflareCacheSettings |
| `cloudflare_argo_tiered_caching` | consumed by CloudflareCacheSettings |
| `cloudflare_authenticated_origin_pulls` | consumed by CloudflareAuthenticatedOriginPulls |
| `cloudflare_authenticated_origin_pulls_certificate` | consumed by CloudflareAuthenticatedOriginPullsCertificate |
| `cloudflare_authenticated_origin_pulls_hostname_certificate` | consumed by CloudflareAuthenticatedOriginPullsCertificate |
| `cloudflare_authenticated_origin_pulls_settings` | consumed by CloudflareAuthenticatedOriginPulls |
| `cloudflare_bot_management` | consumed by CloudflareBotManagement |
| `cloudflare_certificate_authorities_hostname_associations` | consumed by CloudflareZoneTlsSettings |
| `cloudflare_certificate_pack` | consumed by CloudflareCertificatePack |
| `cloudflare_custom_hostname` | consumed by CloudflareCustomHostname |
| `cloudflare_custom_hostname_fallback_origin` | consumed by CloudflareCustomHostnameFallbackOrigin |
| `cloudflare_custom_ssl` | consumed by CloudflareCustomSslCertificate |
| `cloudflare_d1_database` | consumed by CloudflareD1Database |
| `cloudflare_dns_record` | consumed by CloudflareDnsRecord, CloudflareDnsZone |
| `cloudflare_email_routing_address` | consumed by CloudflareEmailRoutingAddress |
| `cloudflare_email_routing_catch_all` | consumed by CloudflareEmailRoutingZone |
| `cloudflare_email_routing_dns` | consumed by CloudflareEmailRoutingZone |
| `cloudflare_email_routing_rule` | consumed by CloudflareEmailRoutingRule |
| `cloudflare_email_routing_settings` | consumed by CloudflareEmailRoutingZone |
| `cloudflare_healthcheck` | consumed by CloudflareHealthcheck |
| `cloudflare_hostname_tls_setting` | consumed by CloudflareZoneTlsSettings |
| `cloudflare_hyperdrive_config` | consumed by CloudflareHyperdriveConfig |
| `cloudflare_list` | consumed by CloudflareList |
| `cloudflare_list_item` | consumed by CloudflareListItem |
| `cloudflare_load_balancer` | consumed by CloudflareLoadBalancer |
| `cloudflare_load_balancer_monitor` | consumed by CloudflareLoadBalancerMonitor |
| `cloudflare_load_balancer_pool` | consumed by CloudflareLoadBalancerPool |
| `cloudflare_logpush_job` | consumed by CloudflareLogpushJob |
| `cloudflare_logpush_ownership_challenge` | consumed by CloudflareLogpushJob |
| `cloudflare_managed_transforms` | consumed by CloudflareZoneSettings |
| `cloudflare_mtls_certificate` | consumed by CloudflareMtlsCertificate |
| `cloudflare_notification_policy` | consumed by CloudflareNotificationPolicy |
| `cloudflare_notification_policy_webhooks` | consumed by CloudflareNotificationWebhook |
| `cloudflare_origin_ca_certificate` | consumed by CloudflareOriginCaCertificate |
| `cloudflare_origin_cloud_region` | consumed by CloudflareZoneSettings |
| `cloudflare_origin_tls_compliance_modes` | consumed by CloudflareZoneTlsSettings |
| `cloudflare_pages_domain` | consumed by CloudflarePagesProject |
| `cloudflare_pages_project` | consumed by CloudflarePagesProject |
| `cloudflare_queue` | consumed by CloudflareQueue |
| `cloudflare_queue_consumer` | consumed by CloudflareQueue |
| `cloudflare_r2_bucket` | consumed by CloudflareR2Bucket |
| `cloudflare_r2_bucket_cors` | consumed by CloudflareR2Bucket |
| `cloudflare_r2_bucket_event_notification` | consumed by CloudflareR2Bucket |
| `cloudflare_r2_bucket_lifecycle` | consumed by CloudflareR2Bucket |
| `cloudflare_r2_bucket_lock` | consumed by CloudflareR2Bucket |
| `cloudflare_r2_custom_domain` | consumed by CloudflareR2Bucket |
| `cloudflare_r2_managed_domain` | consumed by CloudflareR2Bucket |
| `cloudflare_regional_tiered_cache` | consumed by CloudflareCacheSettings |
| `cloudflare_ruleset` | consumed by CloudflareRuleset |
| `cloudflare_secrets_store` | consumed by CloudflareSecretsStore |
| `cloudflare_secrets_store_secret` | consumed by CloudflareSecretsStoreSecret |
| `cloudflare_snippet` | consumed by CloudflareSnippet |
| `cloudflare_snippet_rules` | consumed by CloudflareSnippetRules |
| `cloudflare_tiered_cache` | consumed by CloudflareCacheSettings |
| `cloudflare_total_tls` | consumed by CloudflareZoneTlsSettings |
| `cloudflare_turnstile_widget` | consumed by CloudflareTurnstileWidget |
| `cloudflare_universal_ssl_setting` | consumed by CloudflareZoneTlsSettings |
| `cloudflare_url_normalization_settings` | consumed by CloudflareZoneSettings |
| `cloudflare_waiting_room` | consumed by CloudflareWaitingRoom |
| `cloudflare_waiting_room_event` | consumed by CloudflareWaitingRoomEvent |
| `cloudflare_waiting_room_rules` | consumed by CloudflareWaitingRoom |
| `cloudflare_waiting_room_settings` | consumed by CloudflareZoneSettings |
| `cloudflare_web_analytics_rule` | consumed by CloudflareWebAnalyticsSite |
| `cloudflare_web_analytics_site` | consumed by CloudflareWebAnalyticsSite |
| `cloudflare_workers_cron_trigger` | consumed by CloudflareWorker |
| `cloudflare_workers_custom_domain` | consumed by CloudflareWorker |
| `cloudflare_workers_kv` | consumed by CloudflareWorkersKvPair |
| `cloudflare_workers_kv_namespace` | consumed by CloudflareKvNamespace |
| `cloudflare_workers_route` | consumed by CloudflareWorker |
| `cloudflare_workers_script` | consumed by CloudflareWorker |
| `cloudflare_workers_script_subdomain` | consumed by CloudflareWorker |
| `cloudflare_workflow` | consumed by CloudflareWorkflow |
| `cloudflare_zero_trust_access_ai_controls_mcp_portal` | consumed by CloudflareZeroTrustMcpPortal |
| `cloudflare_zero_trust_access_ai_controls_mcp_server` | consumed by CloudflareZeroTrustMcpServer |
| `cloudflare_zero_trust_access_application` | consumed by CloudflareZeroTrustAccessApplication |
| `cloudflare_zero_trust_access_group` | consumed by CloudflareZeroTrustAccessGroup |
| `cloudflare_zero_trust_access_identity_provider` | consumed by CloudflareZeroTrustAccessIdentityProvider |
| `cloudflare_zero_trust_access_infrastructure_target` | consumed by CloudflareZeroTrustAccessInfrastructureTarget |
| `cloudflare_zero_trust_access_key_configuration` | consumed by CloudflareZeroTrustOrganization |
| `cloudflare_zero_trust_access_policy` | consumed by CloudflareZeroTrustAccessPolicy |
| `cloudflare_zero_trust_access_service_token` | consumed by CloudflareZeroTrustAccessServiceToken |
| `cloudflare_zero_trust_device_custom_profile` | consumed by CloudflareZeroTrustDeviceCustomProfile |
| `cloudflare_zero_trust_device_custom_profile_local_domain_fallback` | consumed by CloudflareZeroTrustDeviceCustomProfile |
| `cloudflare_zero_trust_device_default_profile` | consumed by CloudflareZeroTrustDeviceDefaultProfile |
| `cloudflare_zero_trust_device_default_profile_certificates` | consumed by CloudflareZeroTrustDeviceDefaultProfile |
| `cloudflare_zero_trust_device_default_profile_local_domain_fallback` | consumed by CloudflareZeroTrustDeviceDefaultProfile |
| `cloudflare_zero_trust_device_posture_rule` | consumed by CloudflareZeroTrustDevicePostureRule |
| `cloudflare_zero_trust_dns_location` | consumed by CloudflareZeroTrustDnsLocation |
| `cloudflare_zero_trust_gateway_logging` | consumed by CloudflareZeroTrustGatewaySettings |
| `cloudflare_zero_trust_gateway_pacfile` | consumed by CloudflareZeroTrustGatewaySettings |
| `cloudflare_zero_trust_gateway_policy` | consumed by CloudflareZeroTrustGatewayPolicy |
| `cloudflare_zero_trust_gateway_settings` | consumed by CloudflareZeroTrustGatewaySettings |
| `cloudflare_zero_trust_list` | consumed by CloudflareZeroTrustList |
| `cloudflare_zero_trust_organization` | consumed by CloudflareZeroTrustOrganization |
| `cloudflare_zero_trust_tunnel_cloudflared` | consumed by CloudflareZeroTrustTunnel |
| `cloudflare_zero_trust_tunnel_cloudflared_config` | consumed by CloudflareZeroTrustTunnel |
| `cloudflare_zero_trust_tunnel_cloudflared_route` | consumed by CloudflareZeroTrustTunnelRoute |
| `cloudflare_zero_trust_tunnel_cloudflared_virtual_network` | consumed by CloudflareZeroTrustTunnelVirtualNetwork |
| `cloudflare_zone` | consumed by CloudflareDnsZone |
| `cloudflare_zone_auto_origin_tls_kex` | consumed by CloudflareZoneTlsSettings |
| `cloudflare_zone_cache_reserve` | consumed by CloudflareCacheSettings |
| `cloudflare_zone_cache_variants` | consumed by CloudflareCacheSettings |
| `cloudflare_zone_dns_settings` | consumed by CloudflareDnsZone |
| `cloudflare_zone_dnssec` | consumed by CloudflareDnsZone |
| `cloudflare_zone_hold` | consumed by CloudflareDnsZone |
| `cloudflare_zone_setting` | consumed by CloudflareZoneSettings |
| `cloudflare_zone_subscription` | consumed by CloudflareDnsZone |

### Planned (94)

| Resource | Recorded reason |
|---|---|
| `cloudflare_account` | judged as a planned CloudflareAccount kind (foundational for multi-account and MSP automation) |
| `cloudflare_account_dns_settings` | judged as a planned CloudflareAccountDnsSettings kind (account-level DNS settings singleton with apply/revert semantics) |
| `cloudflare_account_member` | judged as a planned CloudflareAccountMember kind (memberships with roles) |
| `cloudflare_account_subscription` | folds into the planned CloudflareAccount kind (the plan is an attribute of the account) |
| `cloudflare_ai_search_instance` | judged as a planned CloudflareAiSearchInstance kind (managed retrieval-augmented search instance) |
| `cloudflare_ai_search_namespace` | judged as a planned CloudflareAiSearchNamespace kind (grouping namespace referenced by search instances) |
| `cloudflare_ai_search_token` | folds into the planned CloudflareAiSearchInstance kind (endpoint access token of the instance) |
| `cloudflare_api_shield` | judged as a planned CloudflareApiShield kind (the zone's API Shield configuration umbrella: auth characteristics, validation defaults) |
| `cloudflare_api_shield_operation` | folds into the planned CloudflareApiShield kind (endpoint registrations are per-zone configuration rows meaningless outside the shield) |
| `cloudflare_api_shield_schema_validation_settings` | folds into the planned CloudflareApiShield kind (zone validation defaults) |
| `cloudflare_api_token` | judged as a planned CloudflareApiToken kind (user-scoped tokens; the API and ownership model differ from account tokens) |
| `cloudflare_calls_sfu_app` | judged as a planned CloudflareCallsSfuApp kind (realtime SFU application credentials provisioned per application) |
| `cloudflare_calls_turn_app` | judged as a planned CloudflareCallsTurnApp kind (TURN service credentials provisioned per application) |
| `cloudflare_client_certificate` | judged as a planned CloudflareClientCertificate kind (Cloudflare-issued client certificates for API Shield mTLS) |
| `cloudflare_cloud_connector_rules` | judged as a planned CloudflareCloudConnectorRules kind (zone-singleton ordered routing table to cloud storage origins) |
| `cloudflare_connectivity_directory_service` | judged as a planned CloudflareConnectivityDirectoryService kind (service directory entries referencing tunnels) |
| `cloudflare_content_scanning` | judged as a planned CloudflareContentScanning kind (malicious-upload scanning enablement plus expressions) |
| `cloudflare_content_scanning_expression` | folds into the planned CloudflareContentScanning kind (custom expression rows of the parent) |
| `cloudflare_custom_csr` | judged as a planned CloudflareCustomCsr kind (CSR generation referenced during certificate issuance) |
| `cloudflare_custom_origin_trust_store` | judged as a planned CloudflareCustomOriginTrustStore kind (uploaded CA bundle for origin verification) |
| `cloudflare_custom_page_asset` | folds into the planned CloudflareCustomPage kind (uploaded assets exist only to serve custom pages) |
| `cloudflare_custom_pages` | judged as a planned CloudflareCustomPage kind (one instance per error-page identifier with independent lifecycle) |
| `cloudflare_dns_firewall` | judged as a planned CloudflareDnsFirewall kind (standalone DNS firewall clusters with their own nameserver IPs and lifecycle) |
| `cloudflare_flagship_app` | judged as a planned CloudflareFlagshipApp kind (feature-flags application container) |
| `cloudflare_flagship_flag` | judged as a planned CloudflareFlagshipFlag kind (flags with independent lifecycle referenced by SDKs) |
| `cloudflare_google_tag_gateway` | judged as a planned CloudflareGoogleTagGateway kind (zone-singleton tag-gateway configuration) |
| `cloudflare_image_variant` | judged as a planned CloudflareImagesVariant kind (named transformation preset referenced by delivery URLs) |
| `cloudflare_leaked_credential_check` | judged as a planned CloudflareLeakedCredentialCheck kind (enablement plus custom detection rules in one kind) |
| `cloudflare_leaked_credential_check_rule` | folds into the planned CloudflareLeakedCredentialCheck kind (detection rows are meaningless without the check) |
| `cloudflare_load_balancer_monitor_group` | judged as a planned CloudflareLoadBalancerMonitorGroup kind (monitor groups referenced by pools) |
| `cloudflare_oauth_client` | judged as a planned CloudflareOauthClient kind (OAuth clients against Cloudflare as identity provider) |
| `cloudflare_observatory_scheduled_test` | judged as a planned CloudflareObservatoryScheduledTest kind (scheduled speed tests per page) |
| `cloudflare_organization` | judged as a planned CloudflareOrganization kind (organization hierarchy container) |
| `cloudflare_organization_profile` | folds into the planned CloudflareOrganization kind (the profile is settings of the organization) |
| `cloudflare_pipeline` | judged as a planned CloudflarePipeline kind (ingestion pipeline referencing sinks and streams) |
| `cloudflare_pipeline_sink` | judged as a planned CloudflarePipelineSink kind (pipeline destination, e.g. an R2 bucket) |
| `cloudflare_pipeline_stream` | judged as a planned CloudflarePipelineStream kind (pipeline stream input) |
| `cloudflare_r2_bucket_sippy` | folds into the existing CloudflareR2Bucket kind's planned depth expansion (incremental-migration setting of a bucket) |
| `cloudflare_r2_data_catalog` | folds into the existing CloudflareR2Bucket kind's planned depth expansion (per-bucket data catalog toggle) |
| `cloudflare_schema_validation_operation_settings` | folds into the planned CloudflareApiShield kind (per-operation validation overrides) |
| `cloudflare_schema_validation_schemas` | judged as a planned CloudflareSchemaValidationSchema kind (uploaded OpenAPI schema with its own upload and activation lifecycle) |
| `cloudflare_schema_validation_settings` | folds into the planned CloudflareApiShield kind (v2 zone validation defaults) |
| `cloudflare_share` | judged as a planned CloudflareShare kind (cross-account resource sharing) |
| `cloudflare_share_recipient` | folds into the planned CloudflareShare kind (recipient rows of the share) |
| `cloudflare_share_resource` | folds into the planned CloudflareShare kind (resource rows of the share) |
| `cloudflare_spectrum_application` | judged as a planned CloudflareSpectrumApplication kind (TCP/UDP proxying applications) |
| `cloudflare_stream_key` | judged as a planned CloudflareStreamSigningKey kind (signing key referenced by application code for signed playback URLs) |
| `cloudflare_stream_live_input` | judged as a planned CloudflareStreamLiveInput kind (live ingest endpoint with keys and independent lifecycle) |
| `cloudflare_stream_watermark` | judged as a planned CloudflareStreamWatermarkProfile kind (created once, referenced at every upload) |
| `cloudflare_stream_webhook` | judged as a planned CloudflareStreamWebhook kind (account-level notification webhook singleton) |
| `cloudflare_token_validation_config` | judged as a planned CloudflareTokenValidationConfig kind (JWT credential sets referenced by validation rules) |
| `cloudflare_token_validation_rules` | folds into the planned CloudflareApiShield kind (zone-level JWT rules spanning multiple credential configs) |
| `cloudflare_user_agent_blocking_rule` | judged as a planned CloudflareUserAgentBlockingRule kind (standalone user-agent blocking rule object) |
| `cloudflare_user_group` | judged as a planned CloudflareUserGroup kind (permission groups referenced by policies) |
| `cloudflare_user_group_members` | folds into the planned CloudflareUserGroup kind (membership rows of the group) |
| `cloudflare_worker` | folds into the existing CloudflareWorker kind's planned depth expansion (the modern script-settings surface of the worker the kind already owns) |
| `cloudflare_worker_version` | folds into the existing CloudflareWorker kind's planned depth expansion (versions are deployment artifacts of the worker, not independent objects) |
| `cloudflare_workers_deployment` | folds into the existing CloudflareWorker kind's planned depth expansion (gradual-deployment configuration of the worker) |
| `cloudflare_workers_for_platforms_dispatch_namespace` | judged as a planned CloudflareWorkersDispatchNamespace kind (Workers for Platforms dispatch namespace referenced by dispatch workers) |
| `cloudflare_zero_trust_access_custom_page` | judged as a planned CloudflareZeroTrustAccessCustomPage kind (branded block and login pages referenced by applications) |
| `cloudflare_zero_trust_access_mtls_certificate` | judged as a planned CloudflareZeroTrustAccessMtlsCertificate kind (client CA certificate with associated hostnames and independent lifecycle) |
| `cloudflare_zero_trust_access_mtls_hostname_settings` | folds into the planned CloudflareZeroTrustAccessMtlsCertificate kind (hostname settings exist in service of the mTLS certificates) |
| `cloudflare_zero_trust_access_short_lived_certificate` | folds into the existing CloudflareZeroTrustAccessApplication kind's planned depth expansion (SSH short-lived certificate issuance is a per-application toggle) |
| `cloudflare_zero_trust_access_tag` | judged as a planned CloudflareZeroTrustAccessTag kind (label referenced by name across applications) |
| `cloudflare_zero_trust_device_deployment_groups` | judged as a planned CloudflareZeroTrustDeviceDeploymentGroup kind (WARP client rollout rings with independent membership lifecycle) |
| `cloudflare_zero_trust_device_ip_profile` | judged as a planned CloudflareZeroTrustDeviceIpProfile kind (IP-scoped WARP profile variant) |
| `cloudflare_zero_trust_device_managed_networks` | judged as a planned CloudflareZeroTrustDeviceManagedNetwork kind (network fingerprints referenced by profile targeting) |
| `cloudflare_zero_trust_device_posture_integration` | judged as a planned CloudflareZeroTrustDevicePostureIntegration kind (third-party posture providers with credentials and lifecycle) |
| `cloudflare_zero_trust_device_settings` | judged as a planned CloudflareZeroTrustDeviceSettings kind (account-wide device enrollment settings singleton) |
| `cloudflare_zero_trust_device_subnet` | judged as a planned CloudflareZeroTrustDeviceSubnet kind (virtual device subnets referencing virtual networks) |
| `cloudflare_zero_trust_dex_rule` | folds into the planned CloudflareZeroTrustDexTest kind (targeting rules exist in service of the tests) |
| `cloudflare_zero_trust_dex_test` | judged as a planned CloudflareZeroTrustDexTest kind (synthetic experience tests) |
| `cloudflare_zero_trust_dlp_custom_entry` | folds into the planned CloudflareZeroTrustDlpCustomProfile kind (custom detection rows of the profile) |
| `cloudflare_zero_trust_dlp_custom_profile` | judged as a planned CloudflareZeroTrustDlpCustomProfile kind (custom detection profile) |
| `cloudflare_zero_trust_dlp_data_class` | folds into the planned CloudflareZeroTrustDlpSettings kind (classification taxonomy row) |
| `cloudflare_zero_trust_dlp_data_tag` | folds into the planned CloudflareZeroTrustDlpSettings kind (tag taxonomy row) |
| `cloudflare_zero_trust_dlp_data_tag_category` | folds into the planned CloudflareZeroTrustDlpSettings kind (tag taxonomy row) |
| `cloudflare_zero_trust_dlp_dataset` | judged as a planned CloudflareZeroTrustDlpDataset kind (exact-data-match and wordlist datasets with upload lifecycle) |
| `cloudflare_zero_trust_dlp_entry` | folds into the planned CloudflareZeroTrustDlpCustomProfile kind (generic entry rows inside a profile) |
| `cloudflare_zero_trust_dlp_integration_entry` | folds into the planned CloudflareZeroTrustDlpSettings kind (integration-sourced entries managed at account level) |
| `cloudflare_zero_trust_dlp_predefined_entry` | folds into the planned CloudflareZeroTrustDlpPredefinedProfile kind (enablement rows of predefined entries) |
| `cloudflare_zero_trust_dlp_predefined_profile` | judged as a planned CloudflareZeroTrustDlpPredefinedProfile kind (enablement and configuration of predefined profiles) |
| `cloudflare_zero_trust_dlp_sensitivity_group` | folds into the planned CloudflareZeroTrustDlpSettings kind (sensitivity taxonomy row) |
| `cloudflare_zero_trust_dlp_sensitivity_level` | folds into the planned CloudflareZeroTrustDlpSettings kind (sensitivity taxonomy row) |
| `cloudflare_zero_trust_dlp_sensitivity_level_order` | folds into the planned CloudflareZeroTrustDlpSettings kind (ordering of the sensitivity taxonomy) |
| `cloudflare_zero_trust_dlp_settings` | judged as a planned CloudflareZeroTrustDlpSettings kind (account-level DLP configuration including the sensitivity and tag taxonomy) |
| `cloudflare_zero_trust_gateway_certificate` | judged as a planned CloudflareZeroTrustGatewayCertificate kind (TLS inspection certificates with activation lifecycle) |
| `cloudflare_zero_trust_gateway_proxy_endpoint` | judged as a planned CloudflareZeroTrustGatewayProxyEndpoint kind (proxy endpoints with allowed-IP configuration) |
| `cloudflare_zero_trust_network_hostname_route` | judged as a planned CloudflareZeroTrustNetworkHostnameRoute kind (hostname-based private network routes referencing tunnels and virtual networks) |
| `cloudflare_zero_trust_risk_behavior` | judged as a planned CloudflareZeroTrustRiskScoring kind (account risk-behavior configuration) |
| `cloudflare_zero_trust_risk_scoring_integration` | folds into the planned CloudflareZeroTrustRiskScoring kind (SIEM/IdP integration exists in service of risk scoring) |
| `cloudflare_zero_trust_tunnel_warp_connector` | judged as a planned CloudflareZeroTrustWarpConnector kind (site-to-site WARP connector tunnels) |
| `cloudflare_zero_trust_tunnel_warp_connector_config` | folds into the planned CloudflareZeroTrustWarpConnector kind (configuration of the connector, mirroring the cloudflared pattern) |
| `cloudflare_zone_lockdown` | judged as a planned CloudflareZoneLockdown kind (URL lockdown rules, many per zone with independent lifecycle) |

### Deferred (45)

| Resource | Recorded reason |
|---|---|
| `cloudflare_account_dns_settings_internal_view` | Enterprise-only entitlement: internal DNS views |
| `cloudflare_address_map` | requires the BYOIP entitlement |
| `cloudflare_byo_ip_prefix` | requires the BYOIP entitlement |
| `cloudflare_cloudforce_one_request` | threat-intelligence request workflow, not provisioned infrastructure |
| `cloudflare_cloudforce_one_request_asset` | asset attachment of the threat-intelligence request workflow, not provisioned infrastructure |
| `cloudflare_cloudforce_one_request_message` | message on a threat-intelligence request ticket, not provisioned infrastructure |
| `cloudflare_cloudforce_one_request_priority` | priority row of the threat-intelligence request workflow, not provisioned infrastructure |
| `cloudflare_dls_prefix_binding` | requires the BYOIP and Data Localization Suite entitlements |
| `cloudflare_dns_zone_transfers_acl` | Enterprise-only entitlement: secondary DNS zone transfers |
| `cloudflare_dns_zone_transfers_incoming` | Enterprise-only entitlement: secondary DNS zone transfers |
| `cloudflare_dns_zone_transfers_outgoing` | Enterprise-only entitlement: secondary DNS zone transfers |
| `cloudflare_dns_zone_transfers_peer` | Enterprise-only entitlement: secondary DNS zone transfers |
| `cloudflare_dns_zone_transfers_tsig` | Enterprise-only entitlement: secondary DNS zone transfers |
| `cloudflare_email_security_block_sender` | Enterprise add-on entitlement: cloud email security |
| `cloudflare_email_security_impersonation_registry` | Enterprise add-on entitlement: cloud email security |
| `cloudflare_email_security_trusted_domains` | Enterprise add-on entitlement: cloud email security |
| `cloudflare_image` | per-image content object; application data rather than provisioned infrastructure |
| `cloudflare_keyless_certificate` | Enterprise-only entitlement: Keyless SSL |
| `cloudflare_logpull_retention` | Enterprise-only entitlement: Logpull API |
| `cloudflare_magic_network_monitoring_configuration` | Enterprise-only entitlement: Magic Network Monitoring |
| `cloudflare_magic_network_monitoring_rule` | Enterprise-only entitlement: Magic Network Monitoring |
| `cloudflare_magic_transit_cf1_site` | Enterprise-only entitlement: Magic Transit |
| `cloudflare_magic_transit_connector` | Enterprise-only entitlement: Magic Transit |
| `cloudflare_magic_transit_site` | Enterprise-only entitlement: Magic Transit |
| `cloudflare_magic_transit_site_acl` | Enterprise-only entitlement: Magic Transit |
| `cloudflare_magic_transit_site_lan` | Enterprise-only entitlement: Magic Transit |
| `cloudflare_magic_transit_site_wan` | Enterprise-only entitlement: Magic Transit |
| `cloudflare_magic_wan_gre_tunnel` | Enterprise-only entitlement: Magic WAN |
| `cloudflare_magic_wan_ipsec_tunnel` | Enterprise-only entitlement: Magic WAN |
| `cloudflare_magic_wan_static_route` | Enterprise-only entitlement: Magic WAN |
| `cloudflare_moq_relay` | experimental Media-over-QUIC relay; revisit when the product stabilizes |
| `cloudflare_page_rule` | legacy surface superseded by the Rulesets engine, which the catalog already models; Cloudflare is migrating users off Page Rules |
| `cloudflare_page_shield_policy` | Enterprise add-on entitlement: Page Shield CSP policies |
| `cloudflare_regional_hostname` | Enterprise add-on entitlement: Regional Services |
| `cloudflare_registrar_domain` | interactive, account-singular domain transfers; not fleet-deployable infrastructure |
| `cloudflare_sso_connector` | Enterprise-only entitlement: dashboard single sign-on |
| `cloudflare_stream` | per-video content object; application data rather than provisioned infrastructure |
| `cloudflare_stream_audio_track` | content-follower of a video; application data rather than provisioned infrastructure |
| `cloudflare_stream_caption_language` | content-follower of a video; application data rather than provisioned infrastructure |
| `cloudflare_stream_download` | content-follower of a video; application data rather than provisioned infrastructure |
| `cloudflare_user` | manages the calling user's own profile; not a deployable infrastructure object |
| `cloudflare_vulnerability_scanner_credential` | nascent scanning surface; revisit on demand |
| `cloudflare_vulnerability_scanner_credential_set` | nascent scanning surface; revisit on demand |
| `cloudflare_vulnerability_scanner_target_environment` | nascent scanning surface; revisit on demand |
| `cloudflare_web3_hostname` | Cloudflare has de-emphasized its Web3 gateways; revisit on demand |

### Excluded as deprecated (7)

| Resource | Recorded reason |
|---|---|
| `cloudflare_api_shield_discovery_operation` | deprecated in the provider schema |
| `cloudflare_api_shield_operation_schema_validation_settings` | deprecated in the provider schema |
| `cloudflare_api_shield_schema` | deprecated in the provider schema |
| `cloudflare_filter` | deprecated in the provider schema |
| `cloudflare_firewall_rule` | deprecated in the provider schema |
| `cloudflare_rate_limit` | deprecated in the provider schema |
| `cloudflare_snippets` | deprecated in the provider schema |
