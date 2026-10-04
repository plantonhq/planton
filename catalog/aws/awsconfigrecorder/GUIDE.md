# AwsConfigRecorder — Kind Guide

Authored operational judgment for the Config recorder singleton: the
design decisions behind the spec's shape, and what to know before
operating configuration recording in production.

## Design decisions

- **One component for the recording posture.** The recorder, its
  start/stop state, the delivery channel, and retention are four
  provider resources but ONE regional decision — AWS allows one of
  each per region, they only work together, and their teardown is
  order-coupled. The status resource folds to a `recording_enabled`
  bool (its whole schema was a name and a boolean).
- **The names are AWS constants.** Recorder and channel are named
  `default` (the regional convention), and the retention singleton's
  name is API-computed — `metadata.name` never reaches AWS.
- **The recording-group shapes are enforced early.** The provider
  rejects invalid combinations (inclusion + exclusion, strategies that
  contradict `all_supported`) at plan time; the spec's CELs mirror
  those rules so the manifest fails first.
- **Aggregation split out.** The aggregator and its authorization
  reference NO recorder — aggregation works in an account with zero
  recorders — so they are their own kind, not arms here.

## Operating configuration recording in production

- **Scope is the bill.** Per-item recording charges for every
  configuration item; `all_supported` in a busy account is the
  expensive default. Start from an inclusion list of the types your
  rules actually evaluate, and record GLOBAL types (IAM) in exactly
  one region — every extra region multiplies their items.
- **Create order matters and is encoded.** Channel needs recorder;
  starting needs a channel; AWS refuses to delete a channel while the
  recorder runs. Both engines encode the ordering — but a MANUALLY
  stopped recorder drifts against `recording_enabled` and the next
  apply restarts it.
- **The singleton collides.** A recorder already present in the region
  (console-enabled, or another tool's) fails creation. Adopt it by
  import, or remove it — never work around with a second name.
- **Stopping is not losing.** `recording_enabled: false` keeps the
  recorder, channel, and history; recording (and its bill) pauses
  until re-enabled.
- **Daily recording is the cost lever for noisy types.** The
  recording-mode override records chatty types (EC2 instances during
  autoscaling) as daily snapshots while the rest stay continuous.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
