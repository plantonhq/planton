variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareCacheSettings specification"
  type = object({
    zone_id               = string
    smart_tiered_cache    = optional(bool)
    tiered_caching        = optional(bool)
    cache_reserve         = optional(bool)
    regional_tiered_cache = optional(bool)
    argo_smart_routing    = optional(bool)
    cache_variants = optional(object({
      avif = optional(list(string), [])
      bmp  = optional(list(string), [])
      gif  = optional(list(string), [])
      jp2  = optional(list(string), [])
      jpeg = optional(list(string), [])
      jpg  = optional(list(string), [])
      jpg2 = optional(list(string), [])
      png  = optional(list(string), [])
      tif  = optional(list(string), [])
      tiff = optional(list(string), [])
      webp = optional(list(string), [])
    }))
  })
}