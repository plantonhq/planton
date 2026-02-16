.PHONY: help preview up destroy refresh cancel

help:
	@echo "Available targets:"
	@echo "  preview  - Preview changes to the stack"
	@echo "  up       - Deploy the stack"
	@echo "  destroy  - Destroy the stack"
	@echo "  refresh  - Refresh the stack state"
	@echo "  cancel   - Cancel the current stack operation"

preview:
	@echo "Running Pulumi preview..."
	pulumi preview

up:
	@echo "Deploying stack..."
	pulumi up

destroy:
	@echo "Destroying stack..."
	pulumi destroy

refresh:
	@echo "Refreshing stack state..."
	pulumi refresh

cancel:
	@echo "Cancelling current operation..."
	pulumi cancel
