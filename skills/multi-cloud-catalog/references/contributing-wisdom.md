# Contributing Wisdom Back

The moment research or a live deployment teaches you something the pack does
not teach -- a trap no guide names, a judgment comparison with no home, a
wrong or thin fact on a page -- you are holding knowledge the next agent
will need. Do not let it die in the conversation: offer to contribute it as
a pull request. The file you were reading is the file you edit; the
catalog's CI checks the mechanics so a reviewer only has to judge the
wisdom.

## When to offer

Offer exactly once per distinct lesson, right after the moment that taught
it:

- A failure or surprise in this session that a component's `GUIDE.md` never
  warned about.
- A choose-between-components judgment you had to work out yourself because
  no guide or pattern owns the comparison.
- A generated page whose facts disagree with observed behavior, or whose
  field docs are too thin to act on.

Do not offer for knowledge a well-trained model already has without this
catalog -- general cloud lore earns no place in the wisdom layer.

## Route by content class, and verify first

Ground the claim before proposing it: re-read the pack surfaces it touches,
and confirm the behavior that taught it (the command output or failure is
your evidence).

| What you learned | Where it goes | How it travels |
|---|---|---|
| Judgment about one component | That kind's `GUIDE.md`, beside its `reference.md` | This workflow -- the chat-friendly lane |
| Multi-component composition wisdom | A pattern under `_patterns/` | This workflow |
| Catalog-wide wisdom (alternatives, conventions) | The catalog root `GUIDE.md` | This workflow |
| A wrong or thin FACT (field docs, defaults, validation) | The proto comment or rule it derives from -- generated pages are never hand-edited | Needs the repo toolchain; without one, file it as a GitHub issue with the full evidence instead |
| An operational truth the module enforces or prints (a name budget, the pod state a waiting job shows, a recovery loop, a TLS name a job needs) | The proto comment of the field it belongs to -- it is a fact about what the component does, and only `reference.md`, `GUIDE.md`, and `_patterns/` reach an installed agent (see `pack-layout.md`, "What the pack does not carry") | Same as a fact: repo toolchain or a GitHub issue |

One boundary is mechanical, so learn it once: **editing an existing guide
or pattern changes no generated file** -- CI passes as-is. **Adding a NEW
`GUIDE.md` does** (the kind's page head gains its Guide link, the indexes
gain a row), so a new-guide PR must either include the
`make generate-reference` output (requires a repo checkout with Go) or say
plainly in its description that a maintainer must run it -- the freshness
check will hold the PR until someone does, which is correct.

## The drafting bar

With a repo checkout, the authoring standards are
`_rules/docs/write-planton-component-guide.mdc` and
`write-planton-architecture-pattern.mdc` -- follow them. Without a checkout,
this digest is the bar:

1. **Internal-first.** Write only what could not be known without this
   platform: conventions, flag semantics, failure modes observed here. Cut
   anything a well-trained model could have written anyway.
2. **Judgment framing.** "When X vs when Y, and what breaks if you choose
   wrong" -- with the concrete failure named. Never feature lists; the
   generated page already enumerates the spec.
3. **Diagram consequence.** When a choice affects composition, say what it
   renders as: a dedicated component is a visible node; a buried flag is
   nothing.
4. **Ground every claim** in what this session actually observed or read.
   Short and true beats long and plausible.
5. **Mechanical rules CI will enforce for you**: any complete YAML manifest
   (one declaring `apiVersion:` and `kind:`) is machine-validated against
   the real schema; every relative link must resolve; spec fields are
   spelled camelCase, `fieldPath` strings snake_case.

That last point is confidence, not fear: placement, manifests, and links
are checked by machinery on every pull request, so a mechanically broken
contribution is caught by CI -- a reviewer only ever debates the wisdom.

## Delivery -- always draft first, mutate never before yes

Sanitize the draft (no tokens, account IDs, or company-internal names),
show it to the user as a diff against the pack file, and get an explicit
yes. Nothing is created on GitHub before that yes. Then pick the first lane
that fits:

**Inside a planton repo checkout** (working tree at the repo root): edit
the file, branch, and open the PR from the checkout as usual.

**With `gh` but no checkout -- the clone-free lane.** The repo is large;
never clone it for a one-file edit. The pack copy you researched from is
the current text, and the API does the rest (replace `<you>` with
`gh api user --jq .login`):

```
gh repo fork plantonhq/planton --clone=false               # one-time
gh api -X POST repos/<you>/planton/merge-upstream -f branch=main
gh api "repos/plantonhq/planton/contents/<path>?ref=main" --jq .content | base64 -d > current.md
# diff current.md against your pack copy; re-draft if main moved ahead
gh api -X POST repos/<you>/planton/git/refs \
  -f ref=refs/heads/<branch> \
  -f sha="$(gh api repos/<you>/planton/git/ref/heads/main --jq .object.sha)"
gh api -X PUT repos/<you>/planton/contents/<path> \
  -f branch=<branch> -f message="docs(catalog): <what the wisdom adds>" \
  -f sha="$(gh api "repos/<you>/planton/contents/<path>?ref=<branch>" --jq .sha)" \
  -f content="$(base64 -i draft.md)"
gh pr create --repo plantonhq/planton --head <you>:<branch> \
  --title "docs(catalog): <what the wisdom adds>" --body-file body.md
```

(A brand-new file is the same `PUT` without the `sha` field.) If `gh` is
installed but unauthenticated, `gh auth login` is interactive -- hand that
one step to the user.

**With neither**: give the user the direct web-edit URL --
`https://github.com/plantonhq/planton/edit/main/<path>` (GitHub forks
automatically for non-writers) -- plus the finished draft to paste. The
draft is the value; the delivery mechanism is secondary.

## The pull request itself

- Title in the repo's commit convention: `docs(catalog): ...` -- that
  prefix is how maintainers triage wisdom contributions (fork contributors
  cannot set labels, and never need to).
- The repo's PR template includes a catalog knowledge-routing checklist --
  fill it honestly.
- Body: the wisdom in one paragraph, then the evidence that grounded it
  (sanitized command output, the failure observed, the pack surfaces read).
- Expectation to set for first-time contributors: GitHub holds a first-time
  fork PR's CI runs until a maintainer approves them, so the checks appear
  after that click, not instantly. The PR is not being ignored.
- Report the PR URL back to the user so they can follow the review.
