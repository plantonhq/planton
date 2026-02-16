# AwsRedshiftCluster Pulumi Examples

## 1. Minimal Dev Cluster

Single-node `dc2.large` for development. AWS manages the master password via
Secrets Manager. No final snapshot on deletion.

```go
package main

import (
    "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/redshift"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        subnetGroup, err := redshift.NewSubnetGroup(ctx, "dev-subnet-group", &redshift.SubnetGroupArgs{
            Name:      pulumi.String("dev-analytics-subnet-group"),
            SubnetIds: pulumi.StringArray{
                pulumi.String("subnet-aaa"),
                pulumi.String("subnet-bbb"),
            },
        })
        if err != nil {
            return err
        }

        cluster, err := redshift.NewCluster(ctx, "dev-analytics", &redshift.ClusterArgs{
            ClusterIdentifier:      pulumi.String("dev-analytics"),
            ClusterType:            pulumi.String("single-node"),
            NodeType:               pulumi.String("dc2.large"),
            DatabaseName:           pulumi.String("dev"),
            MasterUsername:         pulumi.String("admin"),
            ManageMasterPassword:   pulumi.BoolPtr(true),
            ClusterSubnetGroupName: subnetGroup.Name,
            Encrypted:              pulumi.BoolPtr(true),
            SkipFinalSnapshot:      pulumi.Bool(true),
        })
        if err != nil {
            return err
        }

        ctx.Export("endpoint", cluster.Endpoint)
        ctx.Export("clusterArn", cluster.Arn)
        return nil
    })
}
```

## 2. Production Multi-Node Cluster

Two-node `ra3.xlplus` cluster with KMS encryption, CloudWatch logging, managed
security group, and enhanced VPC routing.

```go
package main

import (
    "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
    "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/redshift"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        // Subnet group
        subnetGroup, err := redshift.NewSubnetGroup(ctx, "prod-subnet-group", &redshift.SubnetGroupArgs{
            Name: pulumi.String("prod-warehouse-subnet-group"),
            SubnetIds: pulumi.StringArray{
                pulumi.String("subnet-aaa"),
                pulumi.String("subnet-bbb"),
            },
        })
        if err != nil {
            return err
        }

        // Security group
        sg, err := ec2.NewSecurityGroup(ctx, "prod-redshift-sg", &ec2.SecurityGroupArgs{
            NamePrefix:  pulumi.String("prod-warehouse-redshift-"),
            Description: pulumi.String("Managed SG for prod Redshift cluster"),
            VpcId:       pulumi.String("vpc-12345"),
        })
        if err != nil {
            return err
        }

        _, err = ec2.NewSecurityGroupRule(ctx, "prod-redshift-ingress", &ec2.SecurityGroupRuleArgs{
            Type:                  pulumi.String("ingress"),
            FromPort:              pulumi.Int(5439),
            ToPort:                pulumi.Int(5439),
            Protocol:              pulumi.String("tcp"),
            SourceSecurityGroupId: pulumi.String("sg-app-layer"),
            SecurityGroupId:       sg.ID(),
        })
        if err != nil {
            return err
        }

        // Parameter group
        paramGroup, err := redshift.NewParameterGroup(ctx, "prod-params", &redshift.ParameterGroupArgs{
            Name:   pulumi.String("prod-warehouse-params"),
            Family: pulumi.String("redshift-1.0"),
            Parameters: redshift.ParameterGroupParameterArray{
                &redshift.ParameterGroupParameterArgs{
                    Name:  pulumi.String("require_ssl"),
                    Value: pulumi.String("true"),
                },
                &redshift.ParameterGroupParameterArgs{
                    Name:  pulumi.String("enable_user_activity_logging"),
                    Value: pulumi.String("true"),
                },
            },
        })
        if err != nil {
            return err
        }

        // Cluster
        cluster, err := redshift.NewCluster(ctx, "prod-warehouse", &redshift.ClusterArgs{
            ClusterIdentifier:      pulumi.String("prod-warehouse"),
            ClusterType:            pulumi.String("multi-node"),
            NodeType:               pulumi.String("ra3.xlplus"),
            NumberOfNodes:          pulumi.Int(2),
            DatabaseName:           pulumi.String("warehouse"),
            MasterUsername:         pulumi.String("warehouse_admin"),
            ManageMasterPassword:   pulumi.BoolPtr(true),
            Port:                   pulumi.Int(5439),
            ClusterSubnetGroupName: subnetGroup.Name,
            VpcSecurityGroupIds:    pulumi.StringArray{sg.ID().ToStringOutput()},
            EnhancedVpcRouting:     pulumi.BoolPtr(true),
            Encrypted:              pulumi.BoolPtr(true),
            KmsKeyId:               pulumi.String("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123"),
            IamRoles: pulumi.StringArray{
                pulumi.String("arn:aws:iam::123456789012:role/redshift-s3-role"),
            },
            DefaultIamRoleArn:                  pulumi.String("arn:aws:iam::123456789012:role/redshift-s3-role"),
            AutomatedSnapshotRetentionPeriod:    pulumi.Int(7),
            SkipFinalSnapshot:                   pulumi.Bool(false),
            FinalSnapshotIdentifier:             pulumi.String("prod-warehouse-final"),
            PreferredMaintenanceWindow:          pulumi.String("sat:03:00-sat:04:00"),
            AllowVersionUpgrade:                 pulumi.BoolPtr(true),
            ClusterParameterGroupName:           paramGroup.Name,
        })
        if err != nil {
            return err
        }

        // Logging
        _, err = redshift.NewLogging(ctx, "prod-logging", &redshift.LoggingArgs{
            ClusterIdentifier:  cluster.ID(),
            LogDestinationType: pulumi.String("cloudwatch"),
            LogExports: pulumi.StringArray{
                pulumi.String("connectionlog"),
                pulumi.String("useractivitylog"),
                pulumi.String("userlog"),
            },
        })
        if err != nil {
            return err
        }

        ctx.Export("endpoint", cluster.Endpoint)
        ctx.Export("clusterArn", cluster.Arn)
        ctx.Export("securityGroupId", sg.ID())
        return nil
    })
}
```

## 3. Using the OpenMCF Module

When using the OpenMCF Pulumi module (the `module/` package), resource creation
is handled internally. You pass the `AwsRedshiftClusterStackInput` protobuf
message and call `Resources()`:

```go
package main

import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    module "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsredshiftcluster/v1/iac/pulumi/module"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        return module.Resources(ctx, stackInput)
    })
}
```

The module reads the `target` (AwsRedshiftCluster manifest) and `provider_config`
from the stack input, then creates all resources with proper tagging, conditional
logic, and output exports.
