.PHONY: clean clean-terraform clean-pulumi

IMAGE_REPO=ghcr.io/plantonhq/planton/base-images/iac-runner

# Note: Images are built by GitHub Actions workflows:
#   - .github/workflows/iac-runner-terraform.yml
#   - .github/workflows/iac-runner-pulumi.yml
#
# Pulumi variants use self-hosted runners with persistent Go cache storage.
# Local builds are not recommended as they won't have the pre-warmed caches.

# Clean up local terraform images
clean-terraform:
	docker rmi $(IMAGE_REPO):terraform-v0.0.1 2>/dev/null || true

# Clean up local pulumi images (all providers)
clean-pulumi:
	docker rmi $(IMAGE_REPO):pulumi-aws-v0.0.1 2>/dev/null || true
	docker rmi $(IMAGE_REPO):pulumi-gcp-v0.0.1 2>/dev/null || true
	docker rmi $(IMAGE_REPO):pulumi-azure-v0.0.1 2>/dev/null || true
	docker rmi $(IMAGE_REPO):pulumi-kubernetes-v0.0.1 2>/dev/null || true
	docker rmi $(IMAGE_REPO):pulumi-cloudflare-v0.0.1 2>/dev/null || true
	docker rmi $(IMAGE_REPO):pulumi-digitalocean-v0.0.1 2>/dev/null || true
	docker rmi $(IMAGE_REPO):pulumi-atlas-v0.0.1 2>/dev/null || true
	docker rmi $(IMAGE_REPO):pulumi-auth0-v0.0.1 2>/dev/null || true
	docker rmi $(IMAGE_REPO):pulumi-civo-v0.0.1 2>/dev/null || true
	docker rmi $(IMAGE_REPO):pulumi-confluent-v0.0.1 2>/dev/null || true
	docker rmi $(IMAGE_REPO):pulumi-snowflake-v0.0.1 2>/dev/null || true

# Clean all local images
clean: clean-terraform clean-pulumi
