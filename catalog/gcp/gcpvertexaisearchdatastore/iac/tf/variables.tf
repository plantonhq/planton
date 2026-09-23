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
  description = "GcpVertexAiSearchDataStore specification"
  type = object({
    # The GCP project the store lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Where the store and its engines live: "global", "us", or "eu"
    # (Discovery Engine multi-regions, not Compute regions). An engine must
    # be in the same location as its stores. Immutable.
    location = string

    # The store's id -- 1-63 characters, RFC 1034 (lowercase letters, digits,
    # hyphens; starts with a letter). Defaults to metadata.name. Immutable.
    data_store_id = optional(string, "")

    # Human-readable name shown in the console (up to 128 characters).
    # Defaults to metadata.name. Mutable.
    display_name = optional(string, "")

    # The industry the corpus belongs to: GENERIC for most content, MEDIA for
    # videos/articles/music with media recommendation, HEALTHCARE_FHIR for
    # FHIR stores. Fixes which engines can use the store. Immutable.
    industry_vertical = string

    # What the store holds: NO_CONTENT for structured records (JSON with a
    # schema), CONTENT_REQUIRED for unstructured documents (PDF, HTML, DOCX
    # and their metadata), PUBLIC_WEBSITE for a public website crawled by
    # URL pattern. Empty lets Google default (NO_CONTENT). Immutable.
    content_config = optional(string, "")

    # The solutions the store enrolls in -- an engine can only use stores
    # enrolled in its solution: SOLUTION_TYPE_SEARCH, SOLUTION_TYPE_CHAT,
    # SOLUTION_TYPE_RECOMMENDATION, SOLUTION_TYPE_GENERATIVE_CHAT. Empty lets
    # Google default (search). Immutable.
    solution_types = optional(list(string), [])

    # True means every document carries access-control information that is
    # ingested with it and enforced at search time (needs the location's
    # ACL config / identity provider). Documents in an ACL-enabled store
    # cannot be read back through GetDocument. Immutable.
    acl_enabled = optional(bool, false)

    # Create an ADVANCED site search store (PUBLIC_WEBSITE only; ignored
    # otherwise): domain-verified crawling, sitemaps, and more index quota
    # versus basic site search over public pages. Input-only; Google does
    # not read it back. Immutable.
    create_advanced_site_search = optional(bool, false)

    # Advanced site search behavior; applies only with
    # create_advanced_site_search. Immutable.
    advanced_site_search_config = optional(object({
      # Do not index the site when the store is created; indexing starts on
      # demand (the console's "Recrawl" or the API).
      disable_initial_index = optional(bool, false)

      # Do not refresh the index automatically; content is recrawled only on
      # demand.
      disable_automatic_refresh = optional(bool, false)
    }))

    # Do not create Google's default schema. Required when declaring a
    # custom schema below, and only then: a store with no schema accepts no
    # documents until one exists. Input-only.
    skip_default_schema_creation = optional(bool, false)

    # The store's schema (with skip_default_schema_creation). Omit to let
    # Google create the default schema and infer fields from the data.
    schema = optional(object({
      # The schema's id within the store, e.g. "default_schema" or your own.
      schema_id = string

      # The schema as a compact JSON string (Google normalizes it): a JSON
      # Schema object whose properties may carry Vertex AI Search's
      # keyPropertyMapping, indexable, searchable, retrievable, dynamicFacetable
      # annotations; "datetime_detection" and "geolocation_detection" at the
      # top level ask Google to infer those types.
      json_schema = string
    }))

    # How documents are parsed and chunked. Omit for Google's default: the
    # digital parser, whole documents. Immutable.
    document_processing_config = optional(object({
      # Layout-based chunking. Omit to index whole documents.
      chunking_config = optional(object({
        # Token limit per chunk, 100-500 (Google defaults to 500). Sent only
        # when set.
        chunk_size = optional(number)

        # Prepend the headings above a chunk (section, subsection) to chunks cut
        # from the middle of a document, so a passage keeps its context.
        include_ancestor_headings = optional(bool, false)
      }))

      # The parser applied to every file type not overridden below. Omit for
      # Google's default (the digital parser).
      default_parsing_config = optional(object({
        # True picks the digital parser: plain text extraction from digital
        # (born-electronic) documents, Google's default. It has no settings.
        digital_parsing = optional(bool, false)

        # The layout parser: structure-aware parsing for headings, tables, and
        # figures -- the parser to pick for RAG quality.
        layout_parsing_config = optional(object({
          # Keep the processed (layout-parsed) document available through the
          # GetProcessedDocument API, so an application can fetch the parsed
          # structure instead of the raw file.
          enable_get_processed_document = optional(bool, false)

          # Have an LLM describe each image during parsing and add the description
          # to the indexed text, so image content becomes searchable.
          enable_image_annotation = optional(bool, false)

          # Refine the detected PDF layout with an LLM (better reading order and
          # section boundaries on complex pages, at higher parsing cost).
          enable_llm_layout_parsing = optional(bool, false)

          # Have an LLM describe each table during parsing and add the description
          # to the indexed text.
          enable_table_annotation = optional(bool, false)

          # HTML classes whose elements are dropped from the parsed content (menus,
          # footers, cookie banners).
          exclude_html_classes = optional(list(string), [])

          # HTML element names dropped from the parsed content, e.g. "nav",
          # "footer", "script".
          exclude_html_elements = optional(list(string), [])

          # HTML element ids dropped from the parsed content.
          exclude_html_ids = optional(list(string), [])

          # Structured content types the parser must extract from the document.
          # Google supports "shareholder-structure" today.
          structured_content_types = optional(list(string), [])
        }))

        # The OCR parser for scanned PDFs.
        ocr_parsing_config = optional(object({
          # On pages that already carry a text layer, use that native text instead
          # of the OCR output -- faster and more accurate for mixed documents.
          use_native_text = optional(bool, false)
        }))
      }))

      # Per-file-type parser overrides, one per file type.
      parsing_config_overrides = optional(list(object({
        # The file type this override applies to: "pdf" (digital, OCR, or layout
        # parsing), "html" (digital or layout), "docx", "pptx", "xlsm", "xlsx"
        # (digital or layout). One override per file type.
        file_type = string

        # The parser for this file type.
        parsing_config = object({
          # True picks the digital parser: plain text extraction from digital
          # (born-electronic) documents, Google's default. It has no settings.
          digital_parsing = optional(bool, false)

          # The layout parser: structure-aware parsing for headings, tables, and
          # figures -- the parser to pick for RAG quality.
          layout_parsing_config = optional(object({
            # Keep the processed (layout-parsed) document available through the
            # GetProcessedDocument API, so an application can fetch the parsed
            # structure instead of the raw file.
            enable_get_processed_document = optional(bool, false)

            # Have an LLM describe each image during parsing and add the description
            # to the indexed text, so image content becomes searchable.
            enable_image_annotation = optional(bool, false)

            # Refine the detected PDF layout with an LLM (better reading order and
            # section boundaries on complex pages, at higher parsing cost).
            enable_llm_layout_parsing = optional(bool, false)

            # Have an LLM describe each table during parsing and add the description
            # to the indexed text.
            enable_table_annotation = optional(bool, false)

            # HTML classes whose elements are dropped from the parsed content (menus,
            # footers, cookie banners).
            exclude_html_classes = optional(list(string), [])

            # HTML element names dropped from the parsed content, e.g. "nav",
            # "footer", "script".
            exclude_html_elements = optional(list(string), [])

            # HTML element ids dropped from the parsed content.
            exclude_html_ids = optional(list(string), [])

            # Structured content types the parser must extract from the document.
            # Google supports "shareholder-structure" today.
            structured_content_types = optional(list(string), [])
          }))

          # The OCR parser for scanned PDFs.
          ocr_parsing_config = optional(object({
            # On pages that already carry a text layer, use that native text instead
            # of the OCR output -- faster and more accurate for mixed documents.
            use_native_text = optional(bool, false)
          }))
        })
      })), [])
    }))

    # Customer-managed encryption key protecting the store's data: a
    # GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}.
    # Omit for Google-managed encryption. Updatable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # The URL patterns a PUBLIC_WEBSITE store crawls (INCLUDE) or leaves out
    # (EXCLUDE). Basic site search needs no domain verification for public
    # pages; advanced site search needs the domain verified in Search
    # Console. Add or remove a pattern by editing the list; any other change
    # replaces that target site.
    target_sites = optional(list(object({
      # The URI pattern, e.g. "cloud.google.com/docs/*" or "www.example.com".
      # Without a wildcard and with exact_match false, every page whose address
      # contains the pattern is included.
      provided_uri_pattern = string

      # INCLUDE (Google's default when empty) crawls pages matching the
      # pattern; EXCLUDE removes matching pages from an included site.
      type = optional(string, "")

      # True matches the pattern exactly (or the one specific page it names);
      # false (Google's default) matches every page containing the pattern.
      exact_match = optional(bool, false)
    })), [])

    # Public sitemap URIs an advanced site search store reads, e.g.
    # "https://www.example.com/sitemap.xml". Each is one immutable sitemap
    # resource.
    sitemap_uris = optional(list(string), [])

    # What happens to the store, its schema, target sites, and sitemaps when
    # this resource is destroyed:
    #   "" / "DELETE" -- everything is deleted, documents included
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and stays in GCP
    # A store an engine still uses cannot be deleted; destroy the engine
    # first.
    deletion_policy = optional(string, "")
  })
}
