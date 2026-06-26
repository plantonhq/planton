package module

const (
	// OpApplicationId is the Access application ID.
	OpApplicationId = "application_id"
	// OpAud is the application's audience (AUD) tag, used to validate Access JWTs.
	OpAud = "aud"
	// OpDomain is the primary protected domain.
	OpDomain = "domain"
	// OpSaasClientId is the issued OAuth client ID (SaaS/OIDC apps).
	OpSaasClientId = "saas_client_id"
	// OpSaasClientSecret is the issued OAuth client secret (SaaS/OIDC apps).
	OpSaasClientSecret = "saas_client_secret"
	// OpSaasPublicKey is the IdP-facing public key (SaaS/SAML apps).
	OpSaasPublicKey = "saas_public_key"
	// OpSaasSsoEndpoint is the SSO endpoint URL (SaaS/SAML apps).
	OpSaasSsoEndpoint = "saas_sso_endpoint"
	// OpSaasIdpEntityId is the IdP entity ID (SaaS/SAML apps).
	OpSaasIdpEntityId = "saas_idp_entity_id"
)
