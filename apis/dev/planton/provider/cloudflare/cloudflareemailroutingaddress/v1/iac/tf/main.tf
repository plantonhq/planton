# An account-scoped Email Routing destination address. Creating it sends a
# verification email to the mailbox; it is usable as a forwarding target only
# after the owner clicks the verification link.
resource "cloudflare_email_routing_address" "main" {
  account_id = var.spec.account_id
  email      = var.spec.email
}
