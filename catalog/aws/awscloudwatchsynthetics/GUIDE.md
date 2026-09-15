# AwsCloudwatchSynthetics — Operational Guide

Live-earned judgment lands here as proof runs and adopter operations teach it; the notes below are the forge-time seed.

## The zip layout is the runtime's contract

Node.js runtimes REQUIRE `nodejs/node_modules/<fileName>.js` inside the zip with handler `<fileName>.handler`; Python runtimes use `python/<fileName>.py`. A zip with the wrong layout creates a canary that lands in CREATE_FAILED — and AWS's only repair is delete-and-recreate, which the provider performs automatically.

## Never put secrets in environment variables

AWS does not return canary environment variables on reads (write-only), and they surface in the Lambda console. Scripts that need credentials should read Secrets Manager or Parameter Store at run time under the execution role.

## start_canary is the cost lever

A READY canary costs nothing; runs bill per run. Keep `start_canary: false` in preproduction manifests and flip it (an in-place update — the provider calls StartCanary) when monitoring should begin.

## Group joins are name-based on purpose

The association resource joins a canary ARN to a group NAME. Groups shared across teams live in one owning instance (or pre-exist); every canary instance joins by name — no instance ever fights another over the group object.

## Retention trims artifact cost

Success artifacts at 31 days (the default) accumulate screenshots fast on frequent schedules. `success_retention_period: 7` with failures at 31 is the common production posture.

## Deleting a canary leaves its Lambda layer behind (observed 2026-08-26)

AWS builds each canary's script into a Lambda layer named `cwsyn-<canary>-<uuid>`. The provider's destroy passes `delete_lambda: true`, which removes the canary's Lambda FUNCTION, but the layer version survives — AWS's DeleteCanary documents that layers are not cleaned up. A dormant layer version costs nothing, but if you audit the account after removing canaries, expect the `cwsyn-` layers and delete their versions by hand (`aws lambda delete-layer-version`).
