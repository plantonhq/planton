package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	pkgerrors "github.com/pkg/errors"
)

// vpcVerifier verifies an AwsVpc via DescribeVpcs. It exists so an AwsVpc can be
// used as a deployed E2E prerequisite (e.g. for AwsSubnet) and confirmed live. A
// deleted VPC returns the typed InvalidVpcID.NotFound error (the "absent" signal).
//
// When the outputs report secondary CIDR associations (the
// secondary_ipv4/ipv6_cidr_association_ids maps, keyed by the module's
// for_each keys), existence asserts each association id is present AND
// "associated" in the VPC's own CidrBlockAssociationSet -- an association can
// describe in "failed"/"disassociated" states, so id presence alone is not
// posture. Absence needs only the VPC probe: associations are children of
// the VPC and cannot outlive it.
type vpcVerifier struct{}

func (*vpcVerifier) IDOutputKey() string { return "vpc_id" }

func (*vpcVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := vpcExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsvpc verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsvpc %q not found after deploy", id)
	}
	return nil
}

func (*vpcVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := vpcExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsvpc verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsvpc %q still exists after destroy", id)
	}
	return nil
}

func (v *vpcVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	vpcId := stringOutputMap(outputs, "vpc_id")
	if vpcId == "" {
		return pkgerrors.New("awsvpc verify-exists: no vpc_id in outputs")
	}
	if err := v.VerifyExists(ctx, cfg, vpcId, region); err != nil {
		return err
	}

	wantIpv4 := vpcAssociationIds(outputs, "secondary_ipv4_cidr_association_ids")
	wantIpv6 := vpcAssociationIds(outputs, "secondary_ipv6_cidr_association_ids")
	if len(wantIpv4) == 0 && len(wantIpv6) == 0 {
		return nil
	}

	out, err := vpcClient(cfg, region).DescribeVpcs(ctx, &ec2.DescribeVpcsInput{VpcIds: []string{vpcId}})
	if err != nil || len(out.Vpcs) == 0 {
		return pkgerrors.Wrapf(err, "awsvpc %q: describing for association posture failed", vpcId)
	}
	vpc := out.Vpcs[0]

	associated := map[string]bool{}
	for _, assoc := range vpc.CidrBlockAssociationSet {
		if assoc.AssociationId != nil && assoc.CidrBlockState != nil &&
			assoc.CidrBlockState.State == types.VpcCidrBlockStateCodeAssociated {
			associated[*assoc.AssociationId] = true
		}
	}
	for _, assoc := range vpc.Ipv6CidrBlockAssociationSet {
		if assoc.AssociationId != nil && assoc.Ipv6CidrBlockState != nil &&
			assoc.Ipv6CidrBlockState.State == types.VpcCidrBlockStateCodeAssociated {
			associated[*assoc.AssociationId] = true
		}
	}
	for key, associationId := range wantIpv4 {
		if !associated[associationId] {
			return pkgerrors.Errorf("awsvpc %q: secondary IPv4 CIDR association %s (key %q) not in associated state",
				vpcId, associationId, key)
		}
	}
	for key, associationId := range wantIpv6 {
		if !associated[associationId] {
			return pkgerrors.Errorf("awsvpc %q: secondary IPv6 CIDR association %s (key %q) not in associated state",
				vpcId, associationId, key)
		}
	}
	return nil
}

func (v *vpcVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	vpcId := stringOutputMap(outputs, "vpc_id")
	if vpcId == "" {
		return pkgerrors.New("awsvpc verify-absent: no vpc_id in outputs")
	}
	return v.VerifyAbsent(ctx, cfg, vpcId, region)
}

// vpcAssociationIds reads one of the association-id output maps (for_each
// key -> association id). A missing or empty map means the spec declared no
// secondary CIDRs of that family.
func vpcAssociationIds(outputs map[string]interface{}, outputName string) map[string]string {
	raw, ok := outputs[outputName]
	if !ok || raw == nil {
		return nil
	}
	ids := map[string]string{}
	switch m := raw.(type) {
	case map[string]interface{}:
		for key, id := range m {
			if s, isStr := id.(string); isStr && s != "" {
				ids[key] = s
			}
		}
	case map[string]string:
		for key, id := range m {
			if id != "" {
				ids[key] = id
			}
		}
	}
	return ids
}

func vpcClient(cfg aws.Config, region string) *ec2.Client {
	return ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		if region != "" {
			o.Region = region
		}
	})
}

func vpcExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	out, err := vpcClient(cfg, region).DescribeVpcs(ctx, &ec2.DescribeVpcsInput{VpcIds: []string{id}})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidVpcID.NotFound" {
			return false, nil
		}
		return false, err
	}
	return len(out.Vpcs) > 0, nil
}
