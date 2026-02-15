.PHONY: up preview destroy refresh

STACK ?= dev

up:
	pulumi up --stack $(STACK) --yes

preview:
	pulumi preview --stack $(STACK)

destroy:
	pulumi destroy --stack $(STACK) --yes

refresh:
	pulumi refresh --stack $(STACK) --yes
