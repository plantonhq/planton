package aws

import (
	"testing"

	"buf.build/go/protovalidate"
)

// TestAwsWebIdentityProviderConfig_TokenXorFile exercises the message-level CEL
// (aws.web_identity.token_xor_file): exactly one of web_identity_token (inline JWT) or
// web_identity_token_file (path the classic provider re-reads) must be set.
func TestAwsWebIdentityProviderConfig_TokenXorFile(t *testing.T) {
	const roleArn = "arn:aws:iam::123456789012:role/customer-oidc"

	cases := []struct {
		name    string
		token   string
		file    string
		wantErr bool
	}{
		{"inline token only is valid", "eyJhbGciOiJSUzI1NiJ9.payload.sig", "", false},
		{"token file only is valid", "", "/var/run/planton/web-identity-token", false},
		{"both set is invalid", "eyJhbGciOiJSUzI1NiJ9.payload.sig", "/var/run/planton/web-identity-token", true},
		{"neither set is invalid", "", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &AwsWebIdentityProviderConfig{
				WebIdentityToken:     tc.token,
				WebIdentityTokenFile: tc.file,
				RoleArn:              roleArn,
			}
			err := protovalidate.Validate(cfg)
			if tc.wantErr && err == nil {
				t.Fatalf("expected a validation error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no validation error, got %v", err)
			}
		})
	}
}
