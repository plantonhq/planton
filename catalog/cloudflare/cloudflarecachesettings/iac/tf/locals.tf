locals {
  zone_id = var.spec.zone_id

  # Cache variants: only extensions with at least one MIME type are sent. The
  # Terraform provider keeps the API's singular extension names (the Pulumi SDK
  # pluralizes them).
  cache_variants_value = var.spec.cache_variants == null ? null : {
    for extension, mime_types in {
      avif = var.spec.cache_variants.avif
      bmp  = var.spec.cache_variants.bmp
      gif  = var.spec.cache_variants.gif
      jp2  = var.spec.cache_variants.jp2
      jpeg = var.spec.cache_variants.jpeg
      jpg  = var.spec.cache_variants.jpg
      jpg2 = var.spec.cache_variants.jpg2
      png  = var.spec.cache_variants.png
      tif  = var.spec.cache_variants.tif
      tiff = var.spec.cache_variants.tiff
      webp = var.spec.cache_variants.webp
    } : extension => mime_types if length(mime_types) > 0
  }
}
