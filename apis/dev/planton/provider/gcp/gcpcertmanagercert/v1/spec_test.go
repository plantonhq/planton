package gcpcertmanagercertv1

import (
	"testing"

	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGcpCertManagerCertSpec_Validation(t *testing.T) {
	tests := []struct {
		name    string
		spec    *GcpCertManagerCertSpec
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid spec with primary domain only",
			spec: &GcpCertManagerCertSpec{
				GcpProjectId:      "my-project",
				PrimaryDomainName: "example.com",
				CloudDnsZoneId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "example-zone",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid spec with wildcard domain",
			spec: &GcpCertManagerCertSpec{
				GcpProjectId:      "my-project",
				PrimaryDomainName: "*.example.com",
				CloudDnsZoneId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "example-zone",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid spec with alternate domains",
			spec: &GcpCertManagerCertSpec{
				GcpProjectId:      "my-project",
				PrimaryDomainName: "example.com",
				AlternateDomainNames: []string{
					"www.example.com",
					"api.example.com",
				},
				CloudDnsZoneId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "example-zone",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid spec with wildcard alternates",
			spec: &GcpCertManagerCertSpec{
				GcpProjectId:      "my-project",
				PrimaryDomainName: "example.com",
				AlternateDomainNames: []string{
					"*.example.com",
					"*.services.example.com",
				},
				CloudDnsZoneId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "example-zone",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing gcp project id",
			spec: &GcpCertManagerCertSpec{
				PrimaryDomainName: "example.com",
				CloudDnsZoneId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "example-zone",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing primary domain name",
			spec: &GcpCertManagerCertSpec{
				GcpProjectId: "my-project",
				CloudDnsZoneId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "example-zone",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid spec without cloud dns zone id (external DNS)",
			spec: &GcpCertManagerCertSpec{
				GcpProjectId:      "my-project",
				PrimaryDomainName: "example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid primary domain pattern",
			spec: &GcpCertManagerCertSpec{
				GcpProjectId:      "my-project",
				PrimaryDomainName: "invalid domain!",
				CloudDnsZoneId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "example-zone",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate alternate domain names",
			spec: &GcpCertManagerCertSpec{
				GcpProjectId:      "my-project",
				PrimaryDomainName: "example.com",
				AlternateDomainNames: []string{
					"www.example.com",
					"www.example.com", // duplicate
				},
				CloudDnsZoneId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "example-zone",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.spec.GcpProjectId == "" {
					assert.Empty(t, tt.spec.GcpProjectId, "gcp_project_id should be empty")
				}
				if tt.spec.PrimaryDomainName == "" {
					assert.Empty(t, tt.spec.PrimaryDomainName, "primary_domain_name should be empty")
				}
			} else {
				assert.NotEmpty(t, tt.spec.GcpProjectId, "gcp_project_id should not be empty")
				assert.NotEmpty(t, tt.spec.PrimaryDomainName, "primary_domain_name should not be empty")
			}
		})
	}
}

func TestCertificateType_Values(t *testing.T) {
	tests := []struct {
		name     string
		certType CertificateType
		expected string
	}{
		{
			name:     "MANAGED type",
			certType: CertificateType_MANAGED,
			expected: "MANAGED",
		},
		{
			name:     "LOAD_BALANCER type",
			certType: CertificateType_LOAD_BALANCER,
			expected: "LOAD_BALANCER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.certType.String())
		})
	}
}

func TestGcpCertManagerCert_Structure(t *testing.T) {
	cert := &GcpCertManagerCert{
		ApiVersion: "gcp.planton.dev/v1",
		Kind:       "GcpCertManagerCert",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-cert",
			Id:   "cert-001",
			Org:  "test-org",
			Env:  "production",
		},
		Spec: &GcpCertManagerCertSpec{
			GcpProjectId:      "test-project",
			PrimaryDomainName: "test.example.com",
			AlternateDomainNames: []string{
				"www.test.example.com",
			},
			CloudDnsZoneId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "test-zone",
				},
			},
			CertificateType:  func() *CertificateType { t := CertificateType_MANAGED; return &t }(),
			ValidationMethod: func() *string { s := "DNS"; return &s }(),
		},
	}

	require.NotNil(t, cert)
	assert.Equal(t, "gcp.planton.dev/v1", cert.ApiVersion)
	assert.Equal(t, "GcpCertManagerCert", cert.Kind)
	assert.Equal(t, "test-cert", cert.Metadata.Name)
	assert.Equal(t, "test-project", cert.Spec.GcpProjectId)
	assert.Equal(t, "test.example.com", cert.Spec.PrimaryDomainName)
	assert.Len(t, cert.Spec.AlternateDomainNames, 1)
	assert.Equal(t, CertificateType_MANAGED, *cert.Spec.CertificateType)
	assert.Equal(t, "DNS", *cert.Spec.ValidationMethod)
}

func TestGcpCertManagerCertStatus_Structure(t *testing.T) {
	status := &GcpCertManagerCertStatus{
		Outputs: &GcpCertManagerCertStackOutputs{
			CertificateId:         "cert-123",
			CertificateName:       "projects/test/locations/global/certificates/test-cert",
			CertificateDomainName: "test.example.com",
			CertificateStatus:     "ACTIVE",
			DnsValidationRecords: []*DnsValidationRecord{
				{
					RecordName: "_acme-challenge.test.example.com.",
					RecordType: "CNAME",
					RecordData: "abc123.authorize.certificatemanager.goog.",
					Domain:     "test.example.com",
				},
			},
		},
	}

	require.NotNil(t, status)
	require.NotNil(t, status.Outputs)
	assert.Equal(t, "cert-123", status.Outputs.CertificateId)
	assert.Equal(t, "projects/test/locations/global/certificates/test-cert", status.Outputs.CertificateName)
	assert.Equal(t, "test.example.com", status.Outputs.CertificateDomainName)
	assert.Equal(t, "ACTIVE", status.Outputs.CertificateStatus)
	require.Len(t, status.Outputs.DnsValidationRecords, 1)
	assert.Equal(t, "CNAME", status.Outputs.DnsValidationRecords[0].RecordType)
	assert.Equal(t, "test.example.com", status.Outputs.DnsValidationRecords[0].Domain)
}
