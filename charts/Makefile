# =============================================================================
# InfraCharts Build & Publish
# =============================================================================
# Discovers all charts (provider/chart-name directories containing values.yaml)
# and validates/publishes them with `planton chart`.
#
# Usage (run from this charts/ directory):
#   make build          Build (validate) all charts (default)
#   make build-verbose  Build all charts with full output on failure
#   make publish        Publish all charts to the Planton platform
#
# IMPORTANT: `planton chart` is provided by the Planton *Platform* CLI. The
# `InfraChart` kind is a platform API, so the open-source `planton` binary built
# from this repo does not carry the `chart` command. Keep this Makefile scoped to
# charts/ and out of the repo's build graph until CLI convergence folds `chart`
# into the open-source command tree.
# =============================================================================

CHARTS := $(shell find . -mindepth 2 -name values.yaml -not -path '*/build/*' \
            | sed 's|/values.yaml||' | sed 's|^\./||' | sort)

.DEFAULT_GOAL := build

.PHONY: build
build:
	@passed=0; failed=0; failures=""; i=0; \
	total=$(words $(CHARTS)); \
	for chart in $(CHARTS); do \
	  i=$$((i+1)); \
	  if planton chart build "$$chart" --no-browser >/dev/null 2>&1; then \
	    passed=$$((passed+1)); \
	    printf "[%d/%d] ✔ %s\n" "$$i" "$$total" "$$chart"; \
	  else \
	    failed=$$((failed+1)); \
	    failures="$$failures  $$chart\n"; \
	    printf "[%d/%d] ✘ %s\n" "$$i" "$$total" "$$chart"; \
	  fi; \
	done; \
	echo ""; \
	echo "Build complete: $$passed passed, $$failed failed out of $$total charts"; \
	if [ $$failed -gt 0 ]; then \
	  echo ""; echo "Failed charts:"; printf "$$failures"; exit 1; \
	fi

.PHONY: build-verbose
build-verbose:
	@passed=0; failed=0; failures=""; \
	total=$(words $(CHARTS)); \
	for chart in $(CHARTS); do \
	  echo "--- $$chart ---"; \
	  if planton chart build "$$chart" --no-browser 2>&1; then \
	    passed=$$((passed+1)); \
	  else \
	    failed=$$((failed+1)); \
	    failures="$$failures  $$chart\n"; \
	  fi; \
	  echo ""; \
	done; \
	echo "Build complete: $$passed passed, $$failed failed out of $$total charts"; \
	if [ $$failed -gt 0 ]; then \
	  echo ""; echo "Failed charts:"; printf "$$failures"; exit 1; \
	fi

.PHONY: publish
publish:
	@passed=0; failed=0; failures=""; i=0; \
	total=$(words $(CHARTS)); \
	for chart in $(CHARTS); do \
	  i=$$((i+1)); \
	  if planton chart publish "$$chart" --platform 2>&1 | grep -q "applied"; then \
	    passed=$$((passed+1)); \
	    printf "[%d/%d] ✔ %s\n" "$$i" "$$total" "$$chart"; \
	  else \
	    failed=$$((failed+1)); \
	    failures="$$failures  $$chart\n"; \
	    printf "[%d/%d] ✘ %s\n" "$$i" "$$total" "$$chart"; \
	  fi; \
	done; \
	echo ""; \
	echo "Publish complete: $$passed passed, $$failed failed out of $$total charts"; \
	if [ $$failed -gt 0 ]; then \
	  echo ""; echo "Failed charts:"; printf "$$failures"; exit 1; \
	fi
