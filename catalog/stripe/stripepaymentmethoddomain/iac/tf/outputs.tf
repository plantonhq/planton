# StripePaymentMethodDomain Outputs
# Maps to the StripePaymentMethodDomainStackOutputs protobuf message: the registration's id and,
# wallet by wallet, its status and Stripe's reason when it is inactive.

output "id" {
  description = "The registration's Stripe id (pmd_...)"
  value       = stripe_payment_method_domain.this.id
}

output "enabled" {
  description = "Whether wallets may appear on the domain"
  value       = stripe_payment_method_domain.this.enabled
}

output "apple_pay_status" {
  description = "active or inactive"
  value       = try(stripe_payment_method_domain.this.apple_pay.status, null)
}

output "apple_pay_error_message" {
  description = "Stripe's reason Apple Pay is inactive, when it is"
  value       = try(stripe_payment_method_domain.this.apple_pay.status_details.error_message, null)
}

output "google_pay_status" {
  description = "active or inactive"
  value       = try(stripe_payment_method_domain.this.google_pay.status, null)
}

output "google_pay_error_message" {
  description = "Stripe's reason Google Pay is inactive, when it is"
  value       = try(stripe_payment_method_domain.this.google_pay.status_details.error_message, null)
}

output "link_status" {
  description = "active or inactive"
  value       = try(stripe_payment_method_domain.this.link.status, null)
}

output "link_error_message" {
  description = "Stripe's reason Link is inactive, when it is"
  value       = try(stripe_payment_method_domain.this.link.status_details.error_message, null)
}

output "paypal_status" {
  description = "active or inactive"
  value       = try(stripe_payment_method_domain.this.paypal.status, null)
}

output "paypal_error_message" {
  description = "Stripe's reason PayPal is inactive, when it is"
  value       = try(stripe_payment_method_domain.this.paypal.status_details.error_message, null)
}

output "amazon_pay_status" {
  description = "active or inactive"
  value       = try(stripe_payment_method_domain.this.amazon_pay.status, null)
}

output "amazon_pay_error_message" {
  description = "Stripe's reason Amazon Pay is inactive, when it is"
  value       = try(stripe_payment_method_domain.this.amazon_pay.status_details.error_message, null)
}

output "klarna_status" {
  description = "active or inactive"
  value       = try(stripe_payment_method_domain.this.klarna.status, null)
}

output "klarna_error_message" {
  description = "Stripe's reason Klarna is inactive, when it is"
  value       = try(stripe_payment_method_domain.this.klarna.status_details.error_message, null)
}
