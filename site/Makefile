# Makefile for Next.js project

YARN := yarn
PORT ?= 3000
# Use venv Python if available, otherwise fall back to python3
PYTHON := $(shell [ -f .venv/bin/python ] && echo ".venv/bin/python" || echo "python3")

# R2 Configuration
R2_BUCKET := planton-assets
R2_ENDPOINT := https://074755a78d8e8f77c119a90a125e8a06.r2.cloudflarestorage.com
R2_PROFILE := r2

.PHONY: deps install lint typecheck build run dev devs clean clean-build preview-site capture sync-assets process-images check-images inbox

deps: install

install:
	$(YARN) install

lint: install
	$(YARN) lint

# The site has no test suite; the type checker is the one that catches a
# renamed prop or a data record a page reads wrong. It runs before every build.
typecheck: install
	$(YARN) typecheck

build: install lint typecheck
	NODE_NO_WARNINGS=1 NEXT_TELEMETRY_DISABLED=1 $(YARN) build
	mkdir -p out/_site && cp -r out/_next out/_site/_next

# Design-review captures of the built export (see scripts/capture-pages.mjs).
#   make capture OUT=/tmp/captures            all scenes
#   make capture OUT=/tmp/after COMPARE=/tmp/before   and prove zero visual change
OUT ?= /tmp/site-captures
capture:
	node scripts/capture-pages.mjs --out $(OUT) $(if $(COMPARE),--compare $(COMPARE),) $(if $(ONLY),--only $(ONLY),)

# Development server
devs: run
	
# Alias for convenience
dev: run

run:
	$(YARN) dev

clean:
	rm -rf .next node_modules .yarn dist out .pnp.cjs .pnp.loader.mjs
	rm -rf packages/website-shell/dist packages/website-shell/node_modules

clean-build: clean install build

preview: preview-site

# Build and locally preview the static export from ./out. NODE_NO_WARNINGS
# keeps Node's own notices (an npm internal loading an ES module) out of the
# terminal; a warning that shows here should be the site's, never the runtime's.
preview-site: build
	cd out && NODE_NO_WARNINGS=1 npx --yes serve -l $(PORT)

# =============================================================================
# Image Processing & Assets Sync
# =============================================================================

# Sync assets folder to Cloudflare R2
# Mirrors content/assets/ to assets.planton.ai/site/
sync-assets:
	aws s3 sync content/assets/ s3://$(R2_BUCKET)/site/ \
		--profile $(R2_PROFILE) \
		--endpoint-url $(R2_ENDPOINT) \
		--delete \
		--exclude "_inbox/*" \
		--exclude "_rules/*" \
		--exclude "*.md" \
		--exclude ".gitignore" \
		--exclude ".gitkeep" \
		--exclude ".DS_Store"

# Process images from inbox (interactive - prompts for context)
# Usage: make process-images CONTEXT="service-hub" DESC="feature-overview"
process-images:
	@if [ -z "$(CONTEXT)" ]; then \
		echo "Usage: make process-images CONTEXT=\"service-hub\" DESC=\"description\""; \
		exit 1; \
	fi
	$(PYTHON) -m tools.image_processor process --context "$(CONTEXT)" --description "$(DESC)"

# Check if all required tools are installed
check-images:
	$(PYTHON) -m tools.image_processor check

# List images in inbox
inbox:
	$(PYTHON) -m tools.image_processor inbox

# Install Python dependencies for image processor
install-image-tools:
	pip install -r tools/image_processor/requirements.txt

# =============================================================================
# Legacy (deprecated - use sync-assets instead)
# =============================================================================

# Sync media assets in content/media to Cloudflare R2 via rclone
# DEPRECATED: Use sync-assets instead
sync-media:
	@echo "WARNING: sync-media is deprecated. Use sync-assets instead."
	@echo "The new workflow uses content/assets/ instead of content/media/"
	@echo ""
	rclone sync content/media/ r2:planton-ai-media --progress --transfers 8 --checkers 16 --s3-acl private
