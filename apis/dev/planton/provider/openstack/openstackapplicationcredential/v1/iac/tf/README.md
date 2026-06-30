# OpenStackApplicationCredential Terraform Module

This Terraform module provisions an OpenStack Identity application credential.

## Resources Created

- `openstack_identity_application_credential_v3` -- Keystone application credential

## Important Notes

- All fields are ForceNew (immutable resource)
- The `secret` output is marked as sensitive
