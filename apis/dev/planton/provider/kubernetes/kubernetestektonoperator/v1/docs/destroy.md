# Destroying Kubernetes Tekton Operator

## Known Limitation

The Tekton Operator creates `TektonInstallerSet` resources with finalizers that can only be removed by the operator controller itself. When running `pulumi destroy` or `terraform destroy`, the operator deployment is deleted before it can clean up these resources, causing the destroy operation to hang or timeout.

This is a fundamental race condition in the Tekton Operator's design that affects all IaC tools (Pulumi, Terraform, Helm, etc.).

## Required Manual Step Before Destroy

**You must run these commands before attempting `pulumi destroy` or `terraform destroy`:**

```bash
# 1. Delete TektonConfig to trigger operator cleanup
kubectl delete tektonconfigs.operator.tekton.dev --all --ignore-not-found

# 2. Wait for TektonInstallerSets to be removed by the operator (up to 5 minutes)
echo "Waiting for TektonInstallerSets to be cleaned up..."
for i in {1..60}; do
  count=$(kubectl get tektoninstallersets.operator.tekton.dev --no-headers 2>/dev/null | wc -l | tr -d ' ')
  if [ "$count" = "0" ]; then
    echo "All TektonInstallerSets removed."
    break
  fi
  echo "  $count TektonInstallerSets remaining, waiting..."
  sleep 5
done

# 3. If any InstallerSets still exist, remove their finalizers
kubectl get tektoninstallersets.operator.tekton.dev -o name 2>/dev/null | \
  xargs -r -I{} kubectl patch {} -p '{"metadata":{"finalizers":[]}}' --type=merge

# 4. Delete remaining Tekton CRs
kubectl delete tektonpipelines.operator.tekton.dev --all --ignore-not-found
kubectl delete tektontriggers.operator.tekton.dev --all --ignore-not-found
kubectl delete tektondashboards.operator.tekton.dev --all --ignore-not-found
kubectl delete tektonchains.operator.tekton.dev --all --ignore-not-found
kubectl delete tektonresults.operator.tekton.dev --all --ignore-not-found
kubectl delete tektonhubs.operator.tekton.dev --all --ignore-not-found

echo "Tekton cleanup complete. You can now run pulumi destroy or terraform destroy."
```

## One-Liner Version

```bash
kubectl delete tektonconfigs.operator.tekton.dev --all --ignore-not-found && \
sleep 30 && \
kubectl get tektoninstallersets.operator.tekton.dev -o name 2>/dev/null | xargs -r -I{} kubectl patch {} -p '{"metadata":{"finalizers":[]}}' --type=merge
```

## Why This Happens

1. Tekton Operator's webhook automatically creates `TektonInstallerSet` resources
2. These resources have the finalizer `tektoninstallersets.operator.tekton.dev`
3. Finalizers can only be removed by the operator controller
4. During IaC destroy, the operator deployment is deleted before it can process finalizer removal
5. CRDs cannot be deleted while instances with finalizers exist
6. Result: Destroy hangs indefinitely waiting for the `tektoninstallersets.operator.tekton.dev` CRD deletion

## What If Destroy Already Failed?

If you already attempted destroy and it's stuck or failed, run:

```bash
# Remove finalizers from any stuck TektonInstallerSet resources
kubectl get tektoninstallersets.operator.tekton.dev -o name 2>/dev/null | \
  xargs -r -I{} kubectl patch {} -p '{"metadata":{"finalizers":[]}}' --type=merge

# Force delete any stuck CRDs
kubectl delete crd tektoninstallersets.operator.tekton.dev --ignore-not-found

# Clean up namespaces if stuck
kubectl delete namespace tekton-operator tekton-pipelines --ignore-not-found
```

Then retry your destroy command.

## References

- [Tekton Operator GitHub](https://github.com/tektoncd/operator)
- [Kubernetes Finalizers Documentation](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/)
- [Tekton Operator Helm Chart Uninstall Notes](https://github.com/tektoncd/operator/tree/main/charts/tekton-operator#uninstalling)

