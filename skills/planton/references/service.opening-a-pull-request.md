---
title: Opening a Pull Request — The One Way You Write to a Repository
description: How the assistant writes a fix back to a service's repository — never a commit of its own, always a pull request the platform opens on a branch of its own (open_repo_pull_request; CLI planton service repo pr), the consent beat that must precede it (show the exact files and the pull request's text, ask, then act), the receipt the platform writes into every pull request, what the answer tells you (whether the pull request builds on its own), merging only on the person's explicit ask pinned to the head they saw (merge_repo_pull_request; CLI planton service repo merge), and every refusal's next step. Read when a fix is known and the person wants it applied, when someone says "open the PR" or "merge it", when a write is refused, or when you are tempted to say "commit this and push".
---

# Opening a Pull Request — The One Way You Write to a Repository

You read repositories freely (`references/service.reading-a-repository.md`); you write them one way only: a pull request the platform opens for you, on a branch of its own, after the person has seen exactly what it will contain and said yes. Never a commit to the default branch. Never a push from your sandbox with a token. Never a pull request the person did not ask for. The pull request is the repository's own review boundary — GitHub's protections, reviewers, and checks apply to it as to anyone — and that is why it is the one door.

## The beat, in order

1. **Know the fix exactly.** You have read the file at the commit that failed (`read_repo_files` with the run's sha) and you can write the whole file as it should read. A fix you cannot write whole is not ready for a pull request — keep diagnosing.
2. **Show, then ask — never the reverse.** Before any tool call, show the person the exact change (the lines that differ, in the file's own context) and the pull request's title and description in plain words: what broke, why, what this changes. Then ask: *"Want me to open that as a pull request?"* One yes per pull request. A yes to "fix it" is not a yes to "open a pull request" — ask the second question.
3. **Open it.** `open_repo_pull_request` with the service, the `run` you are fixing (the platform names the branch `planton/<service>/<run tail>` and writes the run into the receipt), the `files` as whole texts (a file the fix deletes is `{path, removed: true}` with no content — name each file; a rename is a removal plus the new file), `expected_base_sha` = the commit you read the files at, the `title` and `body` you showed, and `confirm: true`. On the desktop: edit the clone in your workspace and run `planton service repo pr <service> --run <id> --file yarn.lock --title "…" --body-file pr.md --expected-base-sha <sha>`; the CLI reads each `--file` whole from the working copy (or `repo/path=local/path`), and each `--remove <path>` removes one file.
4. **Relay the answer.** The link. The branch. And `will_build`, which the platform judged by the service's own trigger rules over exactly the files you changed: *"this pull request builds on its own"* or *"no run starts for this branch — the run starts when it merges."* Say which; never guess.
5. **Merge only when asked, and ask once more.** "Merge it" means: say what merges into which branch (*"merge #42 into main — the lockfile change"*), hear yes, then `merge_repo_pull_request` with the `number`, `expected_head_sha` = the head you were shown (a pull request that gained commits since is refused, not merged), and `confirm: true`. CLI: `planton service repo merge <service> 42 --sha <head> [--method squash]` (it asks in the terminal unless `--force`). You merge only pull requests the platform opened; anyone else's is a human act on GitHub.

## What the platform writes for you

Every pull request the platform opens carries three things you never fake: the commit is authored by the organization's GitHub App **with the requester as co-author** (the repository's history credits the person beside the platform); the description ends with a **receipt the platform writes** — *opened through Planton at the request of Priya Sharma (priya@acme.io), fixing run `svcpipe_…`* — so a pull request can never misstate who said yes; and each replaced file **keeps its mode** (an executable script stays executable). Your `body` is the diagnosis in the person's language; the receipt is appended under it, never by you.

## What the platform refuses, and the next step

Every refusal is a sentence with its next step. Relay it; never work around it.

- **"the base branch moved … and yarn.lock changed in between — re-read at <sha>"** — someone pushed to the base since you read. Re-read the file at the new head, recompose, show again, ask again.
- **"nothing to propose: every file … already reads exactly as asked"** — the fix is already on the branch (someone pushed it, a file you remove is already gone, or a rerun would now pass). Say so; offer to watch the next run.
- **"'…' is not a file on the base branch (a folder, a submodule, or a file already gone)"** — you asked to remove a folder. Read its listing (`read_repo_files` / `planton service repo ls`) and name each file to remove.
- **"branch 'planton/storefront/aa1001' already exists and pull request … is open from it"** — the pull request exists; hand over that link instead of opening a second one.
- **"the head branch … is the default branch; the platform never writes the default branch directly"** — you named the default branch as `branch`. Name the run instead, or another branch.
- **"run … belongs to service 'checkout', not storefront"** — the run and the service disagree; use the run's own service.
- **"you are not allowed to write to repositories through this connection"** — the person lacks write-through on the connection (developers hold it by default; an admin grants it). Fall back to the file-and-diff hand-off: give them the exact file and the commit to make in their own checkout, and say why you could not open it.
- **"the GitHub App installation … lacks the 'Contents: write' permission"** — a GitHub organization owner must accept the App's updated permissions. Say exactly that; if the answer names a branch the change already landed on, tell the person they can open the pull request from it by hand.
- **On merge: "gained commits since it was reviewed"** — read the pull request again, show what changed, and merge with the head they actually saw. **"At least 1 approving review is required …"** (GitHub's words) — the repository's rules decide; name who can review, never route around it. **"which the platform did not open"** — merge it on GitHub.

## What never happens

- No write without the person's yes on the exact change and the pull request's text — the tool's `confirm` is the record that the yes happened, never a box you tick to make it work.
- No commit to the default branch, ever; no `git push` from your sandbox with a minted token (the token is for reading and reproducing — `references/service.reading-a-repository.md`).
- No secret, token, or credential in a file you send, and no file you did not read first.
- No merge you were not asked for, and none of a pull request you did not open.
- No promise a refusal has just denied: relay the sentence and its next step.
