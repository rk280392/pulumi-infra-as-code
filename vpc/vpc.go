package vpc

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateVPC(ctx *pulumi.Context, prefix, vpcCIDR string) (*ec2.Vpc, error) {
	vpc, err := ec2.NewVpc(ctx, prefix+"-vpc", &ec2.VpcArgs{
		CidrBlock: pulumi.String(vpcCIDR),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("eks-vpc"),
		},
	})
	if err != nil {
		return nil, err
	}
	return vpc, nil
}
