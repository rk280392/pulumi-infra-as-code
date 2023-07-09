package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateVPC(ctx *pulumi.Context, prefix string) (pulumi.Output, error) {
	vpc, err := ec2.NewVpc(ctx, prefix+"-vpc", &ec2.VpcArgs{
		CidrBlock:       pulumi.String("172.16.0.0/16"),
		InstanceTenancy: pulumi.String("default"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("eks-vpc"),
		},
	})
	if err != nil {
		return vpc.ID(), err
	}
	return vpc.ID(), nil
}
