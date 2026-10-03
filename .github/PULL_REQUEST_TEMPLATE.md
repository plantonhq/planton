## What does this PR do?

<!-- One or two sentences. Link the issue if one exists. -->

## Catalog knowledge routing

<!-- Delete this section if the PR does not touch the Infra Catalog.
     Full routing table: CONTRIBUTING.md, "Contributing Catalog Knowledge". -->

- [ ] Fact fixes went into proto comments / validation rules (never into
      generated `reference.md` files), and `make generate-reference` was run
- [ ] Kind judgment went into the kind's `GUIDE.md`; composition wisdom
      went into `catalog/patterns/`
- [ ] `go test ./pkg/explain/refgen/` passes (reference freshness + authored
      knowledge checks)

## How was this tested?

<!-- Targeted builds/tests run, and their results. -->
