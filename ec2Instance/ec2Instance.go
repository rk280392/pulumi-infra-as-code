package ec2Instance

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateEIPs(ctx *pulumi.Context, prefix string) (*ec2.Eip, error) {
	eip, err := ec2.NewEip(ctx, prefix+"-eip", &ec2.EipArgs{
		Vpc: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	return eip, nil
}

func CreateSecurityGroup(ctx *pulumi.Context, prefix, internetCIDR string, vpcID pulumi.IDOutput) (*ec2.SecurityGroup, error) {
	sgs, err := ec2.NewSecurityGroup(ctx, prefix+"-sg", &ec2.SecurityGroupArgs{
		VpcId: vpcID,
		Egress: ec2.SecurityGroupEgressArray{
			ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String(internetCIDR)},
			},
		},
		Ingress: ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(80),
				ToPort:     pulumi.Int(80),
				CidrBlocks: pulumi.StringArray{pulumi.String(internetCIDR)},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return sgs, nil
}
