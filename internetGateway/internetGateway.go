package internetGateway

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateInternetGateway(ctx *pulumi.Context, prefix string, vpcID pulumi.IDOutput) (*ec2.InternetGateway, error) {
	igw, err := ec2.NewInternetGateway(ctx, prefix+"-igw", &ec2.InternetGatewayArgs{
		VpcId: vpcID,
	})
	if err != nil {
		return nil, err
	}
	return igw, nil
}
