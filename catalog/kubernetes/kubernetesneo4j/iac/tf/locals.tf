# Computed values for the KubernetesNeo4j module. Every resolution here has
# an exact twin in the Pulumi module's locals.go / values.go / secrets.go —
# keep them in lockstep.
#
# HCL DISCIPLINE (applies to every conditional object in this file):
# conditional entries are written as `key = cond ? value : null` inside ONE
# object literal, pruned with `{ for k, v in {...} : k => v if v != null }`.
# The two tempting alternatives are both broken: `cond ? {...} : {}`
# ternaries fail plan-time type unification when branches carry different
# attributes, and merge() over primitive-only sibling objects silently
# UNIFIES them into map(string) — numbers and booleans arrive in the chart
# values as strings. The null-prune form preserves every value's type.
#
# Optional nested blocks are read with try(): HCL's && does NOT
# short-circuit, so chained null checks still dereference the null. Optional
# SCALARS inside optional blocks are read with try(coalesce(x), null) — the
# null-safe read (coalesce rejects a lone null; try turns that rejection
# into the fallback).

locals {
  # Chart identity — must stay byte-identical with the Pulumi module's vars:
  # cross-engine chart-name drift deploys two different products from one
  # manifest.
  helm_chart_name = "neo4j"
  helm_chart_repo = "https://helm.neo4j.com/neo4j"

  # Release name — metadata.name, NOT a fixed chart name: several Neo4j
  # servers coexist in one cluster (and Enterprise cluster members are each
  # their own release). The chart names its always-created ClusterIP
  # Service after the release (neo4j.fullname = the release name when no
  # name overrides are set), which is what makes local.service_name below
  # deterministic.
  release_name = var.metadata.name

  # Chart version resolved to the pinned default when unset, so both
  # engines install the same chart whether or not the platform's defaulting
  # middleware ran — mirror of the Pulumi module's DefaultChartVersion.
  # Chart versions track Neo4j calendar releases.
  chart_version = coalesce(var.spec.chart_version, "2026.6.0")

  namespace = var.spec.namespace

  # Resource-identity labels stamped on the module-created satellites
  # (namespace, the auth Secret — never injected into the chart's own
  # resources; Helm owns those).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesNeo4j"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- server identity ----------------------------------------------------
  # neo4j.name is REQUIRED by the chart (its neo4j.name helper fails the
  # install when empty — nothing defaults it to the release name), so the
  # module always renders it: cluster_name when set (Enterprise members
  # sharing it form one cluster), else metadata.name.
  neo4j_name = var.spec.cluster_name != "" ? var.spec.cluster_name : var.metadata.name

  edition = coalesce(var.spec.edition, "community")

  # ---- auth -----------------------------------------------------------------
  # The chart contract: neo4j.passwordFromSecret names a Secret carrying key
  # NEO4J_AUTH with value "neo4j/<password>", and the chart LOOKS IT UP at
  # template time — the Secret must exist BEFORE the release (main.tf wires
  # the explicit dependency). The module materializes that Secret itself
  # unless the existing_secret arm references one the user owns: with the
  # password arm it carries the declared value, with auth absent it carries
  # the password random_password.admin generates. The chart is never left
  # to mint a credential nobody can find afterwards.
  auth_password           = try(coalesce(var.spec.auth.password), null)
  auth_existing_secret    = try(coalesce(var.spec.auth.existing_secret), null)
  create_auth_secret      = local.auth_existing_secret == null
  generate_admin_password = local.create_auth_secret && local.auth_password == null

  # The admin password that lands in the module's Secret: declared, or
  # generated. one() over the splat is null when the generator has no
  # instance, so a declared password never indexes a resource that does
  # not exist.
  admin_password = local.auth_password != null ? local.auth_password : one(random_password.admin[*].result)

  # The Secret name rendered into neo4j.passwordFromSecret (and exported as
  # auth_secret_name): the module-materialized "<name>-auth", or the
  # referenced existing Secret. Never empty.
  auth_secret_name = local.create_auth_secret ? "${var.metadata.name}-auth" : local.auth_existing_secret

  # ---- service --------------------------------------------------------------
  # DELIBERATE OVERRIDE OF THE CHART DEFAULT: the chart ships
  # services.neo4j.spec.type: LoadBalancer, which would provision a cloud
  # load balancer (or hang Pending) on every install. This component pins it
  # to ClusterIP unless spec.service.type says otherwise — exposure composes
  # from first-class kinds (KubernetesIngress, Gateway API) over the
  # exported service handle instead.
  service_type = try(coalesce(var.spec.service.type), null) != null ? var.spec.service.type : "ClusterIP"

  # The main ClusterIP Service the chart always creates — neo4j.fullname =
  # the release name (templates/neo4j-svc.yaml). Feeds the endpoint outputs.
  service_name = local.release_name

  # ---- data volume ------------------------------------------------------------
  # The chart REQUIRES volumes.data.mode (its values.yaml ships mode: "" and
  # the templates fail without one), so the module ALWAYS renders the data
  # volume: "dynamic" with the declared StorageClass, else
  # "defaultStorageClass" for the cluster default — size resolved to the
  # spec default (10Gi) either way.
  data_volume_size   = try(coalesce(var.spec.data_volume.size), null) != null ? var.spec.data_volume.size : "10Gi"
  data_storage_class = try(var.spec.data_volume.storage_class, "")

  # ---- neo4j.conf ----------------------------------------------------------------
  # The typed memory block renders as the three neo4j.conf memory keys,
  # merged over the free-form config map — TYPED KEYS WIN on collision (the
  # memory block is the declared interface for those keys; a user config
  # duplicate is silently overridden). Mirror of the Pulumi module's config
  # rendering.
  memory_config = try(var.spec.memory, null) == null ? {} : {
    for k, v in {
      "server.memory.heap.initial_size" = var.spec.memory.heap_initial != "" ? var.spec.memory.heap_initial : null
      "server.memory.heap.max_size"     = var.spec.memory.heap_max != "" ? var.spec.memory.heap_max : null
      "server.memory.pagecache.size"    = var.spec.memory.page_cache != "" ? var.spec.memory.page_cache : null
    } : k => v if v != null
  }
  neo4j_config = merge(var.spec.config, local.memory_config)

  # ---- container resources ---------------------------------------------------------
  # The chart's primary resources shape is flat {cpu, memory} applied to
  # BOTH requests and limits; it also accepts a full-format limits sub-map.
  # The module renders requests into the flat keys and declared limits into
  # the limits sub-map. NOTE the chart REJECTS installs below its floor
  # (500m CPU / 2Gi memory) — the module never defaults below that; when
  # spec.resources is empty nothing renders and the chart's own defaults
  # (1000m/2Gi) apply.
  neo4j_resources = try(var.spec.resources, null) == null ? null : {
    for k, v in {
      cpu    = try(var.spec.resources.requests.cpu, "") != "" ? var.spec.resources.requests.cpu : null
      memory = try(var.spec.resources.requests.memory, "") != "" ? var.spec.resources.requests.memory : null
      limits = try(var.spec.resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = var.spec.resources.limits.cpu != "" ? var.spec.resources.limits.cpu : null
          memory = var.spec.resources.limits.memory != "" ? var.spec.resources.limits.memory : null
        } : lk => lv if lv != null
      }
    } : k => v if v != null && v != {}
  }

  # ---- ssl -------------------------------------------------------------------------
  # Both privateKey.secretName and publicCertificate.secretName point at the
  # ONE resolved scope Secret: the chart mounts private.key and public.crt
  # from it (its subPath defaults). cert-manager Secrets carry
  # tls.key/tls.crt instead — the README documents the key bridge; the
  # module does NOT silently rewrite key names.
  ssl_bolt_secret  = try(var.spec.ssl.bolt.secret, "")
  ssl_https_secret = try(var.spec.ssl.https.secret, "")

  # ---- typed chart values (twin of the Pulumi module's buildHelmValues) --
  helm_values = {
    for k, v in {
      neo4j = {
        for nk, nv in {
          name    = local.neo4j_name
          edition = local.edition

          # The chart's own shape for license acceptance is the STRING
          # "yes"/"no" (not a bool); its default is "no", so the module
          # renders only the affirmative.
          acceptLicenseAgreement = var.spec.accept_license_agreement ? "yes" : null

          # The password itself NEVER appears here — only the Secret name.
          passwordFromSecret = local.auth_secret_name

          resources = local.neo4j_resources
        } : nk => nv if nv != null
      }

      volumes = {
        data = {
          for dk, dv in {
            mode = local.data_storage_class != "" ? "dynamic" : "defaultStorageClass"
            dynamic = local.data_storage_class != "" ? {
              storageClassName = local.data_storage_class
              accessModes      = ["ReadWriteOnce"]
              requests         = { storage = local.data_volume_size }
            } : null
            defaultStorageClass = local.data_storage_class == "" ? {
              accessModes = ["ReadWriteOnce"]
              requests    = { storage = local.data_volume_size }
            } : null
          } : dk => dv if dv != null
        }
      }

      config = length(local.neo4j_config) > 0 ? local.neo4j_config : null

      apoc_config = length(var.spec.apoc_config) > 0 ? var.spec.apoc_config : null

      # useNeo4jDefaultJvmArguments renders only when explicitly declared
      # (the chart default is already true).
      jvm = (length(var.spec.additional_jvm_arguments) > 0 || try(var.spec.use_default_jvm_arguments, null) != null) ? {
        for jk, jv in {
          additionalJvmArguments      = length(var.spec.additional_jvm_arguments) > 0 ? var.spec.additional_jvm_arguments : null
          useNeo4jDefaultJvmArguments = try(var.spec.use_default_jvm_arguments, null)
        } : jk => jv if jv != null
      } : null

      # The ClusterIP override lives here — see local.service_type above.
      services = {
        neo4j = {
          for sk, sv in {
            spec        = { type = local.service_type }
            annotations = length(try(var.spec.service.annotations, {})) > 0 ? var.spec.service.annotations : null
          } : sk => sv if sv != null
        }
      }

      ssl = (local.ssl_bolt_secret != "" || local.ssl_https_secret != "") ? {
        for sk, sv in {
          bolt = local.ssl_bolt_secret != "" ? {
            privateKey        = { secretName = local.ssl_bolt_secret }
            publicCertificate = { secretName = local.ssl_bolt_secret }
          } : null
          https = local.ssl_https_secret != "" ? {
            privateKey        = { secretName = local.ssl_https_secret }
            publicCertificate = { secretName = local.ssl_https_secret }
          } : null
        } : sk => sv if sv != null
      } : null

      # ---- scheduling ------------------------------------------------------
      # The chart reads nodeSelector at the TOP level; tolerations,
      # podAntiAffinity, and priorityClassName live under podSpec.
      # podAntiAffinity renders only when explicitly declared (chart
      # default is already true).
      nodeSelector = length(try(var.spec.scheduling.node_selector, {})) > 0 ? var.spec.scheduling.node_selector : null
      podSpec = try(var.spec.scheduling, null) == null ? null : {
        for pk, pv in {
          tolerations = length(var.spec.scheduling.tolerations) > 0 ? [
            for t in var.spec.scheduling.tolerations : {
              for tk, tv in {
                key               = t.key != "" ? t.key : null
                operator          = t.operator != "" ? t.operator : null
                value             = t.value != "" ? t.value : null
                effect            = t.effect != "" ? t.effect : null
                tolerationSeconds = try(t.toleration_seconds, null)
              } : tk => tv if tv != null
            }
          ] : null
          podAntiAffinity   = try(var.spec.scheduling.pod_anti_affinity, null)
          priorityClassName = var.spec.scheduling.priority_class_name != "" ? var.spec.scheduling.priority_class_name : null
        } : pk => pv if pv != null
      }

      serviceMonitor = var.spec.service_monitor_enabled ? { enabled = true } : null

      # The chart's separated image fields are image.registry / repository /
      # tag. It FAILS when any separated field is set while repository is
      # empty, so the module resolves the spec's documented repository
      # default ("neo4j") whenever the block renders anything.
      image = (try(var.spec.image, null) != null && (try(var.spec.image.registry, "") != "" || try(var.spec.image.repository, "") != "" || try(var.spec.image.tag, "") != "")) ? {
        for ik, iv in {
          registry   = var.spec.image.registry != "" ? var.spec.image.registry : null
          repository = var.spec.image.repository != "" ? var.spec.image.repository : "neo4j"
          tag        = var.spec.image.tag != "" ? var.spec.image.tag : null
        } : ik => iv if iv != null
      } : null
    } : k => v if v != null && v != {}
  }
}
