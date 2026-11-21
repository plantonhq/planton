# GcpColabRuntime Guide

The judgment this guide protects: `desiredState` sets the runtime's state when it is created and whenever the setting changes -- not on every apply -- and a runtime bills compute while it runs and disk while it exists. Decide whether the manifest or the user controls start and stop, and destroy runtimes nobody uses.

## What a runtime is

A Colab Enterprise runtime is a notebook VM assigned to exactly one user (`runtimeUser`), created from a `GcpColabRuntimeTemplate` that fixes its machine, disk, network, and image. Colab Enterprise requires a template for every runtime. The runtime's id defaults to `metadata.name` and is always sent to Google (the provider never reads it back).

## Who controls start and stop

`desiredState` is a client-side control: when the runtime is created, and on a deploy that changes the setting, the provider calls start or stop to match. A start or stop made in Colab Enterprise changes nothing the deploy compares, so it is never undone by a later deploy.

- **`STOPPED`** -- the runtime exists with its disk, and compute does not bill. Good for parking a GPU runtime until someone needs it. A user who starts it in Colab Enterprise keeps it running (and billing) until the template's idle shutdown stops it, or the setting changes.
- **`RUNNING`** -- the runtime is started at creation, and when the setting changes to it.
- **Unset** -- the same as `RUNNING`: clearing `STOPPED` on a parked runtime starts it, and compute bills again. The template's idle shutdown still stops an idle runtime.

## Upgrades

Google releases new Colab images. `autoUpgrade: true` upgrades the runtime when Google reports it upgradable, during a deploy that changes this runtime -- the provider never upgrades on its own, so a runtime nobody redeploys drifts behind. A user-defined runtime past its expiration time is upgraded by such a deploy whether or not `autoUpgrade` is set.

## One runtime per person

Only the runtime user can connect notebooks. Declare one runtime per person (a team chart lists them), each referencing the template for their role.

## Destroy

Under `DELETE`, destroy removes the runtime and its disk -- notebooks saved only on that disk are lost; keep notebooks in source control or Cloud Storage. `PREVENT` makes destroy fail.
