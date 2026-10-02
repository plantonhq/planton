variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "GcpDocumentAiProcessor specification"
  type = object({
    # The GCP project the processor lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Where documents are processed and stored: the multi-region "us" or
    # "eu" (Document AI's main locations), or a region Google lists for
    # Document AI. Your documents never leave it. Immutable.
    location = string

    # What the processor extracts -- one of the types Google lists through
    # fetchProcessorTypes for your project and location, for example:
    #   OCR_PROCESSOR          text, layout, and handwriting (Enterprise OCR)
    #   FORM_PARSER_PROCESSOR  key-value pairs, tables, and checkboxes
    #   LAYOUT_PARSER_PROCESSOR document structure and chunks for RAG
    #   INVOICE_PROCESSOR, EXPENSE_PROCESSOR, ID_PROOFING_PROCESSOR, and
    #   other specialized parsers; CUSTOM_EXTRACTION_PROCESSOR for a
    #   generative extractor you train in the console
    # Some types need an allowlist or a specific location. Immutable.
    type = string

    # The processor's display name -- unique among the project's processors
    # in the location. Defaults to metadata.name. Immutable.
    display_name = optional(string, "")

    # Customer-managed encryption key for the documents and results the
    # processor stores: a GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # in a location compatible with the processor's. The Document AI service
    # agent needs roles/cloudkms.cryptoKeyEncrypterDecrypter on it. Omit for
    # Google-managed encryption. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # The processor version that serves requests which do not name one --
    # its short id, e.g. "pretrained-ocr-v2.1-2024-08-07" for a Google
    # version or the id of a version you trained. The modules compose the
    # full path under this processor. Omit to keep Google's default.
    #
    # Use a concrete version, never the "stable" or "rc" channel alias:
    # Google answers with the version the alias currently resolves to, so a
    # declared alias would never match what Google stores and every deploy
    # would set the default again. Pinning a version is also what makes
    # extraction results reproducible. Changing it re-points the default in
    # place; destroying the block leaves the last default set (Google has no
    # "unset").
    default_version = optional(string, "")

    # What happens to the processor when this resource is destroyed:
    #   "" / "DELETE" -- the processor, its versions, and any version you
    #                    trained are deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the processor leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
