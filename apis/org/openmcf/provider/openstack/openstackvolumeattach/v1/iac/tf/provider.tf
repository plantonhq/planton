terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 3.0"
    }
  }
}

# The OpenStack provider is configured via OS_* environment variables
# (OS_AUTH_URL, OS_USERNAME, OS_PASSWORD, etc.) which are set by the
# OpenMCF providerenvvars layer from the OpenStackProviderConfig proto.
provider "openstack" {}
