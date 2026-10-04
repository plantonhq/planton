package verify

// Spec-map readers for the object/vector data kinds. Manifests arrive in
// snake_case or camelCase depending on authoring surface — read both, the
// same convention as the search helpers.

// seaweedfsS3Map returns the spec's s3 block (or nil).
func seaweedfsS3Map(spec map[string]interface{}) map[string]interface{} {
	s3, _ := spec["s3"].(map[string]interface{})
	return s3
}

// seaweedfsS3Dedicated reports whether the gateway runs as its own
// Deployment (spec.s3.dedicated declared).
func seaweedfsS3Dedicated(spec map[string]interface{}) bool {
	s3 := seaweedfsS3Map(spec)
	if s3 == nil {
		return false
	}
	_, ok := s3["dedicated"].(map[string]interface{})
	return ok
}

// seaweedfsCredentialsSecret resolves the S3 credentials Secret the
// verifier reads: the referenced existing config secret, else the
// chart-generated "<name>-s3-secret" — or "" when auth is explicitly off
// (the kind's default is auth ON).
func seaweedfsCredentialsSecret(spec map[string]interface{}, resourceName string) string {
	s3 := seaweedfsS3Map(spec)
	enabled := true
	auth := true
	if s3 != nil {
		if v, ok := s3["enabled"].(bool); ok {
			enabled = v
		}
		if v, ok := s3["enable_auth"].(bool); ok {
			auth = v
		}
		if v, ok := s3["enableAuth"].(bool); ok {
			auth = v
		}
		if existing, _ := s3["existing_config_secret"].(string); existing != "" {
			return existing
		}
		if existing, _ := s3["existingConfigSecret"].(string); existing != "" {
			return existing
		}
	}
	if !enabled || !auth {
		return ""
	}
	return resourceName + "-s3-secret"
}

// seaweedfsBuckets lists the declared bucket names.
func seaweedfsBuckets(spec map[string]interface{}) []string {
	s3 := seaweedfsS3Map(spec)
	if s3 == nil {
		return nil
	}
	raw, _ := s3["buckets"].([]interface{})
	names := make([]string, 0, len(raw))
	for _, entry := range raw {
		bucket, _ := entry.(map[string]interface{})
		if name, _ := bucket["name"].(string); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// seaweedfsAdminEnabled reports whether the admin console is declared on.
func seaweedfsAdminEnabled(spec map[string]interface{}) bool {
	admin, _ := spec["admin"].(map[string]interface{})
	if admin == nil {
		return false
	}
	enabled, _ := admin["enabled"].(bool)
	return enabled
}

// qdrantApiKeySecretName resolves the chart-owned "<name>-apikey" Secret
// when any read-write key arm is declared (generate or existing — the
// chart copies referenced key material into its own Secret either way), or
// "" for an unauthenticated cluster.
func qdrantApiKeySecretName(spec map[string]interface{}, resourceName string) string {
	apiKey, _ := spec["api_key"].(map[string]interface{})
	if apiKey == nil {
		apiKey, _ = spec["apiKey"].(map[string]interface{})
	}
	if apiKey == nil {
		return ""
	}
	return resourceName + "-apikey"
}
