# Filing Platform Gaps

The moment you explain that Planton cannot do something the user needs, you
are holding the best bug report the platform will ever receive: the goal, the
exact attempt, and the exact shortfall are all in this conversation. Do not
let that context die in the chat — offer to file it.

## When to offer

Offer exactly once per distinct gap, right after explaining the shortfall:

- A catalog kind or spec field the user needs does not exist.
- A chart/build/deploy behavior blocks a legitimate architecture.
- Validation rejects something the cloud provider actually allows (or
  accepts something it rejects).

Do not offer for user mistakes the skill can fix, or for questions.

Phrase it as a service, not a deflection: "Planton doesn't support X today.
I can file a GitHub issue on the open-source repo with the full context from
this session so the maintainers can fix it — want me to draft it?"

## Preflight — gh availability

1. `gh --version` — installed?
2. `gh auth status` — authenticated?

If `gh` is missing, offer to install it (`brew install gh` on macOS — an
install is a mutation: exact command, confirm, run). `gh auth login` is
interactive, so hand that single step to the user with the exact command and
what to expect; verify with `gh auth status` afterwards. If the user prefers
not to install or authenticate, produce the finished issue body as text they
can paste at the repository's new-issue page instead — the draft is the
value; `gh` is only the delivery.

## Drafting — written for two audiences

The issue will be read by maintainers AND by the coding agent they hand it
to for planning the fix. Write for both: complete, specific, reproducible.

- **Title**: the capability gap in one line, not the user's task.
- **What the user was trying to accomplish** — the goal in product terms.
- **What was attempted** — the chart/manifest snippets and commands, verbatim.
- **What Planton did** — errors and build output, verbatim.
- **What was expected instead** — and, when you know it, what the cloud
  provider itself supports (API/console reference).
- **Versions** — CLI (`planton version`), app version if known.

Sanitize before showing the draft: no tokens, keys, account IDs, or
company-internal names. Replace with placeholders (`<account-id>`).

## Filing — a mutation like any other

1. Show the complete draft (title + body) to the user.
2. Wait for an explicit yes; apply any edits they ask for.
3. File against the open-source repo with the platform-gap template's label:

   ```
   gh issue create --repo plantonhq/planton \
     --title "<title>" --body-file <draft-file> --label platform-gap
   ```

4. Report the issue URL back so the user can follow it.
