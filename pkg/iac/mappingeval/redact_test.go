//go:build !codegen

package mappingeval_test

import (
	"testing"

	"github.com/plantonhq/planton/pkg/iac/mappingeval"
)

// TestRedactSeedFingerprints pins keep-vs-strip: exactly the platform's
// seeding fingerprints leave; every realistic account signal stays.
func TestRedactSeedFingerprints(t *testing.T) {
	tag := func(key, value string) map[string]any {
		return map[string]any{"Key": key, "Value": value}
	}
	scan := &mappingeval.Scan{
		Region: "us-west-2",
		Resources: []mappingeval.ScannedResource{
			{
				TypeName:   "AWS::EC2::VPC",
				Identifier: "vpc-1",
				Properties: map[string]any{
					"VpcId": "vpc-1",
					"Tags": []any{
						// Fingerprints -- must be stripped.
						tag("planton.ai/resource", "true"),
						tag("planton.ai/resource-kind", "AwsVpc"),
						tag("planton.ai/environment", "prod"),
						tag("e2e-catalog-kind", "awsvpc"),
						tag("managed-by", "planton-e2e"),
						// Realistic signals -- must stay.
						tag("Name", "orders-prod-vpc"),
						tag("managed-by", "terraform"),
						tag("team", "payments"),
					},
				},
			},
			{
				// No Tags property at all -- must pass through untouched.
				TypeName:   "AWS::EC2::SubnetRouteTableAssociation",
				Identifier: "rtbassoc-1",
				Properties: map[string]any{"Id": "rtbassoc-1"},
			},
		},
	}

	mappingeval.RedactSeedFingerprints(scan)

	kept, _ := scan.Resources[0].Properties["Tags"].([]any)
	if len(kept) != 3 {
		t.Fatalf("want exactly the 3 realistic tags kept, got %d: %v", len(kept), kept)
	}
	wantKeys := map[string]string{"Name": "orders-prod-vpc", "managed-by": "terraform", "team": "payments"}
	for _, raw := range kept {
		entry := raw.(map[string]any)
		key := entry["Key"].(string)
		wantValue, expected := wantKeys[key]
		if !expected || entry["Value"].(string) != wantValue {
			t.Fatalf("unexpected surviving tag %v", entry)
		}
	}
	if _, hasTags := scan.Resources[1].Properties["Tags"]; hasTags {
		t.Fatal("a resource without Tags must not gain a Tags property")
	}
}
