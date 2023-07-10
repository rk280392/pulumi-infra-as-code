package natGateway

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateNatGateway(ctx *pulumi.Context, prefix string, subnetID, eip pulumi.IDOutput) (*ec2.NatGateway, error) {
	ngw, err := ec2.NewNatGateway(ctx, prefix+"-ngw", &ec2.NatGatewayArgs{
		AllocationId: eip,
		SubnetId:     subnetID,
	})
	if err != nil {
		return nil, err
	}
	return ngw, nil
}
