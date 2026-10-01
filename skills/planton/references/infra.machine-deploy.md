# Deploying From This Machine — the Ambient-Login Offer

On a signed-in Planton — a hosted or self-hosted instance, where deploys
normally run on infrastructure the organization operates — the machine you
are working on can become the deploy engine for THIS person's own work:
deployments run right here using the cloud login already on the machine,
the credentials never leave it, and every chart, project, and deployment
record still lands in the organization. For a brand-new team this is the
fastest honest path from conversation to real infrastructure — nothing
needs to be wired into the organization first — and for some people it is
the permanent preference, not a stopgap.

Your job is to NOTICE when this path exists and OFFER it in the user's
language. The user does not know the capability exists, and the moment they
need it — "how do I actually deploy this?" — is exactly when silence about
it forces them into setup ceremony they never needed.

## The probe — once, at the deploy moment, for the chart's provider

Two read-only commands answer everything. Run them ONCE, at the natural
moment — the architecture is built and deployment is the next step, the
user asks to deploy, or a deploy fails for want of a connection. Never
sweep providers at conversation start, and never re-probe an answer you
already have.

1. **Where is this Planton?** `planton instance current` prints the
   instance and its endpoint. A loopback endpoint (`127.0.0.1`,
   `localhost`) means Planton itself runs on this machine — deployments
   already run here, and there is nothing to offer; skip the rest of this
   reference.
2. **Is there a login for the chart's cloud?**

   ```
   planton connect aws detect -o json
   ```

   (`gcp`, `azure`, and `kubernetes` have the same verb.) Detection is the
   CLI's own business — it reads the machine's standard credential chain
   and confirms the login with the provider. You never open `~/.aws`,
   `~/.kube`, or any other credential home yourself: the filesystem
   boundary applies to credential files doubly. A successful probe answers
   the account and the detected handle (the AWS profile, gcloud
   configuration, az subscription, or kubectl context). An `error` answer
   means no usable login — no offer, no complaint, no install homework;
   proceed with the normal handoff.

Judgment, not a gate: when your earlier grounding showed the organization
already has a working connection for this cloud, deploys already work —
don't push the machine path; mention it only if the user asks for it. When
you could not read the org's connections (not every environment can), a
detected login is reason enough to offer.

## The offer — in their language, never plumbing

The offer names the provider and the detected handle, says where
credentials stay and where records go, and states the ownership boundary in
one plain sentence. It speaks entirely in what the user can observe — the
machine, the login, where their work is saved; the machinery that makes it
work is yours alone, and when a command's own output names its internals,
you translate into the user's terms rather than echo. The shape:

> You're signed in to AWS on this machine as **acme-dev**. I can set this
> machine up to run your deployments using that login — your credentials
> never leave this machine, and everything we build is saved to your
> organization. Only deployments you start will run here; teammates can
> see the machine is connected but can't use it. Want me to set that up?

Calibrate the wording to the person, keep the three facts (credentials stay
here, records go there, only your own deploys run here), and make it a real
question. One offer per conversation: a "no" or a silence is final — do not
re-offer, and never make the setup a precondition for finishing the
composition work they actually asked for.

## Consent — a clear yes, then one command

The user's clear yes in conversation IS the consent. Run the consent form
of the same verb — `-y` carries the yes they already gave you (the terminal
prompt would be a second ask for the same thing):

```
planton connect aws detect --this-machine -y
```

This is a mutation with everything that implies: state what you are about
to run before running it, and never run it on a vague or stale yes. The
command sets the machine up silently (the one slow step is a one-time
download — narrate the progress lines it prints, they are written for the
user) and records the login as a connection only this person can deploy
through. Success prints a `Connection` line — that slug is the proof;
report it plainly ("this machine is set up — deployments you start will
run here"). A declined terminal prompt prints `Cancelled.` and exits 0, so
never read a bare exit code as success; the `Connection` line is the
answer.

## Deploying — on their explicit ask, one confirmation

With the machine connected (or the org already wired), deploying a composed
chart is yours to perform when the user explicitly asks — the same
one-confirmation discipline as every mutation. One precondition first: when
the organization's catalog policy disables kinds this chart uses, the user
hears the disclosure BEFORE the deploy starts — which components, that the
rest deploys now, and that an Infrastructure Admin can enable them
(`catalog-availability.md`); a policy refusal mid-pipeline means the
disclosure was skipped, not that something broke.

```
planton chart install <name> <chart-dir> -m "why, like a commit message" --plain
```

The install creates the project and starts its deployment pipeline
immediately; follow the pipeline output and narrate what happens — which
resource is deploying, what failed and why, in the user's vocabulary. The
org and environment ride your CLI context (`--org`/`-e` when they don't).
A working copy of a deployed project keeps its own rules
(`references/infra.deployed-projects.md`) — there the save verb is the deploy.

## When it doesn't work — every failure has a way forward

- **The setup command refuses or fails**: its messages are written for the
  user — surface them as they are, and offer what they suggest (usually:
  open the Planton desktop app, then try again). Never retry in a loop.
- **The CLI doesn't know `--this-machine`** (older installation): that is
  a fact, not a fault. Say the machine-deploy path needs a newer Planton
  app, and offer the alternative — an organization cloud account connected
  through the console — without making it homework.
- **A deploy is refused because the connection belongs to someone else's
  machine**: the refusal says whose and what to do; relay it faithfully.
  Machine connections serve only their owner's own interactive work —
  automated pipelines and teammates ride organization cloud accounts, and
  that is the boundary working, not a bug.
- **No login was detected**: nothing failed. Compose, hand off normally,
  and let the user know deploys can run from the organization's cloud
  accounts once one is connected — or from this machine the moment it has
  a cloud login.
