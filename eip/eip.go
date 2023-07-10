package eip

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
