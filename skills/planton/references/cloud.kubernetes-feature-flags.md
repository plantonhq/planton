# Feature Flags for Your Own Services

A person asking for "feature flags", "a kill switch", "dark launches" or "turning a feature on for one customer" is asking for one thing: a way to change what their running service does without shipping it again. The catalog's answer is an OpenFeature flag engine on their cluster, its flags declared as a typed flag file, and their services evaluating through an OpenFeature SDK. This reference is the judgment. Kind facts (every field, default, validation and output) come from the catalog pack (`catalog.kind-grounding.md`); never recall them.

## Choose the engine

Both are OpenFeature engines, so the service's code is the same OpenFeature SDK either way, and moving engines later changes the provider line, not the call sites.

| Ask | Choose |
|---|---|
| Rich targeting (rules by any attribute, percentages, progressive and scheduled rollouts), notifications on every flag change, evaluation events exported to analytics | **GO Feature Flag** (`KubernetesGoFeatureFlag` with `KubernetesGoFeatureFlagFlagFile`) |
| The CNCF project's own daemon, gRPC sync to in-process providers, flags as OpenFeature Operator `FeatureFlag` resources, JSONLogic targeting | **flagd** (`KubernetesFlagd` with `KubernetesFlagdFlagFile`) |

With no stated preference, GO Feature Flag is the default recommendation: its flag file is the more expressive of the two, and its relay serves REST, OFREP and in-process providers from one place.

## Compose it: the engine and its flags are two resources

Always two: the engine (deployed rarely) and a flag file (edited often). A flag flip edits only the flag file, with no restart and no re-apply of the engine: GO Feature Flag's relay re-reads it on its poll, and flagd sees it when the kubelet syncs the mounted ConfigMap, typically within a minute or two. Never put flags inline in the engine's manifest, and never in the same chart install as a service that changes on every push.

- **Same namespace.** Put the flag file in the engine's namespace, or name its namespace in the engine's source.
- **Give the GO Feature Flag relay and its flag file different names.** The relay's chart already creates a ConfigMap named after the relay, holding its own configuration, and the flag file renders its ConfigMap under the flag file's name. A relay and a flag file both called `flags` collide on install. Use names like `flags` (the relay) and `release-flags` (the file). flagd has no such collision.
- **Install order.** Set `startWithRetrieverError: true` on the relay, so it starts and serves every flag at the caller's default until the file arrives. When the flag file sits in a namespace another resource of the same chart creates, name its ConfigMap in the relay's retriever literally rather than by `valueFrom`. A reference both ways is a dependency cycle no deploy can satisfy.
- **Say the flip time the person will see.** GO Feature Flag's relay polls its retrievers every 60 seconds by default (`flagSource.pollingIntervalMs`); lower it when "within seconds" matters. flagd's flip time is the kubelet's ConfigMap sync.
- **Write rules the engine keeps.** The flag file's schema enforces every rule the engine checks at load (one value type per flag, rules naming real variations, rollouts that ramp forward, real dates), except the query syntax. A query that does not parse makes the engine drop the whole flag, and every evaluation silently returns the caller's default. After every edit, evaluate the flag both ways (below).

## Wire a service to it

The service reads the engine's address from the engine's outputs: `status.outputs.api_endpoint` for GO Feature Flag, or the evaluation/OFREP ports for flagd, carried by `valueFrom` into the service's environment (`infra.config-references.md`). It then evaluates through the OpenFeature SDK for its language with the engine's provider.

- **Prefer in-process evaluation** (GO Feature Flag's in-process provider, or flagd's in-process sync). The service loads the flag configuration and evaluates locally, so an answer never waits on the network, and the last configuration keeps answering while the engine is away.
- **Fail closed in the code.** The default value at every call site is the safe answer, which for a new feature is "off". The default is what callers get before the first load and when a flag is dropped.
- **Pass the context the rules read.** For "on for one customer", pass that customer's stable identifier as the targeting key and as the attribute the rule reads (`org in ["acme"]`). A percentage rollout buckets on the targeting key, so a customer keeps their answer as it widens.

## Keep it private

Without API keys, anyone who can reach the engine's Service evaluates every flag and reads the whole flag configuration, which customer lists in targeting rules make sensitive.
- In-cluster only, no route, and a namespace network policy admitting only the services that read it: no keys needed.
- Reachable from outside its namespace or the cluster: set `authorizedKeys` (GO Feature Flag) with `$secret/` references, and give each consumer its key through its secret environment.

## Prove it

1. Evaluate the flag both ways over OFREP from inside the namespace: `POST /ofrep/v1/evaluate/flags/<flag>` with a context the rule matches (`TARGETING_MATCH`) and one it does not (`DEFAULT`).
2. Flip the flag file, apply it, and time the answer changing with no restart.
3. Then read the service's own behavior, not only the engine's.

## What this is not

- **Not Planton's own early-release features.** A self-hosted Planton has no flag source by design (`self-hosted.reading-a-platform.md`, "Features in early release"); this engine on the person's cluster does not change that.
- **Not entitlements or plans.** A flag hides something new from the customers it isn't ready for, then retires. What a customer's plan includes, forever, belongs in their product's own entitlement model.
- **Not old-versus-new switches.** A flag hides something new. A flag that chooses between an old and a new implementation keeps two code paths alive. Recommend shipping the new one behind the flag and deleting the old when the flag retires.
