# Every resolution here has an exact twin in the Pulumi module's
# locals.go / values.go — keep them in lockstep.
#
# HCL DISCIPLINE (the type-unification class): every conditional block
# below is a SINGLE-ATTRIBUTE ternary against {} or rides the null-prune
# merge idiom — a ternary whose true branch bundles attributes of more
# than one type cannot unify against {}. Tolerations entries mix
# attribute sets per element → the jsonencode/jsondecode seam.

locals {
  # Chart identity — must stay byte-identical with the Pulumi module's
  # vars (cross-engine chart-name drift deploys two different products
  # from one manifest). SERVING TRUTH: the chart's home is the OCI
  # registry — the classic index at charts.deliveryhero.io stalls at
  # 0.31.6 (2024) while ghcr.io serves the live line. The Terraform
  # provider takes repository = the OCI registry path plus the bare
  # chart name; the Pulumi twin passes the joined "oci://.../locust"
  # string as the chart reference. Same chart bytes, different wiring.
  helm_oci_repo      = "oci://ghcr.io/deliveryhero/helm-charts"
  helm_chart_name    = "locust"
  helm_chart_version = "0.35.0"

  # The official image and the release this kind is built against.
  default_image_repository = "locustio/locust"
  default_image_tag        = "2.32.2"

  # The web-UI/REST port and the worker-connect port.
  web_port         = 8089
  master_bind_port = 5557

  # In-pod mount paths: the chart's locustfile path (chart default,
  # pinned explicitly — the master's -f argument derives from it), the
  # login-backend code and the credential files.
  locustfile_mount_path      = "/mnt/locust"
  web_auth_code_mount_path   = "/opt/planton/web-auth-code"
  web_auth_secret_mount_path = "/opt/planton/web-auth"

  # The pod-template annotation carrying the module's content hash —
  # the chart checksums only its OWN ConfigMaps, so module-owned
  # script changes roll the pods through this annotation instead.
  checksum_annotation = "planton.dev/config-checksum"

  # Name budget: the longest derived child name is the module's own
  # `<name>-locustfile` ConfigMap (11-char suffix) — 63 - 11 = 52.
  name_budget = 52

  release_name = var.metadata.name
  namespace    = var.spec.namespace

  # Resource-identity labels stamped on the module-created satellites
  # (the namespace, the script ConfigMaps and the auth Secret — never
  # injected into the chart's own resources; Helm owns those).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesLocust"
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {},
  )

  # The load-test name — labels every chart resource (`load_test:
  # <name>` rides the Deployments' IMMUTABLE selector labels). Empty in
  # the spec = the release name.
  load_test_name = coalesce(try(var.spec.load_test.name, ""), local.release_name)

  # ------------------------- script delivery -----------------------------
  # Inline scripts render into module-owned ConfigMaps; the
  # existing-ConfigMaps arm mounts the user's own. Either way the chart
  # values name the ConfigMaps EXPLICITLY — the chart's bundled-example
  # defaults are a fragile literal-string coupling (they render empty
  # the moment loadtest.name changes) and are never engaged.
  inline_scripts       = try(var.spec.load_test.inline, null)
  existing_config_maps = try(var.spec.load_test.existing_config_maps, null)
  module_owned_scripts = local.inline_scripts != null

  lib_files = local.module_owned_scripts ? try(local.inline_scripts.lib_files, {}) : {}

  locustfile_config_map = local.module_owned_scripts ? "${local.release_name}-locustfile" : local.existing_config_maps.locustfile_config_map
  locustfile_name       = local.module_owned_scripts ? "main.py" : coalesce(try(local.existing_config_maps.locustfile_name, "main.py"), "main.py")
  lib_config_map        = local.module_owned_scripts ? (length(local.lib_files) > 0 ? "${local.release_name}-lib" : "") : try(local.existing_config_maps.lib_config_map, "")

  # --------------------------- web-UI login ------------------------------
  # The secured default: an ABSENT web_ui_auth block means the login is
  # ON with a module-generated credential — the chart's own default (an
  # open UI that can fire load at any reachable host) never ships.
  # Headless runs start no web UI, so there is nothing to protect and
  # the login machinery is skipped.
  headless          = try(var.spec.load_test.headless, false)
  web_login_enabled = local.headless ? false : try(var.spec.web_ui_auth.enabled, null) == false ? false : true

  web_username       = local.web_login_enabled ? coalesce(try(var.spec.web_ui_auth.username, "locust"), "locust") : ""
  auth_secret_name   = local.web_login_enabled ? "${local.release_name}-auth" : ""
  web_auth_code_name = local.web_login_enabled ? "${local.release_name}-web-auth" : ""

  # The login tag floor: the chart renders the modern `--web-login`
  # flag only for image tags >= 2.21.0; BELOW it the chart falls onto
  # `--web-auth=user:password` — credentials as a LITERAL POD ARGUMENT,
  # which this module refuses to render. Non-numeric tags (e.g.
  # "latest") cannot prove the floor, so they fail too (the
  # helm_release precondition).
  image_tag               = coalesce(try(var.spec.image.tag, ""), local.default_image_tag)
  image_tag_version       = try(regex("^v?(\\d+)\\.(\\d+)", local.image_tag), null)
  image_tag_login_capable = local.image_tag_version != null ? (tonumber(local.image_tag_version[0]) > 2 || (tonumber(local.image_tag_version[0]) == 2 && tonumber(local.image_tag_version[1]) >= 21)) : false

  # The module-owned login backend delivered next to the user's
  # locustfile when the web-UI login is on. WHY CODE: Locust's
  # --web-login flag protects every web and REST route behind a
  # session, but deliberately leaves the credential backend to
  # locustfile code — the documented extension seam (the shape follows
  # upstream's own examples/web_ui_auth/basic.py at the pin). The
  # chart's `master.auth` username/password values feed ONLY the legacy
  # pre-2.21 code path that renders credentials as pod arguments —
  # never engaged by this module. PARITY: byte-identical with the
  # Pulumi module's webauth.go.
  # parity: webauth.go webAuthBackendPy
  web_auth_backend_py = <<EOT
"""Platform-managed login for the Locust web UI.

Locust protects its web routes behind a session when started with
--web-login, and delegates the credential backend to locustfile code.
This module implements a single-credential backend: the username,
password and the Flask session-signing key are read from
Secret-projected files, so nothing secret rides rendered values or
process arguments, and sessions survive pod restarts.
"""

import os
import secrets

from flask import Blueprint, redirect, request, session, url_for
from flask_login import UserMixin, login_user

from locust import events

_AUTH_DIR = "/opt/planton/web-auth"


def _read(name):
    with open(os.path.join(_AUTH_DIR, name)) as f:
        return f.read().strip()


class _AuthUser(UserMixin):
    def __init__(self, username):
        self.username = username

    def get_id(self):
        return self.username


@events.init.add_listener
def _planton_web_login(environment, **_kwargs):
    if not environment.web_ui:
        return

    username = _read("username")
    password = _read("password")

    web_ui = environment.web_ui
    web_ui.app.config["SECRET_KEY"] = _read("flask-secret-key")
    web_ui.login_manager.user_loader(
        lambda user_id: _AuthUser(user_id) if user_id == username else None
    )

    base_path = environment.parsed_options.web_base_path
    web_ui.auth_args = {
        "username_password_callback": base_path + "/planton/login",
    }

    blueprint = Blueprint("planton_auth", __name__, url_prefix=base_path)

    @blueprint.route("/planton/login", methods=["POST"])
    def _login():
        username_ok = secrets.compare_digest(
            request.form.get("username", ""), username
        )
        password_ok = secrets.compare_digest(
            request.form.get("password", ""), password
        )
        if username_ok and password_ok:
            login_user(_AuthUser(username))
            return redirect(url_for("locust.index"))
        session["auth_error"] = "Invalid username or password"
        return redirect(url_for("locust.login"))

    web_ui.app.register_blueprint(blueprint)
EOT

  # The master's `-f`: the locustfile plus the login backend,
  # comma-separated absolute paths (Locust's own multi-locustfile
  # form). Command-line -f overrides the LOCUST_LOCUSTFILE env, so the
  # master loads BOTH files while workers (default env) load the
  # locustfile alone.
  master_args = ["-f", "${local.locustfile_mount_path}/${local.locustfile_name},${local.web_auth_code_mount_path}/planton_auth.py"]

  # The `<name>-auth` Secret's three keys projected as files under the
  # login backend's read path (the chart's mount_external_secret seam).
  web_auth_secret_mount = {
    mountPath = local.web_auth_secret_mount_path
    files = {
      (local.auth_secret_name) = ["username", "password", "flask-secret-key"]
    }
  }

  # ------------------------- content checksum ----------------------------
  # Hashes every module-owned input that reaches the pods OUTSIDE
  # chart-rendered resources — the chart's own checksum annotations
  # cover only ITS ConfigMaps/Secret, so without this hash a locustfile
  # edit would update the ConfigMap and roll nothing. The
  # record-separator join is expressible in BOTH engines — the hashes
  # must stay byte-identical across the twins (Pulumi:
  # configChecksum in locals.go).
  checksum_parts = concat(
    [
      "locustfile-configmap=${local.locustfile_config_map}",
      "locustfile-name=${local.locustfile_name}",
      "lib-configmap=${local.lib_config_map}",
      "web-login=${local.web_login_enabled}",
    ],
    local.module_owned_scripts ? ["locustfile-content=${local.inline_scripts.locustfile_content}"] : [],
    local.module_owned_scripts ? [for name in sort(keys(local.lib_files)) : "lib:${name}=${local.lib_files[name]}"] : [],
    local.web_login_enabled ? ["web-auth-code=${local.web_auth_backend_py}", "web-username=${local.web_username}"] : [],
  )
  config_checksum = sha256(join("\u001e", local.checksum_parts))

  # ------------------------------ loadtest --------------------------------
  loadtest_block = merge(
    {
      name                           = local.load_test_name
      locust_locustfile              = local.locustfile_name
      locust_locustfile_path         = local.locustfile_mount_path
      locust_locustfile_configmap    = local.locustfile_config_map
      locust_lib_configmap           = local.lib_config_map
      pip_requirementsfile_configmap = try(var.spec.load_test.pip_requirements_config_map, "")
      headless                       = local.headless
      # The target host — resolved literal (a reference resolves
      # before the module runs). Empty stays empty: the locustfile
      # must then declare its own host.
      locust_host = try(var.spec.load_test.target_host, "")
    },
    length(try(var.spec.load_test.pip_packages, [])) > 0 ? { pip_packages = var.spec.load_test.pip_packages } : {},
    length(try(var.spec.load_test.environment, {})) > 0 ? { environment = var.spec.load_test.environment } : {},
    length(try(var.spec.load_test.env_from_secrets, [])) > 0 ? { environment_load_from_secrets = var.spec.load_test.env_from_secrets } : {},
    length(try(var.spec.load_test.env_from_secret_keys, [])) > 0 ? { environment_external_secret = { for entry in var.spec.load_test.env_from_secret_keys : entry.secret_name => entry.keys } } : {},
    # The chart takes tags as ONE space-joined string (rendered into
    # `--tags`/`--exclude-tags`); the CEL forbids whitespace inside a
    # tag, so the join is unambiguous.
    length(try(var.spec.load_test.tags, [])) > 0 ? { tags = join(" ", var.spec.load_test.tags) } : {},
    length(try(var.spec.load_test.exclude_tags, [])) > 0 ? { excludeTags = join(" ", var.spec.load_test.exclude_tags) } : {},
    local.web_login_enabled ? { mount_external_secret = local.web_auth_secret_mount } : {},
  )

  # ------------------------------- master ---------------------------------
  master_resources = {
    requests = {
      cpu    = try(var.spec.master.resources.requests.cpu, "")
      memory = try(var.spec.master.resources.requests.memory, "")
    }
    limits = {
      cpu    = try(var.spec.master.resources.limits.cpu, "")
      memory = try(var.spec.master.resources.limits.memory, "")
    }
  }
  master_resources_block = (
    local.master_resources.requests.cpu != "" || local.master_resources.requests.memory != "" ||
    local.master_resources.limits.cpu != "" || local.master_resources.limits.memory != ""
    ) ? {
    resources = merge(
      local.master_resources.requests.cpu != "" || local.master_resources.requests.memory != "" ? {
        requests = merge(
          local.master_resources.requests.cpu != "" ? { cpu = local.master_resources.requests.cpu } : {},
          local.master_resources.requests.memory != "" ? { memory = local.master_resources.requests.memory } : {},
        )
      } : {},
      local.master_resources.limits.cpu != "" || local.master_resources.limits.memory != "" ? {
        limits = merge(
          local.master_resources.limits.cpu != "" ? { cpu = local.master_resources.limits.cpu } : {},
          local.master_resources.limits.memory != "" ? { memory = local.master_resources.limits.memory } : {},
        )
      } : {},
    )
  } : {}

  master_tolerations = [for t in try(var.spec.master.scheduling.tolerations, []) : jsondecode(jsonencode(merge(
    t.key != "" ? { key = t.key } : {},
    t.operator != "" ? { operator = t.operator } : {},
    t.value != "" ? { value = t.value } : {},
    t.effect != "" ? { effect = t.effect } : {},
    try(t.toleration_seconds, null) != null ? { tolerationSeconds = t.toleration_seconds } : {},
  )))]

  master_block = merge(
    {
      logLevel = coalesce(try(var.spec.master.log_level, "INFO"), "INFO")
      # The module's content hash — pod-template annotation; see the
      # checksum comment above.
      annotations = { (local.checksum_annotation) = local.config_checksum }
      # The login wiring: auth.enabled renders `--web-login` (the tag
      # floor guarantees the modern path — the legacy
      # credentials-as-arguments path never renders).
      auth = { enabled = local.web_login_enabled }
    },
    local.master_resources_block,
    try(var.spec.master.pdb_enabled, false) ? { pdb = { enabled = true } } : {},
    local.web_login_enabled ? { args = local.master_args } : {},
    length(try(var.spec.master.scheduling.node_selector, {})) > 0 ? { nodeSelector = var.spec.master.scheduling.node_selector } : {},
    length(local.master_tolerations) > 0 ? { tolerations = local.master_tolerations } : {},
  )

  # ------------------------------- workers --------------------------------
  worker_resources = {
    requests = {
      cpu    = try(var.spec.workers.resources.requests.cpu, "")
      memory = try(var.spec.workers.resources.requests.memory, "")
    }
    limits = {
      cpu    = try(var.spec.workers.resources.limits.cpu, "")
      memory = try(var.spec.workers.resources.limits.memory, "")
    }
  }
  worker_resources_block = (
    local.worker_resources.requests.cpu != "" || local.worker_resources.requests.memory != "" ||
    local.worker_resources.limits.cpu != "" || local.worker_resources.limits.memory != ""
    ) ? {
    resources = merge(
      local.worker_resources.requests.cpu != "" || local.worker_resources.requests.memory != "" ? {
        requests = merge(
          local.worker_resources.requests.cpu != "" ? { cpu = local.worker_resources.requests.cpu } : {},
          local.worker_resources.requests.memory != "" ? { memory = local.worker_resources.requests.memory } : {},
        )
      } : {},
      local.worker_resources.limits.cpu != "" || local.worker_resources.limits.memory != "" ? {
        limits = merge(
          local.worker_resources.limits.cpu != "" ? { cpu = local.worker_resources.limits.cpu } : {},
          local.worker_resources.limits.memory != "" ? { memory = local.worker_resources.limits.memory } : {},
        )
      } : {},
    )
  } : {}

  worker_tolerations = [for t in try(var.spec.workers.scheduling.tolerations, []) : jsondecode(jsonencode(merge(
    t.key != "" ? { key = t.key } : {},
    t.operator != "" ? { operator = t.operator } : {},
    t.value != "" ? { value = t.value } : {},
    t.effect != "" ? { effect = t.effect } : {},
    try(t.toleration_seconds, null) != null ? { tolerationSeconds = t.toleration_seconds } : {},
  )))]

  # Autoscaling arm resolution (the spec's oneof guarantees at most
  # one).
  worker_hpa  = try(var.spec.workers.hpa, null)
  worker_keda = try(var.spec.workers.keda, null)

  # CHART CONTRACT: the KEDA ScaledObject reuses worker.hpa
  # minReplicas/maxReplicas for its bounds while worker.hpa.enabled
  # stays false, and the worker Deployment still renders `replicas` on
  # the KEDA arm (the template gates it on hpa.enabled only) — the
  # module pins replicas to the KEDA floor so a Helm upgrade resets
  # scaling to the floor, not to an unrelated count.
  keda_min_replicas = local.worker_keda != null ? coalesce(try(local.worker_keda.min_replicas, 1), 1) : 1

  # The default KEDA trigger reads the LIVE USER COUNT from the
  # master's own stats API — rendered explicitly and pinned (never the
  # chart's tpl'd default), with the spec-level CEL guaranteeing that
  # API is reachable (login off, non-headless). custom_triggers
  # replaces it wholesale.
  keda_triggers = local.worker_keda == null ? "" : (
    try(local.worker_keda.custom_triggers, "") != "" ? local.worker_keda.custom_triggers : <<-EOT
      - type: metrics-api
        metadata:
          activationTargetValue: "0"
          targetValue: "${coalesce(try(local.worker_keda.target_users_per_worker, 50), 50)}"
          url: "${local.web_endpoint}/stats/requests"
          format: json
          valueLocation: user_count
    EOT
  )

  worker_block = merge(
    {
      logLevel    = coalesce(try(var.spec.workers.log_level, "INFO"), "INFO")
      annotations = { (local.checksum_annotation) = local.config_checksum }
    },
    local.worker_resources_block,
    try(var.spec.workers.pdb_enabled, false) ? { pdb = { enabled = true } } : {},
    length(try(var.spec.workers.scheduling.node_selector, {})) > 0 ? { nodeSelector = var.spec.workers.scheduling.node_selector } : {},
    length(local.worker_tolerations) > 0 ? { tolerations = local.worker_tolerations } : {},
    # Exactly one of the three shapes below contributes replica/scaling
    # keys (single-attribute ternaries; the hpa/keda objects are
    # single-typed maps and unify cleanly).
    local.worker_hpa != null ? {
      hpa = {
        enabled                        = true
        minReplicas                    = coalesce(try(local.worker_hpa.min_replicas, 1), 1)
        maxReplicas                    = local.worker_hpa.max_replicas
        targetCPUUtilizationPercentage = coalesce(try(local.worker_hpa.target_cpu_utilization_percent, 40), 40)
      }
    } : {},
    local.worker_keda != null ? { replicas = local.keda_min_replicas } : {},
    local.worker_keda != null ? {
      hpa = {
        enabled     = false
        minReplicas = local.keda_min_replicas
        maxReplicas = local.worker_keda.max_replicas
      }
    } : {},
    local.worker_keda != null ? {
      keda = {
        enabled         = true
        pollingInterval = coalesce(try(local.worker_keda.polling_interval_seconds, 15), 15)
        cooldownPeriod  = coalesce(try(local.worker_keda.cooldown_period_seconds, 30), 30)
        triggers        = local.keda_triggers
      }
    } : {},
    local.worker_hpa == null && local.worker_keda == null ? { replicas = coalesce(try(var.spec.workers.replicas, 1), 1) } : {},
  )

  # ------------------------------- service --------------------------------
  # Ports pinned to the chart defaults so the rendered Service (and the
  # exported endpoints) never drift with a chart-default change.
  service_block = merge(
    {
      type       = coalesce(try(var.spec.service.type, "ClusterIP"), "ClusterIP")
      port       = local.web_port
      targetPort = local.web_port
    },
    length(try(var.spec.service.annotations, {})) > 0 ? { annotations = var.spec.service.annotations } : {},
  )

  # --------------------------- values document ----------------------------
  helm_values = merge(
    {
      # Deterministic child names (`<name>-master`, `<name>-worker`,
      # the bare `<name>` Service) — the release name never
      # double-prefixes and the import map stays exact (the fullname
      # re-pin discipline).
      fullnameOverride = local.release_name
      # The COMBINED image form; the tag always pinned explicitly so
      # the deployed Locust version is declared, never inherited.
      image = {
        repository = coalesce(try(var.spec.image.repository, ""), local.default_image_repository)
        tag        = local.image_tag
      }
      loadtest = local.loadtest_block
      master   = local.master_block
      worker   = local.worker_block
      service  = local.service_block
    },
    length(try(var.spec.image_pull_secrets, [])) > 0 ? { imagePullSecrets = [for name in var.spec.image_pull_secrets : { name = name }] } : {},
    local.web_login_enabled ? { extraConfigMaps = { (local.web_auth_code_name) = local.web_auth_code_mount_path } } : {},
  )

  # ------------------------------- re-pins --------------------------------
  # The deliberate exceptions to the escape hatch's last-word contract
  # (the THIRD values document; twin of the Pulumi module's re-pins):
  # the deterministic names, the script wiring, the login wiring — and
  # the chart's values-rendered-Secret path, NULLED so credentials can
  # never ride rendered values (Helm null-deletion; the Pulumi twin
  # force-empties the same key after its own merge).
  helm_values_repins = merge(
    {
      fullnameOverride = local.release_name
      loadtest = merge(
        {
          locust_locustfile_configmap = local.locustfile_config_map
          locust_locustfile           = local.locustfile_name
          locust_locustfile_path      = local.locustfile_mount_path
          locust_lib_configmap        = local.lib_config_map
          environment_secret          = null
        },
        local.web_login_enabled ? { mount_external_secret = local.web_auth_secret_mount } : {},
      )
      # Branch-shaped conditionals decompose into merges of
      # complementary single-attribute ternaries (the HCL
      # type-unification class — a branch bundling auth + args cannot
      # unify against a branch without args).
      master = merge(
        { auth = { enabled = local.web_login_enabled } },
        local.web_login_enabled ? { args = local.master_args } : {},
      )
    },
    local.web_login_enabled ? { extraConfigMaps = { (local.web_auth_code_name) = local.web_auth_code_mount_path } } : {},
  )

  # ------------------------------- outputs --------------------------------
  master_service       = local.release_name
  web_endpoint         = "http://${local.release_name}.${local.namespace}.svc.cluster.local:${local.web_port}"
  master_bind_endpoint = "${local.release_name}.${local.namespace}.svc.cluster.local:${local.master_bind_port}"
  port_forward_command = "kubectl port-forward svc/${local.release_name} -n ${local.namespace} ${local.web_port}:${local.web_port}"

  # The credential handles are honest: with the login disabled (or a
  # headless run, which starts no web UI) no module-owned credential
  # exists, so the handles export EMPTY rather than names that point at
  # nothing (Pulumi twin exports the same empties).
  web_ui_password_secret_output_name = local.web_login_enabled ? local.auth_secret_name : ""
  web_ui_password_secret_output_key  = local.web_login_enabled ? "password" : ""
}
