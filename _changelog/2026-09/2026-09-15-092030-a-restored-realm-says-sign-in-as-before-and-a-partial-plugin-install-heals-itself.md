# A restored realm says "sign in as before", and a partial plugin install heals itself

**Date**: September 15, 2026
**Type**: Bug Fix
**Components**: planton-operator (the identity component's Ready sentence, the shared sub-operator gate, the backup plugin's install)

## Summary

Two things the first live restore left on the list. The identity component reported a restored realm with the first-run sentence ("the first visitor creates the admin account using the setup code"), true of an empty realm and wrong for one that came back with its people; it now says the realm was restored from a named server, existing users sign in as before, no setup code applies, and where the re-established master admin's credential lives. And the sub-operator gate learned a per-release switch, on for the backup plugin: while its controller Deployment stands unready, the release is re-applied on every pass, so an install refused on one object heals the moment the cause is removed instead of starving forever, and a refusal that persists is reported in `status.backup` on every pass rather than in one that the next overwrites. Releases with co-owned content (Tekton) keep their never-re-apply rule.

## Proof

Component tests: the restored-realm sentence names the source and sign-in as before and never the first-visitor journey; the gate re-applies an unready controller once per pass under the switch and never a serving one; the whole component package green.
