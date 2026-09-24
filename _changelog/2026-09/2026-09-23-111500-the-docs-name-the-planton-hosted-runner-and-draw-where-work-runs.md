# The docs name the Planton-hosted runner and draw where each kind of work runs

## What changed

- **A new page, Planton-Hosted Runners** (`runner/planton-hosted-runners.md`): on Planton's hosted product a new organization's deploys, live cloud operations, and builds run on runners Planton operates until it adds one of its own. The page says what such a runner can access (a short-lived credential per job, read-only, scoped to that job's organization, lasting an hour at most and reissued only while the job runs), draws the order each kind of work picks its runner in, says when a runner of your own is worth adding, walks through moving deploys onto it (a state backend of your own, `planton tofu|pulumi state migrate-backend`, naming the runner on the connection), and shows the default-runner verbs.
- **The runner overview** points a hosted reader at that page before teaching the runner they run themselves, and its third getting-started step says what a default runner governs.
- **The deployment page's "Default Runner Binding" becomes "Default Runner"**: the three-level resolution chain it drew (connection, organization default, a "platform default set by platform operators") matched none of the three real orders and taught a default that governs deploys, which it never did. The retired `set-default --platform` and `get-default --platform` examples are gone, and the runner listing shows the organization's own runners and their capabilities instead of the staff-only platform view.
- **State backends** stops calling migration manual: the page names `state migrate-backend` and says why moving every resource off platform-managed state is what lets a runner of your own carry your deploys.

## Why

The hosted product runs every stranger's first deploy and build on runners Planton operates, and the docs never said so -- the only runner they described was one you install, so the natural reading was that nothing works until you do. The one chain they drew described a platform-operator default that customers can never set and implied a default runner moves deploys. The pages now describe the product as it behaves, in the words the console and the CLI use.

## How to check

```bash
make -C site build                                       # the internal link gate and the llms coverage gate pass
rg -n -- "--platform|Platform default|platform-level default" site/public/docs/runner/   # zero hits
```
