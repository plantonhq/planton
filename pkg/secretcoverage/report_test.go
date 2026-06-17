package secretcoverage

import "testing"

func TestBuildReport(t *testing.T) {
	findings := []Finding{
		{Kind: "AwsRdsInstance", Provider: "aws", Path: "spec.password", Class: Covered},
		{Kind: "AwsRdsInstance", Provider: "aws", Path: "spec.kms_key_id", Class: Exempt, ExemptReason: "kms id"},
		{Kind: "AwsEcsService", Provider: "aws", Path: "spec.container.env.secrets", Class: Gap},
		{Kind: "GcpCloudSql", Provider: "gcp", Path: "spec.root_password", Class: Covered},
		{Kind: "GcpCloudSql", Provider: "gcp", Path: "spec.name", Class: NotSensitive},
	}

	report := BuildReport(findings)

	if report.Covered != 2 {
		t.Errorf("covered = %d, want 2", report.Covered)
	}
	if report.Exempt != 1 {
		t.Errorf("exempt = %d, want 1", report.Exempt)
	}
	if report.Gap != 1 {
		t.Errorf("gap = %d, want 1", report.Gap)
	}

	// Overall coverage = (covered+exempt)/(covered+exempt+gap) = 3/4 = 75%.
	if report.CoveragePercent != 75 {
		t.Errorf("coveragePercent = %.1f, want 75", report.CoveragePercent)
	}

	// Providers sorted by name: aws then gcp.
	if len(report.Providers) != 2 {
		t.Fatalf("providers = %d, want 2", len(report.Providers))
	}
	if report.Providers[0].Provider != "aws" || report.Providers[1].Provider != "gcp" {
		t.Errorf("providers not sorted: %+v", report.Providers)
	}
	// aws: covered=1 exempt=1 gap=1 -> 66.7%; gcp: covered=1 -> 100% (NotSensitive ignored).
	if report.Providers[1].Provider == "gcp" && report.Providers[1].CoveragePercent != 100 {
		t.Errorf("gcp coverage = %.1f, want 100", report.Providers[1].CoveragePercent)
	}

	// Gaps list carries the one gap ID.
	if len(report.Gaps) != 1 {
		t.Errorf("gaps = %v, want 1 entry", report.Gaps)
	}
}
