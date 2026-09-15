# The Machine Learning compute cluster spells its VM sizes as Azure does

## What changed

- **`AzureMachineLearningComputeCluster`'s presets, catalog page, guide, spec comment, and cost model spell VM sizes in Azure's own form** -- `Standard_NC6s_v3`, `Standard_DS3_v2`, `Standard_DS2_v2` -- where they were all-caps (`STANDARD_NC6S_V3`). The rest of the Azure catalog, including the online deployment's presets and the AKS node pool's, already used Azure's spelling; this kind was the one that did not. The cost estimate regenerates with the same prose fix and the same numbers.
- The live e2e lane's manifests keep their proven spelling; Azure accepts either, and a proven fixture is not re-spelled for cosmetics.

## Why

A preset is a starting point people copy, and the catalog's specimen pictures read the preset as written. One kind spelling a SKU in a form Azure's own console never shows made it the one uppercase SKU on the whole Azure sheet -- a small thing that reads as carelessness on a picture a customer judges in seconds.

## How to check

```bash
grep -rn 'STANDARD_' catalog/azure/azuremachinelearningcomputecluster/presets catalog/azure/azuremachinelearningcomputecluster/catalog.md   # nothing
go test ./pkg/presetvalidity/
```
