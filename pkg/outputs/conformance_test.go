//go:build !codegen
// +build !codegen

package outputs

import (
	"path/filepath"
	"testing"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
)

// TestStackOutputsConformance is the standing guard against the systemic IaC
// output-drift class: an engine emits output names/shapes that do not flatten
// onto the kind's StackOutputs proto, silently leaving those proto fields empty.
// (The original bug: the Postgres tofu module emitted a flat
// "password_secret_name" output, which flattens to the key "password_secret_name"
// -- with no dot -- and therefore never populated the proto's nested
// password_secret{name,key} field, while the Pulumi module emitted the correct
// "password_secret.name". See the openmcf-postgres-iac-parity work.)
//
// Why this also enforces tofu<->pulumi parity: both engines feed the SAME generic
// transformer (TransformRaw -> Flatten -> populateMessage). So a single
// conformance bar per kind -- "this representative output set fully populates the
// proto with nothing left unmapped" -- when satisfied by each engine's emitted
// output set, guarantees the two engines produce the same typed StackOutputs.
//
// To extend coverage: add a case with the raw output shape an engine emits (scalars
// as strings; nested objects as map[string]interface{}, exactly how Terraform state
// and the Pulumi automation API surface them) and the proto fields it must populate.
func TestStackOutputsConformance(t *testing.T) {
	// A module dir with no transform override forces the generic reflection path,
	// which is the convention every in-repo module relies on (0 of 364 use an override).
	genericModuleDir := filepath.Join("testdata", "modules", "empty")

	cases := []struct {
		name string
		kind cloudresourcekind.CloudResourceKind
		// rawOutputs mirrors the post-Flatten-input shape both engines emit.
		rawOutputs map[string]interface{}
		// mustPopulate lists StackOutputs proto fields that MUST be set.
		mustPopulate []string
	}{
		{
			name: "KubernetesPostgres",
			kind: cloudresourcekind.CloudResourceKind_KubernetesPostgres,
			rawOutputs: map[string]interface{}{
				"namespace":            "gosilver-prod",
				"service":              "gosilver-prod-postgres-master",
				"port_forward_command": "kubectl port-forward -n gosilver-prod service/gosilver-prod-postgres-master 8080:8080",
				"kube_endpoint":        "gosilver-prod-postgres-master.gosilver-prod.svc.cluster.local",
				"external_hostname":    "gosilver-prod-postgres.planton.live",
				// Nested objects -- the shape that flattens to password_secret.name etc.
				"password_secret": map[string]interface{}{
					"name": "postgres.db-gosilver-prod-postgres.credentials.postgresql.acid.zalan.do",
					"key":  "password",
				},
				"username_secret": map[string]interface{}{
					"name": "postgres.db-gosilver-prod-postgres.credentials.postgresql.acid.zalan.do",
					"key":  "username",
				},
			},
			mustPopulate: []string{
				"namespace", "service", "port_forward_command", "kube_endpoint",
				"external_hostname", "password_secret", "username_secret",
			},
		},
		{
			// AwsSubnet: flat scalar outputs from both engines (subnet id/arn, AZ,
			// CIDR, route table id, region) must each land on the StackOutputs proto.
			name: "AwsSubnet",
			kind: cloudresourcekind.CloudResourceKind_AwsSubnet,
			rawOutputs: map[string]interface{}{
				"subnet_id":         "subnet-0abc123",
				"subnet_arn":        "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-0abc123",
				"availability_zone": "us-west-2a",
				"cidr_block":        "10.0.1.0/24",
				"route_table_id":    "rtb-0abc123",
				"region":            "us-west-2",
			},
			mustPopulate: []string{
				"subnet_id", "subnet_arn", "availability_zone",
				"cidr_block", "route_table_id", "region",
			},
		},
		{
			// AwsInternetGateway: flat scalar outputs from both engines (gateway
			// id/arn, attached vpc id, region) must each land on the StackOutputs proto.
			name: "AwsInternetGateway",
			kind: cloudresourcekind.CloudResourceKind_AwsInternetGateway,
			rawOutputs: map[string]interface{}{
				"internet_gateway_id":  "igw-0abc123",
				"internet_gateway_arn": "arn:aws:ec2:us-west-2:123456789012:internet-gateway/igw-0abc123",
				"vpc_id":               "vpc-0abc123",
				"region":               "us-west-2",
			},
			mustPopulate: []string{
				"internet_gateway_id", "internet_gateway_arn", "vpc_id", "region",
			},
		},
		{
			// AwsEgressOnlyInternetGateway: flat scalar outputs from both engines
			// (gateway id, attached vpc id, region) must each land on the StackOutputs
			// proto. An egress-only gateway has no ARN, so none is emitted.
			name: "AwsEgressOnlyInternetGateway",
			kind: cloudresourcekind.CloudResourceKind_AwsEgressOnlyInternetGateway,
			rawOutputs: map[string]interface{}{
				"egress_only_internet_gateway_id": "eigw-0abc123",
				"vpc_id":                          "vpc-0abc123",
				"region":                          "us-west-2",
			},
			mustPopulate: []string{
				"egress_only_internet_gateway_id", "vpc_id", "region",
			},
		},
		{
			// AwsNatGateway: flat scalar outputs from both engines (gateway id,
			// public/private ip, ENI id, subnet id, region) must each land on the
			// StackOutputs proto. A NAT gateway has no ARN, so none is emitted.
			name: "AwsNatGateway",
			kind: cloudresourcekind.CloudResourceKind_AwsNatGateway,
			rawOutputs: map[string]interface{}{
				"nat_gateway_id":       "nat-0abc123",
				"public_ip":            "52.10.20.30",
				"private_ip":           "10.0.0.10",
				"network_interface_id": "eni-0abc123",
				"subnet_id":            "subnet-0abc123",
				"region":               "us-west-2",
			},
			mustPopulate: []string{
				"nat_gateway_id", "public_ip", "private_ip",
				"network_interface_id", "subnet_id", "region",
			},
		},
		{
			// AwsVpc: flat scalar outputs from both engines (vpc id/arn, primary and
			// IPv6 CIDR, owner, the route-table/default-resource ids, region) must
			// each land on the thin StackOutputs proto.
			name: "AwsVpc",
			kind: cloudresourcekind.CloudResourceKind_AwsVpc,
			rawOutputs: map[string]interface{}{
				"vpc_id":                    "vpc-0abc123",
				"vpc_arn":                   "arn:aws:ec2:us-west-2:123456789012:vpc/vpc-0abc123",
				"cidr_block":                "10.0.0.0/16",
				"ipv6_cidr_block":           "2600:1f18:abcd:1200::/56",
				"owner_id":                  "123456789012",
				"main_route_table_id":       "rtb-0abc123",
				"default_security_group_id": "sg-0abc123",
				"default_network_acl_id":    "acl-0abc123",
				"default_route_table_id":    "rtb-0abc123",
				"region":                    "us-west-2",
			},
			mustPopulate: []string{
				"vpc_id", "vpc_arn", "cidr_block", "ipv6_cidr_block", "owner_id",
				"main_route_table_id", "default_security_group_id",
				"default_network_acl_id", "default_route_table_id", "region",
			},
		},
		{
			// Guards the externaldns tofu module's output rename to solver_sa: the
			// module previously emitted "service_account_name", which does not flatten
			// onto the KubernetesExternalDnsStackOutputs.solver_sa proto field (the
			// Pulumi module already exported "solver_sa"). Both engines now emit the
			// same three outputs.
			name: "KubernetesExternalDns",
			kind: cloudresourcekind.CloudResourceKind_KubernetesExternalDns,
			rawOutputs: map[string]interface{}{
				"namespace":    "external-dns",
				"release_name": "gosilver-in-external-dns",
				"solver_sa":    "gosilver-in-external-dns",
			},
			mustPopulate: []string{"namespace", "release_name", "solver_sa"},
		},
		{
			// CloudflareR2Bucket: both engines emit the same three flat outputs
			// (bucket name, path-style S3 URL, custom domain URL), each of which
			// must land on the StackOutputs proto.
			name: "CloudflareR2Bucket",
			kind: cloudresourcekind.CloudResourceKind_CloudflareR2Bucket,
			rawOutputs: map[string]interface{}{
				"bucket_name":       "media-assets",
				"bucket_url":        "https://00000000000000000000000000000000.r2.cloudflarestorage.com/media-assets",
				"custom_domain_url": "https://media.example.com",
			},
			mustPopulate: []string{"bucket_name", "bucket_url", "custom_domain_url"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidateOverride(tc.kind, genericModuleDir, tc.rawOutputs)
			if err != nil {
				t.Fatalf("ValidateOverride failed: %v", err)
			}
			if len(result.SchemaErrors) != 0 {
				t.Fatalf("unexpected schema errors: %v", result.SchemaErrors)
			}
			if result.DryRun == nil {
				t.Fatal("expected a dry-run result")
			}

			// Core invariant: every emitted output lands on a proto field. A
			// regression to a flat/mismatched output name surfaces here.
			if len(result.DryRun.UnmappedOutputs) != 0 {
				t.Errorf("%s: outputs did not map onto the StackOutputs proto: %v",
					tc.kind.String(), result.DryRun.UnmappedOutputs)
			}

			populated := make(map[string]bool, len(result.DryRun.PopulatedFields))
			for _, f := range result.DryRun.PopulatedFields {
				populated[f.ProtoField] = true
			}
			for _, field := range tc.mustPopulate {
				if !populated[field] {
					t.Errorf("%s: expected proto field %q to be populated, but it was not",
						tc.kind.String(), field)
				}
			}
		})
	}
}

// TestStackOutputsConformance_DetectsFlatSecretDrift proves the guard actually
// catches the historical drift: the pre-fix Postgres tofu module emitted flat
// "password_secret_name"/"password_secret_key" outputs, which do NOT flatten onto
// the proto's password_secret{name,key} field. The guard must flag both the
// unmapped output and the unpopulated proto field.
func TestStackOutputsConformance_DetectsFlatSecretDrift(t *testing.T) {
	genericModuleDir := filepath.Join("testdata", "modules", "empty")
	kind := cloudresourcekind.CloudResourceKind_KubernetesPostgres

	flatDriftOutputs := map[string]interface{}{
		"namespace":            "gosilver-prod",
		"password_secret_name": "postgres.db-gosilver-prod-postgres.credentials.postgresql.acid.zalan.do",
		"password_secret_key":  "password",
	}

	result, err := ValidateOverride(kind, genericModuleDir, flatDriftOutputs)
	if err != nil {
		t.Fatalf("ValidateOverride failed: %v", err)
	}
	if result.DryRun == nil {
		t.Fatal("expected a dry-run result")
	}

	if len(result.DryRun.UnmappedOutputs) == 0 {
		t.Error("expected the flat password_secret_name/_key outputs to be reported as unmapped, but none were")
	}
	for _, f := range result.DryRun.PopulatedFields {
		if f.ProtoField == "password_secret" {
			t.Error("flat outputs must NOT populate the nested password_secret proto field")
		}
	}
}
