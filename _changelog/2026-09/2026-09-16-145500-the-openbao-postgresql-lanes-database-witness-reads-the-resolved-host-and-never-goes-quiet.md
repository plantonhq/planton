# The OpenBao PostgreSQL lane's database witness reads the resolved host and never goes quiet

**Date**: September 16, 2026
**Type**: Fix (E2E verifier)
**Components**: `catalog/kubernetes/aa_e2e/verify/openbao.go`, `catalog/kubernetes/aa_e2e/verify/openbao_test.go`

## Summary

The first live run of `behavioral-postgresql` passed with its third witness silently off. The verifier proves PostgreSQL storage from three witnesses -- no data volume claimed, `sys/leader` holding the lock, and the lock table and KV table present in the declared database read with psql on the CloudNativePG primary -- and the third one needs the cluster's name. The extractor took it from `host.valueFrom.name`, on the assumption that the verifier reads the scenario manifest as authored. It does not: the harness resolves references before the verify phases and hands the verifier a manifest whose host is the literal `<cluster>-rw` the `KubernetesPostgres` kind exports as `rw_service`. The extractor found no reference, left the cluster empty, and the witness printed "host declared literally ... skipped" and returned nil.

Two changes. The extractor now derives the cluster from the host in whichever form it has: the reference's name when the manifest is read as authored, otherwise the first DNS label with CloudNativePG's `-rw` suffix stripped (`e2e-bao-pg-rw`, `pg-rw.data`, `pg-rw.data.svc.cluster.local` all name their cluster; a host that is no read-write Service names none). And the witness no longer skips: a host that names no CloudNativePG cluster in the vault's namespace, or a missing `database`, is a refusal that says what to declare -- every lane on this engine composes a `KubernetesPostgres` beside the vault, and a witness that can go quiet is a proof that can die unnoticed.

## Verification

`go test ./catalog/kubernetes/aa_e2e/verify/` (the shape cases gain the resolved literal and the qualified Service name; `TestCnpgClusterFromRwHost` covers the suffix edges); `go vet`; `make e2e-build`; `make e2e-vet`. Live on the `planton-e2e` kind cluster, both engines: `behavioral-postgresql` PASS on Pulumi (211 s) and Terraform (166 s) with the third witness speaking -- `openbao_ha_locks` holds 1 row and `openbao_kv_store` exists in database `openbao` on the CNPG primary.
