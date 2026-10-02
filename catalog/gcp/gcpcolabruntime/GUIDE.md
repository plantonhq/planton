# GcpColabRuntime Guide

The judgment this guide protects: `desiredState` is a policy the block enforces on every apply, and a runtime bills compute while it runs and disk while it exists. Decide whether the manifest or the user controls start and stop, and destroy runtimes nobody uses.

## What a runtime is

A Colab Enterprise runtime is a notebook VM assigned to exactly one user (`runtimeUser`), created from a `GcpColabRuntimeTemplate` that fixes its machine, disk, network, and image. Colab Enterprise requires a template for every runtime. The runtime's id defaults to `metadata.name` and is always sent to Google (the provider never reads it back).

## Who controls start and stop

`desiredState` is a client-side control: on every apply the provider compares the runtime's state and calls start or stop to match.

- **`STOPPED`** -- the runtime exists with its disk, and compute does not bill. Good for parking a GPU runtime until someone needs it -- but a user who starts it in the console will see it stopped again on the next apply.
- **`RUNNING`** -- the block starts it on every apply.
- **Unset** -- the runtime starts at creation and the user starts and stops it freely. The template's idle shutdown still stops it when idle.

## Upgrades

Google releases new Colab images. `autoUpgrade: true` upgrades the runtime when it is started and Google reports it upgradable, so long-lived runtimes do not drift behind.

## One runtime per person

Only the runtime user can connect notebooks. Declare one runtime per person (a team chart lists them), each referencing the template for their role.

## Destroy

Under `DELETE`, destroy removes the runtime and its disk -- notebooks saved only on that disk are lost; keep notebooks in source control or Cloud Storage. `PREVENT` makes destroy fail.
