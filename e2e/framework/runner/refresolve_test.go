package runner

import (
	"os"
	"path/filepath"
	"testing"

	awsiamrolev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsiamrole/v1"
	awssubnetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awssubnet/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/internal/manifest"
)

const subnetManifestWithRef = `apiVersion: aws.planton.dev/v1
kind: AwsSubnet
metadata:
  name: ref-subnet
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      fieldPath: status.outputs.vpc_id
  availabilityZone: us-west-2a
  cidrBlock: 10.0.1.0/24
`

func writeTempManifest(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manifest.yaml")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatalf("failed to write temp manifest: %v", err)
	}
	return path
}

func TestResolveManifestRefs_ResolvesVpcIdFromPrerequisite(t *testing.T) {
	manifestPath := writeTempManifest(t, subnetManifestWithRef)

	depOutputs := map[cloudresourcekind.CloudResourceKind]map[string]interface{}{
		cloudresourcekind.CloudResourceKind_AwsVpc: {
			"vpc_id":   "vpc-resolved123",
			"vpc_cidr": "10.0.0.0/16",
		},
	}

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}
	if resolvedPath == manifestPath {
		t.Fatal("expected a new resolved manifest path, got the original")
	}

	obj, err := manifest.LoadManifest(resolvedPath)
	if err != nil {
		t.Fatalf("failed to load resolved manifest: %v", err)
	}
	subnet, ok := obj.(*awssubnetv1.AwsSubnet)
	if !ok {
		t.Fatalf("resolved manifest is not an AwsSubnet: %T", obj)
	}
	if got := subnet.GetSpec().GetVpcId().GetValue(); got != "vpc-resolved123" {
		t.Errorf("vpc_id value = %q, want %q", got, "vpc-resolved123")
	}
	if subnet.GetSpec().GetVpcId().GetValueFrom() != nil {
		t.Error("vpc_id should be a literal after resolution, but value_from is still set")
	}
}

const roleManifestWithRepeatedRefs = `apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: ref-role
spec:
  region: us-west-2
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: ec2.amazonaws.com
        Action: sts:AssumeRole
  managedPolicyArns:
    - value: arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore
    - valueFrom:
        kind: AwsIamPolicy
        name: my-policy
        fieldPath: status.outputs.policy_arn
`

// The repeated case: a list mixing a literal (an AWS-managed policy ARN) with a
// reference to a deployed prerequisite. Each element resolves independently and
// the literal passes through untouched.
func TestResolveManifestRefs_ResolvesRepeatedRefsFromPrerequisite(t *testing.T) {
	manifestPath := writeTempManifest(t, roleManifestWithRepeatedRefs)

	depOutputs := map[cloudresourcekind.CloudResourceKind]map[string]interface{}{
		cloudresourcekind.CloudResourceKind_AwsIamPolicy: {
			"policy_arn":  "arn:aws:iam::123456789012:policy/my-policy",
			"policy_name": "my-policy",
		},
	}

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}
	if resolvedPath == manifestPath {
		t.Fatal("expected a new resolved manifest path, got the original")
	}

	obj, err := manifest.LoadManifest(resolvedPath)
	if err != nil {
		t.Fatalf("failed to load resolved manifest: %v", err)
	}
	role, ok := obj.(*awsiamrolev1.AwsIamRole)
	if !ok {
		t.Fatalf("resolved manifest is not an AwsIamRole: %T", obj)
	}
	arns := role.GetSpec().GetManagedPolicyArns()
	if len(arns) != 2 {
		t.Fatalf("managed_policy_arns length = %d, want 2", len(arns))
	}
	if got := arns[0].GetValue(); got != "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore" {
		t.Errorf("literal element = %q, want the untouched AWS-managed ARN", got)
	}
	if got := arns[1].GetValue(); got != "arn:aws:iam::123456789012:policy/my-policy" {
		t.Errorf("resolved element = %q, want the prerequisite's policy_arn", got)
	}
	if arns[1].GetValueFrom() != nil {
		t.Error("resolved element should be a literal, but value_from is still set")
	}
}

// A repeated ref whose kind has no deployed prerequisite is left untouched
// rather than erroring -- matching the singular behavior.
func TestResolveManifestRefs_RepeatedRefWithoutPrerequisiteLeftUntouched(t *testing.T) {
	manifestPath := writeTempManifest(t, roleManifestWithRepeatedRefs)

	depOutputs := map[cloudresourcekind.CloudResourceKind]map[string]interface{}{
		cloudresourcekind.CloudResourceKind_AwsVpc: {
			"vpc_id": "vpc-unrelated",
		},
	}

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}
	if resolvedPath != manifestPath {
		t.Errorf("expected original path when no matching prerequisite is deployed, got %q", resolvedPath)
	}
}

func TestResolveManifestRefs_NoDependenciesReturnsOriginal(t *testing.T) {
	manifestPath := writeTempManifest(t, subnetManifestWithRef)

	resolvedPath, err := ResolveManifestRefs(manifestPath, nil)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}
	if resolvedPath != manifestPath {
		t.Errorf("expected original path when there are no dependencies, got %q", resolvedPath)
	}
}

func TestResolveManifestRefs_MissingOutputErrors(t *testing.T) {
	manifestPath := writeTempManifest(t, subnetManifestWithRef)

	depOutputs := map[cloudresourcekind.CloudResourceKind]map[string]interface{}{
		cloudresourcekind.CloudResourceKind_AwsVpc: {
			"vpc_cidr": "10.0.0.0/16", // vpc_id intentionally absent
		},
	}

	if _, err := ResolveManifestRefs(manifestPath, depOutputs); err == nil {
		t.Fatal("expected an error when the prerequisite output is missing, got nil")
	}
}
