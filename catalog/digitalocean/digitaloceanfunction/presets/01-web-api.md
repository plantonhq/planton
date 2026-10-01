# Hello-World Function

This preset deploys a DigitalOcean Functions app from DigitalOcean's public Node.js hello-world sample. Both engines create an App Platform app with one functions component -- there is no standalone Functions resource.

Runtime, memory, timeout, entrypoint, and cron schedules are **not** on this spec. They live in the repo's `project.yml`, which App Platform reads at deploy time -- from the repository root when `sourceDirectory` is unset (as here), or from `sourceDirectory` when set. Putting those knobs on the spec would silently do nothing.

The sample uses a public git clone URL so the preset applies without a linked GitHub account. Switch to `github` (owner/repo plus optional `deployOnPush`) only when the DigitalOcean account has GitHub connected.

## When to Use

- HTTP functions whose source is a Functions-style repo (`project.yml` + a `packages/` tree)
- Accounts that do not have GitHub linked yet
- Checking that a Functions deploy works before pointing at your own repo

## Key Configuration Choices

- **App name** (`appName`) -- the App Platform app's own name: 2-32 characters, starts with a letter, unique across every app in the DigitalOcean account. Renaming updates the app in place, but the default `<name>-<hash>.ondigitalocean.app` URL changes with it.
- **Function name** (`functionName`) -- the functions component name inside the app, same 2-32 character rule.
- **Git clone URL** (`git`) -- public HTTPS clone. No GitHub OAuth required.
- **Source directory** (`sourceDirectory`) -- leave unset when `project.yml` is at the repository root (the sample). Set it, for example `functions/api`, only when `project.yml` lives in a subdirectory. A wrong directory fails the build minutes into the deploy, not at validation.
- **Environment variables** (`envs`) -- `plaintext` or `secret`. Secrets are stored in App Platform's secret store.

To change runtime or add a schedule, edit `project.yml` in the repo, not this manifest.
