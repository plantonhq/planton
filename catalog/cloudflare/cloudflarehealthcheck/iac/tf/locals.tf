locals {
  # The converter emits header values as list WRAPPERS ({values = [...]}) --
  # the provider wants a plain map of string lists, so unwrap here. Empty
  # strings and empty containers are dropped rather than sent.
  http_config = var.spec.http_config != null ? {
    method           = var.spec.http_config.method
    path             = var.spec.http_config.path
    port             = var.spec.http_config.port
    expected_codes   = length(var.spec.http_config.expected_codes) > 0 ? var.spec.http_config.expected_codes : null
    expected_body    = var.spec.http_config.expected_body != "" ? var.spec.http_config.expected_body : null
    follow_redirects = var.spec.http_config.follow_redirects
    allow_insecure   = var.spec.http_config.allow_insecure
    header = length(var.spec.http_config.headers) > 0 ? {
      for name, wrapper in var.spec.http_config.headers : name => wrapper.values
    } : null
  } : null

  tcp_config = var.spec.tcp_config != null ? {
    method = var.spec.tcp_config.method
    port   = var.spec.tcp_config.port
  } : null
}
