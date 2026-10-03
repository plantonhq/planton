output "certificate_pack_id" {
  description = "The certificate pack identifier"
  value       = cloudflare_certificate_pack.main.id
}

# No status output: pack issuance is asynchronous (initializing ->
# pending_validation -> active), and a point-in-time phase is never a stable
# output -- it flips on the first refresh after the transition and
# re-plans forever. Read issuance status from the Cloudflare API instead.

# No primary_certificate output: the server populates it asynchronously
# after the order (absent at create, "0" appears seconds later, then the
# real certificate id as issuance progresses -- measured live 2026-08-29).
# A transitioning value is never a stable output. Read the pack's
# certificates from the Cloudflare API instead.

output "zone_id" {
  description = "The Cloudflare zone the pack was ordered in (a pack's API identity is zone_id + certificate_pack_id)"
  value       = cloudflare_certificate_pack.main.zone_id
}
