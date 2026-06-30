.PHONY: build
build:
	go build -o pulumi .

.PHONY: preview
preview: build
	pulumi preview --stack dev

.PHONY: up
up: build
	pulumi up --stack dev --yes

.PHONY: destroy
destroy: build
	pulumi destroy --stack dev --yes
