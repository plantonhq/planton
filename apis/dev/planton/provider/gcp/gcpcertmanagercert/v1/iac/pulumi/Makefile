.PHONY: install
install:
	@echo "Installing Pulumi dependencies..."
	go mod download

.PHONY: up
up:
	@echo "Running Pulumi up..."
	pulumi up --yes

.PHONY: destroy
destroy:
	@echo "Running Pulumi destroy..."
	pulumi destroy --yes

.PHONY: preview
preview:
	@echo "Running Pulumi preview..."
	pulumi preview

.PHONY: refresh
refresh:
	@echo "Running Pulumi refresh..."
	pulumi refresh --yes

