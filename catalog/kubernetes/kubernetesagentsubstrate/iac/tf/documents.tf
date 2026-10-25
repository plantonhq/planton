# The composition (compose.go, workloads.go, splice.go, patch.go): the
# published release files, the typed settings applied onto upstream's own
# objects, upstream's own components applied where a setting selects them,
# and the documents grouped into the four phases main.tf applies in order.
#
# PATTERN: a patch that applies only to one item of a list is written as
# merge(item, { for k, v in PATCH : k => v if CONDITION }), the HCL form of
# a conditional merge that keeps every item's own shape. applyFlags is
# spelled out per container: a flag upstream already passes keeps its
# position and takes the new value; every other flag is appended in order.

# A dropped connection, HTTP 429 or a 5xx is retried (honoring Retry-After),
# as the Pulumi twin's fetchText does: a plan fetches every release file, so
# a host limiting a burst of plans slows the deploy down. The three reads run
# in the Go module's order, each depending on the one before: the seal, then
# the image list and its install shape, then every other file, so a release
# Planton never published, or one of another shape, is refused by name
# before anything else is read.
data "http" "checksums" {
  url                = "${local.version_url}${local.checksums_file}"
  request_timeout_ms = 120000

  retry {
    attempts     = 3
    min_delay_ms = 1000
    max_delay_ms = 8000
  }

  lifecycle {
    postcondition {
      condition     = self.status_code == 200
      error_message = self.status_code == 429 ? "observed: ${self.url} answered HTTP 429 on every attempt\nmeaning: the host is limiting how often this address may download; a burst of plans from one address can reach the limit\nnext step: wait a minute and re-run" : self.status_code >= 500 ? "observed: ${self.url} answered HTTP ${self.status_code} on every attempt\nmeaning: the host is failing, not the release\nnext step: re-run once it answers 200 (`curl -I ${self.url}`)" : "observed: ${self.url} answered HTTP ${self.status_code}\nmeaning: Planton has not published Agent Substrate ${local.version}: every published release carries SHA256SUMS, uploaded last\nnext step: set spec.version to a release Planton publishes (its ${local.release_base_url}<version>/${local.checksums_file} answers 200), or leave it empty for ${local.release}"
    }
  }
}

data "http" "images" {
  url                = "${local.version_url}${local.images_file}"
  request_timeout_ms = 120000
  depends_on         = [data.http.checksums]

  retry {
    attempts     = 3
    min_delay_ms = 1000
    max_delay_ms = 8000
  }

  lifecycle {
    postcondition {
      condition     = self.status_code == 200
      error_message = self.status_code == 429 ? "observed: ${self.url} answered HTTP 429 on every attempt\nmeaning: the host is limiting how often this address may download; a burst of plans from one address can reach the limit\nnext step: wait a minute and re-run" : self.status_code >= 500 ? "observed: ${self.url} answered HTTP ${self.status_code} on every attempt\nmeaning: the host is failing, not the release\nnext step: re-run once it answers 200 (`curl -I ${self.url}`)" : "observed: ${self.url} answered HTTP ${self.status_code}\nmeaning: Agent Substrate ${local.version} is published without ${local.images_file}, which every release of its install shape carries\nnext step: report it to Planton; a published release never changes, so re-running will not help"
    }
    postcondition {
      condition     = sha256(self.response_body) == try(local.seal[local.images_file], "")
      error_message = "observed: ${self.url} does not match the release's SHA256SUMS\nmeaning: the download was altered on its way, or the file changed after the release was sealed\nnext step: re-run; if it repeats, report it to Planton"
    }
    postcondition {
      condition     = coalesce(try(jsondecode(self.response_body).shape, ""), local.version) == local.release
      error_message = "observed: Agent Substrate ${local.version} has the install shape of ${coalesce(try(jsondecode(self.response_body).shape, ""), local.version)}, and this ${local.kind_name} composes ${local.release}'s\nmeaning: ${local.version} installs differently from what this Planton catalog release knows how to compose, so nothing was applied\nnext step: set spec.version to ${local.release} or a release of its shape, or deploy this resource from a Planton catalog release that composes ${coalesce(try(jsondecode(self.response_body).shape, ""), local.version)}'s shape"
    }
  }
}

data "http" "release_file" {
  for_each = local.release_files

  url                = "${local.version_url}${each.value}"
  request_timeout_ms = 120000
  depends_on         = [data.http.images]

  retry {
    attempts     = 3
    min_delay_ms = 1000
    max_delay_ms = 8000
  }

  lifecycle {
    postcondition {
      condition     = self.status_code == 200
      error_message = self.status_code == 429 ? "observed: ${self.url} answered HTTP 429 on every attempt\nmeaning: the host is limiting how often this address may download; a burst of plans from one address can reach the limit\nnext step: wait a minute and re-run" : self.status_code >= 500 ? "observed: ${self.url} answered HTTP ${self.status_code} on every attempt\nmeaning: the host is failing, not the release\nnext step: re-run once it answers 200 (`curl -I ${self.url}`)" : "observed: ${self.url} answered HTTP ${self.status_code}\nmeaning: Agent Substrate ${local.version} is published without ${each.value}, which every release of its install shape carries\nnext step: report it to Planton; a published release never changes, so re-running will not help"
    }
    postcondition {
      condition     = sha256(self.response_body) == try(local.seal[each.value], "")
      error_message = "observed: ${self.url} does not match the release's SHA256SUMS\nmeaning: the download was altered on its way, or the file changed after the release was sealed\nnext step: re-run; if it repeats, report it to Planton"
    }
  }
}

locals {
  # The release's seal: path -> SHA-256, from SHA256SUMS (parseChecksums).
  seal = {
    for line in split("\n", data.http.checksums.response_body) : split("  ", line)[1] => split("  ", line)[0]
    if length(split("  ", line)) == 2
  }
}

locals {
  # ---- the published files, parsed --------------------------------------------
  # Split on separator lines; each piece gets back the newline the split
  # consumed, so a block scalar that ends a document keeps its final newline
  # (the Go module splits before the separator and keeps it).
  #
  # YAML 1.1 reads a leading-zero file mode (defaultMode: 0600) as octal, as
  # Kubernetes and the Go module read it; yamldecode follows YAML 1.2 and
  # would read 600 decimal, a mode the API server refuses. So each file's
  # octal mode literals are rewritten as their decimal value before decoding
  # (up to four distinct ones per file; the preflight refuses more).
  release_bodies = { for role, path in local.release_files : role => data.http.release_file[role].response_body }
  octal_modes = {
    for role, body in local.release_bodies : role => distinct([
      for m in regexall("(?m)^[ \\t]*(?:defaultMode|mode):[ \\t]+(0[0-7]+)[ \\t]*$", body) : m[0]
    ])
  }
  decoded_bodies = {
    for role, body in local.release_bodies : role => replace(replace(replace(replace(body,
      "/(?m)^([ \\t]*(?:defaultMode|mode):[ \\t]+)${try(local.octal_modes[role][0], "x")}([ \\t]*)$/", "$${1}${try(parseint(local.octal_modes[role][0], 8), 0)}$${2}"),
      "/(?m)^([ \\t]*(?:defaultMode|mode):[ \\t]+)${try(local.octal_modes[role][1], "x")}([ \\t]*)$/", "$${1}${try(parseint(local.octal_modes[role][1], 8), 0)}$${2}"),
      "/(?m)^([ \\t]*(?:defaultMode|mode):[ \\t]+)${try(local.octal_modes[role][2], "x")}([ \\t]*)$/", "$${1}${try(parseint(local.octal_modes[role][2], 8), 0)}$${2}"),
    "/(?m)^([ \\t]*(?:defaultMode|mode):[ \\t]+)${try(local.octal_modes[role][3], "x")}([ \\t]*)$/", "$${1}${try(parseint(local.octal_modes[role][3], 8), 0)}$${2}")
  }
  documents_by_file = {
    for role, path in local.release_files : role => [
      for doc in split("\n---", "\n${local.decoded_bodies[role]}") : yamldecode("${doc}\n")
      if trimspace(doc) != "" && can(yamldecode("${doc}\n").kind)
    ] if role != "setup_sql"
  }
  images = jsondecode(data.http.images.response_body).images

  # ---- the pod-certificate controller ------------------------------------------
  podcert_upstream  = one([for d in local.documents_by_file.pod_certificates : d if d.kind == "Deployment" && d.metadata.name == local.pod_certificates_name])
  podcert_container = one([for c in local.podcert_upstream.spec.template.spec.containers : c if c.name == "controller"])
  podcert_size10_args = [
    for op in yamldecode(local.documents_by_file.podcert_size10[0].patches[0].patch) : op.value
  ]
  podcert_base_args   = concat(local.podcert_container.args, local.size10 ? local.podcert_size10_args : [])
  podcert_flag_values = { for f in local.pod_cert_flags : f.name => f.value }
  podcert_args = concat(
    [for a in local.podcert_base_args : substr(a, 0, 2) == "--" && contains(keys(local.podcert_flag_values), trimprefix(split("=", a)[0], "--")) ? "--${trimprefix(split("=", a)[0], "--")}=${local.podcert_flag_values[trimprefix(split("=", a)[0], "--")]}" : a],
    [for f in local.pod_cert_flags : "--${f.name}=${f.value}" if !contains([for a in local.podcert_base_args : trimprefix(split("=", a)[0], "--") if substr(a, 0, 2) == "--"], f.name)],
  )
  podcert_doc = merge(local.podcert_upstream, {
    spec = merge(local.podcert_upstream.spec, {
      template = merge(local.podcert_upstream.spec.template, {
        spec = merge(local.podcert_upstream.spec.template.spec, {
          containers = [for c in local.podcert_upstream.spec.template.spec.containers : merge(c, { for k, v in merge(
            {
              args = local.podcert_args
              env  = [for e in try(c.env, []) : merge(e, { for k2, v2 in { value = tostring(local.pod_certs.workers_per_signer) } : k2 => v2 if e.name == "WORKERS_PER_SIGNER" && local.pod_certs.workers_per_signer != 1 })]
            },
            { for k3, v3 in { resources = local.res.pod_certs } : k3 => v3 if length(local.res.pod_certs) > 0 },
          ) : k => v if c.name == "controller" })]
        })
      })
    })
  })

  # ---- resources, per component (setResources) --------------------------------
  resource_sources = {
    api_server = local.api.resources
    identity   = local.identity.resources
    controller = local.controller.resources
    node_agent = local.node_agent.resources
    router     = local.router.resources
    egress     = local.egress.resources
    provider   = local.egress.provider_resources
    pod_certs  = local.pod_certs.resources
    postgres   = local.db.resources
  }
  res = {
    for component, r in local.resource_sources : component => {
      for field, cm in { requests = try(r.requests, null), limits = try(r.limits, null) } : field => {
        for k, v in { cpu = try(cm.cpu, ""), memory = try(cm.memory, "") } : k => v if v != ""
      } if length({ for k, v in { cpu = try(cm.cpu, ""), memory = try(cm.memory, "") } : k => v if v != "" }) > 0
    }
  }

  # ---- the bundled database (patchPostgres) ------------------------------------
  postgres_config_upstream = one([for d in local.documents_by_file.postgres : d if d.kind == "ConfigMap" && d.metadata.name == "postgres-config"])
  postgres_config_data     = merge(local.postgres_config_upstream.data, local.size10 ? yamldecode(local.decoded_bodies["postgres_size10"]).data : {})
  postgres_upstream        = one([for d in local.documents_by_file.postgres : d if d.kind == "StatefulSet" && d.metadata.name == local.postgres_name])
  postgres_container       = one([for c in local.postgres_upstream.spec.template.spec.containers : c if c.name == local.postgres_name])
  postgres_size10_resources = {
    requests = merge(local.postgres_container.resources.requests, { cpu = "80", memory = "140Gi" })
    limits   = merge({ for k, v in local.postgres_container.resources.limits : k => v if k != "cpu" }, { memory = "140Gi" })
  }
  postgres_hook = local.postgres_container.lifecycle.postStart.exec.command
  postgres_doc = merge(local.postgres_upstream, {
    spec = merge(local.postgres_upstream.spec, {
      volumeClaimTemplates = [for claim in local.postgres_upstream.spec.volumeClaimTemplates : merge(claim, { for k, v in {
        metadata = merge(try(claim.metadata, {}), { labels = merge(try(claim.metadata.labels, {}), { (local.minor_label) = join(".", slice(split(".", trimprefix(local.version, "v")), 0, 2)) }) })
        spec = merge(try(claim.spec, {}),
          { for k2, v2 in { storageClassName = local.db.storage_class } : k2 => v2 if local.db.storage_class != "" },
          { resources = merge(try(claim.spec.resources, {}), { requests = merge(try(claim.spec.resources.requests, {}), { storage = local.db.volume_size }) }) },
        )
      } : k => v if claim.metadata.name == "data" })]
      template = merge(local.postgres_upstream.spec.template, {
        spec = merge(local.postgres_upstream.spec.template.spec, {
          containers = [for c in local.postgres_upstream.spec.template.spec.containers : merge(c, { for k, v in merge(
            { for k2, v2 in { resources = local.postgres_size10_resources } : k2 => v2 if local.size10 },
            { for k2, v2 in { resources = local.res.postgres } : k2 => v2 if length(local.res.postgres) > 0 },
            {
              lifecycle = merge(try(c.lifecycle, {}), { postStart = merge(try(c.lifecycle.postStart, {}), { exec = merge(try(c.lifecycle.postStart.exec, {}), {
                command = [local.postgres_hook[0], local.postgres_hook[1], "${replace(local.postgres_hook[2], "/\\n+$/", "")}\n${local.setup_command}\n"]
              }) }) })
              volumeMounts = concat(try(c.volumeMounts, []), [{ name = local.postgres_setup_config_map, mountPath = local.postgres_setup_mount_path, readOnly = true }])
            },
          ) : k => v if c.name == local.postgres_name })]
          volumes = concat(local.postgres_upstream.spec.template.spec.volumes, [{ name = local.postgres_setup_config_map, configMap = { name = local.postgres_setup_config_map } }])
        })
      })
    })
  })

  # ---- the API server (patchAPIServer) -----------------------------------------
  api_upstream    = one([for d in local.documents_by_file.api_server : d if d.kind == "Deployment" && d.metadata.name == local.api_server_name])
  api_container   = one([for c in local.api_upstream.spec.template.spec.containers : c if c.name == local.api_server_name])
  api_flag_values = { for f in local.api_flags : f.name => f.value }
  api_args = concat(
    [for a in local.api_container.args : substr(a, 0, 2) == "--" && contains(keys(local.api_flag_values), trimprefix(split("=", a)[0], "--")) ? "--${trimprefix(split("=", a)[0], "--")}=${local.api_flag_values[trimprefix(split("=", a)[0], "--")]}" : a],
    [for f in local.api_flags : "--${f.name}=${f.value}" if !contains([for a in local.api_container.args : trimprefix(split("=", a)[0], "--") if substr(a, 0, 2) == "--"], f.name)],
  )
  api_password_env = [for e in [
    { name = "ATE_API_POSTGRES_READ_WRITE_CONNECTION_STRING", valueFrom = { secretKeyRef = { name = try(local.db_read_write_secret.name, ""), key = try(local.db_read_write_secret.key, "") } } },
    { name = "ATE_API_POSTGRES_OWNER_CONNECTION_STRING", valueFrom = { secretKeyRef = { name = local.db_owner_secret.name, key = local.db_owner_secret.key } } },
  ] : e if local.db_external_passwords]
  storage_env_names = [for e in local.storage_env : e.name]
  api_doc = merge(local.api_upstream, {
    spec = merge(local.api_upstream.spec, {
      replicas = local.api.replicas
      template = merge(local.api_upstream.spec.template, {
        spec = merge(local.api_upstream.spec.template.spec, {
          containers = [for c in local.api_upstream.spec.template.spec.containers : merge(c, { for k, v in merge(
            { args = local.api_args },
            { for k2, v2 in { resources = local.res.api_server } : k2 => v2 if length(local.res.api_server) > 0 },
            { for k2, v2 in { env = concat(try(c.env, []), local.api_password_env) } : k2 => v2 if local.db_external_passwords },
          ) : k => v if c.name == local.api_server_name })]
          initContainers = [for c in local.api_upstream.spec.template.spec.initContainers : merge(c, { for k, v in {
            env = concat([for e in try(c.env, []) : e if !contains(local.storage_env_names, e.name)], local.storage_env)
          } : k => v if c.name == "objectstore-plugin" })]
        })
      })
    })
  })

  # ---- the identity server (patchIdentityServer) -------------------------------
  identity_upstream    = one([for d in local.documents_by_file.identity_server : d if d.kind == "Deployment" && d.metadata.name == local.identity_server_name])
  identity_container   = one([for c in local.identity_upstream.spec.template.spec.containers : c if c.name == local.identity_server_name])
  identity_flag_values = { for f in local.identity_flags : f.name => f.value }
  identity_args = concat(
    [for a in local.identity_container.args : substr(a, 0, 2) == "--" && contains(keys(local.identity_flag_values), trimprefix(split("=", a)[0], "--")) ? "--${trimprefix(split("=", a)[0], "--")}=${local.identity_flag_values[trimprefix(split("=", a)[0], "--")]}" : a],
    [for f in local.identity_flags : "--${f.name}=${f.value}" if !contains([for a in local.identity_container.args : trimprefix(split("=", a)[0], "--") if substr(a, 0, 2) == "--"], f.name)],
  )
  identity_doc = merge(local.identity_upstream, {
    spec = merge(local.identity_upstream.spec, {
      replicas = local.identity.replicas
      template = merge(local.identity_upstream.spec.template, {
        spec = merge(local.identity_upstream.spec.template.spec, {
          containers = [for c in local.identity_upstream.spec.template.spec.containers : merge(c, { for k, v in merge(
            { args = local.identity_args },
            { for k2, v2 in { resources = local.res.identity } : k2 => v2 if length(local.res.identity) > 0 },
          ) : k => v if c.name == local.identity_server_name })]
        })
      })
    })
  })

  # ---- the controller (patchController) ----------------------------------------
  controller_upstream    = one([for d in local.documents_by_file.controller : d if d.kind == "Deployment" && d.metadata.name == local.controller_name])
  controller_container   = one([for c in local.controller_upstream.spec.template.spec.containers : c if c.name == local.controller_name])
  controller_flag_values = { for f in local.controller_flags : f.name => f.value }
  controller_args = concat(
    [for a in local.controller_container.args : substr(a, 0, 2) == "--" && contains(keys(local.controller_flag_values), trimprefix(split("=", a)[0], "--")) ? "--${trimprefix(split("=", a)[0], "--")}=${local.controller_flag_values[trimprefix(split("=", a)[0], "--")]}" : a],
    [for f in local.controller_flags : "--${f.name}=${f.value}" if !contains([for a in local.controller_container.args : trimprefix(split("=", a)[0], "--") if substr(a, 0, 2) == "--"], f.name)],
  )
  controller_doc = merge(local.controller_upstream, {
    spec = merge(local.controller_upstream.spec, {
      template = merge(local.controller_upstream.spec.template, {
        spec = merge(local.controller_upstream.spec.template.spec, {
          containers = [for c in local.controller_upstream.spec.template.spec.containers : merge(c, { for k, v in merge(
            { args = local.controller_args },
            { for k2, v2 in { resources = local.res.controller } : k2 => v2 if length(local.res.controller) > 0 },
          ) : k => v if c.name == local.controller_name })]
        })
      })
    })
  })

  # ---- the node agent (patchNodeAgent) -----------------------------------------
  node_upstream    = one([for d in local.documents_by_file.node_agent : d if d.kind == "DaemonSet" && d.metadata.name == local.node_agent_name])
  node_container   = one([for c in local.node_upstream.spec.template.spec.containers : c if c.name == local.node_agent_name])
  node_flag_values = { for f in local.node_agent_flags : f.name => f.value }
  node_args_flagged = concat(
    [for a in local.node_container.args : substr(a, 0, 2) == "--" && contains(keys(local.node_flag_values), trimprefix(split("=", a)[0], "--")) ? "--${trimprefix(split("=", a)[0], "--")}=${local.node_flag_values[trimprefix(split("=", a)[0], "--")]}" : a],
    [for f in local.node_agent_flags : "--${f.name}=${f.value}" if !contains([for a in local.node_container.args : trimprefix(split("=", a)[0], "--") if substr(a, 0, 2) == "--"], f.name)],
  )
  # Off GKE: upstream's kind overlay drops the node image credential
  # provider's two flags and its host mounts.
  node_args           = [for a in local.node_args_flagged : a if local.gke || length(regexall("^--image-credential-provider-", a)) == 0]
  credential_mounts   = ["/run/image-credential-provider/bin", "/run/image-credential-provider/config.yaml"]
  credential_volumes  = ["image-credential-provider-bin", "image-credential-provider-config"]
  node_sidecar        = one([for c in local.node_upstream.spec.template.spec.initContainers : c if c.name == "objectstore-plugin"])
  staging_flag_values = { for f in local.staging_flags : f.name => f.value }
  node_sidecar_args = concat(
    [for a in local.node_sidecar.args : substr(a, 0, 2) == "--" && contains(keys(local.staging_flag_values), trimprefix(split("=", a)[0], "--")) ? "--${trimprefix(split("=", a)[0], "--")}=${local.staging_flag_values[trimprefix(split("=", a)[0], "--")]}" : a],
    [for f in local.staging_flags : "--${f.name}=${f.value}" if !contains([for a in local.node_sidecar.args : trimprefix(split("=", a)[0], "--") if substr(a, 0, 2) == "--"], f.name)],
  )
  node_spec = local.node_upstream.spec.template.spec
  node_doc = merge(local.node_upstream, {
    spec = merge(local.node_upstream.spec, {
      template = merge(local.node_upstream.spec.template, {
        spec = merge(local.node_spec,
          {
            containers = [for c in local.node_spec.containers : merge(c, { for k, v in merge(
              {
                args         = local.node_args
                volumeMounts = [for m in try(c.volumeMounts, []) : m if local.gke || !contains(local.credential_mounts, m.mountPath)]
              },
              { for k2, v2 in { resources = local.res.node_agent } : k2 => v2 if length(local.res.node_agent) > 0 },
            ) : k => v if c.name == local.node_agent_name })]
            initContainers = [for c in local.node_spec.initContainers : merge(c, { for k, v in {
              args = local.node_sidecar_args
              env  = concat([for e in try(c.env, []) : e if !contains(local.storage_env_names, e.name)], local.storage_env)
            } : k => v if c.name == "objectstore-plugin" })]
            volumes = [for vol in local.node_spec.volumes : merge(vol, { for k, v in {
              emptyDir = merge(try(vol.emptyDir, {}), { sizeLimit = "${local.node_agent.staging_gib + 1}Gi" })
            } : k => v if vol.name == "asset-staging" && local.node_agent.staging_gib != 8 }) if local.gke || !contains(local.credential_volumes, vol.name)]
          },
          { for k, v in { nodeSelector = merge(try(local.node_spec.nodeSelector, {}), local.node_agent.node_selector) } : k => v if length(local.node_agent.node_selector) > 0 },
          { for k, v in { tolerations = concat(try(local.node_spec.tolerations, []), local.node_agent.tolerations) } : k => v if length(local.node_agent.tolerations) > 0 },
        )
      })
    })
  })

  # ---- the router (patchRouter; agentgateway below) ----------------------------
  router_upstream    = one([for d in local.documents_by_file.router : d if d.kind == "Deployment" && d.metadata.name == local.router_name])
  router_container   = one([for c in local.router_upstream.spec.template.spec.containers : c if c.name == local.router_name])
  router_flag_values = { for f in local.router_flags : f.name => f.value }
  router_args = concat(
    [for a in local.router_container.args : substr(a, 0, 2) == "--" && contains(keys(local.router_flag_values), trimprefix(split("=", a)[0], "--")) ? "--${trimprefix(split("=", a)[0], "--")}=${local.router_flag_values[trimprefix(split("=", a)[0], "--")]}" : a],
    [for f in local.router_flags : "--${f.name}=${f.value}" if !contains([for a in local.router_container.args : trimprefix(split("=", a)[0], "--") if substr(a, 0, 2) == "--"], f.name)],
  )
  router_envoy_doc = merge(local.router_upstream, {
    spec = merge(local.router_upstream.spec, {
      replicas = local.router.replicas
      template = merge(local.router_upstream.spec.template, {
        spec = merge(local.router_upstream.spec.template.spec, {
          containers = [for c in local.router_upstream.spec.template.spec.containers : merge(c, { for k, v in merge(
            { args = local.router_args },
            { for k2, v2 in { resources = local.res.router } : k2 => v2 if length(local.res.router) > 0 },
          ) : k => v if c.name == local.router_name })]
        })
      })
    })
  })

  # ---- the egress gateway (patchEgress; agentgateway below) --------------------
  egress_upstream    = one([for d in local.documents_by_file.egress : d if d.kind == "Deployment" && d.metadata.name == local.egress_name])
  egress_extproc     = one([for c in local.egress_upstream.spec.template.spec.containers : c if c.name == "ext-proc"])
  egress_minter      = one([for c in local.egress_upstream.spec.template.spec.initContainers : c if c.name == "sdsmint"])
  egress_flag_values = { for f in local.egress_flags : f.name => f.value }
  egress_args = concat(
    [for a in local.egress_extproc.args : substr(a, 0, 2) == "--" && contains(keys(local.egress_flag_values), trimprefix(split("=", a)[0], "--")) ? "--${trimprefix(split("=", a)[0], "--")}=${local.egress_flag_values[trimprefix(split("=", a)[0], "--")]}" : a],
    [for f in local.egress_flags : "--${f.name}=${f.value}" if !contains([for a in local.egress_extproc.args : trimprefix(split("=", a)[0], "--") if substr(a, 0, 2) == "--"], f.name)],
  )
  minter_flag_values = { for f in local.minter_flags : f.name => f.value }
  minter_args = concat(
    [for a in local.egress_minter.args : substr(a, 0, 2) == "--" && contains(keys(local.minter_flag_values), trimprefix(split("=", a)[0], "--")) ? "--${trimprefix(split("=", a)[0], "--")}=${local.minter_flag_values[trimprefix(split("=", a)[0], "--")]}" : a],
    [for f in local.minter_flags : "--${f.name}=${f.value}" if !contains([for a in local.egress_minter.args : trimprefix(split("=", a)[0], "--") if substr(a, 0, 2) == "--"], f.name)],
  )
  egress_envoy_doc = merge(local.egress_upstream, {
    spec = merge(local.egress_upstream.spec, {
      replicas = local.egress.replicas
      template = merge(local.egress_upstream.spec.template, {
        spec = merge(local.egress_upstream.spec.template.spec, {
          containers = [for c in local.egress_upstream.spec.template.spec.containers : merge(c,
            { for k, v in merge(
              { args = local.egress_args },
              { for k2, v2 in { resources = local.res.egress } : k2 => v2 if length(local.res.egress) > 0 },
            ) : k => v if c.name == "ext-proc" },
            { for k, v in { args = concat(try(c.args, []), ["--concurrency", tostring(local.egress.envoy_concurrency)]) } : k => v if c.name == "envoy" && local.egress.envoy_concurrency > 0 },
          )]
          initContainers = [for c in local.egress_upstream.spec.template.spec.initContainers : merge(c, { for k, v in { args = local.minter_args } : k => v if c.name == "sdsmint" })]
        })
      })
    })
  })

  # The additional ext_proc splice (spliceExtProc): both filter markers and
  # the cluster marker in Envoy's configuration, each replaced by upstream's
  # block indented like the marker.
  egress_config_upstream = one([for d in local.documents_by_file.egress : d if d.kind == "ConfigMap" && d.metadata.name == local.egress_name])
  extproc_server_name    = "${local.egress.extproc_service}.${local.egress.extproc_namespace}.svc"
  extproc_filter_lines   = split("\n", local.extproc_filter)
  extproc_cluster_lines  = split("\n", format(local.extproc_cluster, local.extproc_server_name, local.extproc_server_name, "${local.extproc_server_name}.cluster.local", local.egress.extproc_port))
  egress_envoy_yaml = local.egress.extproc_service == "" ? local.egress_config_upstream.data["envoy.yaml"] : replace(
    replace(local.egress_config_upstream.data["envoy.yaml"], "/(?m)^([ \\t]*)#ATE_MITM_EXTPROC_FILTER.*$/", join("\n", [for l in local.extproc_filter_lines : l == "" ? "" : "$${1}${l}"])),
    "/(?m)^([ \\t]*)#ATE_MITM_EXTPROC_CLUSTER.*$/", join("\n", [for l in local.extproc_cluster_lines : l == "" ? "" : "$${1}${l}"]),
  )
  egress_config_doc = merge(local.egress_config_upstream, {
    data = merge(local.egress_config_upstream.data, { "envoy.yaml" = local.egress_envoy_yaml })
  })

  # ---- the credential provider (patchCredentialProvider) -----------------------
  provider_upstream    = one([for d in local.documents_by_file.credential_provider : d if d.kind == "Deployment" && d.metadata.name == local.credential_provider_name])
  provider_container   = one([for c in local.provider_upstream.spec.template.spec.containers : c if c.name == local.credential_provider_name])
  provider_flag_values = { for f in local.provider_flags : f.name => f.value }
  provider_args = concat(
    [for a in local.provider_container.args : substr(a, 0, 2) == "--" && contains(keys(local.provider_flag_values), trimprefix(split("=", a)[0], "--")) ? "--${trimprefix(split("=", a)[0], "--")}=${local.provider_flag_values[trimprefix(split("=", a)[0], "--")]}" : a],
    [for f in local.provider_flags : "--${f.name}=${f.value}" if !contains([for a in local.provider_container.args : trimprefix(split("=", a)[0], "--") if substr(a, 0, 2) == "--"], f.name)],
  )
  provider_doc = merge(local.provider_upstream, {
    spec = merge(local.provider_upstream.spec, {
      template = merge(local.provider_upstream.spec.template, {
        spec = merge(local.provider_upstream.spec.template.spec, {
          containers = [for c in local.provider_upstream.spec.template.spec.containers : merge(c, { for k, v in merge(
            { args = local.provider_args },
            { for k2, v2 in { resources = local.res.provider } : k2 => v2 if length(local.res.provider) > 0 },
          ) : k => v if c.name == local.credential_provider_name })]
        })
      })
    })
  })

  # ---- agentgateway (applyAgentgateway): upstream's two components ------------
  # components/agentgateway patches, in upstream's order: [0] delete the
  # router's Envoy ConfigMap, [1] router: replace container 0, remove
  # container 1, add a volume, [2] router: point envoy-config at the
  # agentgateway ConfigMap and the metrics port annotation (kustomize's merge
  # of the volumes collapses the added one into upstream's volume of the same
  # name, which the router already declares), [3] egress:
  # replace container 0, rename volume 0's ConfigMap, remove container 1, add
  # a volume, [4] the router Service's stats port and its plaintext http and
  # connect ports, [5] the router NetworkPolicy's first ingress rule admits
  # those plaintext ports. components/agentgateway-egress: [0] drop the minter
  # and the Envoy pod identity, [1] the egress configuration, [2] mount the
  # egress interception CA.
  agw             = local.documents_by_file.agentgateway[0].patches
  agw_egress      = local.documents_by_file.agentgateway_egress[0].patches
  agw_router_ops  = yamldecode(local.agw[1].patch)
  agw_router_smp  = yamldecode(local.agw[2].patch).spec.template
  agw_egress_ops  = yamldecode(local.agw[3].patch)
  agw_service_ops = yamldecode(local.agw[4].patch)
  agw_policy_ops  = yamldecode(local.agw[5].patch)
  agw_egress_smp  = yamldecode(local.agw_egress[0].patch).spec.template.spec
  agw_egress_cfg  = yamldecode(local.agw_egress[1].patch).data
  agw_egress_ops2 = yamldecode(local.agw_egress[2].patch)

  router_agw_containers = concat([local.agw_router_ops[0].value], slice(local.router_upstream.spec.template.spec.containers, 2, length(local.router_upstream.spec.template.spec.containers)))
  router_agw_doc = merge(local.router_upstream, {
    spec = merge(local.router_upstream.spec, {
      replicas = local.router.replicas
      template = merge(local.router_upstream.spec.template, {
        metadata = merge(local.router_upstream.spec.template.metadata, {
          annotations = merge(local.router_upstream.spec.template.metadata.annotations, local.agw_router_smp.metadata.annotations)
        })
        spec = merge(local.router_upstream.spec.template.spec, {
          containers = [for i, c in local.router_agw_containers : merge(c, { for k, v in { resources = local.res.router } : k => v if i == 0 && length(local.res.router) > 0 })]
          volumes = concat(
            [for vol in local.router_upstream.spec.template.spec.volumes : merge(vol, { for k, v in {
              configMap = merge(try(vol.configMap, {}), try(one([for p in local.agw_router_smp.spec.volumes : p.configMap if p.name == vol.name]), {}))
            } : k => v if contains([for p in local.agw_router_smp.spec.volumes : p.name], vol.name) }, { for k, v in local.agw_router_ops[2].value : k => v if vol.name == local.agw_router_ops[2].value.name })],
            [for added in [local.agw_router_ops[2].value] : added if !contains([for vol in local.router_upstream.spec.template.spec.volumes : vol.name], added.name)],
          )
        })
      })
    })
  })
  router_service_upstream = one([for d in local.documents_by_file.router : d if d.kind == "Service" && d.metadata.name == local.router_name])
  router_service_agw_doc = merge(local.router_service_upstream, {
    spec = merge(local.router_service_upstream.spec, { ports = concat(local.router_service_upstream.spec.ports, [for op in local.agw_service_ops : op.value]) })
  })
  router_policy_upstream = one([for d in local.documents_by_file.network_policies : d if d.kind == "NetworkPolicy" && d.metadata.name == local.router_name])
  router_policy_agw_doc = merge(local.router_policy_upstream, {
    spec = merge(local.router_policy_upstream.spec, {
      ingress = [for i, rule in local.router_policy_upstream.spec.ingress : merge(rule, { for k, v in {
        ports = concat(try(rule.ports, []), [for op in local.agw_policy_ops : op.value])
      } : k => v if i == 0 })]
    })
  })

  egress_agw_spec = local.egress_upstream.spec.template.spec
  egress_agw_containers = concat(
    [merge(local.agw_egress_ops[0].value, {
      volumeMounts = concat(local.agw_egress_ops[0].value.volumeMounts, [local.agw_egress_ops2[0].value])
    }, { for k, v in { resources = local.res.egress } : k => v if length(local.res.egress) > 0 })],
    slice(local.egress_agw_spec.containers, 2, length(local.egress_agw_spec.containers)),
  )
  egress_agw_volumes = concat(
    [for i, vol in local.egress_agw_spec.volumes : merge(vol, { for k, v in { configMap = merge(try(vol.configMap, {}), { name = local.agw_egress_ops[1].value }) } : k => v if i == 0 })
    if !contains([for p in local.agw_egress_smp.volumes : p.name], vol.name)],
    [local.agw_egress_ops[3].value, local.agw_egress_ops2[1].value],
  )
  egress_agw_doc = merge(local.egress_upstream, {
    spec = merge(local.egress_upstream.spec, {
      replicas = local.egress.replicas
      template = merge(local.egress_upstream.spec.template, {
        spec = merge(
          { for k, v in local.egress_agw_spec : k => v if !contains(["initContainers", "securityContext"], k) },
          { securityContext = { for k, v in local.egress_agw_spec.securityContext : k => v if !contains(keys(local.agw_egress_smp.securityContext), k) } },
          { containers = local.egress_agw_containers, volumes = local.egress_agw_volumes },
        )
      })
    })
  })

  agw_config_documents   = { for d in local.documents_by_file.agentgateway_config : d.metadata.name => d }
  agw_egress_config_name = "atenet-egress-agentgateway-substrate-config"
  agw_provider_lines     = local.egress.provider == "disabled" ? [] : split("\n", format(local.agentgateway_providers, jsonencode(local.egress.provider_name), jsonencode(local.egress.provider_address), jsonencode(local.provider_server_name)))
  agw_egress_config_yaml = replace(local.agw_egress_cfg["config.yaml"], "/(?m)^([ \\t]*)#ATE_AGENTGATEWAY_CREDENTIAL_PROVIDERS.*\\n/",
  join("", [for l in local.agw_provider_lines : "$${1}${l}\n"]))
  agw_config_docs = {
    for name, d in local.agw_config_documents : name => merge(d, { for k, v in {
      data = merge(d.data, { "config.yaml" = local.agw_egress_config_yaml })
    } : k => v if name == local.agw_egress_config_name })
  }

  # ---- the Cloud SQL Auth Proxy sidecar (applyCloudSQLSidecar) -----------------
  cloudsql_sidecar = yamldecode(local.decoded_bodies["cloudsql_patch"]).spec.template.spec.initContainers[0]
  api_doc_final = merge(local.api_doc, { for k, v in {
    spec = merge(local.api_doc.spec, {
      template = merge(local.api_doc.spec.template, {
        spec = merge(local.api_doc.spec.template.spec, {
          initContainers = concat(local.api_doc.spec.template.spec.initContainers, [local.cloudsql_sidecar])
        })
      })
    })
  } : k => v if local.db_kind == "cloudsql" })

  # ---- documents by identity, then by phase -------------------------------------
  id_of = {
    for role, docs in local.documents_by_file : role => [
      for d in docs : join("//", concat([try(d.apiVersion, ""), d.kind, try(d.metadata.name, "")], try(d.metadata.namespace, "") != "" ? [d.metadata.namespace] : []))
    ]
  }

  namespace_docs_raw = merge(
    {
      for d in local.documents_by_file.namespace : "${d.apiVersion}//${d.kind}//${d.metadata.name}" => merge(d, {
        metadata = merge(d.metadata, { labels = merge(try(d.metadata.labels, {}), local.identity_labels) })
      }) if local.create_namespace
    },
    {
      for d in local.documents_by_file.pod_certificates : "${d.apiVersion}//${d.kind}//${d.metadata.name}" => merge(d, {
        metadata = merge(d.metadata, { labels = merge(try(d.metadata.labels, {}), local.identity_labels) })
      }) if d.kind == "Namespace"
    },
  )

  config_map_docs = {
    for name, data in merge(
      {
        "ate-api-server-envvars" = local.api_env_config_map
        "ate-api-config"         = { "apiconfig.textproto" = local.api_config_text }
        "ate-api-authentication" = { "authentication.yaml" = local.authentication_yaml }
      },
      local.egress.provider == "kubernetes" ? { "k8s-credential-provider-namespace-policy" = { "namespace-policy.yaml" = local.policy_yaml } } : {},
    ) : "v1//ConfigMap//${name}//${local.namespace}" => { apiVersion = "v1", kind = "ConfigMap", metadata = { name = name, namespace = local.namespace }, data = data }
  }

  prerequisite_docs_raw = merge(
    { for i, d in local.documents_by_file.controller_role : local.id_of.controller_role[i] => d },
    local.config_map_docs,
    { for i, d in local.documents_by_file.pod_certificates : local.id_of.pod_certificates[i] => d if d.kind != "Namespace" && !(d.kind == "Deployment" && d.metadata.name == local.pod_certificates_name) },
    { "apps/v1//Deployment//${local.pod_certificates_name}//${local.pod_certificates_namespace}" = local.podcert_doc },
    { for i, d in local.documents_by_file.sandbox_validation : local.id_of.sandbox_validation[i] => d },
  )

  database_docs_raw = merge(
    { for i, d in local.documents_by_file.otel_config : local.id_of.otel_config[i] => merge(d, { data = local.otel_data }) },
    { for id, d in merge(
      { "v1//ConfigMap//${local.postgres_setup_config_map}//${local.namespace}" = { apiVersion = "v1", kind = "ConfigMap", metadata = { name = local.postgres_setup_config_map, namespace = local.namespace }, data = { "setup.sql" = local.release_bodies["setup_sql"] } } },
      { for i, d in local.documents_by_file.postgres : local.id_of.postgres[i] => d if !(d.kind == "StatefulSet" && d.metadata.name == local.postgres_name) && !(d.kind == "ConfigMap" && d.metadata.name == "postgres-config") },
      { "v1//ConfigMap//postgres-config//${local.namespace}" = merge(local.postgres_config_upstream, { data = local.postgres_config_data }) },
      { "apps/v1//StatefulSet//${local.postgres_name}//${local.namespace}" = local.postgres_doc },
    ) : id => d if local.db_kind == "bundled" },
  )

  # Upstream's system files, minus the namespace ate-controller.yaml repeats,
  # with the patched objects swapped in by identity.
  system_roles = concat(
    ["api_server", "identity_server", "controller", "node_agent", "router", "network_policies", "egress"],
    local.gke ? ["router_monitoring"] : [],
    local.egress.provider == "kubernetes" ? ["credential_provider"] : [],
  )
  system_upstream = merge([
    for role in local.system_roles : { for i, d in local.documents_by_file[role] : local.id_of[role][i] => d if d.kind != "Namespace" }
  ]...)
  sa_annotations = {
    (local.api_server_name) = local.api_server_gsa
    (local.node_agent_name) = local.node_agent_gsa
  }
  system_docs_raw = merge(
    {
      for id, d in local.system_upstream : id => d
      if !(local.router.agentgateway && d.kind == "ConfigMap" && d.metadata.name == "atenet-router-envoy-config")
    },
    {
      for id, d in local.system_upstream : id => merge(d, {
        metadata = merge(d.metadata, { annotations = merge(try(d.metadata.annotations, {}), { "iam.gke.io/gcp-service-account" = local.sa_annotations[d.metadata.name] }) })
      }) if d.kind == "ServiceAccount" && try(local.sa_annotations[d.metadata.name], "") != ""
    },
    { "apps/v1//Deployment//${local.api_server_name}//${local.namespace}" = local.api_doc_final },
    { "apps/v1//Deployment//${local.identity_server_name}//${local.namespace}" = local.identity_doc },
    { "apps/v1//Deployment//${local.controller_name}//${local.namespace}" = local.controller_doc },
    { "apps/v1//DaemonSet//${local.node_agent_name}//${local.namespace}" = local.node_doc },
    { for id, d in {
      "apps/v1//Deployment//${local.router_name}//${local.namespace}" = local.router_envoy_doc
      "apps/v1//Deployment//${local.egress_name}//${local.namespace}" = local.egress_envoy_doc
    } : id => d if !local.router.agentgateway },
    { for id, d in {
      "apps/v1//Deployment//${local.router_name}//${local.namespace}"                 = local.router_agw_doc
      "v1//Service//${local.router_name}//${local.namespace}"                         = local.router_service_agw_doc
      "apps/v1//Deployment//${local.egress_name}//${local.namespace}"                 = local.egress_agw_doc
      "networking.k8s.io/v1//NetworkPolicy//${local.router_name}//${local.namespace}" = local.router_policy_agw_doc
    } : id => d if local.router.agentgateway },
    { "v1//ConfigMap//${local.egress_name}//${local.namespace}" = local.egress_config_doc },
    { for name, d in local.agw_config_docs : "v1//ConfigMap//${name}//${local.namespace}" => d if local.router.agentgateway },
    { for id, d in { "apps/v1//Deployment//${local.credential_provider_name}//${local.namespace}" = local.provider_doc } : id => d if local.egress.provider == "kubernetes" },
  )

  runtime_docs_raw = { for i, d in local.documents_by_file.sandbox_gvisor : local.id_of.sandbox_gvisor[i] => d }

  # ---- passes over every document (compose.go) ----------------------------------
  # cordon (upstream's component): the control-plane Deployments and the
  # database get the node selector, toleration and spread upstream's patches
  # name.
  cordon_patches     = local.documents_by_file.cordon[0].patches
  cordon_deployments = local.cordon_patches[0].target.name
  cordon_deploy      = yamldecode(local.cordon_patches[0].patch).spec.template
  cordon_database    = yamldecode(local.cordon_patches[1].patch).spec.template.spec

  workload_kinds = ["Deployment", "DaemonSet", "StatefulSet"]

  # The mirror registry over every image a document runs (mirrorImages).
  all_images = distinct(compact(flatten(concat(
    [for d in values(merge(local.prerequisite_docs_raw, local.database_docs_raw, local.system_docs_raw)) : [
      for c in concat(try(d.spec.template.spec.initContainers, []), try(d.spec.template.spec.containers, [])) : try(c.image, "")
    ] if contains(local.workload_kinds, d.kind)],
    values(local.images),
  ))))

  # The preview gates (applyPreview): every container that runs one of
  # upstream's components, the release's images bar the Envoy dataplane,
  # takes them; none leaves the binaries' default.
  component_images = [for name, image in local.images : image if name != "envoy-dataplane"]
  preview_args     = length(local.preview) > 0 ? ["--preview=${join(",", local.preview)}"] : []
  mirrored = {
    for image in local.all_images : image => local.registry == "" ? image : "${local.registry}/${(
      length(split("/", image)) == 1 ? "library/${image}" :
      length(regexall("[.:]", split("/", image)[0])) > 0 || split("/", image)[0] == "localhost" ? join("/", slice(split("/", image), 1, length(split("/", image)))) :
      image
    )}"
  }

  prerequisite_docs = local.processed.prerequisites
  database_docs     = local.processed.database
  system_docs       = local.processed.system
  runtime_docs      = local.runtime_docs_raw
  namespace_docs    = local.namespace_docs_raw

  processed = {
    for phase, docs in {
      prerequisites = local.prerequisite_docs_raw
      database      = local.database_docs_raw
      system        = local.system_docs_raw
      } : phase => merge(
      { for id, d in docs : id => d if !contains(local.workload_kinds, d.kind) },
      {
        for id, d in docs : id => merge(d, {
          metadata = merge(d.metadata,
            { labels = merge(try(d.metadata.labels, {}), { (local.owner_label) = local.kind_name }) },
            { for k, v in { annotations = merge(try(d.metadata.annotations, {}), { (local.version_annotation) = local.version }) } : k => v if d.metadata.name == local.api_server_name },
          )
          spec = merge(d.spec, {
            template = merge(d.spec.template, {
              metadata = merge(d.spec.template.metadata,
                local.cordon && d.kind == "Deployment" && length(regexall("^(?:${local.cordon_deployments})$", d.metadata.name)) > 0 ? {
                  labels = merge(try(d.spec.template.metadata.labels, {}), local.cordon_deploy.metadata.labels)
                } : {},
                length(merge(
                  { for k, v in { (local.config_hash_annotation) = try(local.config_hashes[d.metadata.name], "") } : k => v if contains(keys(local.config_hashes), d.metadata.name) },
                  { for k, v in { (local.env_hash_annotation) = local.env_hash } : k => v if d.metadata.name == local.api_server_name },
                  )) > 0 ? {
                  annotations = merge(try(d.spec.template.metadata.annotations, {}),
                    { for k, v in { (local.config_hash_annotation) = try(local.config_hashes[d.metadata.name], "") } : k => v if contains(keys(local.config_hashes), d.metadata.name) },
                    { for k, v in { (local.env_hash_annotation) = local.env_hash } : k => v if d.metadata.name == local.api_server_name },
                  )
                } : {},
              )
              spec = merge(d.spec.template.spec,
                {
                  containers = [for c in d.spec.template.spec.containers : merge(c,
                    { for k, v in { args = concat(try(c.args, []), local.preview_args) } : k => v if length(local.preview_args) > 0 && contains(local.component_images, c.image) },
                    { image = local.mirrored[c.image] },
                  )]
                },
                { for k, v in { initContainers = [for c in try(d.spec.template.spec.initContainers, []) : merge(c,
                  { for k2, v2 in { args = concat(try(c.args, []), local.preview_args) } : k2 => v2 if length(local.preview_args) > 0 && contains(local.component_images, c.image) },
                  { image = local.mirrored[c.image] },
                )] } : k => v if length(try(d.spec.template.spec.initContainers, [])) > 0 },
                { for k, v in { imagePullSecrets = local.pull_secrets } : k => v if length(local.pull_secrets) > 0 },
                { for k, v in {
                  nodeSelector              = merge(try(d.spec.template.spec.nodeSelector, {}), local.cordon_deploy.spec.nodeSelector)
                  tolerations               = local.cordon_deploy.spec.tolerations
                  topologySpreadConstraints = local.cordon_deploy.spec.topologySpreadConstraints
                } : k => v if local.cordon && d.kind == "Deployment" && length(regexall("^(?:${local.cordon_deployments})$", d.metadata.name)) > 0 },
                { for k, v in {
                  nodeSelector = merge(try(d.spec.template.spec.nodeSelector, {}), local.cordon_database.nodeSelector)
                  tolerations  = local.cordon_database.tolerations
                } : k => v if local.cordon && d.kind == "StatefulSet" && d.metadata.name == local.postgres_name },
              )
            })
          })
        }) if contains(local.workload_kinds, d.kind)
      },
    )
  }
}
