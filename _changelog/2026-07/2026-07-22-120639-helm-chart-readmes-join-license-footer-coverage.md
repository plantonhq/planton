# Helm Chart READMEs Join License-Footer Coverage

**Date**: July 22, 2026
**Type**: Enhancement
**Components**: Legal/Licensing, CI Guards, Helm Charts

## Summary

The three helm chart READMEs (`helm/planton`, `helm/planton-operator`,
`helm/planton-runner`) now end with the canonical license footer, and the
license-footer guard and lint workflow cover `helm/<chart>/README.md` —
completing footer coverage across every published README surface (627 files
total).

## Problem Statement / Motivation

When the license footer landed across all kind and Infra Chart READMEs,
the helm charts were deferred: unrelated work was in flight in `helm/` and
sweeping it into the footer commit would have mixed unrelated changes. That
work has landed, so the deferred surface closes.

## Solution / What's New

- Canonical footer appended to all three helm chart READMEs. Notably, the
  guard's first extended run **discovered a third chart**
  (`helm/planton-runner`) beyond the two that were tracked — the exact class
  of drift the path-shaped guard exists to catch.
- `hack/guards/ensure_license_footers.sh`: helm READMEs added to the scoped
  set (`helm/<chart>/README.md`).
- `lint.license-footers.yaml`: `helm/**` added to the path triggers; header
  updated.

## Validation

- Guard run over the full scope (560 kind + 64 Infra Chart + 3 helm
  READMEs): green — after the guard itself caught the untracked third chart
  and it was footered.
- `actionlint` and `shellcheck`: zero findings.

## Impact

Every README that fronts a published Planton artifact now carries traveling
attribution, with CI enforcement across all three surfaces. No consumer-
visible behavior change.

## Related Work

- The original footer rollout (kinds + Infra Charts) and its guard.
- NOTICE / TRADEMARKS / CLA — the completed legal-posture bundle this closes
  out.

---

**Status**: ✅ Production Ready
