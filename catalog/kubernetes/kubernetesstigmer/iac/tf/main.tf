# KubernetesStigmer: the stigmer-operator chart, the Secrets the declaration
# gives values for, the restore check, the cleanup wait, and the one Stigmer
# resource, in that order. A destroy runs it backwards: the Stigmer goes
# first, terraform_data.cleanup_wait waits for the operator to remove what it
# installed, and only then do the Secrets and the chart go. Pulumi twin:
# iac/pulumi/module/main.go, which applies the same documents
# (module/parity_test.go holds them together).
#
# THE CHART OWNS ITS DEFINITION: the Stigmer definition is a template of the
# chart's release that Helm keeps on uninstall (helm.sh/resource-policy:
# keep), so the module adds no CRD handling. The Stigmer is applied with
# kubectl_manifest, which needs no definition at plan time: on a cluster that
# has never had Stigmer, the definition arrives with the release the Stigmer
# depends on.

# The operator's namespace, when this resource owns it. Its labels carry
# planton.ai/kind, the ownership mark the preflight reads.
resource "kubernetes_namespace_v1" "operator" {
  count = local.create_namespace ? 1 : 0

  metadata {
    name   = local.operator_namespace
    labels = local.labels
  }

  depends_on = [terraform_data.preflight]
}

# The operator. The release waits for the operator to be Available: an
# operator that stops at startup (beside another copy, say) fails this
# deploy, instead of a Stigmer that never reconciles.
resource "helm_release" "operator" {
  name             = local.operator_release
  namespace        = local.operator_namespace
  repository       = local.chart_repository
  chart            = local.chart_name
  version          = var.spec.version
  create_namespace = false
  values           = [yamlencode(local.chart_values)]

  depends_on = [terraform_data.preflight, kubernetes_namespace_v1.operator]
}

# The Secrets the declaration gives values for, under the names the operator
# reads, each holding only the keys declared.
resource "kubernetes_secret_v1" "declared" {
  for_each = local.secrets

  metadata {
    name      = each.key
    namespace = local.namespace
    labels    = local.labels
  }
  type = "Opaque"
  data = each.value

  depends_on = [terraform_data.preflight]
}

# The restore check (scripts.tf keys_check_script), Tekton's gate shape: a
# ConfigMap whose creation runs the check, re-created (force_new), and so
# re-checked, whenever the namespace or the archive moves, which is when a
# new database could be created from an archive. A failing check fails the
# create, so the next apply checks again. The Stigmer depends on it.
resource "kubectl_manifest" "keys_gate" {
  yaml_body = yamlencode({
    apiVersion = "v1"
    kind       = "ConfigMap"
    metadata = {
      name      = local.keys_gate
      namespace = local.namespace
      labels    = local.labels
    }
    data = {
      purpose     = "Records that the restore check passed for this namespace and archive: a restore found the keys Secret it needs. Deleting it changes nothing."
      namespace   = local.namespace
      recoverFrom = try(var.spec.database.bundled.recover_from.server_name, "")
    }
  })

  server_side_apply = true
  force_conflicts   = true
  force_new         = true

  # The script rides the environment and the command only evals it, so the
  # deploy log shows one line rather than the whole script.
  provisioner "local-exec" {
    interpreter = ["/bin/sh", "-c"]
    command     = "eval \"$STIGMER_KEYS_CHECK\""
    environment = {
      STIGMER_KEYS_CHECK   = local.keys_check_script
      STIGMER_NAMESPACE    = local.namespace
      STIGMER_NAME         = local.stigmer_name
      STIGMER_KEYS_SECRET  = local.keys_secret
      STIGMER_RECOVER_FROM = try(var.spec.database.bundled.recover_from.server_name, "")
    }
  }

  depends_on = [helm_release.operator, kubernetes_secret_v1.declared]
}

# The cleanup wait (scripts.tf cleanup_wait_script). It depends on the
# release, every Secret and the operator's namespace, and the Stigmer depends
# on it, so a destroy deletes the Stigmer first, then runs this while the
# operator and the bucket's credentials are still there.
#
# PARITY-EXCEPTION: a destroy-time step can read only its own state, so the
# script and its inputs live in this terraform_data's input, which nobody
# with access to the cluster can change. The Pulumi twin is a marker
# ConfigMap whose delete hook runs the same script from the program
# (module/gates.go cleanupMarker), the one object the engines do not share.
# Because the script is stored in state, a fix to it reaches a destroy after
# one apply.
resource "terraform_data" "cleanup_wait" {
  input = {
    script             = local.cleanup_wait_script
    namespace          = local.namespace
    name               = local.stigmer_name
    operator_namespace = local.operator_namespace
    operator_release   = local.operator_release
    wait_seconds       = tostring(local.cleanup_wait_seconds)
  }

  # The kubeconfig is read by the script from the tofu process environment
  # (KUBECONFIG, KUBE_CONFIG_PATH, KUBE_CTX), never from state, so a destroy
  # from another machine reaches the cluster the way every other step does.
  provisioner "local-exec" {
    when        = destroy
    interpreter = ["/bin/sh", "-c"]
    command     = "eval \"$STIGMER_CLEANUP_WAIT\""
    environment = {
      STIGMER_CLEANUP_WAIT       = self.output.script
      STIGMER_NAMESPACE          = self.output.namespace
      STIGMER_NAME               = self.output.name
      STIGMER_OPERATOR_NAMESPACE = self.output.operator_namespace
      STIGMER_OPERATOR_RELEASE   = self.output.operator_release
      STIGMER_WAIT_SECONDS       = self.output.wait_seconds
    }
  }

  depends_on = [helm_release.operator, kubernetes_secret_v1.declared, kubernetes_namespace_v1.operator]
}

# The Stigmer, from only the keys the spec declares (locals.tf).
resource "kubectl_manifest" "stigmer" {
  yaml_body         = yamlencode(local.stigmer_document)
  server_side_apply = true
  force_conflicts   = true

  depends_on = [kubectl_manifest.keys_gate, terraform_data.cleanup_wait]
}
