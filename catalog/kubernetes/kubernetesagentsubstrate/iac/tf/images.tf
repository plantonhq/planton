# The release's images (release.go): every image upstream's manifests name by
# placeholder, plus the worker images, each pinned at the release's tag by the
# digest upstream's registry serves for it, exactly as upstream's installer
# pins them (cmd/ate-setup/internal/images/prebuilt.go, Pin).
#
# The digest is read the way any registry client reads it: an anonymous pull
# token for the repository from ghcr.io, then a HEAD on the tag's manifest
# accepting a multi-platform index, whose Docker-Content-Digest is the digest
# a container runtime resolves. A lookup that fails lands in images_missing,
# and the preflight refuses the apply by name (preflight.tf).

locals {
  # Every image: value in the release's files, quotes stripped, mapped to the
  # image it names ("" for a literal image upstream pins itself, "?" for a
  # placeholder this module does not know how to resolve).
  image_values = distinct(flatten([
    for role, body in local.release_bodies : [
      for m in regexall("(?m)^[ \\t]*(?:-[ \\t]+)?image:[ \\t]*(\"[^\"\\n]*\"|'[^'\\n]*'|[^\\s#]+)", body) : trim(m[0], "\"'")
    ] if role != "setup_sql"
  ]))
  image_of_value = {
    for value in local.image_values : value => (
      value == local.envoy_placeholder ? local.envoy_image :
      startswith(value, local.ko_image_prefix) ? element(split("/", value), length(split("/", value)) - 1) :
      startswith(value, "ko://") || startswith(value, "$${") ? "?" : ""
    )
  }
  unknown_placeholders = [for value, image in local.image_of_value : value if image == "?"]

  # The release's images, by name, sorted: the placeholders' and the workers'.
  image_names = sort(distinct(concat(
    local.worker_images,
    [for value, image in local.image_of_value : image if image != "" && image != "?"],
  )))
}

data "http" "registry_token" {
  for_each = toset(local.image_names)

  url                = "https://${split("/", local.upstream_registry)[0]}/token?scope=repository:${join("/", slice(split("/", local.upstream_registry), 1, length(split("/", local.upstream_registry))))}/${each.key}:pull"
  request_timeout_ms = 60000

  retry {
    attempts     = 3
    min_delay_ms = 1000
    max_delay_ms = 8000
  }
}

data "http" "image_digest" {
  for_each = toset(local.image_names)

  url    = "https://${split("/", local.upstream_registry)[0]}/v2/${join("/", slice(split("/", local.upstream_registry), 1, length(split("/", local.upstream_registry))))}/${each.key}/manifests/${local.version}"
  method = "HEAD"
  request_headers = {
    Authorization = "Bearer ${try(jsondecode(data.http.registry_token[each.key].response_body).token, "")}"
    Accept        = "application/vnd.oci.image.index.v1+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json"
  }
  request_timeout_ms = 60000

  retry {
    attempts     = 3
    min_delay_ms = 1000
    max_delay_ms = 8000
  }
}

locals {
  image_digests = {
    for name in local.image_names : name => try(
      data.http.image_digest[name].response_headers["Docker-Content-Digest"],
      data.http.image_digest[name].response_headers["docker-content-digest"],
      "",
    )
  }
  # Images the registry would not resolve, with the answer it gave. ghcr.io
  # answers an image that does not exist as unauthorized to an anonymous
  # reader, so 401 and 403 mean "not published" too.
  images_missing = [
    for name in local.image_names : "${local.upstream_registry}/${name}:${local.version} (HTTP ${data.http.image_digest[name].status_code})"
    if data.http.image_digest[name].status_code != 200 || local.image_digests[name] == ""
  ]

  # name -> the image pinned at the release (resolveImages), and each
  # placeholder value -> the same, for the containers that name it.
  images = { for name in local.image_names : name => "${local.upstream_registry}/${name}:${local.version}@${local.image_digests[name]}" }
  pinned_image = {
    for value, image in local.image_of_value : value => local.images[image] if image != "" && image != "?"
  }
}
