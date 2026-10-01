# GcpMonitoringNotificationChannel Guide

The judgment this guide protects: a channel is the LAST hop of the paging
path. Every mistake here is silent by construction — a wrong address, an
unverified channel, or a disabled endpoint drops notifications without any
incident-side symptom.

## Two label maps that must never be confused

The provider's `labels` argument is the channel's CONFIGURATION (the email
address, the Slack channel name); `user_labels` is freeform metadata. The
spec names them by what they are: `channelLabels` for configuration,
`labels` for metadata. Putting the email address in `labels` creates a
channel that validates, deploys, and notifies nobody — GCP treats a
missing configuration key as a server-side error only for some types.

## Credentials have exactly one home

`authToken`, `password`, and `serviceKey` live in `sensitiveLabels`, where
the platform enforces managed-secret handling and GCP redacts them on
read. Validation refuses the same keys in `channelLabels` — a credential
in a plain map would be shown in plans and kept outside Pulumi's secret
encryption.

## Verification is part of the deploy, not an afterthought

SMS and email channels deliver ONLY after verification. The
`verification_status` output is the honest signal: `UNVERIFIED` means the
channel exists but pages nobody. Bake the verification step into the
runbook for new on-call addresses; types that need none report
`VERIFICATION_STATUS_UNSPECIFIED`, which is normal.

## Deleting a channel policies still use

The default (`forceDelete: false`) makes such a delete FAIL — the right
posture, because GCP otherwise removes the channel from every referencing
policy in the same operation and the policies keep evaluating with one
fewer delivery path. Set `forceDelete: true` only when retiring a channel
deliberately, and re-point the policies first.

## Teardown discipline

`PREVENT` is the right `deletionPolicy` once production policies page
through this channel — destroying the delivery endpoint is equivalent to
disabling the alerting it serves. `ABANDON` keeps the channel delivering
while dropping it from management.
