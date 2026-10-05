# Contributing to Planton

First, thanks for contributing to `Planton` and helping make it better. We appreciate the help!
This repository is one of many across the `Planton` ecosystem, and we welcome contributions to them all.

## Code of Conduct

Please make sure to read and observe our [Contributor Code of Conduct](CODE-OF-CONDUCT.md).

## Communications

You are welcome to join the [Discord Server](https://discord.gg/pwcSapdQAp) for questions and a community of like-minded folks.
We discuss features and file bugs on GitHub via [Issues](https://github.com/plantonhq/planton/issues) as well as [Discussions](https://github.com/plantonhq/planton/discussions).

We welcome contributions from the community to enhance **Planton**. Whether you want to fix bugs, add new
features, or improve documentation, your efforts are appreciated and will help make this project better for everyone.

## How to Contribute

1. **Fork the Repository**

   Start by forking the [Planton GitHub repository](https://github.com/plantonhq/planton) to your own
   GitHub account.

2. **Clone the Repository**

   Clone your forked repository to your local machine:

   ```bash
   git clone https://github.com/yourusername/planton.git
   ```

3. **Create a Branch**

   Create a new branch for your feature or bug fix:

   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **Make Changes**

    - Implement your changes, following the project's coding standards and guidelines.
    - Ensure that any new code is well-documented and adheres to the existing style.

5. **Run Tests**

    - Before committing your changes, run existing tests to ensure nothing is broken.
    - Add new tests if you're introducing new features or modifying existing functionality.

6. **Commit Changes**

   Commit your changes with clear and descriptive messages:

   ```bash
   git commit -m "Add feature X to improve Y"
   ```

7. **Push to GitHub**

   Push your branch to your forked repository:

   ```bash
   git push origin feature/your-feature-name
   ```

8. **Create a Pull Request**

    - Go to the original repository and click on "New Pull Request."
    - Select your branch and provide a detailed description of your changes.
    - Include any relevant issue numbers or context that helps reviewers understand your contribution.

9. **Review Process**

    - Your pull request will be reviewed by the maintainers.
    - Be prepared to make adjustments based on feedback.
    - Once approved, your changes will be merged into the main branch.

## Contribution Guidelines

- **Coding Standards**: Follow the established coding conventions and style guides for the project.
- **Documentation**: Update or add documentation to reflect your changes, especially in code comments and README files.
- **Commit Messages**: Write clear and concise commit messages that explain the "what" and "why" of your changes.
- **Issue Reporting**: If you encounter a bug or have a feature request, please open an issue before working on it to
  discuss the best approach.

## Contributing Catalog Knowledge

The Infra Catalog carries three layers of knowledge -- facts (the protos),
verified data (the cost/controls/permissions fact-sheets), and judgment (guides and
patterns) -- and your contribution lands in a different place depending on which kind
it is. Route it right and every surface (reference pages, CLI, indexes, the console,
the packaged skill) improves at once:

| You want to fix or add | Where it goes | Why |
|---|---|---|
| A wrong or missing **fact** (a field's meaning, a default, an alias like "Redis-compatible", a validation) | The proto comment or validation rule in the kind's `spec.proto` / `outputs.proto`, then run `make generate-reference` | The `reference.md` pages are generated -- never edit them by hand; fixing the source regenerates every surface |
| A **cost fact or price** (billing dimensions, a unit price, an estimate) | The kind's `cost.yaml` for billing facts; `catalog/_pricing/` for price books, derivation rules, and models -- estimates are GENERATED (`make generate-cost-estimates`), never edited | Every figure is recomputed and gate-checked (`lint.catalog-data`); a hand-typed dollar amount anywhere else is a second source of truth that will drift |
| A **control-posture claim** (what a component enforces or exposes) | The kind's `controls.yaml`, citing the central control catalog and crosswalks in `catalog/_compliance/` | Posture is evidence-backed and conformance-gated; it maps onto framework requirements but never becomes "compliant" or "certified" |
| A **provisioning-permission claim** (what the runner needs) | The kind's `iac/permissions.yaml` | Machine-checked against the providers' own published action/scope inventories -- an invented permission fails CI |
| **Judgment about one kind** (operational gotchas, when to choose it, what it pairs with) | `GUIDE.md` beside that kind's `reference.md` (create it if absent, then run `make generate-reference` so the page links it) | Authoring standard: `_rules/docs/write-catalog-kind-guide.mdc` |
| **How kinds compose** (multi-kind wiring, trade-offs, failure modes) | A pattern in `catalog/_patterns/` | Authoring standard: `_rules/docs/write-planton-architecture-pattern.mdc` |
| **Catalog-wide wisdom** (finding alternatives, cross-provider conventions) | `catalog/_docs/GUIDE.md` | The catalog's own guide |
| **The page format itself** (headings, tables, the search grammar) | The markdown renderer in `pkg/explain`, then `make generate-reference` | Format quality is measured, never asserted: run the eval in `_rules/docs/evaluate-planton-catalog-research.mdc` and include the numbers in your PR |

The file you edit is exactly the file agents and users read -- in the repository,
the release archive, and the packaged skill alike; nothing is renamed on the way.
CI keeps authored knowledge honest: embedded complete manifests must validate
against their schemas, declared kind names must resolve, and links must not break
(`go test ./pkg/explain/refgen/` runs the same checks locally).

## Licensing

This project is licensed under [Apache-2.0](LICENSE), and contributions are accepted
under the same terms — inbound equals outbound. This is what the license itself says
(section 5): any contribution you intentionally submit for inclusion is licensed
under Apache-2.0, without additional terms.

By contributing, you affirm that you have the right to submit the code you
contribute — it is your own work, or you have permission to contribute it under
these terms.

Two files travel with the project wherever it goes: [LICENSE](LICENSE) and
[NOTICE](NOTICE) (the attribution every redistribution carries). For use of the
Planton name and logo, see [TRADEMARKS.md](TRADEMARKS.md).

On your first pull request, the CLA bot will ask you to sign the
[Contributor License Agreement](CLA.md) — a one-time comment on the PR. The
CLA is what lets the project be stewarded for the long term (its grants are
spelled out in [CLA.md](CLA.md)); your signature is recorded in this
repository, and you keep every right to use your own contributions however
you wish.

## Community and Support

We encourage you to join our community and contribute to the project:

- **GitHub Issues**: Report bugs or request new features by opening an issue
  on [GitHub](https://github.com/plantonhq/planton/issues).
- **Discussions**: Engage with other users and contributors in
  our [GitHub Discussions](https://github.com/plantonhq/planton/discussions) forum.
