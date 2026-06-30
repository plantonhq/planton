.PHONY: up preview destroy refresh

# OpenFGA Authorization Model Pulumi Module
# NOTE: OpenFGA does not have a Pulumi provider. This module is a pass-through placeholder.
# Use Terraform/Tofu provisioner for actual deployments.

MANIFEST_PATH ?= ../hack/manifest.yaml

up:
	@echo "WARNING: OpenFGA does not have a Pulumi provider."
	@echo "This module is a pass-through placeholder and will not create resources."
	@echo "Use: planton apply --manifest $(MANIFEST_PATH) --provisioner tofu"
	pulumi up --stack dev --yes

preview:
	@echo "WARNING: OpenFGA does not have a Pulumi provider."
	@echo "This module is a pass-through placeholder."
	pulumi preview --stack dev

destroy:
	@echo "WARNING: OpenFGA does not have a Pulumi provider."
	@echo "No resources to destroy from this module."
	pulumi destroy --stack dev --yes

refresh:
	pulumi refresh --stack dev --yes
