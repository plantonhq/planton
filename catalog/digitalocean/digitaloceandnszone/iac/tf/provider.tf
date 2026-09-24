terraform {
  required_version = ">= 1.5"

  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.99"
    }
  }
}

provider "digitalocean" {
  token = var.digitalocean_token

  # DigitalOcean's DNS API deadlocks on CONCURRENT record writes to one
  # domain: two record creates (or deletes) in flight together fail one of
  # them with "422 Error 1213 (40001): Deadlock found when trying to get
  # lock" (live-verified: 1 in 8 parallel creates, on a fresh and on a
  # settled domain alike), and the provider does not retry 422s. Terraform
  # cannot serialize the instances of one for_each resource, so this module
  # serializes at the client instead: a static limit of one request per
  # second starts every API call at least a second after the previous one,
  # and record writes take a fraction of that. A zone of N records applies
  # in roughly 2N seconds -- the right trade for a resource whose write
  # failures otherwise look random. The Pulumi module serializes the same
  # records with an explicit dependency chain.
  requests_per_second = 1
}
